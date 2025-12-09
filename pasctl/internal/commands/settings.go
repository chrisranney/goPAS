package commands

import (
	"fmt"
	"strconv"
	"strings"

	"pasctl/internal/output"
)

// SetCommand handles setting configuration options.
type SetCommand struct{}

func (c *SetCommand) Name() string {
	return "set"
}

func (c *SetCommand) Description() string {
	return "Set configuration options"
}

func (c *SetCommand) Usage() string {
	return `set <option> <value>

Options:
  output <format>       Set output format: table, json, yaml

Examples:
  set output json
  set output table
  set output yaml
`
}

func (c *SetCommand) Subcommands() []string {
	return []string{"output"}
}

func (c *SetCommand) Execute(execCtx *ExecutionContext, args []string) error {
	if len(args) == 0 {
		fmt.Println(c.Usage())
		return nil
	}

	switch args[0] {
	case "output":
		if len(args) < 2 {
			return fmt.Errorf("output format required: table, json, or yaml")
		}
		return c.setOutput(execCtx, args[1])
	default:
		return fmt.Errorf("unknown option: %s", args[0])
	}
}

func (c *SetCommand) setOutput(execCtx *ExecutionContext, format string) error {
	format = strings.ToLower(format)
	switch format {
	case "table":
		execCtx.Formatter.SetFormat(output.FormatTable)
	case "json":
		execCtx.Formatter.SetFormat(output.FormatJSON)
	case "yaml":
		execCtx.Formatter.SetFormat(output.FormatYAML)
	default:
		return fmt.Errorf("invalid output format: %s (use: table, json, yaml)", format)
	}

	output.PrintSuccess("Output format set to: %s", format)
	return nil
}

// ConfigCommand handles viewing and setting persistent configuration.
type ConfigCommand struct{}

func (c *ConfigCommand) Name() string {
	return "config"
}

func (c *ConfigCommand) Description() string {
	return "View or modify configuration"
}

func (c *ConfigCommand) Usage() string {
	return `config [option] [value]

View current configuration or set a specific option.

Options:
  default-server <url>   Set default server URL
  default-auth <method>  Set default auth method (cyberark, ldap, radius, windows)
  output <format>        Set default output format (table, json, yaml)
  history-size <n>       Set history size
  insecure-ssl <bool>    Enable/disable SSL verification
  timeout <seconds>      Set request timeout

Examples:
  config                           Show all settings
  config default-server https://cyberark.example.com
  config default-auth ldap
  config output json
  config insecure-ssl true
`
}

func (c *ConfigCommand) Execute(execCtx *ExecutionContext, args []string) error {
	if len(args) == 0 {
		return c.showConfig(execCtx)
	}

	if len(args) < 2 {
		return fmt.Errorf("value required for option: %s", args[0])
	}

	option := args[0]
	value := args[1]

	switch option {
	case "default-server":
		execCtx.Config.DefaultServer = value
	case "default-auth":
		switch strings.ToLower(value) {
		case "cyberark", "ldap", "radius", "windows":
			execCtx.Config.DefaultAuthType = strings.ToLower(value)
		default:
			return fmt.Errorf("invalid auth method: %s", value)
		}
	case "output":
		switch strings.ToLower(value) {
		case "table", "json", "yaml":
			execCtx.Config.OutputFormat = strings.ToLower(value)
		default:
			return fmt.Errorf("invalid output format: %s", value)
		}
	case "history-size":
		n, err := strconv.Atoi(value)
		if err != nil || n < 0 {
			return fmt.Errorf("invalid history size: %s", value)
		}
		execCtx.Config.HistorySize = n
	case "insecure-ssl":
		switch strings.ToLower(value) {
		case "true", "yes", "1":
			execCtx.Config.InsecureSSL = true
		case "false", "no", "0":
			execCtx.Config.InsecureSSL = false
		default:
			return fmt.Errorf("invalid boolean value: %s", value)
		}
	case "timeout":
		n, err := strconv.Atoi(value)
		if err != nil || n < 0 {
			return fmt.Errorf("invalid timeout: %s", value)
		}
		execCtx.Config.Timeout = n
	default:
		return fmt.Errorf("unknown option: %s", option)
	}

	// Save config
	if err := execCtx.Config.Save(); err != nil {
		output.PrintWarning("Config saved to memory but failed to write to disk: %v", err)
	} else {
		output.PrintSuccess("Config updated: %s = %s", option, value)
	}

	return nil
}

func (c *ConfigCommand) showConfig(execCtx *ExecutionContext) error {
	fmt.Println()
	fmt.Printf("  %s\n", output.Header("Configuration"))
	fmt.Println()
	fmt.Printf("  Default Server:   %s\n", valueOrDefault(execCtx.Config.DefaultServer, "(not set)"))
	fmt.Printf("  Default Auth:     %s\n", valueOrDefault(execCtx.Config.DefaultAuthType, "cyberark"))
	fmt.Printf("  Output Format:    %s\n", valueOrDefault(execCtx.Config.OutputFormat, "table"))
	fmt.Printf("  History Size:     %d\n", execCtx.Config.HistorySize)
	fmt.Printf("  Insecure SSL:     %s\n", boolToStr(execCtx.Config.InsecureSSL))
	fmt.Printf("  Timeout:          %ds\n", execCtx.Config.Timeout)
	fmt.Println()

	return nil
}

func valueOrDefault(value, defaultVal string) string {
	if value == "" {
		return output.Dim(defaultVal)
	}
	return value
}

// ClearCommand clears the screen.
type ClearCommand struct{}

func (c *ClearCommand) Name() string {
	return "clear"
}

func (c *ClearCommand) Description() string {
	return "Clear the screen"
}

func (c *ClearCommand) Usage() string {
	return "clear\n\nClear the terminal screen."
}

func (c *ClearCommand) Execute(execCtx *ExecutionContext, args []string) error {
	fmt.Print("\033[H\033[2J")
	return nil
}

// HistoryCommand shows command history.
type HistoryCommand struct {
	historyFunc func() []string
}

func NewHistoryCommand(historyFunc func() []string) *HistoryCommand {
	return &HistoryCommand{historyFunc: historyFunc}
}

func (c *HistoryCommand) Name() string {
	return "history"
}

func (c *HistoryCommand) Description() string {
	return "Show command history"
}

func (c *HistoryCommand) Usage() string {
	return `history [n]

Show command history. Optionally specify number of recent commands to show.

Examples:
  history        Show all history
  history 10     Show last 10 commands
`
}

func (c *HistoryCommand) Execute(execCtx *ExecutionContext, args []string) error {
	if c.historyFunc == nil {
		return fmt.Errorf("history not available")
	}

	history := c.historyFunc()
	if len(history) == 0 {
		output.PrintInfo("No command history")
		return nil
	}

	// Parse optional limit
	limit := len(history)
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err == nil && n > 0 && n < limit {
			limit = n
		}
	}

	start := len(history) - limit
	if start < 0 {
		start = 0
	}

	fmt.Println()
	for i := start; i < len(history); i++ {
		fmt.Printf("  %4d  %s\n", i+1, history[i])
	}
	fmt.Println()

	return nil
}
