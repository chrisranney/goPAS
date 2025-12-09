package commands

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/chrisranney/gopas/pkg/users"

	"pasctl/internal/output"
)

// UsersCommand handles user operations.
type UsersCommand struct{}

func (c *UsersCommand) Name() string {
	return "users"
}

func (c *UsersCommand) Description() string {
	return "Manage CyberArk users"
}

func (c *UsersCommand) Usage() string {
	return `users <subcommand> [options]

Subcommands:
  list                  List users
  get <user-id>         Get user details
  create <username>     Create a new user
  delete <user-id>      Delete a user
  activate <user-id>    Activate a suspended user
  reset-password        Reset a user's password

Options for 'list':
  --search=TERM         Search term
  --type=TYPE           Filter by user type (EPVUser, CPM, etc.)
  --limit=N             Maximum results (default: 25)

Options for 'create':
  --password=PASS       Initial password (required)
  --type=TYPE           User type (default: EPVUser)
  --description=DESC    User description
  --email=EMAIL         User email

Options for 'reset-password':
  --user=ID             User ID (required)
  --password=PASS       New password (required)

Examples:
  users list --search=admin
  users list --type=EPVUser --limit=10
  users get 123
  users create newuser --password=SecurePass123!
  users activate 123
  users reset-password --user=123 --password=NewPass123!
  users delete 123
`
}

func (c *UsersCommand) Subcommands() []string {
	return []string{"list", "get", "create", "delete", "activate", "reset-password"}
}

func (c *UsersCommand) Execute(execCtx *ExecutionContext, args []string) error {
	if err := RequireSession(execCtx); err != nil {
		return err
	}

	if len(args) == 0 {
		fmt.Println(c.Usage())
		return nil
	}

	switch args[0] {
	case "list":
		return c.list(execCtx, args[1:])
	case "get":
		return c.get(execCtx, args[1:])
	case "create":
		return c.create(execCtx, args[1:])
	case "delete":
		return c.delete(execCtx, args[1:])
	case "activate":
		return c.activate(execCtx, args[1:])
	case "reset-password":
		return c.resetPassword(execCtx, args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", args[0])
	}
}

func (c *UsersCommand) list(execCtx *ExecutionContext, args []string) error {
	fs := flag.NewFlagSet("users list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	search := fs.String("search", "", "Search term")
	userType := fs.String("type", "", "Filter by user type")
	limit := fs.Int("limit", 25, "Maximum results")

	if err := fs.Parse(args); err != nil {
		return err
	}

	opts := users.ListOptions{
		Limit: *limit,
	}
	if *search != "" {
		opts.Search = *search
	}
	if *userType != "" {
		opts.UserType = *userType
	}

	result, err := users.List(execCtx.Ctx, execCtx.Session, opts)
	if err != nil {
		return err
	}

	if len(result.Users) == 0 {
		output.PrintInfo("No users found")
		return nil
	}

	if execCtx.Formatter.GetFormat() == output.FormatTable {
		table := output.NewTable("ID", "USERNAME", "TYPE", "SOURCE", "ENABLED", "SUSPENDED")
		for _, user := range result.Users {
			table.AddRow(
				fmt.Sprintf("%d", user.ID),
				user.Username,
				user.UserType,
				user.Source,
				boolToStr(user.EnableUser),
				boolToStr(user.Suspended),
			)
		}
		table.Render()
		fmt.Printf("\nShowing %d of %d users\n", len(result.Users), result.Total)
	} else {
		return execCtx.Formatter.Format(result)
	}

	return nil
}

func (c *UsersCommand) get(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("user ID required")
	}

	userID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid user ID: %s", args[0])
	}

	user, err := users.Get(execCtx.Ctx, execCtx.Session, userID)
	if err != nil {
		return err
	}

	return execCtx.Formatter.Format(user)
}

func (c *UsersCommand) create(execCtx *ExecutionContext, args []string) error {
	fs := flag.NewFlagSet("users create", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	password := fs.String("password", "", "Initial password (required)")
	userType := fs.String("type", "EPVUser", "User type")
	description := fs.String("description", "", "User description")
	email := fs.String("email", "", "User email")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("username required")
	}

	username := fs.Arg(0)

	if *password == "" {
		return fmt.Errorf("--password is required")
	}

	opts := users.CreateOptions{
		Username:        username,
		InitialPassword: *password,
		UserType:        *userType,
		Description:     *description,
	}

	if *email != "" {
		opts.Internet = &users.Internet{
			BusinessEmail: *email,
		}
	}

	user, err := users.Create(execCtx.Ctx, execCtx.Session, opts)
	if err != nil {
		return err
	}

	output.PrintSuccess("User '%s' created with ID: %d", user.Username, user.ID)
	return execCtx.Formatter.Format(user)
}

func (c *UsersCommand) delete(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("user ID required")
	}

	userID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid user ID: %s", args[0])
	}

	// Confirm deletion
	fmt.Printf("Are you sure you want to delete user %d? [y/N]: ", userID)
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		output.PrintInfo("Deletion cancelled")
		return nil
	}

	if err := users.Delete(execCtx.Ctx, execCtx.Session, userID); err != nil {
		return err
	}

	output.PrintSuccess("User %d deleted", userID)
	return nil
}

func (c *UsersCommand) activate(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("user ID required")
	}

	userID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid user ID: %s", args[0])
	}

	user, err := users.ActivateUser(execCtx.Ctx, execCtx.Session, userID)
	if err != nil {
		return err
	}

	output.PrintSuccess("User '%s' (ID: %d) activated", user.Username, user.ID)
	return nil
}

func (c *UsersCommand) resetPassword(execCtx *ExecutionContext, args []string) error {
	fs := flag.NewFlagSet("users reset-password", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	userIDStr := fs.String("user", "", "User ID (required)")
	password := fs.String("password", "", "New password (required)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *userIDStr == "" {
		return fmt.Errorf("--user is required")
	}
	if *password == "" {
		return fmt.Errorf("--password is required")
	}

	userID, err := strconv.Atoi(*userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID: %s", *userIDStr)
	}

	if err := users.ResetPassword(execCtx.Ctx, execCtx.Session, userID, *password); err != nil {
		return err
	}

	output.PrintSuccess("Password reset for user %d", userID)
	return nil
}
