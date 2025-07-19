package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type orderModel struct {
	order Order
}

func (m orderModel) Init() tea.Cmd { return nil }

func (m orderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m orderModel) View() string {
	created := m.order.CreatedAt.Format("2006-01-02 15:04")
	retailer := m.order.Address.CompanyName
	if retailer == "" {
		retailer = m.order.Address.Name
	}

	// Calculate total
	var totalCents int
	for _, item := range m.order.Items {
		totalCents += item.PriceCents * item.Quantity
	}

	s := fmt.Sprintf(
		"Order ID: %s\nStatus: %s\nRetailer: %s\nCreated: %s\nTotal: $%.2f\n\nItems:\n",
		m.order.DisplayID, m.order.State, retailer, created, float64(totalCents)/100,
	)
	for _, item := range m.order.Items {
		s += fmt.Sprintf("  - %s (%s) x%d ($%.2f each)\n",
			item.ProductName, item.Sku, item.Quantity, float64(item.PriceCents)/100)
	}
	s += "\nPress 'q' or 'esc' to quit."
	return s
}

func ShowOrderTUI(order Order) error {
	p := tea.NewProgram(orderModel{order: order})
	_, err := p.Run()
	return err
}
