package app

import (
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
	os.Unsetenv("FAIRE_API_TOKEN")
	defer func() {
		recover()
	}()
	NewFaireClient()
}
