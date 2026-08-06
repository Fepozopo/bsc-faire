package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	apppkg "github.com/Fepozopo/bsc-faire/internal/app"
	"github.com/Fepozopo/bsc-faire/internal/version"
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	osDialog "github.com/sqweek/dialog"
)

// openFileWindow creates a file open dialog using the system's native file manager.
func openFileWindow(parent fyne.Window, callback func(filePath string, e error)) {
	filePath, err := osDialog.File().Load()
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

// checkForUpdates checks GitHub for a newer release and presents all update UI on Fyne's main thread.
func checkForUpdates(w fyne.Window, showNoUpdatesDialog bool) {
	go func() {
		const repo = "Fepozopo/bsc-faire"
		latest, found, err := selfupdate.DetectLatest(repo)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("update check failed: %w", err), w)
			})
			return
		}

		currentVer, _ := semver.Parse(version.Version)
		if !found || latest.Version.Equals(currentVer) {
			if showNoUpdatesDialog {
				fyne.Do(func() {
					dialog.ShowInformation("No Updates", "You are already running the latest version.", w)
				})
			}
			return
		}
		updateMsg := fmt.Sprintf("A new version (%s) is available. Update now or continue using the current version.", latest.Version)
		fyne.Do(func() {
			dialog.NewCustomConfirm(
				"Update Available",
				"Update",
				"Continue",
				widget.NewLabel(updateMsg),
				func(ok bool) {
					if ok {
						exe, err := os.Executable()
						if err != nil {
							dialog.ShowError(fmt.Errorf("could not locate executable: %w", err), w)
							return
						}

						// Show infinite progress bar dialog
						progress := widget.NewProgressBarInfinite()
						progressLabel := widget.NewLabel("Updating application...")
						progressDialog := dialog.NewCustom("Updating", "Cancel", container.NewVBox(progressLabel, progress), w)
						progressDialog.Show()

						go func() {
							err = selfupdate.UpdateTo(latest.AssetURL, exe)
							fyne.Do(func() {
								progressDialog.Hide()
								if err != nil {
									dialog.ShowError(fmt.Errorf("update failed: %w", err), w)
									return
								}
								// Force restart
								cmd := exec.Command(exe, os.Args[1:]...)
								cmd.Env = os.Environ()
								err := cmd.Start()
								if err != nil {
									dialog.ShowError(fmt.Errorf("failed to restart: %w", err), w)
									return
								}
								os.Exit(0)
							})
						}()
					} else {
						// A declined update must not interrupt the user's current work.
						return
					}
				},
				w,
			).Show()
		})
	}()
}

// orderExportConfiguration describes one CSV export action presented in the GUI.
type orderExportConfiguration struct {
	ButtonLabel     string
	FormTitle       string
	ProgressMessage string
	Filename        string
	State           string
	UsesOrderIDs    bool
}

// newOrderExportButton creates an order-export button that uses the live or mock client selected by useMock.
func newOrderExportButton(parent fyne.Window, useMock func() bool, configuration orderExportConfiguration) *widget.Button {
	return widget.NewButton(configuration.ButtonLabel, func() {
		saleSourceEntry := widget.NewEntry()
		saleSourceEntry.SetPlaceHolder("Enter sale source: 21, asc, bjp, bsc, gtg, oat, or sm")

		formItems := []*widget.FormItem{
			widget.NewFormItem("Sale Source", saleSourceEntry),
		}
		orderIDsEntry := widget.NewMultiLineEntry()
		if configuration.UsesOrderIDs {
			orderIDsEntry.SetPlaceHolder("One display ID or bo_ ID per line; commas and semicolons also work")
			orderIDsEntry.SetMinRowsVisible(4)
			formItems = append(formItems, widget.NewFormItem("Order IDs", orderIDsEntry))
		}

		dialog.ShowForm(configuration.FormTitle, "Export", "Cancel", formItems, func(ok bool) {
			if !ok {
				return
			}

			saleSource := strings.TrimSpace(saleSourceEntry.Text)
			apiToken := "mock-token"
			if !useMock() {
				var err error
				apiToken, err = apppkg.GetToken(saleSource)
				if err != nil || apiToken == "" {
					dialog.ShowError(fmt.Errorf("invalid or missing token for sale source %q", saleSource), parent)
					return
				}
			}

			filter := apppkg.OrderExportFilter{State: configuration.State}
			if configuration.UsesOrderIDs {
				filter.State = ""
				filter.OrderIdentifiers = apppkg.ParseOrderIdentifiers(orderIDsEntry.Text)
			}

			outputPath, err := apppkg.DownloadsFilePath(configuration.Filename)
			if err != nil {
				dialog.ShowError(fmt.Errorf("prepare export destination: %w", err), parent)
				return
			}

			progress := widget.NewProgressBarInfinite()
			progressLabel := widget.NewLabel(configuration.ProgressMessage)
			progressDialog := dialog.NewCustom("Exporting", "Cancel", container.NewVBox(progressLabel, progress), parent)
			progressDialog.Show()

			go func() {
				var client apppkg.OrderClient
				if useMock() {
					client = &apppkg.MockFaireClient{Orders: apppkg.MockOrders}
				} else {
					client = apppkg.NewFaireClient()
				}

				count, err := apppkg.ExportOrdersToCSV(client, apiToken, saleSource, outputPath, filter)
				fyne.Do(func() {
					progressDialog.Hide()
					if err != nil {
						dialog.ShowError(fmt.Errorf("export failed: %w", err), parent)
						return
					}
					dialog.ShowInformation("Export Complete", fmt.Sprintf("Exported %d orders to %s", count, outputPath), parent)
				})
			}()
		}, parent)
	})
}
