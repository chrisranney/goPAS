// Package repl provides the REPL (Read-Eval-Print Loop) for pasctl.
package repl

import (
	"strings"
	"unicode"
)

// ParseArgs parses a command line string into arguments, respecting quotes.
func ParseArgs(line string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for _, r := range line {
		switch {
		case r == '"' || r == '\'':
			if inQuotes {
				if r == quoteChar {
					// End of quoted string
					inQuotes = false
					quoteChar = rune(0)
				} else {
					// Different quote character inside quotes
					current.WriteRune(r)
				}
			} else {
				// Start of quoted string
				inQuotes = true
				quoteChar = r
			}
		case unicode.IsSpace(r):
			if inQuotes {
				current.WriteRune(r)
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	// Don't forget the last argument
	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

// ParseFlags extracts flags from args and returns remaining positional args.
func ParseFlags(args []string) (flags map[string]string, positional []string) {
	flags = make(map[string]string)

	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			// Long flag
			arg = strings.TrimPrefix(arg, "--")
			if idx := strings.Index(arg, "="); idx != -1 {
				flags[arg[:idx]] = arg[idx+1:]
			} else {
				flags[arg] = "true"
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			// Short flag
			arg = strings.TrimPrefix(arg, "-")
			if idx := strings.Index(arg, "="); idx != -1 {
				flags[arg[:idx]] = arg[idx+1:]
			} else {
				flags[arg] = "true"
			}
		} else {
			positional = append(positional, arg)
		}
	}

	return flags, positional
}

// JoinArgs joins arguments back into a command string.
func JoinArgs(args []string) string {
	var parts []string
	for _, arg := range args {
		if strings.ContainsAny(arg, " \t\"'") {
			// Quote the argument if it contains spaces or quotes
			arg = "\"" + strings.ReplaceAll(arg, "\"", "\\\"") + "\""
		}
		parts = append(parts, arg)
	}
	return strings.Join(parts, " ")
}

// SplitCommand splits a command line into command and subcommand.
func SplitCommand(line string) (cmd string, subcmd string, rest []string) {
	args := ParseArgs(line)
	if len(args) == 0 {
		return "", "", nil
	}

	cmd = args[0]
	if len(args) > 1 {
		subcmd = args[1]
	}
	if len(args) > 2 {
		rest = args[2:]
	}

	return cmd, subcmd, rest
}
