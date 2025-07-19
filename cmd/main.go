package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Fepozopo/bsc-faire/internal/app"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to the Faire CLI. Type 'help' for commands, 'exit' to quit.")

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
