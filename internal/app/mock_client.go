package app

import (
	"encoding/json"
	"time"
)

// MockFaireClient implements FaireClientInterface for testing and development
type MockFaireClient struct {
	CallCount  int
	FailOnCall map[int]bool // map of call indices to fail (1-based)
	Orders     []Order      // mock orders for testing
}

// MockOrders is a shared set of mock orders for testing/demo
var MockOrders = []Order{
	{
		ID:         "mock123",
		DisplayID:  "MOCK-ORDER-1",
		State:      "NEW",
		RetailerID: "retailer_001",
	},
	{
		ID:         "mock456",
		DisplayID:  "MOCK-ORDER-2",
		State:      "PROCESSING",
		RetailerID: "retailer_002",
	},
	{
		ID:         "mock789",
		DisplayID:  "MOCK-ORDER-3",
		State:      "NEW",
		RetailerID: "retailer_003",
	},
}

func (m *MockFaireClient) AddShipment(payload ShipmentPayload, apiToken string) error {
	time.Sleep(300 * time.Millisecond) // Simulate network/processing delay
	m.CallCount++
	if m.FailOnCall != nil && m.FailOnCall[m.CallCount] {
		return &MockError{"simulated failure"}
	}
	return nil
}

type MockError struct {
	msg string
}

func (e *MockError) Error() string {
	return e.msg
}

// GetAllOrders returns mock orders as JSON ([]byte)
func (m *MockFaireClient) GetAllOrders(apiToken string, limit int, page int, states string) ([]byte, error) {
	time.Sleep(300 * time.Millisecond) // Simulate network/processing delay
	m.CallCount++
	orders := m.Orders
	if orders == nil {
		orders = MockOrders
	}
	resp := Orders{Page: page, Limit: limit, Orders: orders}
	return json.Marshal(resp)
}

// GetOrderByID returns a single mock order as JSON ([]byte)
func (m *MockFaireClient) GetOrderByID(PONumber string, apiToken string) ([]byte, error) {
	time.Sleep(300 * time.Millisecond) // Simulate network/processing delay
	m.CallCount++
	orders := m.Orders
	if orders == nil {
		orders = MockOrders
	}
	for _, order := range orders {
		if order.DisplayID == PONumber || order.ID == PONumber {
			return json.Marshal(order)
		}
	}
	return nil, &MockError{"order not found"}
}
