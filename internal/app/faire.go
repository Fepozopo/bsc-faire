package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// FaireClient sends authenticated requests to the Faire Orders API.
type FaireClient struct {
	BaseURL string
}

// FaireClientInterface defines the Faire operations used by the application.
type FaireClientInterface interface {
	AddShipment(payload ShipmentPayload, apiToken string) error
	GetAllOrders(apiToken string, limit int, page int, excludedStates string) ([]byte, error)
	GetOrderByID(orderIdentifier string, apiToken string) ([]byte, error)
}

// ShipmentRequest is the request body accepted by Faire's shipment endpoint.
type ShipmentRequest struct {
	Shipments []ShipmentPayload `json:"shipments"`
}

// ShipmentPayload describes a shipment that will be attached to an order.
type ShipmentPayload struct {
	OrderID        string `json:"order_id"`
	MakerCostCents int    `json:"maker_cost_cents"`
	Carrier        string `json:"carrier"`
	TrackingCode   string `json:"tracking_code"`
	ShippingType   string `json:"shipping_type"`
	SaleSource     string `json:"sale_source"`
	ErrorMsg       string `json:"error_msg"`
}

// NewFaireClient loads the Faire base URL from the environment and returns a client.
func NewFaireClient() *FaireClient {
	_ = godotenv.Load()
	return &FaireClient{
		BaseURL: os.Getenv("FAIRE_BASE_URL"),
	}
}

// AddShipment adds payload to its order using apiToken and returns an API or transport error.
func (c *FaireClient) AddShipment(payload ShipmentPayload, apiToken string) error {
	endpoint := fmt.Sprintf("%s/orders/%s/shipments", strings.TrimRight(c.BaseURL, "/"), url.PathEscape(payload.OrderID))
	body, err := json.Marshal(ShipmentRequest{Shipments: []ShipmentPayload{payload}})
	if err != nil {
		return fmt.Errorf("marshal shipment request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create shipment request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-FAIRE-ACCESS-TOKEN", apiToken)

	return c.doRequest(req)
}

// GetAllOrders returns one page of orders while excluding the supplied Faire order states.
func (c *FaireClient) GetAllOrders(apiToken string, limit int, page int, excludedStates string) ([]byte, error) {
	endpoint, err := url.Parse(strings.TrimRight(c.BaseURL, "/") + "/orders")
	if err != nil {
		return nil, fmt.Errorf("parse orders endpoint: %w", err)
	}

	query := endpoint.Query()
	query.Set("limit", strconv.Itoa(limit))
	query.Set("page", strconv.Itoa(page))
	query.Set("excluded_states", excludedStates)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create orders request: %w", err)
	}
	req.Header.Set("X-FAIRE-ACCESS-TOKEN", apiToken)

	return c.readResponse(req)
}

// GetOrderByID returns the order identified by a display ID or a Faire bo_ order ID.
func (c *FaireClient) GetOrderByID(orderIdentifier string, apiToken string) ([]byte, error) {
	orderID := OrderIdentifierToOrderID(orderIdentifier)
	if orderID == "" {
		return nil, fmt.Errorf("order identifier cannot be empty")
	}

	endpoint := fmt.Sprintf("%s/orders/%s", strings.TrimRight(c.BaseURL, "/"), url.PathEscape(orderID))
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create order request: %w", err)
	}
	req.Header.Set("X-FAIRE-ACCESS-TOKEN", apiToken)

	return c.readResponse(req)
}

// doRequest sends req and returns a descriptive error unless Faire returns a successful status.
func (c *FaireClient) doRequest(req *http.Request) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("faire API error (%s): %s", resp.Status, strings.TrimSpace(string(body)))
}

// readResponse sends req and returns its body when Faire returns a successful status.
func (c *FaireClient) readResponse(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read Faire API response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("faire API error (%s): %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return body, nil
}
