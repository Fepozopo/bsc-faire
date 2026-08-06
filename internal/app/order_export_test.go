package app

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// exportTestClient is a deterministic OrderClient used to verify export behavior without HTTP.
type exportTestClient struct {
	ordersByID     map[string]Order
	ordersByPage   map[int][]Order
	excludedStates []string
	requestedIDs   []string
}

// GetAllOrders records the inverse filter and returns the configured page of test orders.
func (c *exportTestClient) GetAllOrders(apiToken string, limit int, page int, excludedStates string) ([]byte, error) {
	c.excludedStates = append(c.excludedStates, excludedStates)
	return json.Marshal(Orders{Page: page, Limit: limit, Orders: c.ordersByPage[page]})
}

// GetOrderByID records and returns the configured order for orderIdentifier.
func (c *exportTestClient) GetOrderByID(orderIdentifier string, apiToken string) ([]byte, error) {
	c.requestedIDs = append(c.requestedIDs, orderIdentifier)
	return json.Marshal(c.ordersByID[orderIdentifier])
}

// TestExportOrdersToCSVByState confirms BACKORDERED uses inverse filtering and excludes unexpected order states locally.
func TestExportOrdersToCSVByState(t *testing.T) {
	client := &exportTestClient{ordersByPage: map[int][]Order{
		1: {testOrder("bo_backordered", "BACKORDER-1", OrderStateBackordered), testOrder("bo_new", "NEW-1", OrderStateNew)},
	}}
	filename := filepath.Join(t.TempDir(), "backordered.csv")

	count, err := ExportOrdersToCSV(client, "token", "bsc", filename, OrderExportFilter{State: OrderStateBackordered})
	if err != nil {
		t.Fatalf("ExportOrdersToCSV returned an error: %v", err)
	}
	if count != 1 {
		t.Fatalf("exported order count = %d, want 1", count)
	}

	if len(client.excludedStates) != 1 {
		t.Fatalf("GetAllOrders call count = %d, want 1", len(client.excludedStates))
	}
	if strings.Contains(client.excludedStates[0], OrderStateBackordered) {
		t.Fatalf("inverse filter %q excludes BACKORDERED", client.excludedStates[0])
	}
	for _, state := range faireOrderStates {
		if state != OrderStateBackordered && !strings.Contains(client.excludedStates[0], state) {
			t.Errorf("inverse filter %q does not exclude %s", client.excludedStates[0], state)
		}
	}

	rows := readExportCSV(t, filename)
	if len(rows) != 2 {
		t.Fatalf("CSV row count = %d, want 2", len(rows))
	}
	if rows[1][0] != "bo_backordered" {
		t.Errorf("CSV order ID = %q, want bo_backordered", rows[1][0])
	}
	if rows[1][23] != "BSC" {
		t.Errorf("CSV sale source = %q, want BSC", rows[1][23])
	}
}

// TestExportOrdersToCSVByIdentifiers confirms a user-supplied order list preserves order and deduplicates identifiers.
func TestExportOrdersToCSVByIdentifiers(t *testing.T) {
	identifiers := ParseOrderIdentifiers("FIRST-1, second-2\nFIRST-1")
	wantIdentifiers := []string{"FIRST-1", "second-2"}
	if !reflect.DeepEqual(identifiers, wantIdentifiers) {
		t.Fatalf("ParseOrderIdentifiers() = %#v, want %#v", identifiers, wantIdentifiers)
	}

	client := &exportTestClient{ordersByID: map[string]Order{
		"FIRST-1":  testOrder("bo_first", "FIRST-1", OrderStateNew),
		"second-2": testOrder("bo_second", "SECOND-2", OrderStateBackordered),
	}}
	filename := filepath.Join(t.TempDir(), "selected.csv")

	count, err := ExportOrdersToCSV(client, "token", "asc", filename, OrderExportFilter{OrderIdentifiers: identifiers})
	if err != nil {
		t.Fatalf("ExportOrdersToCSV returned an error: %v", err)
	}
	if count != 2 {
		t.Fatalf("exported order count = %d, want 2", count)
	}
	if !reflect.DeepEqual(client.requestedIDs, wantIdentifiers) {
		t.Errorf("requested identifiers = %#v, want %#v", client.requestedIDs, wantIdentifiers)
	}

	rows := readExportCSV(t, filename)
	if len(rows) != 3 {
		t.Fatalf("CSV row count = %d, want 3", len(rows))
	}
	if rows[1][0] != "bo_first" || rows[2][0] != "bo_second" {
		t.Errorf("CSV order IDs = %q, %q; want bo_first, bo_second", rows[1][0], rows[2][0])
	}
}

// TestDownloadsFilePath returns a created Downloads path under the current user's home directory.
func TestDownloadsFilePath(t *testing.T) {
	homeDirectory := t.TempDir()
	t.Setenv("HOME", homeDirectory)

	path, err := DownloadsFilePath("orders.csv")
	if err != nil {
		t.Fatalf("DownloadsFilePath returned an error: %v", err)
	}
	if want := filepath.Join(homeDirectory, "Downloads", "orders.csv"); path != want {
		t.Errorf("DownloadsFilePath() = %q, want %q", path, want)
	}
	if info, err := os.Stat(filepath.Dir(path)); err != nil || !info.IsDir() {
		t.Errorf("Downloads directory was not created: %v", err)
	}
}

// TestExportOrdersToCSVRejectsAnEmptyFilter prevents an accidental unfiltered order export.
func TestExportOrdersToCSVRejectsAnEmptyFilter(t *testing.T) {
	_, err := ExportOrdersToCSV(&exportTestClient{}, "token", "bsc", filepath.Join(t.TempDir(), "orders.csv"), OrderExportFilter{})
	if err == nil || !strings.Contains(err.Error(), "unsupported order state") {
		t.Fatalf("ExportOrdersToCSV empty filter error = %v, want unsupported state error", err)
	}
}

// roundTripFunc turns a function into an HTTP transport for listener-free client tests.
type roundTripFunc func(*http.Request) (*http.Response, error)

// RoundTrip sends req through the wrapped test function.
func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

// TestFaireClientOrderRequests verifies query encoding, raw Faire ID support, and non-success HTTP errors.
func TestFaireClientOrderRequests(t *testing.T) {
	originalClient := http.DefaultClient
	t.Cleanup(func() { http.DefaultClient = originalClient })
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		statusCode := http.StatusOK
		body := `{"orders":[]}`
		switch r.URL.Path {
		case "/orders":
			if got, want := r.URL.Query().Get("excluded_states"), "NEW,BACKORDERED"; got != want {
				t.Errorf("excluded_states = %q, want %q", got, want)
			}
			if got, want := r.Header.Get("X-FAIRE-ACCESS-TOKEN"), "test-token"; got != want {
				t.Errorf("access token = %q, want %q", got, want)
			}
		case "/orders/bo_abc123":
			body = `{"id":"bo_abc123"}`
		case "/orders/bo_missing":
			statusCode = http.StatusNotFound
			body = "order not found"
		default:
			statusCode = http.StatusNotFound
			body = "not found"
		}
		return &http.Response{
			StatusCode: statusCode,
			Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})}

	client := &FaireClient{BaseURL: "https://faire.test"}
	if _, err := client.GetAllOrders("test-token", 50, 3, "NEW,BACKORDERED"); err != nil {
		t.Fatalf("GetAllOrders returned an error: %v", err)
	}
	if _, err := client.GetOrderByID("BO_ABC123", "test-token"); err != nil {
		t.Fatalf("GetOrderByID with raw Faire ID returned an error: %v", err)
	}
	if _, err := client.GetOrderByID("missing", "test-token"); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("GetOrderByID non-success error = %v, want HTTP 404 error", err)
	}
}

// testOrder returns a minimally populated order that yields one CSV row.
func testOrder(id, displayID, state string) Order {
	var order Order
	if err := json.Unmarshal([]byte(`{
		"id":"`+id+`",
		"display_id":"`+displayID+`",
		"state":"`+state+`",
		"items":[{"sku":"SKU-1","price_cents":125,"quantity":2}]
	}`), &order); err != nil {
		panic(err)
	}
	return order
}

// readExportCSV reads filename and fails the calling test when it is not a valid CSV file.
func readExportCSV(t *testing.T, filename string) [][]string {
	t.Helper()
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("open CSV: %v", err)
	}
	defer func() { _ = file.Close() }()

	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("read CSV: %v", err)
	}
	return rows
}
