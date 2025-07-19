package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	apppkg "github.com/Fepozopo/bsc-faire/internal/app"
	"github.com/joho/godotenv"
	osDialog "github.com/sqweek/dialog"
)

// openFileWindow creates a file open dialog using the system's native file manager
// and calls the given callback function with the selected file path.
// If the user cancels the dialog, the error argument will be set to an error with message "cancelled".
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

	// Main menu buttons
	processBtn := widget.NewButton("Process Shipments CSV", func() {
		openFileWindow(w, func(filePath string, e error) {
			if e != nil {
				dialog.ShowError(e, w)
				return
			}
			if filePath == "" {
				return
			}
			if len(filePath) < 4 || filePath[len(filePath)-4:] != ".csv" {
				dialog.ShowError(fmt.Errorf("please select a .csv file"), w)
				return
			}
			resultCh := make(chan error)
			go func() {
				err := apppkg.ProcessShipments(filePath)
				resultCh <- err
			}()
			err := <-resultCh
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title: func() string {
					if err != nil {
						return "Error"
					} else {
						return "Success"
					}
				}(),
				Content: func() string {
					if err != nil {
						return fmt.Sprintf("Failed to process shipments: %v", err)
					}
					return "Shipments processed successfully!"
				}(),
			})
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to process shipments: %v", err), w)
			} else {
				dialog.ShowInformation("Success", "Shipments processed successfully!", w)
			}
		})
	})
	ordersBtn := widget.NewButton("Get All Orders", func() {
		// Prompt for sale source (sm or bsc)
		entry := widget.NewEntry()
		entry.SetPlaceHolder("Enter sale source: sm or bsc")
		dialog.ShowForm("Get All Orders", "Get", "Cancel",
			[]*widget.FormItem{
				widget.NewFormItem("Sale Source", entry),
			}, func(ok bool) {
				if !ok {
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
					dialog.ShowError(fmt.Errorf("invalid sale source: must be 'sm' or 'bsc'"), w)
					return
				}
				respCh := make(chan struct {
					resp []byte
					err  error
				})
				go func() {
					client := apppkg.NewFaireClient()
					resp, err := client.GetAllOrders(token)
					respCh <- struct {
						resp []byte
						err  error
					}{resp, err}
				}()
				result := <-respCh
				if result.err != nil {
					dialog.ShowError(fmt.Errorf("failed to get orders: %v", result.err), w)
					return
				}
				dialog.ShowInformation("Orders", string(result.resp), w)
			}, w)
	})

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
					dialog.ShowError(fmt.Errorf("invalid sale source: must be 'sm' or 'bsc'"), w)
					return
				}
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
				if result.err != nil {
					dialog.ShowError(fmt.Errorf("failed to get order: %v", result.err), w)
					return
				}
				dialog.ShowInformation("Order", string(result.resp), w)
			}, w)
	})
	quitBtn := widget.NewButton("Quit", func() { os.Exit(0) })

	w.SetContent(container.NewVBox(
		widget.NewLabel("BSC Faire GUI"),
		processBtn,
		ordersBtn,
		orderBtn,
		quitBtn,
	))

	w.Resize(fyne.NewSize(400, 300))
	w.ShowAndRun()
}
