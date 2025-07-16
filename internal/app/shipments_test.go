package app

import (
	"os"
	"testing"
)

func TestParseShipmentsCSV(t *testing.T) {
	f, err := os.CreateTemp("", "shipments-*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("CustomerNumber,PONumber,BillingType,Carrier,TrackingCode,MakerCostCents\n")
	f.WriteString("0090671,ORDER123,Consignee,UPS,TRACK123,1000\n")
	f.WriteString("0090671,ORDER124,Prepaid,FedEx,TRACK124,2000\n")
	f.WriteString("0000000,ORDER125,Prepaid,FedEx,TRACK125,3000\n")
	f.Close()

	shipments, err := ParseShipmentsCSV(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(shipments) != 2 {
		t.Errorf("expected 2 shipments, got %d", len(shipments))
	}
	if shipments[0].PONumber != "ORDER123" || shipments[1].PONumber != "ORDER124" {
		t.Errorf("unexpected PO numbers: %+v", shipments)
	}
}
