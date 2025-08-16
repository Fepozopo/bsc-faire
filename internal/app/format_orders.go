package app

import (
	"fmt"
	"strings"
)

// FormatOrder returns a formatted string for a single order.
func FormatOrder(order Order) string {
	created := order.CreatedAt.Format("2006-01-02")
	retailer := order.Address.CompanyName
	if retailer == "" {
		retailer = order.Address.Name
	}

	var totalCents int
	for _, item := range order.Items {
		totalCents += item.PriceCents * item.Quantity
	}

	s := fmt.Sprintf(
		"Order ID: %s\nStatus: %s\nRetailer: %s\nCreated: %s\nShip By: %s\nTotal: $%.2f\n\nItems:\n",
		order.DisplayID, order.State, retailer, created, order.ShipAfter.Format("2006-01-02"), float64(totalCents)/100,
	)
	for _, item := range order.Items {
		s += fmt.Sprintf("  - %s x%d ($%.2f each) %s\n",
			item.Sku, item.Quantity, float64(item.PriceCents)/100, item.ProductName)
	}
	return s
}

// FormatOrders returns a formatted string for a list of orders.
func FormatOrders(orders []Order) string {
	var b strings.Builder
	for i, order := range orders {
		b.WriteString(FormatOrder(order))
		if i < len(orders)-1 {
			b.WriteString("\n" + strings.Repeat("-", 40) + "\n\n")
		}
	}
	return b.String()
}
