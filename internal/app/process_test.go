package app

import (
	"os"
	"strings"
	"testing"
)

func TestProcessShipments_MockClient(t *testing.T) {
	// Set required API tokens for test
	os.Setenv("BSC_API_TOKEN", "dummy-token")
	os.Setenv("SMD_API_TOKEN", "dummy-token")
	os.Setenv("C21_API_TOKEN", "dummy-token")
	os.Setenv("ASC_API_TOKEN", "dummy-token")
	os.Setenv("BJP_API_TOKEN", "dummy-token")
	os.Setenv("GTG_API_TOKEN", "dummy-token")
	os.Setenv("OAT_API_TOKEN", "dummy-token")
	// Prepare a temporary CSV file
	csvContent := `Source Document Key,PO Numbers,Master Tracking #,Ready Date/Time,Shipment Charges Applied Total,Ship Carrier Name,Billing Type,Recipient Customer ID,Sale Source (UDF)
0083419,88NJSQS3DD,1Z972Y3Y0301141377,7/16/2025,0,UPS,Consignee,0090671,SM
0083510,ATG32GC3XX,1Z972Y3Y0312933410,7/17/2025,17.79,UPS,Prepaid,0090671,BSC
0083511,ATG32GC3XY,1Z972Y3Y0312933411,7/18/2025,18.00,UPS,Prepaid,0090671,BSC
`
	tmpFile, err := os.CreateTemp("", "test_shipments_*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Simulate one failure on the second call
	mockClient := &MockFaireClient{
		FailOnCall: map[int]bool{2: true},
	}

	processed, failed, err := ProcessShipments(tmpFile.Name(), mockClient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(processed) != 2 {
		t.Errorf("expected 2 processed shipments, got %d", len(processed))
	}
	if len(failed) != 1 {
		t.Errorf("expected 1 failed shipment, got %d", len(failed))
	}

	// Check that the log file was created in the logs directory
	logDir := "logs"
	files, err := os.ReadDir(logDir)
	if err != nil {
		t.Fatalf("could not read logs directory: %v", err)
	}
	found := false
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".txt") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected a log file to be created in %s", logDir)
	}
}
