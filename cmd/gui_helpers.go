package main

import (
	"fmt"
	"os"
	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
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

// checkForUpdates checks GitHub for a newer release and optionally confirms when the current version is latest.
func checkForUpdates(w fyne.Window, showNoUpdatesDialog bool) {
	go func() {
		const repo = "Fepozopo/bsc-faire"
		latest, found, err := selfupdate.DetectLatest(repo)
		if err != nil {
			dialog.ShowError(fmt.Errorf("update check failed: %w", err), w)
			return
		}

		currentVer, _ := semver.Parse(version.Version)
		if !found || latest.Version.Equals(currentVer) {
			if showNoUpdatesDialog {
				dialog.ShowInformation("No Updates", "You are already running the latest version.", w)
			}
			return
		}
		updateMsg := fmt.Sprintf("A new version (%s) is available. You must update to continue using the application.", latest.Version)
		dialog.NewCustomConfirm(
			"Update Required",
			"Update",
			"Quit",
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
					os.Exit(0)
				}
			},
			w,
		).Show()
	}()
}
