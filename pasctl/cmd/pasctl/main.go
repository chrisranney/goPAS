// Package main provides the entry point for pasctl.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"pasctl/internal/config"
	"pasctl/internal/repl"
)

var (
	version = "1.0.0"
)

func main() {
	// Command line flags
	var (
		showVersion = flag.Bool("version", false, "Show version")
		showHelp    = flag.Bool("help", false, "Show help")
		command     = flag.String("c", "", "Execute a single command and exit")
		scriptFile  = flag.String("script", "", "Execute commands from a script file")
	)

	flag.Parse()

	if *showVersion {
		fmt.Printf("pasctl version %s\n", version)
		os.Exit(0)
	}

	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load config: %v\n", err)
		cfg = config.Default()
	}

	// Validate config
	cfg.Validate()

	// Create REPL
	r, err := repl.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize pasctl: %v\n", err)
		os.Exit(1)
	}
	defer r.Close()

	// Handle single command mode
	if *command != "" {
		if err := r.RunCommand(*command); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Handle script mode
	if *scriptFile != "" {
		commands, err := readScript(*scriptFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading script: %v\n", err)
			os.Exit(1)
		}
		if err := r.RunScript(commands); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Handle piped input
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Input is being piped
		commands, err := readFromStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
		if err := r.RunScript(commands); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Interactive mode
	if err := r.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf(`pasctl - CyberArk PAS Interactive Shell

Usage:
  pasctl [options]
  pasctl -c "command"
  pasctl --script=file.txt
  echo "command" | pasctl

Options:
  -c "command"      Execute a single command and exit
  --script=FILE     Execute commands from a script file
  --version         Show version information
  --help            Show this help message

Interactive Mode:
  Simply run 'pasctl' without arguments to enter interactive mode.
  Type 'help' for available commands.

Examples:
  pasctl                                       # Start interactive shell
  pasctl -c "safes list"                       # Run single command
  pasctl --script=setup.txt                    # Run commands from file
  echo "accounts list --safe=Prod" | pasctl    # Pipe commands

Configuration:
  Config file: ~/.pasctl/config.json
  History file: ~/.pasctl_history

Environment Variables:
  PASCTL_SERVER   Default server URL
  PASCTL_USER     Default username
  PASCTL_AUTH     Default auth method (cyberark, ldap, radius, windows)

For more information, visit: https://github.com/chrisranney/gopas
`)
}

func readScript(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var commands []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			commands = append(commands, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return commands, nil
}

func readFromStdin() ([]string, error) {
	var commands []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			commands = append(commands, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return commands, nil
}
