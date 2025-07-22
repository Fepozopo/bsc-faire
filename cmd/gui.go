package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	apppkg "github.com/Fepozopo/bsc-faire/internal/app"
	"github.com/joho/godotenv"
	osDialog "github.com/sqweek/dialog"
)

// openFileWindow creates a file open dialog using the system's native file manager.
// When a file is selected, it calls the provided callback with the file path.
// If the user cancels or an error occurs, it shows an error dialog.
func openFileWindow(parent fyne.Window, callback func(filePath string, e error)) {
	filePath, err := osDialog.File().Load() // Use the aliased dialog for the native file open
	if err != nil {
		if err.Error() == "cancelled" {
			dialog.ShowError(fmt.Errorf("file open cancelled: %v", err), parent)
		} else {
			dialog.ShowError(fmt.Errorf("file open failed: %v", err), parent)
		}
		return
	}
	callback(filePath, nil)
}

func RunGUI() {
	myApp := fyneapp.New()
	w := myApp.NewWindow("Faire GUI")

	// Load .env to get API tokens
	godotenv.Load()
	bscToken := os.Getenv("BSC_API_TOKEN")
	smdToken := os.Getenv("SMD_API_TOKEN")

	// Button: Process Shipments CSV
	// - Opens a file dialog for the user to select a CSV file containing shipment data.
	// - Processes the shipments asynchronously.
	// - Displays a dialog with detailed results, including all fields of processed and failed shipments.
	// - Sends a notification with the result summary.
	processBtn := widget.NewButton("Process Shipments CSV", func() {
		openFileWindow(w, func(filePath string, e error) {
			if e != nil {
				// Show error if file selection failed or was cancelled
				dialog.ShowError(e, w)
				return
			}
			if filePath == "" {
				// No file selected, do nothing
				return
			}
			if len(filePath) < 4 || filePath[len(filePath)-4:] != ".csv" {
				// Ensure the selected file is a CSV
				dialog.ShowError(fmt.Errorf("please select a .csv file"), w)
				return
			}
			// Struct for passing results from goroutine
			type resultStruct struct {
				processed []apppkg.ShipmentPayload
				failed    []apppkg.ShipmentPayload
				err       error
			}
			resultCh := make(chan resultStruct)
			// Process shipments in a goroutine to avoid blocking the UI
			go func() {
				processed, failed, err := apppkg.ProcessShipments(filePath)
				resultCh <- resultStruct{processed, failed, err}
			}()
			result := <-resultCh

			// Helper function to format a slice of ShipmentPayloads for display
			var formatPayloads = func(payloads []apppkg.ShipmentPayload) string {
				if len(payloads) == 0 {
					return "  None"
				}
				msg := ""
				for _, p := range payloads {
					msg += fmt.Sprintf(
						"  OrderID: %s\n    MakerCostCents: %d\n    Carrier: %s\n    TrackingCode: %s\n    ShippingType: %s\n",
						p.OrderID, p.MakerCostCents, p.Carrier, p.TrackingCode, p.ShippingType,
					)
				}
				return msg
			}

			// Build the result message for the dialog and notification
			var msg string
			if result.err != nil {
				msg = fmt.Sprintf("Failed to process shipments: %v", result.err)
			} else {
				msg = "Shipments processed successfully!\n\nProcessed Shipments:\n"
				msg += formatPayloads(result.processed)
				msg += "\n\nFailed Shipments:\n"
				msg += formatPayloads(result.failed)
			}
			// Send notification with result summary
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title: func() string {
					if result.err != nil {
						return "Error"
					} else {
						return "Success"
					}
				}(),
				Content: msg,
			})
			// Show dialog with result details
			if result.err != nil {
				dialog.ShowError(fmt.Errorf(msg), w)
			} else {
				// Create a scrollable container for the message
				scroll := container.NewVScroll(widget.NewLabel(msg))
				scroll.SetMinSize(fyne.NewSize(380, 250)) // Adjust size as needed

				// Show a custom dialog with the scrollable content
				dialog.ShowCustom("Success", "OK", scroll, w)
			}
		})
	})

	// Button: Get All Orders
	// - Prompts the user for a sale source ("sm" or "bsc").
	// - Fetches all orders for the selected source asynchronously.
	// - Displays the raw response in a dialog.
	ordersBtn := widget.NewButton("Get All Orders", func() {
		// Prompt for sale source (sm or bsc)
		entry := widget.NewEntry()
		entry.SetPlaceHolder("Enter sale source: sm or bsc")
		dialog.ShowForm("Get All Orders", "Get", "Cancel",
			[]*widget.FormItem{
				widget.NewFormItem("Sale Source", entry),
			}, func(ok bool) {
				if !ok {
					// User cancelled the form
					return
				}
				saleSource := entry.Text
				var token string
				switch saleSource {
				case "sm":
					token = smdToken
				case "bsc":
					token = bscToken
				default:
					// Invalid input, show error
					dialog.ShowError(fmt.Errorf("invalid sale source: must be 'sm' or 'bsc'"), w)
					return
				}
				// Fetch orders asynchronously to avoid blocking the UI
				respCh := make(chan struct {
					resp []byte
					err  error
				})
				go func() {
					client := apppkg.NewFaireClient()
					resp, err := client.GetAllOrders(token, 50, 1, "")
					respCh <- struct {
						resp []byte
						err  error
					}{resp, err}
				}()
				result := <-respCh
				// Show dialog with orders or error
				if result.err != nil {
					dialog.ShowError(fmt.Errorf("failed to get orders: %v", result.err), w)
					return
				}
				dialog.ShowInformation("Orders", string(result.resp), w)
			}, w)
	})

	// Button: Get Order By ID
	// - Prompts the user for a sale source ("sm" or "bsc") and an order ID.
	// - Fetches the order details asynchronously.
	// - Displays the raw order response in a dialog.
	orderBtn := widget.NewButton("Get Order By ID", func() {
		// Prompt for sale source and order ID
		saleSourceEntry := widget.NewEntry()
		saleSourceEntry.SetPlaceHolder("sm or bsc")
		orderIDEntry := widget.NewEntry()
		orderIDEntry.SetPlaceHolder("Order ID (Display ID)")
		dialog.ShowForm("Get Order By ID", "Get", "Cancel",
			[]*widget.FormItem{
				widget.NewFormItem("Sale Source", saleSourceEntry),
				widget.NewFormItem("Order ID", orderIDEntry),
			}, func(ok bool) {
				if !ok {
					// User cancelled the form
					return
				}
				saleSource := saleSourceEntry.Text
				orderID := orderIDEntry.Text
				var token string
				switch saleSource {
				case "sm":
					token = smdToken
				case "bsc":
					token = bscToken
				default:
					// Invalid input, show error
					dialog.ShowError(fmt.Errorf("invalid sale source: must be 'sm' or 'bsc'"), w)
					return
				}
				// Fetch order asynchronously to avoid blocking the UI
				respCh := make(chan struct {
					resp []byte
					err  error
				})
				go func() {
					client := apppkg.NewFaireClient()
					resp, err := client.GetOrderByID(orderID, token)
					respCh <- struct {
						resp []byte
						err  error
					}{resp, err}
				}()
				result := <-respCh
				// Show dialog with order details or error
				if result.err != nil {
					dialog.ShowError(fmt.Errorf("failed to get order: %v", result.err), w)
					return
				}
				dialog.ShowInformation("Order", string(result.resp), w)
			}, w)
	})
	// Button: Quit
	// - Exits the application immediately when clicked.
	quitBtn := widget.NewButton("Quit", func() { os.Exit(0) })

	// Button: Test Process Shipments
	testProcessBtn := widget.NewButton("Test Process Shipments", func() {
		// Example processed and failed shipments
		exampleProcessed := []apppkg.ShipmentPayload{
			{
				OrderID:        "ORDER123",
				MakerCostCents: 1500,
				Carrier:        "UPS",
				TrackingCode:   "1Z999AA10123456784",
				ShippingType:   "Standard",
			},
			{
				OrderID:        "ORDER456",
				MakerCostCents: 2000,
				Carrier:        "FedEx",
				TrackingCode:   "123456789012",
				ShippingType:   "Express",
			},
		}
		exampleFailed := []apppkg.ShipmentPayload{
			{
				OrderID:        "ORDER789",
				MakerCostCents: 1800,
				Carrier:        "USPS",
				TrackingCode:   "9400110200881234567890",
				ShippingType:   "Standard",
			},
		}

		// Use the same formatting function as your real process
		formatPayloads := func(payloads []apppkg.ShipmentPayload) string {
			if len(payloads) == 0 {
				return "  None"
			}
			msg := ""
			for _, p := range payloads {
				msg += fmt.Sprintf(
					"  OrderID: %s\n    MakerCostCents: %d\n    Carrier: %s\n    TrackingCode: %s\n    ShippingType: %s\n",
					p.OrderID, p.MakerCostCents, p.Carrier, p.TrackingCode, p.ShippingType,
				)
			}
			return msg
		}

		msg := "Shipments processed successfully!\n\nProcessed Shipments:\n"
		msg += formatPayloads(exampleProcessed)
		msg += "\n\nFailed Shipments:\n"
		msg += formatPayloads(exampleFailed)

		scroll := container.NewVScroll(widget.NewLabel(msg))
		scroll.SetMinSize(fyne.NewSize(380, 250))
		dialog.ShowCustom("Test Process Result", "OK", scroll, w)
	})

	// Set up the main window layout with all buttons and the application label.
	w.SetContent(container.NewVBox(
		widget.NewLabel("Faire GUI"),
		processBtn,
		widget.NewLabel(""), // Adds a small space
		ordersBtn,
		orderBtn,
		widget.NewLabel(""), // Adds a small space
		testProcessBtn,
		layout.NewSpacer(), // Pushes everything below
		quitBtn,
	))
	// Set initial window size and start the GUI event loop.
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}
