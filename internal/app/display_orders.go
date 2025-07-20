package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// orderModel is a Bubble Tea model for displaying a single order in a TUI.
type orderModel struct {
	order Order
}

// Init implements the tea.Model interface. No initialization needed here.
func (m orderModel) Init() tea.Cmd { return nil }

// Update handles key events for the single order view.
// Pressing 'q' or 'esc' will quit the view.
func (m orderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" {
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the order details as a string for the TUI.
func (m orderModel) View() string {
	created := m.order.CreatedAt.Format("2006-01-02 15:04")
	retailer := m.order.Address.CompanyName
	if retailer == "" {
		retailer = m.order.Address.Name
	}

	// Calculate total order value in cents
	var totalCents int
	for _, item := range m.order.Items {
		totalCents += item.PriceCents * item.Quantity
	}

	// Build the order summary string
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

// ShowOrderTUI launches a Bubble Tea TUI to display a single order.
// Blocks until the user quits the view.
func ShowOrderTUI(order Order) error {
	p := tea.NewProgram(orderModel{order: order})
	_, err := p.Run()
	return err
}

// orderListModel is a Bubble Tea model for displaying and navigating a list of orders.
type orderListModel struct {
	orders   []Order // List of orders to display
	selected int     // Index of the currently selected order
	quitting bool    // Whether the user has chosen to quit
}

// Init implements the tea.Model interface. No initialization needed here.
func (m orderListModel) Init() tea.Cmd {
	return nil
}

// Update handles key events for the order list view.
// Supports navigation (up/down), viewing an order (enter), and quitting (q/esc).
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
			// Show details for selected order in a new TUI.
			return m, func() tea.Msg {
				ShowOrderTUI(m.orders[m.selected])
				return nil
			}
		}
	}
	return m, nil
}

// View renders the list of orders, highlighting the selected one.
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

// ShowOrdersTUI launches a Bubble Tea TUI to display and navigate a list of orders.
// Blocks until the user quits the view.
func ShowOrdersTUI(orders []Order) error {
	p := tea.NewProgram(orderListModel{orders: orders})
	_, err := p.Run()
	return err
}
