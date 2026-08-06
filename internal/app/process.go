package app

import "strings"

// DisplayIDToOrderID converts a display ID (e.g., "BXDMJBWXID") to an order ID (e.g., "bo_bxdmjbwxid").
func DisplayIDToOrderID(displayID string) string {
	return "bo_" + strings.ToLower(displayID)
}

// OrderIDToDisplayID converts a Faire bo_ order ID to its uppercase display ID.
func OrderIDToDisplayID(orderID string) string {
	if !strings.HasPrefix(orderID, "bo_") {
		return orderID // Return as is if it doesn't match expected format
	}
	return strings.ToUpper(strings.TrimPrefix(orderID, "bo_"))
}

// ProcessShipments submits shipments from csvPath and returns the successful and failed payloads.
// client performs Faire API requests, and err reports CSV parsing failures.
func ProcessShipments(csvPath string, client FaireClientInterface) (processed []ShipmentPayload, failed []ShipmentPayload, err error) {
	shipments, parseErr := ParseShipmentsCSV(csvPath)
	if parseErr != nil {
		err = parseErr
		return
	}
	shippingType := "SHIP_ON_YOUR_OWN"
	for _, s := range shipments {
		apiToken, tokenErr := GetToken(s.SaleSource)
		if tokenErr != nil || apiToken == "" {
			// Should not happen due to ParseShipmentsCSV, but skip just in case
			continue
		}
		orderID := DisplayIDToOrderID(s.PONumber)
		payload := ShipmentPayload{
			OrderID:        orderID,
			MakerCostCents: s.MakerCostCents,
			Carrier:        s.Carrier,
			TrackingCode:   s.TrackingCode,
			ShippingType:   shippingType,
			SaleSource:     s.SaleSource,
		}
		addErr := client.AddShipment(payload, apiToken)
		if addErr != nil {
			// Preserve the API error in the result so the GUI can show the user which shipment failed.
			payload.ErrorMsg = addErr.Error()
			failed = append(failed, payload)
		} else {
			processed = append(processed, payload)
		}
	}
	return
}
