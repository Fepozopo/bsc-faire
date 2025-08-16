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
	return FormatOrder(m.order) + "\nPress 'q' or 'esc' to quit."
}

// ShowOrderTUI launches a Bubble Tea TUI to display a single order.
// Blocks until the user quits the view.
func ShowOrderTUI(order Order) error {
	p := tea.NewProgram(orderModel{order: order})
	_, err := p.Run()
	return err
}

// orderListModel is a Bubble Tea model for displaying and navigating a list of orders,
// and for viewing order details in a single TUI program.
type orderListModel struct {
	orders        []Order // List of orders to display
	selected      int     // Index of the currently selected order
	quitting      bool    // Whether the user has chosen to quit
	viewingDetail bool    // Whether currently viewing order details
	detailOrder   Order   // The order being viewed in detail
}

// Init implements the tea.Model interface. No initialization needed here.
func (m orderListModel) Init() tea.Cmd {
	return nil
}

// Update handles key events for the order list and detail views.
// In detail view, pressing 'q', 'esc', or 'b' returns to the list.
func (m orderListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.viewingDetail {
			switch msg.String() {
			case "q", "esc", "b":
				m.viewingDetail = false
			}
		} else {
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
				m.viewingDetail = true
				m.detailOrder = m.orders[m.selected]
			}
		}
	}
	return m, nil
}

// View renders the list of orders, highlighting the selected one, or the detail view if selected.
func (m orderListModel) View() string {
	if m.quitting {
		return ""
	}
	if m.viewingDetail {
		return FormatOrder(m.detailOrder) + "\nPress 'b', 'q', or 'esc' to go back."
	}
	var b strings.Builder
	b.WriteString("Orders:\n\n")
	for i, order := range m.orders {
		cursor := " " // no cursor
		if m.selected == i {
			cursor = ">" // cursor
		}
		line := fmt.Sprintf("%s [%s] %s | %s | %s\n", cursor, order.DisplayID, order.State, order.ShipAfter.Format("2006-01-02"), order.Address.CompanyName)
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
