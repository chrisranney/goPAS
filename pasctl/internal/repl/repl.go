package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"

	"github.com/chrisranney/gopas"

	"pasctl/internal/commands"
	"pasctl/internal/config"
	"pasctl/internal/output"
)

// REPL represents the Read-Eval-Print Loop.
type REPL struct {
	rl       *readline.Instance
	registry *commands.Registry
	session  *gopas.Session
	config   *config.Config
	format   *output.Formatter
	ctx      context.Context
	cancel   context.CancelFunc
}

// New creates a new REPL instance.
func New(cfg *config.Config) (*REPL, error) {
	historyPath, err := config.HistoryPath()
	if err != nil {
		historyPath = ""
	}

	completer := NewCompleter()

	rl, err := readline.NewEx(&readline.Config{
		Prompt:            "\033[36mpasctl>\033[0m ",
		HistoryFile:       historyPath,
		AutoComplete:      completer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize readline: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	r := &REPL{
		rl:       rl,
		registry: commands.NewRegistry(),
		session:  &gopas.Session{},
		config:   cfg,
		format:   output.NewFormatter(output.Format(cfg.OutputFormat)),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Register all commands
	r.registerCommands()

	return r, nil
}

func (r *REPL) registerCommands() {
	// Session commands
	r.registry.Register(&commands.ConnectCommand{})
	r.registry.Register(&commands.DisconnectCommand{})
	r.registry.Register(&commands.StatusCommand{})
	r.registry.Register(&commands.CCPCommand{})

	// Resource commands
	r.registry.Register(&commands.AccountsCommand{})
	r.registry.Register(&commands.SafesCommand{})
	r.registry.Register(&commands.UsersCommand{})
	r.registry.Register(&commands.PlatformsCommand{})

	// Monitoring commands
	r.registry.Register(&commands.PSMCommand{})
	r.registry.Register(&commands.HealthCommand{})

	// Settings commands
	r.registry.Register(&commands.SetCommand{})
	r.registry.Register(&commands.ConfigCommand{})
	r.registry.Register(&commands.ClearCommand{})
	r.registry.Register(commands.NewHistoryCommand(r.getHistory))

	// Help command (needs registry reference)
	r.registry.Register(commands.NewHelpCommand(r.registry))
}

// Run starts the REPL loop.
func (r *REPL) Run() error {
	defer r.rl.Close()

	r.printWelcome()

	for {
		line, err := r.rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				if len(line) == 0 {
					fmt.Println("Use 'exit' to quit")
					continue
				}
				continue
			}
			if err == io.EOF {
				break
			}
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Handle exit commands
		if line == "exit" || line == "quit" {
			r.cleanup()
			fmt.Println("Goodbye!")
			break
		}

		// Execute the command
		if err := r.execute(line); err != nil {
			fmt.Printf("\033[31mError: %v\033[0m\n", err)
		}
	}

	return nil
}

// RunCommand executes a single command (for non-interactive mode).
func (r *REPL) RunCommand(line string) error {
	return r.execute(line)
}

// RunScript executes commands from a script file or stdin.
func (r *REPL) RunScript(commands []string) error {
	for _, line := range commands {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fmt.Printf("pasctl> %s\n", line)
		if err := r.execute(line); err != nil {
			return err
		}
	}
	return nil
}

func (r *REPL) execute(line string) error {
	args := ParseArgs(line)
	if len(args) == 0 {
		return nil
	}

	cmdName := args[0]
	cmdArgs := args[1:]

	// Look up the command
	cmd, exists := r.registry.Get(cmdName)
	if !exists {
		return fmt.Errorf("unknown command: %s (type 'help' for available commands)", cmdName)
	}

	// Build execution context
	execCtx := &commands.ExecutionContext{
		Ctx:       r.ctx,
		Session:   r.session,
		Config:    r.config,
		Formatter: r.format,
	}

	// Execute the command
	return cmd.Execute(execCtx, cmdArgs)
}

func (r *REPL) printWelcome() {
	fmt.Println()
	fmt.Println(output.InfoBold("pasctl - CyberArk PAS Interactive Shell"))
	fmt.Println()
	fmt.Println("Type 'help' for available commands, 'exit' to quit")
	fmt.Println()
}

func (r *REPL) cleanup() {
	if r.session != nil && r.session.IsValid() {
		gopas.CloseSession(r.ctx, r.session)
	}
	r.cancel()
}

func (r *REPL) getHistory() []string {
	// Read history from file
	historyFile := r.rl.Config.HistoryFile
	if historyFile == "" {
		return nil
	}

	file, err := os.Open(historyFile)
	if err != nil {
		return nil
	}
	defer file.Close()

	var result []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

// UpdatePrompt updates the prompt to show connection status.
func (r *REPL) UpdatePrompt() {
	if r.session != nil && r.session.IsValid() {
		r.rl.SetPrompt(fmt.Sprintf("\033[36m%s@pasctl>\033[0m ", r.session.User))
	} else {
		r.rl.SetPrompt("\033[36mpasctl>\033[0m ")
	}
}

// SetSession updates the session reference.
func (r *REPL) SetSession(sess *gopas.Session) {
	r.session = sess
	r.UpdatePrompt()
}

// GetSession returns the current session.
func (r *REPL) GetSession() *gopas.Session {
	return r.session
}

// Close closes the REPL.
func (r *REPL) Close() {
	r.cleanup()
	r.rl.Close()
}
