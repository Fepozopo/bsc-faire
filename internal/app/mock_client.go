package app

// MockFaireClient implements FaireClientInterface for testing and development
type MockFaireClient struct {
	CallCount  int
	FailOnCall map[int]bool // map of call indices to fail (1-based)
}

func (m *MockFaireClient) AddShipment(payload ShipmentPayload, apiToken string) error {
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
