package app

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	// OrderStateNew is the Faire state for an order that has not begun fulfillment.
	OrderStateNew = "NEW"
	// OrderStateBackordered is the Faire state for an order awaiting inventory.
	OrderStateBackordered = "BACKORDERED"
)

const ordersPageSize = 50

// OrderClient retrieves orders needed to build an order-export CSV.
type OrderClient interface {
	GetAllOrders(apiToken string, limit int, page int, excludedStates string) ([]byte, error)
	GetOrderByID(orderIdentifier string, apiToken string) ([]byte, error)
}

// OrderExportFilter chooses either one Faire state or an explicit set of order identifiers.
type OrderExportFilter struct {
	State            string
	OrderIdentifiers []string
}

// faireOrderStates lists every order state supported by the export's inverse filter.
var faireOrderStates = []string{
	OrderStateNew,
	OrderStateBackordered,
	"CANCELED",
	"PROCESSING",
	"PRE_TRANSIT",
	"IN_TRANSIT",
	"DELIVERED",
	"RETURNED",
	"PENDING_RETAILER_CONFIRMATION",
	"DAMAGED_OR_MISSING",
}

// ExportNewOrdersToCSV exports all NEW orders for saleSource to filename and returns the order count.
func (c *FaireClient) ExportNewOrdersToCSV(saleSource, filename string) (int, error) {
	return c.exportOrdersForState(saleSource, filename, OrderStateNew)
}

// ExportBackorderedOrdersToCSV exports all BACKORDERED orders for saleSource to filename and returns the order count.
func (c *FaireClient) ExportBackorderedOrdersToCSV(saleSource, filename string) (int, error) {
	return c.exportOrdersForState(saleSource, filename, OrderStateBackordered)
}

// ExportOrdersByIDsToCSV exports the supplied display IDs or Faire bo_ IDs for saleSource to filename.
func (c *FaireClient) ExportOrdersByIDsToCSV(saleSource, filename string, orderIdentifiers []string) (int, error) {
	token, err := tokenForSaleSource(saleSource)
	if err != nil {
		return 0, err
	}

	return ExportOrdersToCSV(c, token, saleSource, filename, OrderExportFilter{
		OrderIdentifiers: orderIdentifiers,
	})
}

// ExportOrdersToCSV retrieves the orders selected by filter and writes them to filename.
// apiToken authenticates requests, saleSource is recorded in the CSV, and the return value is the order count.
func ExportOrdersToCSV(client OrderClient, apiToken, saleSource, filename string, filter OrderExportFilter) (int, error) {
	if err := filter.validate(); err != nil {
		return 0, err
	}

	var (
		orders []Order
		err    error
	)
	if len(filter.OrderIdentifiers) > 0 {
		orders, err = getOrdersByIdentifiers(client, apiToken, filter.OrderIdentifiers)
	} else {
		orders, err = getOrdersByState(client, apiToken, filter.State)
	}
	if err != nil {
		return 0, err
	}

	if err := writeOrdersCSV(filename, saleSource, orders); err != nil {
		return 0, err
	}
	return len(orders), nil
}

// ParseOrderIdentifiers returns unique non-empty identifiers entered as comma-, semicolon-, or line-separated values.
func ParseOrderIdentifiers(raw string) []string {
	seen := make(map[string]struct{})
	identifiers := make([]string, 0)

	for _, identifier := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r'
	}) {
		identifier = strings.TrimSpace(identifier)
		if identifier == "" {
			continue
		}

		key := strings.ToLower(identifier)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		identifiers = append(identifiers, identifier)
	}

	return identifiers
}

// OrderIdentifierToOrderID converts a display ID to Faire's bo_ ID while preserving a supplied bo_ ID.
func OrderIdentifierToOrderID(orderIdentifier string) string {
	orderIdentifier = strings.TrimSpace(orderIdentifier)
	if orderIdentifier == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(orderIdentifier), "bo_") {
		return strings.ToLower(orderIdentifier)
	}
	return DisplayIDToOrderID(orderIdentifier)
}

// exportOrdersForState resolves the sale-source token and exports the requested Faire state.
func (c *FaireClient) exportOrdersForState(saleSource, filename, state string) (int, error) {
	token, err := tokenForSaleSource(saleSource)
	if err != nil {
		return 0, err
	}
	return ExportOrdersToCSV(c, token, saleSource, filename, OrderExportFilter{State: state})
}

// tokenForSaleSource returns the configured token or a consistent error for an invalid or unconfigured sale source.
func tokenForSaleSource(saleSource string) (string, error) {
	token, err := GetToken(saleSource)
	if err != nil || token == "" {
		return "", fmt.Errorf("invalid or missing token for sale source %q", saleSource)
	}
	return token, nil
}

// validate rejects ambiguous filters and unsupported state-based exports.
func (filter OrderExportFilter) validate() error {
	if len(filter.OrderIdentifiers) > 0 && filter.State != "" {
		return fmt.Errorf("an export can select either a state or order identifiers, not both")
	}
	if len(filter.OrderIdentifiers) > 0 {
		return nil
	}
	if !isKnownOrderState(filter.State) {
		return fmt.Errorf("unsupported order state %q", filter.State)
	}
	return nil
}

// getOrdersByState paginates Faire's inverse state filter and retains only the requested state as a safeguard.
func getOrdersByState(client OrderClient, apiToken, state string) ([]Order, error) {
	excludedStates := excludedStatesFor(state)
	orders := make([]Order, 0)

	for page := 1; ; page++ {
		response, err := client.GetAllOrders(apiToken, ordersPageSize, page, excludedStates)
		if err != nil {
			return nil, fmt.Errorf("get %s orders on page %d: %w", state, page, err)
		}

		var ordersResponse Orders
		if err := json.Unmarshal(response, &ordersResponse); err != nil {
			return nil, fmt.Errorf("parse %s orders on page %d: %w", state, page, err)
		}
		for _, order := range ordersResponse.Orders {
			// Keep the local check because the export must never include a wrong state if Faire ignores a filter.
			if strings.EqualFold(order.State, state) {
				orders = append(orders, order)
			}
		}

		if len(ordersResponse.Orders) < ordersPageSize {
			return orders, nil
		}
	}
}

// getOrdersByIdentifiers retrieves each requested order in user-entered order.
func getOrdersByIdentifiers(client OrderClient, apiToken string, orderIdentifiers []string) ([]Order, error) {
	orders := make([]Order, 0, len(orderIdentifiers))
	for _, orderIdentifier := range orderIdentifiers {
		response, err := client.GetOrderByID(orderIdentifier, apiToken)
		if err != nil {
			return nil, fmt.Errorf("get order %q: %w", orderIdentifier, err)
		}

		var order Order
		if err := json.Unmarshal(response, &order); err != nil {
			return nil, fmt.Errorf("parse order %q: %w", orderIdentifier, err)
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// excludedStatesFor returns all Faire order states except state for Faire's inverse state filter.
func excludedStatesFor(state string) string {
	excludedStates := make([]string, 0, len(faireOrderStates)-1)
	for _, candidate := range faireOrderStates {
		if candidate != state {
			excludedStates = append(excludedStates, candidate)
		}
	}
	return strings.Join(excludedStates, ",")
}

// isKnownOrderState reports whether state can be represented by the inverse state filter.
func isKnownOrderState(state string) bool {
	for _, candidate := range faireOrderStates {
		if candidate == state {
			return true
		}
	}
	return false
}

// writeOrdersCSV writes orders to filename using saleSource in each CSV row.
func writeOrdersCSV(filename, saleSource string, orders []Order) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write(orderCSVHeader); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}

	for _, order := range orders {
		if err := writeOrderCSVRows(writer, saleSource, order); err != nil {
			return err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush CSV: %w", err)
	}
	return nil
}

// writeOrderCSVRows writes one CSV row per item in order.
func writeOrderCSVRows(writer *csv.Writer, saleSource string, order Order) error {
	includesFreeShipping := make([]string, 0, len(order.BrandDiscounts))
	discountPercentages := make([]string, 0, len(order.BrandDiscounts))
	for _, discount := range order.BrandDiscounts {
		includesFreeShipping = append(includesFreeShipping, strconv.FormatBool(discount.IncludesFreeShipping))
		discountPercentages = append(discountPercentages, fmt.Sprintf("%.2f", discount.DiscountPercentage))
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
			strings.ToUpper(saleSource),
			order.SalesRepName,
			order.Notes,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write CSV row for order %q: %w", order.ID, err)
		}
	}
	return nil
}

// orderCSVHeader defines the stable column order for all order-export CSV files.
var orderCSVHeader = []string{
	"id", "display_id", "created_at", "ship_after",
	"address_name", "address_address1", "address_address2", "address_postal_code",
	"address_city", "address_state", "address_state_code", "address_phone_number",
	"address_country", "address_country_code", "address_company_name",
	"is_free_shipping", "brand_discounts_includes_free_shipping", "brand_discounts_discount_percentage",
	"payout_costs_commission_bps", "payout_costs_commission_cents",
	"item_sku", "item_price_cents", "item_quantity", "sale_source", "sales_rep_name", "notes",
}
