package app

import (
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
	Short: "Get all orders (optionally specify sale source: SM or BSC)",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := NewFaireClient()
		var token string
		if len(args) == 0 {
			token = os.Getenv("FAIRE_API_TOKEN")
		} else {
			saleSource := args[0]
			switch saleSource {
			case "SM":
				token = os.Getenv("SMD_API_TOKEN")
			case "BSC":
				token = os.Getenv("BSC_API_TOKEN")
			default:
				return fmt.Errorf("invalid sale source: %s (must be 'SM' or 'BSC')", saleSource)
			}
		}
		resp, err := client.GetAllOrders(token)
		if err != nil {
			return err
		}
		fmt.Println(string(resp))
		return nil
	},
}

var orderCmd = &cobra.Command{
	Use:   "order [sale_source] [orderID]",
	Short: "Get a single order by sale source (SM or BSC) and ID",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := NewFaireClient()
		saleSource := args[0]
		var token string
		switch saleSource {
		case "SM":
			token = os.Getenv("SMD_API_TOKEN")
		case "BSC":
			token = os.Getenv("BSC_API_TOKEN")
		default:
			return fmt.Errorf("invalid sale source: %s (must be 'SM' or 'BSC')", saleSource)
		}
		orderID := args[1]
		resp, err := client.GetOrderByID(orderID, token)
		if err != nil {
			return err
		}
		fmt.Println(string(resp))
		return nil
	},
}
