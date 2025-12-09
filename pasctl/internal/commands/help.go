package commands

import (
	"fmt"
	"strings"

	"pasctl/internal/output"
)

// HelpCommand provides help information.
type HelpCommand struct {
	registry *Registry
}

// NewHelpCommand creates a new help command.
func NewHelpCommand(registry *Registry) *HelpCommand {
	return &HelpCommand{registry: registry}
}

func (c *HelpCommand) Name() string {
	return "help"
}

func (c *HelpCommand) Description() string {
	return "Display help information"
}

func (c *HelpCommand) Usage() string {
	return `help [command]

Display help information for pasctl commands.

Examples:
  help              Show all available commands
  help accounts     Show help for the accounts command
  help safes        Show help for the safes command
`
}

func (c *HelpCommand) Execute(execCtx *ExecutionContext, args []string) error {
	if len(args) == 0 {
		return c.showAllHelp()
	}

	return c.showCommandHelp(args[0])
}

func (c *HelpCommand) showAllHelp() error {
	fmt.Println()
	fmt.Println(output.Header("pasctl - CyberArk PAS Interactive Shell"))
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println()

	// Group commands by category
	categories := map[string][]string{
		"Session": {"connect", "disconnect", "status"},
		"Resources": {
			"accounts", "safes", "users", "platforms",
		},
		"Monitoring": {"psm", "health"},
		"Settings":   {"set", "config"},
		"Other":      {"help", "history", "clear", "exit"},
	}

	for _, cat := range []string{"Session", "Resources", "Monitoring", "Settings", "Other"} {
		cmds := categories[cat]
		fmt.Printf("  %s:\n", output.InfoBold(cat))
		for _, name := range cmds {
			if cmd, ok := c.registry.Get(name); ok {
				fmt.Printf("    %-12s %s\n", output.Bold(name), cmd.Description())
			} else {
				fmt.Printf("    %-12s\n", output.Dim(name))
			}
		}
		fmt.Println()
	}

	fmt.Println("Use 'help <command>' for more information about a specific command.")
	fmt.Println()
	return nil
}

func (c *HelpCommand) showCommandHelp(name string) error {
	cmd, ok := c.registry.Get(name)
	if !ok {
		return fmt.Errorf("unknown command: %s", name)
	}

	fmt.Println()
	fmt.Printf("%s - %s\n", output.Header(cmd.Name()), cmd.Description())
	fmt.Println()
	fmt.Println(output.InfoBold("Usage:"))
	fmt.Println()

	// Indent the usage text
	lines := strings.Split(cmd.Usage(), "\n")
	for _, line := range lines {
		fmt.Printf("  %s\n", line)
	}
	fmt.Println()

	return nil
}
