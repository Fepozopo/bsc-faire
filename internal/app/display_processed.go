package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type processedModel struct {
	processed []ShipmentPayload
	failed    []ShipmentPayload
	quitting  bool
}

func (m processedModel) Init() tea.Cmd {
	return nil
}

func (m processedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m processedModel) View() string {
	var b strings.Builder
	b.WriteString("Shipments Processing Results\n\n")
	b.WriteString(fmt.Sprintf("Processed Shipments (%d):\n", len(m.processed)))
	if len(m.processed) == 0 {
		b.WriteString("  None\n")
	} else {
		for _, p := range m.processed {
			b.WriteString(fmt.Sprintf(
				"  OrderID: %s\n    MakerCostCents: %d\n    Carrier: %s\n    TrackingCode: %s\n    ShippingType: %s\n    SaleSource: %s\n",
				p.OrderID, p.MakerCostCents, p.Carrier, p.TrackingCode, p.ShippingType, p.SaleSource,
			))
		}
	}
	b.WriteString(fmt.Sprintf("\nFailed Shipments (%d):\n", len(m.failed)))
	if len(m.failed) == 0 {
		b.WriteString("  None\n")
	} else {
		for _, p := range m.failed {
			b.WriteString(fmt.Sprintf(
				"  OrderID: %s\n    MakerCostCents: %d\n    Carrier: %s\n    TrackingCode: %s\n    ShippingType: %s\n    SaleSource: %s\n",
				p.OrderID, p.MakerCostCents, p.Carrier, p.TrackingCode, p.ShippingType, p.SaleSource,
			))
		}
	}
	b.WriteString("\nPress q or esc to quit.\n")
	return b.String()
}

func ShowProcessedTUI(processed, failed []ShipmentPayload) error {
	p := tea.NewProgram(processedModel{processed: processed, failed: failed})
	_, err := p.Run()
	return err
}
