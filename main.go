package main

import (
	"fmt"
	"os"

	"github.com/adavis/ccs/cmd"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "switch":
		if err := cmd.Switch(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "current":
		if err := cmd.Current(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version", "-v", "--version":
		fmt.Printf("ccs version %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`ccs - Claude Code Switcher

Usage:
  ccs switch     Switch Claude Code profile (interactive)
  ccs current    Show active profile
  ccs version    Show version
  ccs help       Show this help

Shell Integration:
  Add this function to your .zshrc or .bashrc:
    ccs() {
      eval "$(command ccs switch)"
    }

  Then use: ccs`)
}
