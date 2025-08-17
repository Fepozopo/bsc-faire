package app

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	statesFlag string
	mockFlag   bool
	failsFlag  string
)

func init() {
	rootCmd.AddCommand(processCmd)
	rootCmd.AddCommand(ordersCmd)
	rootCmd.AddCommand(orderCmd)
	rootCmd.AddCommand(exportCmd)

	processCmd.Flags().BoolVar(&mockFlag, "mock", false, "Use mock Faire client (no real API calls)")
	processCmd.Flags().StringVar(&failsFlag, "fails", "", "Comma-separated list of shipment indices to fail (mock only)")

	ordersCmd.Flags().StringVar(&statesFlag, "states", "", "Comma-separated list of order states to include (e.g. NEW,DELIVERED)")
	ordersCmd.Flags().BoolVar(&mockFlag, "mock", false, "Use mock Faire client (no real API calls)")

	orderCmd.Flags().BoolVar(&mockFlag, "mock", false, "Use mock Faire client (no real API calls)")
	exportCmd.Flags().BoolVar(&mockFlag, "mock", false, "Generate only CSV headers (no real API calls)")
}

var processCmd = &cobra.Command{
	Use:   "process [csvfile]",
	Short: "Process shipments from a CSV file and add them to Faire orders",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var client FaireClientInterface
		if mockFlag {
			failMap := map[int]bool{}
			if failsFlag != "" {
				for _, s := range strings.Split(failsFlag, ",") {
					if idx, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
						failMap[idx] = true
					}
				}
			}
			client = &MockFaireClient{FailOnCall: failMap}
		} else {
			client = NewFaireClient()
		}
		processed, failed, err := ProcessShipments(args[0], client)
		if err != nil {
			return err
		}
		return ShowProcessedTUI(processed, failed)
	},
}

var ordersCmd = &cobra.Command{
	Use:   "orders [sale_source]",
	Short: "Get all orders by sale source (21, asc, bjp, bsc, gtg, oat, sm)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var client FaireClientInterface
		if mockFlag {
			// Use shared MockOrders from mock_client.go
			client = &MockFaireClient{Orders: MockOrders}
		} else {
			client = NewFaireClient()
		}
		var token string
		if len(args) == 0 {
			return fmt.Errorf("sale source is required (21, asc, bjp, bsc, gtg, oat, sm)")
		} else if len(args) > 1 {
			return fmt.Errorf("too many arguments, expected 1 (got %d)", len(args))
		} else {
			saleSource := args[0]
			var err error
			token, err = GetToken(saleSource)
			if err != nil || token == "" {
				fmt.Println("Invalid sale source. Must be one of: 21, asc, bjp, bsc, gtg, oat, sm")
				return nil
			}
		}

		// Always use internal pagination to fetch all orders; limit and page flags removed
		limit := 50

		// All possible states
		allStates := []string{
			"DELIVERED", "BACKORDERED", "CANCELED", "NEW", "PROCESSING", "PRE_TRANSIT",
			"IN_TRANSIT", "RETURNED", "PENDING_RETAILER_CONFIRMATION", "DAMAGED_OR_MISSING",
		}

		// Default excluded states
		states := "DELIVERED,BACKORDERED,CANCELED,PRE_TRANSIT,IN_TRANSIT,RETURNED,PENDING_RETAILER_CONFIRMATION,DAMAGED_OR_MISSING"
		if statesFlag != "" {
			// Split by comma, trim spaces, and capitalize
			userStates := strings.Split(statesFlag, ",")
			stateSet := make(map[string]struct{})
			for _, s := range userStates {
				stateSet[strings.ToUpper(strings.TrimSpace(s))] = struct{}{}
			}
			// Exclude everything except the user-provided states
			var filtered []string
			for _, state := range allStates {
				if _, keep := stateSet[state]; !keep {
					filtered = append(filtered, state)
				}
			}
			states = strings.Join(filtered, ",")
		}

		var allOrders []Order
		currPage := 1
		for {
			resp, err := client.GetAllOrders(token, limit, currPage, states)
			if err != nil {
				return err
			}
			var ordersResp Orders
			if err := json.Unmarshal(resp, &ordersResp); err != nil {
				return fmt.Errorf("failed to parse orders: %w", err)
			}
			allOrders = append(allOrders, ordersResp.Orders...)
			if len(ordersResp.Orders) < limit {
				break
			}
			currPage++
		}

		if statesFlag != "" {
			allowed := make(map[string]struct{})
			for _, s := range strings.Split(statesFlag, ",") {
				allowed[strings.ToUpper(strings.TrimSpace(s))] = struct{}{}
			}
			var filtered []Order
			for _, order := range allOrders {
				if _, ok := allowed[strings.ToUpper(order.State)]; ok {
					filtered = append(filtered, order)
				}
			}
			allOrders = filtered
		}
		return ShowOrdersTUI(allOrders)
	},
}

var orderCmd = &cobra.Command{
	Use:   "order [sale_source] [orderID]",
	Short: "Get a single order by sale source (sm or bsc) and ID",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var client FaireClientInterface
		if mockFlag {
			// Use shared MockOrders from mock_client.go
			client = &MockFaireClient{Orders: MockOrders}
		} else {
			client = NewFaireClient()
		}
		if len(args) != 2 {
			return fmt.Errorf("expected 2 arguments (sale source and order ID), got %d", len(args))
		}
		saleSource := args[0]
		var token string
		var err error
		token, err = GetToken(saleSource)
		if err != nil || token == "" {
			fmt.Println("Invalid sale source. Must be one of: 21, asc, bjp, bsc, gtg, oat, sm")
			return nil
		}
		orderID := args[1]
		resp, err := client.GetOrderByID(orderID, token)
		if err != nil {
			return err
		}
		var order Order
		if err := json.Unmarshal(resp, &order); err != nil {
			return err
		}
		if err := ShowOrderTUI(order); err != nil {
			return err
		}
		return nil
	},
}

var exportCmd = &cobra.Command{
	Use:   "export [sale_source]",
	Short: "Export NEW orders to a CSV file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if mockFlag {
			if err := WriteMockOrdersCSV("faire_new_orders.csv"); err != nil {
				return err
			}
			fmt.Printf("Exported CSV headers to faire_new_orders.csv\n")
			return nil
		}
		client := NewFaireClient()
		count, err := client.ExportNewOrdersToCSV(args[0], "faire_new_orders.csv")
		if err != nil {
			return err
		}
		fmt.Printf("Exported %d new orders to faire_new_orders.csv\n", count)
		return nil
	},
}
