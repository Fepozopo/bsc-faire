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
	c21Token := os.Getenv("C21_API_TOKEN")
	ascToken := os.Getenv("ASC_API_TOKEN")
	bjpToken := os.Getenv("BJP_API_TOKEN")
	bscToken := os.Getenv("BSC_API_TOKEN")
	gtgToken := os.Getenv("GTG_API_TOKEN")
	oatToken := os.Getenv("OAT_API_TOKEN")
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
		case "21":
			apiToken = c21Token
		case "asc":
			apiToken = ascToken
		case "bjp":
			apiToken = bjpToken
		case "bsc":
			apiToken = bscToken
		case "gtg":
			apiToken = gtgToken
		case "oat":
			apiToken = oatToken
		case "sm":
			apiToken = smdToken
		default:
			// Should not happen due to ParseShipmentsCSV, but skip just in case
			continue
		}
		orderID := DisplayIDToOrderID(s.PONumber)
		billingType := BillingToShippingType(s.BillingType)
		payload := ShipmentPayload{
			OrderID:        orderID,
			MakerCostCents: s.MakerCostCents,
			Carrier:        s.Carrier,
			TrackingCode:   s.TrackingCode,
			ShippingType:   billingType,
			SaleSource:     s.SaleSource,
		}
		addErr := client.AddShipment(payload, apiToken)
		if addErr != nil {
			failed = append(failed, payload)
		} else {
			processed = append(processed, payload)
		}
	}
	return
}
