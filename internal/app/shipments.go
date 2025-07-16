package app

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

type Shipment struct {
	CustomerNumber string
	PONumber       string
	BillingType    string
	Carrier        string
	TrackingCode   string
	MakerCostCents int
}

func ParseShipmentsCSV(path string) ([]Shipment, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := csv.NewReader(file)
	headers, err := r.Read()
	if err != nil {
		return nil, err
	}

	var idx = make(map[string]int)
	for i, h := range headers {
		idx[h] = i
	}

	var shipments []Shipment
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if record[idx["CustomerNumber"]] != "0090671" {
			continue
		}
		makerCost := 0
		fmt.Sscanf(record[idx["MakerCostCents"]], "%d", &makerCost)
		shipments = append(shipments, Shipment{
			CustomerNumber: record[idx["CustomerNumber"]],
			PONumber:       record[idx["PONumber"]],
			BillingType:    record[idx["BillingType"]],
			Carrier:        record[idx["Carrier"]],
			TrackingCode:   record[idx["TrackingCode"]],
			MakerCostCents: makerCost,
		})
	}
	return shipments, nil
}
