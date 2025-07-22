package app

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	limitFlag  int
	pageFlag   int
	statesFlag string
)

func init() {
	rootCmd.AddCommand(processCmd)
	rootCmd.AddCommand(ordersCmd)
	rootCmd.AddCommand(orderCmd)
	rootCmd.AddCommand(testProcessCmd)

	ordersCmd.Flags().IntVar(&limitFlag, "limit", 50, "Max number of orders to return (10-50)")
	ordersCmd.Flags().IntVar(&pageFlag, "page", 1, "Page number to return (default 1)")
	ordersCmd.Flags().StringVar(&statesFlag, "states", "", "Comma separated list of states to exclude. If set, will exclude all except the provided state.")
}

var processCmd = &cobra.Command{
	Use:   "process [csvfile]",
	Short: "Process shipments from a CSV file and add them to Faire orders",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		processed, failed, err := ProcessShipments(args[0])
		if err != nil {
			return err
		}
		return ShowProcessedTUI(processed, failed)
	},
}

var ordersCmd = &cobra.Command{
	Use:   "orders [sale_source]",
	Short: "Get all orders by sale source (21, asc, bjp, bsc, gtg, oat, smd)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := NewFaireClient()
		var token string
		if len(args) == 0 {
			return fmt.Errorf("sale source is required (21, asc, bjp, bsc, gtg, oat, smd)")
		} else if len(args) > 1 {
			return fmt.Errorf("too many arguments, expected 1 (got %d)", len(args))
		} else {
			saleSource := args[0]
			switch saleSource {
			case "21":
				token = os.Getenv("C21_API_TOKEN")
			case "asc":
				token = os.Getenv("ASC_API_TOKEN")
			case "bjp":
				token = os.Getenv("BJP_API_TOKEN")
			case "bsc":
				token = os.Getenv("BSC_API_TOKEN")
			case "gtg":
				token = os.Getenv("GTG_API_TOKEN")
			case "oat":
				token = os.Getenv("OAT_API_TOKEN")
			case "sm":
				token = os.Getenv("SMD_API_TOKEN")
			default:
				return fmt.Errorf("invalid sale source: %s (must be 21, asc, bjp, bsc, gtg, oat, or smd)", saleSource)
			}
		}

		// Validate limit
		limit := limitFlag
		if limit < 10 || limit > 50 {
			limit = 50
		}
		page := pageFlag
		if page < 1 {
			page = 1
		}

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

		resp, err := client.GetAllOrders(token, limit, page, states)
		if err != nil {
			return err
		}
		var ordersResp Orders
		if err := json.Unmarshal(resp, &ordersResp); err != nil {
			return fmt.Errorf("failed to parse orders: %w", err)
		}

		if statesFlag != "" {
			allowed := make(map[string]struct{})
			for _, s := range strings.Split(statesFlag, ",") {
				allowed[strings.ToUpper(strings.TrimSpace(s))] = struct{}{}
			}
			var filtered []Order
			for _, order := range ordersResp.Orders {
				if _, ok := allowed[strings.ToUpper(order.State)]; ok {
					filtered = append(filtered, order)
				}
			}
			ordersResp.Orders = filtered
		}
		return ShowOrdersTUI(ordersResp.Orders)
	},
}

var orderCmd = &cobra.Command{
	Use:   "order [sale_source] [orderID]",
	Short: "Get a single order by sale source (sm or bsc) and ID",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := NewFaireClient()
		if len(args) != 2 {
			return fmt.Errorf("expected 2 arguments (sale source and order ID), got %d", len(args))
		}
		saleSource := args[0]
		var token string
		switch saleSource {
		case "21":
			token = os.Getenv("C21_API_TOKEN")
		case "asc":
			token = os.Getenv("ASC_API_TOKEN")
		case "bjp":
			token = os.Getenv("BJP_API_TOKEN")
		case "bsc":
			token = os.Getenv("BSC_API_TOKEN")
		case "gtg":
			token = os.Getenv("GTG_API_TOKEN")
		case "oat":
			token = os.Getenv("OAT_API_TOKEN")
		case "sm":
			token = os.Getenv("SMD_API_TOKEN")
		default:
			return fmt.Errorf("invalid sale source: %s (must be 21, asc, bjp, bsc, gtg, oat, or smd)", saleSource)
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

var testProcessCmd = &cobra.Command{
	Use:   "test-process",
	Short: "Preview the processed shipments TUI with sample data",
	RunE: func(cmd *cobra.Command, args []string) error {
		processed := []ShipmentPayload{
			{OrderID: "BXDMJBWXID", MakerCostCents: 1000, Carrier: "UPS", TrackingCode: "1Z999AA10123456784", ShippingType: "SHIP_WITH_FAIRE"},
			{OrderID: "ABCD1234", MakerCostCents: 2000, Carrier: "FedEx", TrackingCode: "123456789", ShippingType: "SHIP_ON_YOUR_OWN"},
		}
		failed := []ShipmentPayload{
			{OrderID: "FAILED001", MakerCostCents: 3000, Carrier: "DHL", TrackingCode: "DHLTRACK001", ShippingType: "SHIP_ON_YOUR_OWN"},
		}
		return ShowProcessedTUI(processed, failed)
	},
}
