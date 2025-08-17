package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	apppkg "github.com/Fepozopo/bsc-faire/internal/app"
	"github.com/Fepozopo/bsc-faire/internal/version"
)

func RunGUI() {
	myApp := fyneapp.New()
	w := myApp.NewWindow(fmt.Sprintf("Faire GUI (version %s)", version.Version))

	// Check for updates on startup (do not show 'No Updates' dialog)
	checkForUpdates(w, false)

	useMock := false
	mockFailsEntry := widget.NewEntry()
	mockFailsEntry.SetPlaceHolder("Mock fail indices (e.g. 1,3,5)")
	mockFailsEntry.Disable() // Start disabled

	mockCheck := widget.NewCheck("Use Mock Server", func(checked bool) {
		useMock = checked
		if checked {
			mockFailsEntry.Enable()
		} else {
			mockFailsEntry.Disable()
		}
	})

	// Button: Process Shipments CSV
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

			fileLabel := widget.NewLabel(fmt.Sprintf("Selected file: %s", filePath))
			var confirmDialog dialog.Dialog

			submitBtn := widget.NewButton("Submit", nil)

			content := container.NewVBox(
				fileLabel,
				submitBtn,
			)
			confirmDialog = dialog.NewCustom("Confirm File", "Cancel", content, w)
			confirmDialog.Show()

			submitBtn.OnTapped = func() {
				confirmDialog.Hide()

				progress := widget.NewProgressBarInfinite()
				progressLabel := widget.NewLabel("Processing shipments...")
				progressDialog := dialog.NewCustom("Processing", "Cancel", container.NewVBox(progressLabel, progress), w)
				progressDialog.Show()

				go func() {
					var client apppkg.FaireClientInterface
					if useMock {
						failMap := map[int]bool{}
						mockFails := mockFailsEntry.Text
						if mockFails != "" {
							for _, s := range strings.Split(mockFails, ",") {
								if idx, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
									failMap[idx] = true
								}
							}
						}
						client = &apppkg.MockFaireClient{FailOnCall: failMap}
					} else {
						client = apppkg.NewFaireClient()
					}
					processed, failed, err := apppkg.ProcessShipments(filePath, client)

					fyne.Do(func() {
						var formatPayloads = func(payloads []apppkg.ShipmentPayload, showError bool) string {
							if len(payloads) == 0 {
								return "  None"
							}
							msg := ""
							for _, p := range payloads {
								msg += fmt.Sprintf(
									"  OrderID: %s\n    MakerCostCents: %.2f\n    Carrier: %s\n    TrackingCode: %s\n    ShippingType: %s\n",
									apppkg.OrderIDToDisplayID(p.OrderID), float32(p.MakerCostCents)/100, p.Carrier, p.TrackingCode, p.ShippingType,
								)
								if showError && p.ErrorMsg != "" {
									msg += fmt.Sprintf("    SaleSource: %s\n    Error: %s\n\n", p.SaleSource, p.ErrorMsg)
								} else {
									msg += fmt.Sprintf("    SaleSource: %s\n\n", p.SaleSource)
								}
							}
							return msg
						}

						total := len(processed) + len(failed)
						successful := len(processed)
						failedCount := len(failed)

						var msg string
						if err != nil {
							msg = fmt.Sprintf("Processed %d shipments: %d successful, %d failed\n\nFailed to process shipments: %v", total, successful, failedCount, err)
						} else {
							summary := fmt.Sprintf("Processed %d shipments: %d successful, %d failed\n\n", total, successful, failedCount)
							msg = summary
							msg += "Failed Shipments:\n"
							msg += formatPayloads(failed, true)
							msg += "\n\nProcessed Shipments:\n"
							msg += formatPayloads(processed, false)
						}

						fyne.CurrentApp().SendNotification(&fyne.Notification{
							Title: func() string {
								if err != nil {
									return "Error"
								} else {
									return "Success"
								}
							}(),
							Content: msg,
						})
						progressDialog.Hide()
						if err != nil {
							dialog.ShowError(fmt.Errorf("%s", msg), w)
						} else {
							entry := widget.NewMultiLineEntry()
							entry.SetText(msg)
							scroll := container.NewVScroll(entry)
							scroll.SetMinSize(fyne.NewSize(380, 250))
							dialog.ShowCustom("Success", "OK", scroll, w)
						}
					})
				}()
			}
		})
	})

	// Button: Get All Orders
	ordersBtn := widget.NewButton("Get All Orders", func() {
		entry := widget.NewEntry()
		entry.SetPlaceHolder("Enter sale source: 21, asc, bjp, bsc, gtg, oat, or sm")
		dialog.ShowForm("Get All Orders", "Get", "Cancel",
			[]*widget.FormItem{
				widget.NewFormItem("Sale Source", entry),
			}, func(ok bool) {
				if !ok {
					return
				}
				saleSource := entry.Text
				token, err := apppkg.GetToken(saleSource)
				if err != nil || token == "" {
					dialog.ShowError(fmt.Errorf("invalid or missing token for sale source '%s'", saleSource), w)
					return
				}
				progress := widget.NewProgressBarInfinite()
				progressLabel := widget.NewLabel("Fetching orders...")
				progressDialog := dialog.NewCustom("Fetching Orders", "Cancel", container.NewVBox(progressLabel, progress), w)
				progressDialog.Show()

				go func() {
					client := apppkg.NewFaireClient()
					var allOrders []apppkg.Order
					currPage := 1
					for {
						resp, err := client.GetAllOrders(token, 50, currPage, "DELIVERED,BACKORDERED,CANCELED,PRE_TRANSIT,IN_TRANSIT,RETURNED,PENDING_RETAILER_CONFIRMATION,DAMAGED_OR_MISSING")
						if err != nil {
							fyne.Do(func() {
								progressDialog.Hide()
								dialog.ShowError(fmt.Errorf("failed to get orders: %v", err), w)
							})
							return
						}
						var ordersResp apppkg.Orders
						if err := json.Unmarshal(resp, &ordersResp); err != nil {
							fyne.Do(func() {
								progressDialog.Hide()
								dialog.ShowError(fmt.Errorf("failed to parse orders: %v", err), w)
							})
							return
						}
						allOrders = append(allOrders, ordersResp.Orders...)
						if len(ordersResp.Orders) < 50 {
							break
						}
						currPage++
					}
					fyne.Do(func() {
						progressDialog.Hide()
						msg := apppkg.FormatOrders(allOrders)
						entry := widget.NewMultiLineEntry()
						entry.SetText(msg)
						scroll := container.NewVScroll(entry)
						scroll.SetMinSize(fyne.NewSize(500, 400))
						dialog.ShowCustom("Orders", "OK", scroll, w)
					})
				}()
			}, w)
	})

	// Button: Export NEW Orders to CSV
	exportBtn := widget.NewButton("Export NEW Orders to CSV", func() {
		entry := widget.NewEntry()
		entry.SetPlaceHolder("Enter sale source: 21, asc, bjp, bsc, gtg, oat, or sm")
		dialog.ShowForm("Export NEW Orders", "Export", "Cancel",
			[]*widget.FormItem{
				widget.NewFormItem("Sale Source", entry),
			}, func(ok bool) {
				if !ok {
					return
				}
				saleSource := entry.Text
				progress := widget.NewProgressBarInfinite()
				progressLabel := widget.NewLabel("Exporting new orders to CSV...")
				progressDialog := dialog.NewCustom("Exporting", "Cancel", container.NewVBox(progressLabel, progress), w)
				progressDialog.Show()
				go func() {
					client := apppkg.NewFaireClient()
					count, err := client.ExportNewOrdersToCSV(saleSource, "faire_new_orders.csv")
					fyne.Do(func() {
						progressDialog.Hide()
						if err != nil {
							dialog.ShowError(fmt.Errorf("export failed: %v", err), w)
						} else {
							dialog.ShowInformation("Export Complete", fmt.Sprintf("Exported %d new orders to faire_new_orders.csv", count), w)
						}
					})
				}()
			}, w)
	})

	// Button: Get Order By ID
	orderBtn := widget.NewButton("Get Order By ID", func() {
		saleSourceEntry := widget.NewEntry()
		saleSourceEntry.SetPlaceHolder("Sale Source (21, asc, bjp, bsc, gtg, oat, sm)")
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
				token, err := apppkg.GetToken(saleSource)
				if err != nil || token == "" {
					dialog.ShowError(fmt.Errorf("invalid or missing token for sale source '%s'", saleSource), w)
					return
				}
				progress := widget.NewProgressBarInfinite()
				progressLabel := widget.NewLabel("Fetching order...")
				progressDialog := dialog.NewCustom("Fetching Order", "Cancel", container.NewVBox(progressLabel, progress), w)
				progressDialog.Show()

				go func() {
					client := apppkg.NewFaireClient()
					resp, err := client.GetOrderByID(orderID, token)
					fyne.Do(func() {
						progressDialog.Hide()
						if err != nil {
							dialog.ShowError(fmt.Errorf("failed to get order: %v", err), w)
							return
						}
						var order apppkg.Order
						if err := json.Unmarshal(resp, &order); err != nil {
							dialog.ShowError(fmt.Errorf("failed to parse order: %v", err), w)
							return
						}
						msg := apppkg.FormatOrder(order)
						entry := widget.NewMultiLineEntry()
						entry.SetText(msg)
						scroll := container.NewVScroll(entry)
						scroll.SetMinSize(fyne.NewSize(500, 400))
						dialog.ShowCustom("Order", "OK", scroll, w)
					})
				}()
			}, w)
	})

	// Button: Launch CLI in new terminal window
	launchCliBtn := widget.NewButton("Launch CLI", func() {
		if err := launchCLIInTerminal(); err != nil {
			dialog.ShowError(err, w)
		}
	})

	// Button: Self-Update
	updateBtn := widget.NewButton("Check for Updates", func() {
		checkForUpdates(w, true)
	})

	quitBtn := widget.NewButton("Quit", func() { os.Exit(0) })

	w.SetContent(container.NewVBox(
		container.NewHBox(
			widget.NewLabel(fmt.Sprintf("Faire GUI (version %s)", version.Version)),
			layout.NewSpacer(),
			mockCheck,
			widget.NewLabel(""),
			container.NewGridWrap(fyne.NewSize(250, mockFailsEntry.MinSize().Height), mockFailsEntry),
		),
		processBtn,
		exportBtn,
		widget.NewLabel(""),
		ordersBtn,
		orderBtn,
		widget.NewLabel(""),
		layout.NewSpacer(),
		launchCliBtn,
		updateBtn,
		quitBtn,
	))
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}
