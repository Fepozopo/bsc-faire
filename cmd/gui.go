package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	apppkg "github.com/Fepozopo/bsc-faire/internal/app"
)

func RunGUI() {
	myApp := fyneapp.New()
	w := myApp.NewWindow("BSC Faire GUI")

	// Main menu buttons
	processBtn := widget.NewButton("Process Shipments CSV", func() {
		fileDialog := dialog.NewFileOpen(
			func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				if reader == nil {
					return
				}
				path := reader.URI().Path()
				reader.Close()
				resultCh := make(chan error)
				go func() {
					err := apppkg.ProcessShipments(path)
					resultCh <- err
				}()
				go func() {
					err := <-resultCh
					if err != nil {
						fyne.CurrentApp().SendNotification(&fyne.Notification{
							Title:   "Error",
							Content: fmt.Sprintf("Failed to process shipments: %v", err),
						})
						dialog.ShowError(fmt.Errorf("failed to process shipments: %v", err), w)
					} else {
						fyne.CurrentApp().SendNotification(&fyne.Notification{
							Title:   "Success",
							Content: "Shipments processed successfully!",
						})
						dialog.ShowInformation("Success", "Shipments processed successfully!", w)
					}
				}()
			}, w)
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
		fileDialog.Show()
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
					token = os.Getenv("SMD_API_TOKEN")
				case "bsc":
					token = os.Getenv("BSC_API_TOKEN")
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
				go func() {
					result := <-respCh
					if result.err != nil {
						dialog.ShowError(fmt.Errorf("failed to get orders: %v", result.err), w)
						return
					}
					dialog.ShowInformation("Orders", string(result.resp), w)
				}()
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
					token = os.Getenv("SMD_API_TOKEN")
				case "bsc":
					token = os.Getenv("BSC_API_TOKEN")
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
				go func() {
					result := <-respCh
					if result.err != nil {
						dialog.ShowError(fmt.Errorf("failed to get order: %v", result.err), w)
						return
					}
					dialog.ShowInformation("Order", string(result.resp), w)
				}()
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
