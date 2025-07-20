package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/joho/godotenv"
)

type FaireClient struct {
	BaseURL string
}

type ShipmentRequest struct {
	Shipments []ShipmentPayload `json:"shipments"`
}

type ShipmentPayload struct {
	OrderID        string `json:"order_id"`
	MakerCostCents int    `json:"maker_cost_cents"`
	Carrier        string `json:"carrier"`
	TrackingCode   string `json:"tracking_code"`
	ShippingType   string `json:"shipping_type"`
}

func NewFaireClient() *FaireClient {
	godotenv.Load()
	return &FaireClient{
		BaseURL: "https://www.faire.com/external-api/v2",
	}
}

func (c *FaireClient) AddShipment(orderID string, payload ShipmentPayload, apiToken string) error {
	url := fmt.Sprintf("%s/orders/%s/shipments", c.BaseURL, orderID)
	body, _ := json.Marshal(ShipmentRequest{Shipments: []ShipmentPayload{payload}})
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-FAIRE-ACCESS-TOKEN", apiToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("faire API error: %s", string(b))
	}
	return nil
}

func (c *FaireClient) GetAllOrders(apiToken string, limit int, page int, states string) ([]byte, error) {
	url := fmt.Sprintf("%s/orders?limit=%d&page=%d&excluded_states=%s", c.BaseURL, limit, page, states)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-FAIRE-ACCESS-TOKEN", apiToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *FaireClient) GetOrderByID(PONumber string, apiToken string) ([]byte, error) {
	orderID := DisplayIDToOrderID(PONumber)
	url := fmt.Sprintf("%s/orders/%s", c.BaseURL, orderID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-FAIRE-ACCESS-TOKEN", apiToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
