package commands

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chrisranney/gopas/pkg/monitoring"

	"pasctl/internal/output"
)

// PSMCommand handles PSM monitoring operations.
type PSMCommand struct{}

func (c *PSMCommand) Name() string {
	return "psm"
}

func (c *PSMCommand) Description() string {
	return "Monitor PSM sessions and recordings"
}

func (c *PSMCommand) Usage() string {
	return `psm <subcommand> [options]

Subcommands:
  sessions              List recorded PSM sessions
  live                  List live (active) PSM sessions
  get <session-id>      Get session details
  terminate <session-id> Terminate a live session
  suspend <session-id>  Suspend a live session
  resume <session-id>   Resume a suspended session
  activities <session-id> View session activities
  properties <session-id> View session properties

Options for 'sessions':
  --from=TIME           Start time (e.g., 2024-01-01, -24h, -7d)
  --to=TIME             End time (e.g., 2024-01-02, now)
  --safe=NAME           Filter by safe name
  --search=TERM         Search term
  --limit=N             Maximum results (default: 25)

Options for 'live':
  --search=TERM         Search term
  --limit=N             Maximum results (default: 25)

Examples:
  psm sessions --from=-24h
  psm sessions --from=2024-01-01 --to=2024-01-31 --safe=Production
  psm live
  psm get abc123
  psm terminate abc123
  psm activities abc123
`
}

func (c *PSMCommand) Subcommands() []string {
	return []string{"sessions", "live", "get", "terminate", "suspend", "resume", "activities", "properties"}
}

func (c *PSMCommand) Execute(execCtx *ExecutionContext, args []string) error {
	if err := RequireSession(execCtx); err != nil {
		return err
	}

	if len(args) == 0 {
		fmt.Println(c.Usage())
		return nil
	}

	switch args[0] {
	case "sessions":
		return c.sessions(execCtx, args[1:])
	case "live":
		return c.live(execCtx, args[1:])
	case "get":
		return c.get(execCtx, args[1:])
	case "terminate":
		return c.terminate(execCtx, args[1:])
	case "suspend":
		return c.suspend(execCtx, args[1:])
	case "resume":
		return c.resume(execCtx, args[1:])
	case "activities":
		return c.activities(execCtx, args[1:])
	case "properties":
		return c.properties(execCtx, args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", args[0])
	}
}

func (c *PSMCommand) sessions(execCtx *ExecutionContext, args []string) error {
	fs := flag.NewFlagSet("psm sessions", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	from := fs.String("from", "", "Start time")
	to := fs.String("to", "", "End time")
	safe := fs.String("safe", "", "Filter by safe name")
	search := fs.String("search", "", "Search term")
	limit := fs.Int("limit", 25, "Maximum results")

	if err := fs.Parse(args); err != nil {
		return err
	}

	opts := monitoring.ListOptions{
		Limit: *limit,
	}
	if *from != "" {
		t, err := parseTime(*from)
		if err != nil {
			return fmt.Errorf("invalid --from time: %w", err)
		}
		opts.FromTime = t.Unix()
	}
	if *to != "" {
		t, err := parseTime(*to)
		if err != nil {
			return fmt.Errorf("invalid --to time: %w", err)
		}
		opts.ToTime = t.Unix()
	}
	if *safe != "" {
		opts.Safe = *safe
	}
	if *search != "" {
		opts.Search = *search
	}

	result, err := monitoring.ListSessions(execCtx.Ctx, execCtx.Session, opts)
	if err != nil {
		return err
	}

	if len(result.Recordings) == 0 {
		output.PrintInfo("No sessions found")
		return nil
	}

	if execCtx.Formatter.GetFormat() == output.FormatTable {
		table := output.NewTable("SESSION ID", "USER", "TARGET", "PROTOCOL", "START", "DURATION")
		for _, s := range result.Recordings {
			startTime := time.Unix(s.Start, 0).Format("2006-01-02 15:04")
			duration := formatSeconds(s.Duration)
			table.AddRow(
				truncate(s.SessionID.String(), 20),
				s.User,
				s.RemoteMachine,
				s.Protocol,
				startTime,
				duration,
			)
		}
		table.Render()
		fmt.Printf("\nShowing %d of %d sessions\n", len(result.Recordings), result.Total)
	} else {
		return execCtx.Formatter.Format(result)
	}

	return nil
}

func (c *PSMCommand) live(execCtx *ExecutionContext, args []string) error {
	fs := flag.NewFlagSet("psm live", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	search := fs.String("search", "", "Search term")
	limit := fs.Int("limit", 25, "Maximum results")

	if err := fs.Parse(args); err != nil {
		return err
	}

	opts := monitoring.ListOptions{
		Limit: *limit,
	}
	if *search != "" {
		opts.Search = *search
	}

	result, err := monitoring.ListLiveSessions(execCtx.Ctx, execCtx.Session, opts)
	if err != nil {
		return err
	}

	if len(result.Recordings) == 0 {
		output.PrintInfo("No active sessions")
		return nil
	}

	if execCtx.Formatter.GetFormat() == output.FormatTable {
		table := output.NewTable("SESSION ID", "USER", "TARGET", "PROTOCOL", "START", "CAN TERMINATE")
		for _, s := range result.Recordings {
			startTime := time.Unix(s.Start, 0).Format("2006-01-02 15:04")
			table.AddRow(
				truncate(s.SessionID.String(), 20),
				s.User,
				s.RemoteMachine,
				s.Protocol,
				startTime,
				boolToStr(s.CanTerminate),
			)
		}
		table.Render()
		fmt.Printf("\nActive sessions: %d\n", len(result.Recordings))
	} else {
		return execCtx.Formatter.Format(result)
	}

	return nil
}

func (c *PSMCommand) get(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("session ID required")
	}

	sessionID := args[0]
	session, err := monitoring.GetSession(execCtx.Ctx, execCtx.Session, sessionID)
	if err != nil {
		return err
	}

	return execCtx.Formatter.Format(session)
}

func (c *PSMCommand) terminate(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("session ID required")
	}

	sessionID := args[0]

	// Confirm termination
	fmt.Printf("Are you sure you want to terminate session %s? [y/N]: ", sessionID)
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		output.PrintInfo("Termination cancelled")
		return nil
	}

	if err := monitoring.TerminateSession(execCtx.Ctx, execCtx.Session, sessionID); err != nil {
		return err
	}

	output.PrintSuccess("Session %s terminated", sessionID)
	return nil
}

func (c *PSMCommand) suspend(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("session ID required")
	}

	sessionID := args[0]
	if err := monitoring.SuspendSession(execCtx.Ctx, execCtx.Session, sessionID); err != nil {
		return err
	}

	output.PrintSuccess("Session %s suspended", sessionID)
	return nil
}

func (c *PSMCommand) resume(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("session ID required")
	}

	sessionID := args[0]
	if err := monitoring.ResumeSession(execCtx.Ctx, execCtx.Session, sessionID); err != nil {
		return err
	}

	output.PrintSuccess("Session %s resumed", sessionID)
	return nil
}

func (c *PSMCommand) activities(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("session ID required")
	}

	sessionID := args[0]
	activities, err := monitoring.GetSessionActivities(execCtx.Ctx, execCtx.Session, sessionID)
	if err != nil {
		return err
	}

	if len(activities) == 0 {
		output.PrintInfo("No activities found for session %s", sessionID)
		return nil
	}

	return execCtx.Formatter.Format(activities)
}

func (c *PSMCommand) properties(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("session ID required")
	}

	sessionID := args[0]
	props, err := monitoring.GetSessionProperties(execCtx.Ctx, execCtx.Session, sessionID)
	if err != nil {
		return err
	}

	if len(props) == 0 {
		output.PrintInfo("No properties found for session %s", sessionID)
		return nil
	}

	return execCtx.Formatter.Format(props)
}

// Helper functions

func parseTime(s string) (time.Time, error) {
	now := time.Now()

	// Handle relative times
	if strings.HasPrefix(s, "-") {
		// Parse duration like -24h, -7d
		s = strings.TrimPrefix(s, "-")

		// Handle days
		if strings.HasSuffix(s, "d") {
			days := strings.TrimSuffix(s, "d")
			var d int
			_, err := fmt.Sscanf(days, "%d", &d)
			if err != nil {
				return time.Time{}, fmt.Errorf("invalid duration: %s", s)
			}
			return now.AddDate(0, 0, -d), nil
		}

		// Handle standard Go duration
		d, err := time.ParseDuration(s)
		if err != nil {
			return time.Time{}, err
		}
		return now.Add(-d), nil
	}

	// Handle "now"
	if strings.ToLower(s) == "now" {
		return now, nil
	}

	// Try various date formats
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02",
		"01/02/2006",
	}

	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

func formatSeconds(secs int64) string {
	if secs == 0 {
		return "-"
	}

	d := time.Duration(secs) * time.Second
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
