package app

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestBillingToShippingType(t *testing.T) {
	if BillingToShippingType("Consignee") != "SHIP_WITH_FAIRE" {
		t.Error("Consignee should map to Consignee")
	}
	if BillingToShippingType("Prepaid") != "SHIP_ON_YOUR_OWN" {
		t.Error("Prepaid should map to SHIP_ON_YOUR_OWN")
	}
	if BillingToShippingType("Other") != "SHIP_ON_YOUR_OWN" {
		t.Error("Other should default to SHIP_ON_YOUR_OWN")
	}
}

func TestNewFaireClientPanic(t *testing.T) {
	os.Unsetenv("BSC_API_TOKEN")
	defer func() {
		recover()
	}()
	NewFaireClient()
}

func TestAddShipment(t *testing.T) {
	// Step 1: Create temp CSV file
	csvContent := `Source Document Key,PO Numbers,Master Tracking #,Ready Date/Time,Shipment Charges Applied Total,Ship Carrier Name,Billing Type,Recipient Customer ID,Sale Source (UDF)
0083419,88NJSQS3DD,1Z972Y3Y0301141377,7/16/2025,0,UPS,Consignee,0090671,SM`

	tmpFile, err := os.CreateTemp("", "test_shipments_*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write([]byte(csvContent)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Step 2: Parse CSV
	shipments, err := ParseShipmentsCSV(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to parse shipments: %v", err)
	}
	t.Logf("Parsed shipments: %+v", shipments)

	if len(shipments) != 1 {
		t.Fatalf("expected 1 shipment, got %d", len(shipments))
	}
	shipment := shipments[0]

	// Step 3: Convert to ShipmentPayload
	payload := ShipmentPayload{
		OrderID:        shipment.PONumber,
		MakerCostCents: shipment.MakerCostCents,
		Carrier:        shipment.Carrier,
		TrackingCode:   shipment.TrackingCode,
		ShippingType:   BillingToShippingType(shipment.BillingType),
	}
	t.Logf("Shipment payload: %+v", payload)

	// Step 4: Mock Faire API endpoint
	var receivedBody ShipmentRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read body: %v", err)
		}
		json.Unmarshal(body, &receivedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Step 5: Call AddShipment
	client := &FaireClient{
		BaseURL: server.URL,
	}
	err = client.AddShipment(payload.OrderID, payload, "dummy-token")
	if err != nil {
		t.Fatalf("AddShipment failed: %v", err)
	}

	// Step 6: Assert payload matches CSV
	if len(receivedBody.Shipments) != 1 {
		t.Fatalf("expected 1 shipment in request, got %d", len(receivedBody.Shipments))
	}
	got := receivedBody.Shipments[0]
	if got.OrderID != shipment.PONumber {
		t.Errorf("OrderID mismatch: got %s, want %s", got.OrderID, shipment.PONumber)
	}
	if got.MakerCostCents != shipment.MakerCostCents {
		t.Errorf("MakerCostCents mismatch: got %d, want %d", got.MakerCostCents, shipment.MakerCostCents)
	}
	if got.Carrier != shipment.Carrier {
		t.Errorf("Carrier mismatch: got %s, want %s", got.Carrier, shipment.Carrier)
	}
	if got.TrackingCode != shipment.TrackingCode {
		t.Errorf("TrackingCode mismatch: got %s, want %s", got.TrackingCode, shipment.TrackingCode)
	}
	wantType := BillingToShippingType(shipment.BillingType)
	if got.ShippingType != wantType {
		t.Errorf("ShippingType mismatch: got %s, want %s", got.ShippingType, wantType)
	}
}
