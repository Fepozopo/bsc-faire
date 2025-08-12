package app

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type FaireClient struct {
	BaseURL string
}

// FaireClientInterface allows mocking of FaireClient for testing
type FaireClientInterface interface {
	AddShipment(payload ShipmentPayload, apiToken string) error
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
	SaleSource     string `json:"sale_source"`
	ErrorMsg       string `json:"error_msg"`
}

func NewFaireClient() *FaireClient {
	godotenv.Load()
	return &FaireClient{
		BaseURL: os.Getenv("FAIRE_BASE_URL"),
	}
}

func (c *FaireClient) AddShipment(payload ShipmentPayload, apiToken string) error {
	url := fmt.Sprintf("%s/orders/%s/shipments", c.BaseURL, payload.OrderID)
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

// ExportNewOrdersToCSV exports new orders for the given sale source to a CSV file.
// Returns the number of exported orders and any error.
func (c *FaireClient) ExportNewOrdersToCSV(saleSource, filename string) (int, error) {
	token, err := GetToken(saleSource)
	if err != nil || token == "" {
		return 0, fmt.Errorf("invalid or missing token for sale source '%s'", saleSource)
	}

	// Paginate through all orders and collect NEW ones
	limit := 50
	page := 1
	states := "DELIVERED,BACKORDERED,CANCELED,PROCESSING,PRE_TRANSIT,IN_TRANSIT,RETURNED,PENDING_RETAILER_CONFIRMATION,DAMAGED_OR_MISSING"
	var newOrders []Order

	for {
		resp, err := c.GetAllOrders(token, limit, page, states)
		if err != nil {
			return 0, err
		}
		var ordersResp Orders
		if err := json.Unmarshal(resp, &ordersResp); err != nil {
			return 0, fmt.Errorf("failed to parse orders: %w", err)
		}
		foundNew := 0
		for _, order := range ordersResp.Orders {
			if strings.ToUpper(order.State) == "NEW" {
				newOrders = append(newOrders, order)
				foundNew++
			}
		}
		if len(ordersResp.Orders) < limit {
			break // last page
		}
		page++
	}

	// Prepare CSV
	file, err := os.Create(filename)
	if err != nil {
		return 0, fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header (add item_sku and item_price_cents)
	header := []string{
		"id", "display_id", "created_at", "ship_after",
		"address_name", "address_address1", "address_address2", "address_postal_code",
		"address_city", "address_state", "address_state_code", "address_phone_number",
		"address_country", "address_country_code", "address_company_name",
		"is_free_shipping", "brand_discounts_includes_free_shipping", "brand_discounts_discount_percentage",
		"payout_costs_commission_bps", "payout_costs_commission_cents",
		"item_sku", "item_price_cents", "item_quantity",
	}
	if err := writer.Write(header); err != nil {
		return 0, fmt.Errorf("failed to write CSV header: %w", err)
	}

	orderCount := 0
	for _, order := range newOrders {
		// Brand discounts fields
		var includesFreeShipping []string
		var discountPercentages []string
		for _, bd := range order.BrandDiscounts {
			includesFreeShipping = append(includesFreeShipping, strconv.FormatBool(bd.IncludesFreeShipping))
			discountPercentages = append(discountPercentages, fmt.Sprintf("%.2f", bd.DiscountPercentage))
		}

		for _, item := range order.Items {
			row := []string{
				order.ID,
				order.DisplayID,
				order.CreatedAt.Format("20060102"),
				order.ShipAfter.Format("20060102"),
				order.Address.Name,
				order.Address.Address1,
				order.Address.Address2,
				order.Address.PostalCode,
				order.Address.City,
				order.Address.State,
				order.Address.StateCode,
				order.Address.PhoneNumber,
				order.Address.Country,
				order.Address.CountryCode,
				order.Address.CompanyName,
				strconv.FormatBool(order.IsFreeShipping),
				strings.Join(includesFreeShipping, ","),
				strings.Join(discountPercentages, ","),
				fmt.Sprintf("%.2f", float64(order.PayoutCosts.CommissionBps)*0.01),
				fmt.Sprintf("%.2f", float64(order.PayoutCosts.CommissionCents)/100.0),
				item.Sku,
				fmt.Sprintf("%.2f", float64(item.PriceCents)/100.0),
				strconv.Itoa(item.Quantity),
			}
			if err := writer.Write(row); err != nil {
				return 0, fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
		orderCount++
	}

	return orderCount, nil
}
