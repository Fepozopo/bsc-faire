package app

import (
	"fmt"

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
	Use:   "orders",
	Short: "Get all orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := NewFaireClient()
		resp, err := client.GetAllOrders()
		if err != nil {
			return err
		}
		fmt.Println(string(resp))
		return nil
	},
}

var orderCmd = &cobra.Command{
	Use:   "order [orderID]",
	Short: "Get a single order by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := NewFaireClient()
		resp, err := client.GetOrderByID(args[0])
		if err != nil {
			return err
		}
		fmt.Println(string(resp))
		return nil
	},
}
