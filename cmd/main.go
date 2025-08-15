package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"slices"

	"github.com/Fepozopo/bsc-faire/internal/app"
	"github.com/Fepozopo/bsc-faire/internal/version"
)

// Entry point: launch GUI by default, CLI only with --cli flag

func main() {
	if slices.Contains(os.Args[1:], "--cli") {
		runCLI()
		return
	}
	RunGUI()
}

func runCLI() {
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
		args := strings.Fields(input)
		cmd := app.RootCmd()
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			fmt.Println("Error:", err)
		}
	}
}
