package app

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestParseShipmentsCSV(t *testing.T) {
	// Create a temporary CSV file with the expected headers and data, including Sale Source (UDF)
	csvContent := `Source Document Key,PO Numbers,Master Tracking #,Shipment Charges Applied Total,Ship Carrier Name,Billing Type,Recipient Customer ID,Sale Source (UDF)
DOC1,ORDER123,TRACK123,10.00,UPS,Consignee,0090671,SM
DOC2,ORDER124,TRACK124,20.50,FedEx,Prepaid,0090671,BSC
DOC3,ORDER125,TRACK125,30.00,DHL,Third Party,0000000,SM
DOC4,ORDER126,TRACK126,5.00,USPS,Consignee,0090671,OTHER
DOC5,ORDER127,TRACK127,7.00,UPS,Consignee,0090671,BSC
DOC6,ORDER128,TRACK128,8.00,UPS,Consignee,0090671,XYZ
DOC7,ORDER129,TRACK129,9.00,UPS,Consignee,0090671,SM
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

	// Log the parsed shipments for debugging
	for i, s := range shipments {
		b, _ := json.MarshalIndent(s, "", "  ")
		t.Logf("Shipment %d:\n%s", i+1, b)
	}

	// Only rows with Sale Source (UDF) == "SM" or "BSC" should be included
	expectedLen := 4 // ORDER123 (SM), ORDER124 (BSC), ORDER127 (BSC), ORDER129 (SM)
	if len(shipments) != expectedLen {
		t.Errorf("expected %d shipments, got %d", expectedLen, len(shipments))
	}

	// Verify the shipments
	expected := []struct {
		PONumber       string
		SaleSource     string
		MakerCostCents int
	}{
		{"ORDER123", "SM", 1000},
		{"ORDER124", "BSC", 2050},
		{"ORDER127", "BSC", 700},
		{"ORDER129", "SM", 900},
	}
	for i, exp := range expected {
		if i >= len(shipments) {
			break
		}
		s := shipments[i]
		if s.PONumber != exp.PONumber {
			t.Errorf("expected PONumber '%s', got '%s'", exp.PONumber, s.PONumber)
		}
		if s.MakerCostCents != exp.MakerCostCents {
			t.Errorf("expected MakerCostCents %d, got %d", exp.MakerCostCents, s.MakerCostCents)
		}
	}

	// Test case for missing required header
	missingHeaderCSV := `Source Document Key,PO Numbers,Master Tracking #,Shipment Charges Applied Total,Ship Carrier Name,Recipient Customer ID,Sale Source (UDF)
DOC1,ORDER123,TRACK123,10.00,UPS,0090671,SM
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
	invalidCostCSV := `Source Document Key,PO Numbers,Master Tracking #,Shipment Charges Applied Total,Ship Carrier Name,Billing Type,Recipient Customer ID,Sale Source (UDF)
DOC1,ORDER123,TRACK123,ABC,UPS,Consignee,0090671,SM
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
