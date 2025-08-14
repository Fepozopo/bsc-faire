package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Fepozopo/bsc-faire/internal/app"
	"github.com/Fepozopo/bsc-faire/internal/version"
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Welcome to the Faire CLI (version %s). Type 'help' for commands, 'exit' to quit.\n", version.Version)

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			break
		}
		if input == "self-update" {
			doSelfUpdate()
			continue
		}
		args := strings.Fields(input)
		cmd := app.RootCmd()
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			fmt.Println("Error:", err)
		}
	}
}

func doSelfUpdate() {
	const repo = "Fepozopo/bsc-faire"
	fmt.Println("Checking for updates...")
	latest, found, err := selfupdate.DetectLatest(repo)
	if err != nil {
		fmt.Println("Error occurred while detecting version:", err)
		return
	}
	if !found {
		fmt.Println("No release found")
		return
	}
	currentVer, _ := semver.Parse(version.Version)
	if latest.Version.Equals(currentVer) {
		fmt.Println("You are running the latest version.")
		return
	}
	fmt.Printf("New version available: %s\n", latest.Version)
	fmt.Println("Updating...")
	exe, err := os.Executable()
	if err != nil {
		fmt.Println("Could not locate executable path:", err)
		return
	}
	err = selfupdate.UpdateTo(latest.AssetURL, exe)
	if err != nil {
		fmt.Println("Update failed:", err)
		return
	}
	fmt.Println("Successfully updated to version", latest.Version)
	fmt.Println("Please restart the CLI.")
}
