package app

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Shipment struct {
	CustomerNumber string
	PONumber       string
	BillingType    string
	Carrier        string
	TrackingCode   string
	MakerCostCents int
	SaleSource     string
}

func ParseShipmentsCSV(path string) ([]Shipment, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	r := csv.NewReader(file)

	headers, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}

	var idx = make(map[string]int)
	for i, h := range headers {
		// Clean up header names from potential leading/trailing spaces if any
		idx[strings.TrimSpace(h)] = i
	}

	// Validate required headers exist
	requiredHeaders := []string{"Source Document Key", "PO Numbers", "Master Tracking #", "Shipment Charges Applied Total", "Ship Carrier Name", "Billing Type", "Recipient Customer ID", "Sale Source (UDF)"}
	for _, rh := range requiredHeaders {
		if _, ok := idx[rh]; !ok {
			return nil, fmt.Errorf("missing required header: %s", rh)
		}
	}

	validSaleSources := map[string]struct{}{
		"21":  {},
		"ASC": {},
		"BJP": {},
		"BSC": {},
		"GTG": {},
		"OAT": {},
		"SM":  {},
	}

	var shipments []Shipment
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read record: %w", err)
		}

		// Clean and retrieve Recipient Customer ID
		recipientCustomerID := record[idx["Recipient Customer ID"]]

		saleSource := record[idx["Sale Source (UDF)"]]
		if recipientCustomerID != "0090671" {
			continue
		}
		if _, ok := validSaleSources[saleSource]; !ok {
			continue
		}

		// Parse Shipment Charges Applied Total
		makerCostStr := record[idx["Shipment Charges Applied Total"]]
		makerCostDollars, err := strconv.ParseFloat(makerCostStr, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse MakerCostCents '%s': %w", makerCostStr, err)
		}
		// Convert dollars to cents (multiply by 100) and cast to int
		makerCostCents := int(makerCostDollars * 100)

		shipments = append(shipments, Shipment{
			CustomerNumber: recipientCustomerID,
			PONumber:       record[idx["PO Numbers"]],
			BillingType:    record[idx["Billing Type"]],
			Carrier:        record[idx["Ship Carrier Name"]],
			TrackingCode:   record[idx["Master Tracking #"]],
			MakerCostCents: makerCostCents,
			SaleSource:     saleSource,
		})
	}
	return shipments, nil
}
