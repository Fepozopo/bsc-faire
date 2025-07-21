package app

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// DisplayIDToOrderID converts a display ID (e.g., "BXDMJBWXID") to an order ID (e.g., "bo_bxdmjbwxid").
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

// ProcessShipments now returns processed and failed shipments as slices of ShipmentPayload
func ProcessShipments(csvPath string) (processed []ShipmentPayload, failed []ShipmentPayload, err error) {
	// Load .env to get API tokens
	godotenv.Load()
	bscToken := os.Getenv("BSC_API_TOKEN")
	smdToken := os.Getenv("SMD_API_TOKEN")

	shipments, parseErr := ParseShipmentsCSV(csvPath)
	if parseErr != nil {
		err = parseErr
		return
	}
	client := NewFaireClient()
	for _, s := range shipments {
		var apiToken string
		switch s.SaleSource {
		case "SM":
			apiToken = smdToken
		case "BSC":
			apiToken = bscToken
		default:
			// Should not happen due to ParseShipmentsCSV, but skip just in case
			continue
		}
		orderID := DisplayIDToOrderID(s.PONumber)
		payload := ShipmentPayload{
			OrderID:        orderID,
			MakerCostCents: s.MakerCostCents,
			Carrier:        s.Carrier,
			TrackingCode:   s.TrackingCode,
			ShippingType:   BillingToShippingType(s.BillingType),
		}
		addErr := client.AddShipment(orderID, payload, apiToken)
		if addErr != nil {
			failed = append(failed, payload)
		} else {
			processed = append(processed, payload)
		}
	}
	return
}
