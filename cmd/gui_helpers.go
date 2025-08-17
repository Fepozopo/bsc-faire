package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

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

// Cross-platform function to launch CLI in a new terminal window
func launchCLIInTerminal() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	binDir := filepath.Dir(exePath)

	switch runtime.GOOS {
	case "darwin":
		// macOS: use a temporary shell script to reliably launch CLI in Terminal with --cli
		tmpScript := filepath.Join(os.TempDir(), "launch_faire_cli.sh")
		scriptContent := fmt.Sprintf("#!/bin/bash\ncd '%s'\n'%s' --cli\n", binDir, exePath)
		if err := os.WriteFile(tmpScript, []byte(scriptContent), 0700); err != nil {
			return err
		}
		cmd := exec.Command("open", "-a", "Terminal", tmpScript)
		fmt.Println("Launching CLI with script:", tmpScript)
		return cmd.Start()
	case "windows":
		// Windows: use a temporary batch script to cd and launch CLI with --cli
		tmpScript := filepath.Join(os.TempDir(), "launch_faire_cli.bat")
		scriptContent := fmt.Sprintf("cd /d \"%s\"\r\n\"%s\" --cli\r\npause\r\n", binDir, exePath)
		if err := os.WriteFile(tmpScript, []byte(scriptContent), 0700); err != nil {
			return err
		}
		cmd := exec.Command("cmd", "/C", "start", "", tmpScript)
		fmt.Println("Launching CLI with script:", tmpScript)
		return cmd.Start()
	case "linux":
		// Linux: use a temporary shell script to cd and launch CLI with --cli
		tmpScript := filepath.Join(os.TempDir(), "launch_faire_cli.sh")
		scriptContent := fmt.Sprintf("#!/bin/bash\ncd '%s'\n'%s' --cli\nexec bash\n", binDir, exePath)
		if err := os.WriteFile(tmpScript, []byte(scriptContent), 0700); err != nil {
			return err
		}
		terminals := [][]string{
			{"gnome-terminal", "--", tmpScript},
			{"x-terminal-emulator", "-e", tmpScript},
			{"xterm", "-e", tmpScript},
		}
		var lastErr error
		for _, term := range terminals {
			cmd := exec.Command(term[0], term[1:]...)
			if err := cmd.Start(); err == nil {
				return nil
			} else {
				lastErr = err
			}
		}
		return lastErr
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}
