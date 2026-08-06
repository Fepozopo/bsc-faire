package app

import (
	"encoding/json"
	"strings"
	"time"
)

// MockFaireClient implements FaireClientInterface for local testing and development.
type MockFaireClient struct {
	CallCount  int
	FailOnCall map[int]bool // Map of one-based shipment call indices that should fail.
	Orders     []Order      // Orders returned when a test does not supply its own set.
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
	{
		ID:         "mock012",
		DisplayID:  "MOCK-ORDER-4",
		State:      OrderStateBackordered,
		RetailerID: "retailer_004",
	},
}

// AddShipment simulates adding a shipment and fails configured calls.
func (m *MockFaireClient) AddShipment(payload ShipmentPayload, apiToken string) error {
	time.Sleep(300 * time.Millisecond) // Simulate network/processing delay
	m.CallCount++
	if m.FailOnCall != nil && m.FailOnCall[m.CallCount] {
		return &MockError{"simulated failure"}
	}
	return nil
}

// MockError is an error returned by MockFaireClient when simulating a failed API call.
type MockError struct {
	msg string
}

// Error returns the simulated failure message.
func (e *MockError) Error() string {
	return e.msg
}

// GetAllOrders returns mock orders as JSON after applying the supplied inverse state filter.
func (m *MockFaireClient) GetAllOrders(apiToken string, limit int, page int, excludedStates string) ([]byte, error) {
	time.Sleep(300 * time.Millisecond) // Match the asynchronous timing of the live client during manual testing.
	m.CallCount++
	orders := m.Orders
	if orders == nil {
		orders = MockOrders
	}

	excluded := make(map[string]struct{})
	for _, state := range strings.Split(excludedStates, ",") {
		excluded[strings.ToUpper(strings.TrimSpace(state))] = struct{}{}
	}
	filteredOrders := make([]Order, 0, len(orders))
	for _, order := range orders {
		if _, isExcluded := excluded[strings.ToUpper(order.State)]; !isExcluded {
			filteredOrders = append(filteredOrders, order)
		}
	}

	resp := Orders{Page: page, Limit: limit, Orders: filteredOrders}
	return json.Marshal(resp)
}

// GetOrderByID returns a single mock order by display ID or internal mock ID as JSON.
func (m *MockFaireClient) GetOrderByID(orderIdentifier string, apiToken string) ([]byte, error) {
	time.Sleep(300 * time.Millisecond) // Simulate network/processing delay
	m.CallCount++
	orders := m.Orders
	if orders == nil {
		orders = MockOrders
	}
	for _, order := range orders {
		if strings.EqualFold(order.DisplayID, orderIdentifier) || strings.EqualFold(order.ID, orderIdentifier) {
			return json.Marshal(order)
		}
	}
	return nil, &MockError{"order not found"}
}
