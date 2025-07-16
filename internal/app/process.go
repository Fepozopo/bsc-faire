package app

import (
	"fmt"
	"strings"
)

// DisplayIDToOrderID converts a display ID (e.g., "BXDmjBWXID") to an order ID (e.g., "bo_bxdmjBWXID").
func DisplayIDToOrderID(displayID string) string {
	return "bo_" + strings.ToLower(displayID)
}

func BillingToShippingType(billing string) string {
	switch billing {
	case "Consignee":
		return "SHIP_WITH_FAIRE"
	case "Prepaid":
		return "SHIP_ON_YOUR_OWN"
	default:
		return "SHIP_ON_YOUR_OWN"
	}
}

func ProcessShipments(csvPath string) error {
	shipments, err := ParseShipmentsCSV(csvPath)
	if err != nil {
		return err
	}
	client := NewFaireClient()
	for _, s := range shipments {
		orderID := DisplayIDToOrderID(s.PONumber)
		payload := ShipmentPayload{
			OrderID:        orderID,
			MakerCostCents: s.MakerCostCents,
			Carrier:        s.Carrier,
			TrackingCode:   s.TrackingCode,
			ShippingType:   BillingToShippingType(s.BillingType),
		}
		err := client.AddShipment(orderID, payload)
		if err != nil {
			fmt.Printf("Failed to add shipment for order %s: %v\n", s.PONumber, err)
		} else {
			fmt.Printf("Shipment added for order %s\n", s.PONumber)
		}
	}
	return nil
}
