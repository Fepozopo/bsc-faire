package app

import (
	"fmt"
	"strings"

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

// Bubble Tea TUI to display and navigate a list of orders
type orderListModel struct {
	orders   []Order
	selected int
	quitting bool
}

func (m orderListModel) Init() tea.Cmd {
	return nil
}

func (m orderListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.orders)-1 {
				m.selected++
			}
		case "enter":
			// Show details for selected order
			return m, func() tea.Msg {
				ShowOrderTUI(m.orders[m.selected])
				return nil
			}
		}
	}
	return m, nil
}

func (m orderListModel) View() string {
	if m.quitting {
		return ""
	}
	var b strings.Builder
	b.WriteString("Orders:\n\n")
	for i, order := range m.orders {
		cursor := " " // no cursor
		if m.selected == i {
			cursor = ">" // cursor
		}
		line := fmt.Sprintf("%s [%s] %s | %s | %s\n", cursor, order.DisplayID, order.State, order.CreatedAt.Format("2006-01-02"), order.Address.CompanyName)
		b.WriteString(line)
	}
	b.WriteString("\nUse ↑/↓ or j/k to move, Enter to view, q to quit.")
	return b.String()
}

func ShowOrdersTUI(orders []Order) error {
	p := tea.NewProgram(orderListModel{orders: orders})
	_, err := p.Run()
	return err
}
