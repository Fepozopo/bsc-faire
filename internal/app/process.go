package app

import (
	"fmt"
	"os"
	"strings"
	"time"
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

	shipments, parseErr := ParseShipmentsCSV(csvPath)
	if parseErr != nil {
		err = parseErr
		return
	}
	fmt.Fprintf(logFile, "INFO: Parsed %d shipments from CSV\n", len(shipments))
	for i, s := range shipments {
		fmt.Fprintf(logFile, "INFO: Processing shipment %d: %+v\n", i+1, s)
		apiToken, tokenErr := GetToken(s.SaleSource)
		if tokenErr != nil || apiToken == "" {
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
			payload.ErrorMsg = addErr.Error() // Attach error message to payload
			failed = append(failed, payload)
		} else {
			fmt.Fprintf(logFile, "INFO: Successfully processed shipment\n")
			processed = append(processed, payload)
		}
	}
	fmt.Fprintf(logFile, "INFO: Finished processing. %d processed, %d failed\n", len(processed), len(failed))
	return
}
