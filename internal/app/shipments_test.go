package app

import (
	"os"
	"strings"
	"testing"
)

func TestParseShipmentsCSV(t *testing.T) {
	// Create a temporary CSV file with the expected headers and data
	csvContent := `Source Document Key,PO Numbers,Master Tracking #,Shipment Charges Applied Total,Ship Carrier Name,Billing Type,Recipient Customer ID
DOC1,ORDER123,TRACK123,10.00,UPS,Consignee,0090671
DOC2,ORDER124,TRACK124,20.50,FedEx,Prepaid,0090671
DOC3,ORDER125,TRACK125,30.00,DHL,Third Party,0000000
DOC4,ORDER126,TRACK126,5.00,USPS,Consignee,0090671
`
	f, err := os.CreateTemp("", "shipments-*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name()) // Clean up the temporary file
	defer f.Close()           // Close the file

	if _, err := f.WriteString(csvContent); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	// Close the file so ParseShipmentsCSV can open and read it
	f.Close()

	shipments, err := ParseShipmentsCSV(f.Name())
	if err != nil {
		t.Fatalf("unexpected error parsing CSV: %v", err)
	}

	// Expected shipments: only those with Recipient Customer ID "0090671"
	expectedLen := 3
	if len(shipments) != expectedLen {
		t.Errorf("expected %d shipments, got %d", expectedLen, len(shipments))
	}

	// Verify the first shipment
	if len(shipments) > 0 {
		s := shipments[0]
		if s.CustomerNumber != "0090671" {
			t.Errorf("expected CustomerNumber '0090671', got '%s'", s.CustomerNumber)
		}
		if s.PONumber != "ORDER123" {
			t.Errorf("expected PONumber 'ORDER123', got '%s'", s.PONumber)
		}
		if s.BillingType != "Consignee" {
			t.Errorf("expected BillingType 'Consignee', got '%s'", s.BillingType)
		}
		if s.Carrier != "UPS" {
			t.Errorf("expected Carrier 'UPS', got '%s'", s.Carrier)
		}
		if s.TrackingCode != "TRACK123" {
			t.Errorf("expected TrackingCode 'TRACK123', got '%s'", s.TrackingCode)
		}
		if s.MakerCostCents != 1000 { // 10.00 * 100
			t.Errorf("expected MakerCostCents 1000, got %d", s.MakerCostCents)
		}
	}

	// Verify the second shipment
	if len(shipments) > 1 {
		s := shipments[1]
		if s.PONumber != "ORDER124" {
			t.Errorf("expected PONumber 'ORDER124', got '%s'", s.PONumber)
		}
		if s.MakerCostCents != 2050 { // 20.50 * 100
			t.Errorf("expected MakerCostCents 2050, got %d", s.MakerCostCents)
		}
	}

	// Verify the third shipment
	if len(shipments) > 2 {
		s := shipments[2]
		if s.PONumber != "ORDER126" {
			t.Errorf("expected PONumber 'ORDER126', got '%s'", s.PONumber)
		}
		if s.MakerCostCents != 500 { // 5.00 * 100
			t.Errorf("expected MakerCostCents 500, got %d", s.MakerCostCents)
		}
	}

	// Test case for missing required header
	missingHeaderCSV := `Source Document Key,PO Numbers,Master Tracking #,Shipment Charges Applied Total,Ship Carrier Name,Recipient Customer ID
DOC1,ORDER123,TRACK123,10.00,UPS,0090671
`
	fMissing, err := os.CreateTemp("", "missing-header-*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file for missing header test: %v", err)
	}
	defer os.Remove(fMissing.Name())
	defer fMissing.Close()

	if _, err := fMissing.WriteString(missingHeaderCSV); err != nil {
		t.Fatalf("failed to write to temp file for missing header test: %v", err)
	}
	fMissing.Close()

	_, err = ParseShipmentsCSV(fMissing.Name())
	if err == nil {
		t.Error("expected error for missing 'Billing Type' header, got nil")
	} else if !strings.Contains(err.Error(), "missing required header: Billing Type") {
		t.Errorf("expected 'missing required header: Billing Type' error, got: %v", err)
	}

	// Test case for invalid MakerCostCents format
	invalidCostCSV := `Source Document Key,PO Numbers,Master Tracking #,Shipment Charges Applied Total,Ship Carrier Name,Billing Type,Recipient Customer ID
DOC1,ORDER123,TRACK123,ABC,UPS,Consignee,0090671
`
	fInvalidCost, err := os.CreateTemp("", "invalid-cost-*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file for invalid cost test: %v", err)
	}
	defer os.Remove(fInvalidCost.Name())
	defer fInvalidCost.Close()

	if _, err := fInvalidCost.WriteString(invalidCostCSV); err != nil {
		t.Fatalf("failed to write to temp file for invalid cost test: %v", err)
	}
	fInvalidCost.Close()

	_, err = ParseShipmentsCSV(fInvalidCost.Name())
	if err == nil {
		t.Error("expected error for invalid MakerCostCents, got nil")
	} else if !strings.Contains(err.Error(), "failed to parse MakerCostCents 'ABC'") {
		t.Errorf("expected 'failed to parse MakerCostCents 'ABC'' error, got: %v", err)
	}
}
