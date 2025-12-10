package commands

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/chrisranney/gopas/pkg/accounts"

	"pasctl/internal/output"
)

// AccountsCommand handles account operations.
type AccountsCommand struct{}

func (c *AccountsCommand) Name() string {
	return "accounts"
}

func (c *AccountsCommand) Description() string {
	return "Manage CyberArk accounts"
}

func (c *AccountsCommand) Usage() string {
	return `accounts <subcommand> [options]

Subcommands:
  list                List accounts
  get <id>            Get account details
  create              Create a new account
  delete <id>         Delete an account
  password <id>       Retrieve account password
  change <id>         Trigger immediate password change (CPM)
  verify <id>         Trigger credential verification (CPM)
  reconcile <id>      Trigger credential reconciliation (CPM)
  activities <id>     View account activity log

Options for 'list':
  --safe=NAME         Filter by safe name
  --search=TERM       Search term
  --limit=N           Maximum results (default: 25)
  --offset=N          Skip first N results

Options for 'create':
  --safe=NAME         Safe name (required)
  --platform=ID       Platform ID (required)
  --address=ADDR      Target address (required)
  --username=USER     Account username (required)
  --secret=PASS       Account password
  --name=NAME         Account name

Options for 'password':
  --reason=TEXT       Reason for access (may be required by policy)

Examples:
  accounts list --safe=Production --limit=10
  accounts get 12_34
  accounts create --safe=Production --platform=WinServerLocal --address=server1 --username=admin
  accounts password 12_34 --reason="Authorized maintenance"
  accounts change 12_34
  accounts verify 12_34
  accounts reconcile 12_34
  accounts delete 12_34
`
}

func (c *AccountsCommand) Subcommands() []string {
	return []string{"list", "get", "create", "delete", "password", "change", "verify", "reconcile", "activities"}
}

func (c *AccountsCommand) Execute(execCtx *ExecutionContext, args []string) error {
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
	case "password":
		return c.password(execCtx, args[1:])
	case "change":
		return c.change(execCtx, args[1:])
	case "verify":
		return c.verify(execCtx, args[1:])
	case "reconcile":
		return c.reconcile(execCtx, args[1:])
	case "activities":
		return c.activities(execCtx, args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", args[0])
	}
}

func (c *AccountsCommand) list(execCtx *ExecutionContext, args []string) error {
	fs := flag.NewFlagSet("accounts list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	safe := fs.String("safe", "", "Filter by safe name")
	search := fs.String("search", "", "Search term")
	limit := fs.Int("limit", 25, "Maximum results")
	offset := fs.Int("offset", 0, "Skip first N results")

	if err := fs.Parse(args); err != nil {
		return err
	}

	opts := accounts.ListOptions{
		Limit:  *limit,
		Offset: *offset,
	}
	if *safe != "" {
		opts.SafeName = *safe
	}
	if *search != "" {
		opts.Search = *search
	}

	result, err := accounts.List(execCtx.Ctx, execCtx.Session, opts)
	if err != nil {
		return err
	}

	if len(result.Value) == 0 {
		output.PrintInfo("No accounts found")
		return nil
	}

	// Format output based on current format setting
	if execCtx.Formatter.GetFormat() == output.FormatTable {
		// Create custom table for better display
		table := output.NewTable("ID", "USERNAME", "ADDRESS", "PLATFORM", "SAFE")
		for _, acc := range result.Value {
			table.AddRow(acc.ID.String(), acc.UserName, acc.Address, acc.PlatformID.String(), acc.SafeName)
		}
		table.Render()
		fmt.Printf("\nShowing %d of %d accounts\n", len(result.Value), result.Count)
	} else {
		return execCtx.Formatter.Format(result)
	}

	return nil
}

func (c *AccountsCommand) get(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("account ID required")
	}

	accountID := args[0]
	account, err := accounts.Get(execCtx.Ctx, execCtx.Session, accountID)
	if err != nil {
		return err
	}

	return execCtx.Formatter.Format(account)
}

func (c *AccountsCommand) create(execCtx *ExecutionContext, args []string) error {
	fs := flag.NewFlagSet("accounts create", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	safe := fs.String("safe", "", "Safe name (required)")
	platform := fs.String("platform", "", "Platform ID (required)")
	address := fs.String("address", "", "Target address (required)")
	username := fs.String("username", "", "Account username (required)")
	secret := fs.String("secret", "", "Account password")
	name := fs.String("name", "", "Account name")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate required fields
	var missing []string
	if *safe == "" {
		missing = append(missing, "--safe")
	}
	if *platform == "" {
		missing = append(missing, "--platform")
	}
	if *address == "" {
		missing = append(missing, "--address")
	}
	if *username == "" {
		missing = append(missing, "--username")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required options: %s", strings.Join(missing, ", "))
	}

	opts := accounts.CreateOptions{
		SafeName:   *safe,
		PlatformID: *platform,
		Address:    *address,
		UserName:   *username,
	}
	if *secret != "" {
		opts.Secret = *secret
	}
	if *name != "" {
		opts.Name = *name
	}

	account, err := accounts.Create(execCtx.Ctx, execCtx.Session, opts)
	if err != nil {
		return err
	}

	output.PrintSuccess("Account created with ID: %s", account.ID)
	return execCtx.Formatter.Format(account)
}

func (c *AccountsCommand) delete(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("account ID required")
	}

	accountID := args[0]

	// Confirm deletion
	fmt.Printf("Are you sure you want to delete account %s? [y/N]: ", accountID)
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		output.PrintInfo("Deletion cancelled")
		return nil
	}

	if err := accounts.Delete(execCtx.Ctx, execCtx.Session, accountID); err != nil {
		return err
	}

	output.PrintSuccess("Account %s deleted", accountID)
	return nil
}

func (c *AccountsCommand) password(execCtx *ExecutionContext, args []string) error {
	fs := flag.NewFlagSet("accounts password", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	reason := fs.String("reason", "", "Reason for access")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("account ID required")
	}

	accountID := fs.Arg(0)
	password, err := accounts.GetPassword(execCtx.Ctx, execCtx.Session, accountID, *reason)
	if err != nil {
		return err
	}

	fmt.Printf("Password: %s\n", password)
	return nil
}

func (c *AccountsCommand) change(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("account ID required")
	}

	accountID := args[0]
	err := accounts.ChangeCredentialsImmediately(execCtx.Ctx, execCtx.Session, accountID, accounts.ChangeCredentialsOptions{})
	if err != nil {
		return err
	}

	output.PrintSuccess("Password change initiated for account %s", accountID)
	return nil
}

func (c *AccountsCommand) verify(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("account ID required")
	}

	accountID := args[0]
	err := accounts.VerifyCredentials(execCtx.Ctx, execCtx.Session, accountID)
	if err != nil {
		return err
	}

	output.PrintSuccess("Credential verification initiated for account %s", accountID)
	return nil
}

func (c *AccountsCommand) reconcile(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("account ID required")
	}

	accountID := args[0]
	err := accounts.ReconcileCredentials(execCtx.Ctx, execCtx.Session, accountID)
	if err != nil {
		return err
	}

	output.PrintSuccess("Credential reconciliation initiated for account %s", accountID)
	return nil
}

func (c *AccountsCommand) activities(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("account ID required")
	}

	accountID := args[0]
	activities, err := accounts.GetActivities(execCtx.Ctx, execCtx.Session, accountID)
	if err != nil {
		return err
	}

	if len(activities) == 0 {
		output.PrintInfo("No activities found for account %s", accountID)
		return nil
	}

	return execCtx.Formatter.Format(activities)
}
