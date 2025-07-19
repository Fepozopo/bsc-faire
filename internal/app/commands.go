package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(processCmd)
	rootCmd.AddCommand(ordersCmd)
	rootCmd.AddCommand(orderCmd)
}

var processCmd = &cobra.Command{
	Use:   "process [csvfile]",
	Short: "Process shipments from a CSV file and add them to Faire orders",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return ProcessShipments(args[0])
	},
}

var ordersCmd = &cobra.Command{
	Use:   "orders [sale_source]",
	Short: "Get all orders by sale source (sm or bsc)",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := NewFaireClient()
		var token string
		if len(args) == 0 {
			return fmt.Errorf("sale source is required (sm or bsc)")
		} else if len(args) > 1 {
			return fmt.Errorf("too many arguments, expected 1 (got %d)", len(args))
		} else {
			saleSource := args[0]
			switch saleSource {
			case "sm":
				token = os.Getenv("SMD_API_TOKEN")
			case "bsc":
				token = os.Getenv("BSC_API_TOKEN")
			default:
				return fmt.Errorf("invalid sale source: %s (must be 'sm' or 'bsc')", saleSource)
			}
		}
		resp, err := client.GetAllOrders(token)
		if err != nil {
			return err
		}
		var ordersResp Orders
		if err := json.Unmarshal(resp, &ordersResp); err != nil {
			return fmt.Errorf("failed to parse orders: %w", err)
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
		case "sm":
			token = os.Getenv("SMD_API_TOKEN")
		case "bsc":
			token = os.Getenv("BSC_API_TOKEN")
		default:
			return fmt.Errorf("invalid sale source: %s (must be 'sm' or 'bsc')", saleSource)
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
