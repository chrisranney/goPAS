// Package commands implements the command system for pasctl.
package commands

import (
	"context"
	"fmt"
	"sort"

	"github.com/chrisranney/gopas"

	"pasctl/internal/config"
	"pasctl/internal/output"
)

// ExecutionContext holds the context for command execution.
type ExecutionContext struct {
	Ctx       context.Context
	Session   *gopas.Session
	Config    *config.Config
	Formatter *output.Formatter
}

// Command represents a command that can be executed.
type Command interface {
	Name() string
	Description() string
	Usage() string
	Execute(execCtx *ExecutionContext, args []string) error
}

// Subcommand represents a command with subcommands.
type Subcommand interface {
	Subcommands() []string
}

// Registry holds all registered commands.
type Registry struct {
	commands map[string]Command
}

// NewRegistry creates a new command registry.
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register adds a command to the registry.
func (r *Registry) Register(cmd Command) {
	r.commands[cmd.Name()] = cmd
}

// Get retrieves a command by name.
func (r *Registry) Get(name string) (Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

// All returns all registered commands.
func (r *Registry) All() map[string]Command {
	return r.commands
}

// Names returns the names of all registered commands in sorted order.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RequireSession checks if a session is available.
func RequireSession(execCtx *ExecutionContext) error {
	if execCtx.Session == nil || !execCtx.Session.IsValid() {
		return fmt.Errorf("not connected - use 'connect' first")
	}
	return nil
}
