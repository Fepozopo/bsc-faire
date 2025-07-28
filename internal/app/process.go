package app

import (
	"fmt"
	"os"
	"strings"
	"time"

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
// and accepts a FaireClientInterface for testability.
func ProcessShipments(csvPath string, client FaireClientInterface) (processed []ShipmentPayload, failed []ShipmentPayload, err error) {
	// Ensure logs directory exists
	logDir := "logs"
	if mkErr := os.MkdirAll(logDir, 0755); mkErr != nil {
		err = fmt.Errorf("failed to create logs directory: %w", mkErr)
		return
	}
	// Create log file with timestamp
	logFileName := fmt.Sprintf("%s/%s.txt", logDir, time.Now().Format("20060102_150405"))
	logFile, fileErr := os.Create(logFileName)
	if fileErr != nil {
		err = fmt.Errorf("failed to create log file: %w", fileErr)
		return
	}
	defer logFile.Close()

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
	fmt.Fprintf(logFile, "INFO: Parsed %d shipments from CSV\n", len(shipments))
	for i, s := range shipments {
		fmt.Fprintf(logFile, "INFO: Processing shipment %d: %+v\n", i+1, s)
		var apiToken string
		switch s.SaleSource {
		case "21":
			apiToken = c21Token
		case "ASC":
			apiToken = ascToken
		case "BJP":
			apiToken = bjpToken
		case "BSC":
			apiToken = bscToken
		case "GTG":
			apiToken = gtgToken
		case "OAT":
			apiToken = oatToken
		case "SM":
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
			fmt.Fprintf(logFile, "ERROR: Failed to add shipment: %v\n", addErr)
			failed = append(failed, payload)
		} else {
			fmt.Fprintf(logFile, "INFO: Successfully processed shipment\n")
			processed = append(processed, payload)
		}
	}
	fmt.Fprintf(logFile, "INFO: Finished processing. %d processed, %d failed\n", len(processed), len(failed))
	return
}
