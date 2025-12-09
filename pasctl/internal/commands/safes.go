package commands

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/chrisranney/gopas/pkg/safemembers"
	"github.com/chrisranney/gopas/pkg/safes"

	"pasctl/internal/output"
)

// SafesCommand handles safe operations.
type SafesCommand struct{}

func (c *SafesCommand) Name() string {
	return "safes"
}

func (c *SafesCommand) Description() string {
	return "Manage CyberArk safes"
}

func (c *SafesCommand) Usage() string {
	return `safes <subcommand> [options]

Subcommands:
  list                  List safes
  get <name>            Get safe details
  create <name>         Create a new safe
  update <name>         Update a safe
  delete <name>         Delete a safe
  members <name>        List safe members
  add-member            Add a member to a safe
  remove-member         Remove a member from a safe

Options for 'list':
  --search=TERM         Search term
  --limit=N             Maximum results (default: 25)
  --include-accounts    Include account counts

Options for 'create':
  --description=DESC    Safe description
  --cpm=NAME            Managing CPM server
  --retention=DAYS      Number of days retention

Options for 'update':
  --description=DESC    Safe description
  --cpm=NAME            Managing CPM server

Options for 'add-member':
  --safe=NAME           Safe name (required)
  --member=NAME         Member name (required)
  --role=ROLE           Permission role: user, admin, auditor (default: user)

Options for 'remove-member':
  --safe=NAME           Safe name (required)
  --member=NAME         Member name (required)

Examples:
  safes list --search=Prod
  safes get MySafe
  safes create MySafe --description="Test safe" --cpm=PasswordManager
  safes update MySafe --description="Updated description"
  safes members MySafe
  safes add-member --safe=MySafe --member=user1 --role=admin
  safes remove-member --safe=MySafe --member=user1
  safes delete MySafe
`
}

func (c *SafesCommand) Subcommands() []string {
	return []string{"list", "get", "create", "update", "delete", "members", "add-member", "remove-member"}
}

func (c *SafesCommand) Execute(execCtx *ExecutionContext, args []string) error {
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
	case "update":
		return c.update(execCtx, args[1:])
	case "delete":
		return c.delete(execCtx, args[1:])
	case "members":
		return c.members(execCtx, args[1:])
	case "add-member":
		return c.addMember(execCtx, args[1:])
	case "remove-member":
		return c.removeMember(execCtx, args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", args[0])
	}
}

func (c *SafesCommand) list(execCtx *ExecutionContext, args []string) error {
	fs := flag.NewFlagSet("safes list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	search := fs.String("search", "", "Search term")
	limit := fs.Int("limit", 25, "Maximum results")
	includeAccounts := fs.Bool("include-accounts", false, "Include account counts")

	if err := fs.Parse(args); err != nil {
		return err
	}

	opts := safes.ListOptions{
		Limit:           *limit,
		IncludeAccounts: *includeAccounts,
	}
	if *search != "" {
		opts.Search = *search
	}

	result, err := safes.List(execCtx.Ctx, execCtx.Session, opts)
	if err != nil {
		return err
	}

	if len(result.Value) == 0 {
		output.PrintInfo("No safes found")
		return nil
	}

	if execCtx.Formatter.GetFormat() == output.FormatTable {
		headers := []string{"NAME", "DESCRIPTION", "CPM", "OLAC"}
		if *includeAccounts {
			headers = append(headers, "ACCOUNTS")
		}
		table := output.NewTable(headers...)

		for _, safe := range result.Value {
			row := []string{
				safe.SafeName,
				truncate(safe.Description, 40),
				safe.ManagingCPM,
				boolToStr(safe.OLACEnabled),
			}
			if *includeAccounts && safe.Accounts != nil {
				row = append(row, fmt.Sprintf("%d", *safe.Accounts))
			}
			table.AddRow(row...)
		}
		table.Render()
		fmt.Printf("\nShowing %d of %d safes\n", len(result.Value), result.Count)
	} else {
		return execCtx.Formatter.Format(result)
	}

	return nil
}

func (c *SafesCommand) get(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("safe name required")
	}

	safeName := args[0]
	safe, err := safes.Get(execCtx.Ctx, execCtx.Session, safeName)
	if err != nil {
		return err
	}

	return execCtx.Formatter.Format(safe)
}

func (c *SafesCommand) create(execCtx *ExecutionContext, args []string) error {
	fs := flag.NewFlagSet("safes create", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	description := fs.String("description", "", "Safe description")
	cpm := fs.String("cpm", "", "Managing CPM server")
	retention := fs.Int("retention", 7, "Number of days retention")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("safe name required")
	}

	safeName := fs.Arg(0)
	if len(safeName) > 28 {
		return fmt.Errorf("safe name cannot exceed 28 characters")
	}

	opts := safes.CreateOptions{
		SafeName:              safeName,
		Description:           *description,
		ManagingCPM:           *cpm,
		NumberOfDaysRetention: *retention,
	}

	safe, err := safes.Create(execCtx.Ctx, execCtx.Session, opts)
	if err != nil {
		return err
	}

	output.PrintSuccess("Safe '%s' created", safe.SafeName)
	return execCtx.Formatter.Format(safe)
}

func (c *SafesCommand) update(execCtx *ExecutionContext, args []string) error {
	fs := flag.NewFlagSet("safes update", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	description := fs.String("description", "", "Safe description")
	cpm := fs.String("cpm", "", "Managing CPM server")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("safe name required")
	}

	safeName := fs.Arg(0)
	opts := safes.UpdateOptions{}

	if *description != "" {
		opts.Description = *description
	}
	if *cpm != "" {
		opts.ManagingCPM = *cpm
	}

	safe, err := safes.Update(execCtx.Ctx, execCtx.Session, safeName, opts)
	if err != nil {
		return err
	}

	output.PrintSuccess("Safe '%s' updated", safe.SafeName)
	return execCtx.Formatter.Format(safe)
}

func (c *SafesCommand) delete(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("safe name required")
	}

	safeName := args[0]

	// Confirm deletion
	fmt.Printf("Are you sure you want to delete safe '%s'? [y/N]: ", safeName)
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		output.PrintInfo("Deletion cancelled")
		return nil
	}

	if err := safes.Delete(execCtx.Ctx, execCtx.Session, safeName); err != nil {
		return err
	}

	output.PrintSuccess("Safe '%s' deleted", safeName)
	return nil
}

func (c *SafesCommand) members(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("safe name required")
	}

	safeName := args[0]
	result, err := safemembers.List(execCtx.Ctx, execCtx.Session, safeName, safemembers.ListOptions{})
	if err != nil {
		return err
	}

	if len(result.Value) == 0 {
		output.PrintInfo("No members found for safe '%s'", safeName)
		return nil
	}

	if execCtx.Formatter.GetFormat() == output.FormatTable {
		table := output.NewTable("MEMBER", "TYPE", "PREDEFINED", "READ-ONLY")
		for _, member := range result.Value {
			table.AddRow(
				member.MemberName,
				member.MemberType,
				boolToStr(member.IsPredefinedUser),
				boolToStr(member.IsReadOnly),
			)
		}
		table.Render()
	} else {
		return execCtx.Formatter.Format(result)
	}

	return nil
}

func (c *SafesCommand) addMember(execCtx *ExecutionContext, args []string) error {
	fs := flag.NewFlagSet("safes add-member", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	safe := fs.String("safe", "", "Safe name (required)")
	member := fs.String("member", "", "Member name (required)")
	role := fs.String("role", "user", "Permission role: user, admin, auditor")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *safe == "" {
		return fmt.Errorf("--safe is required")
	}
	if *member == "" {
		return fmt.Errorf("--member is required")
	}

	// Get permissions based on role
	var perms *safemembers.Permissions
	switch strings.ToLower(*role) {
	case "admin":
		perms = safemembers.DefaultAdminPermissions()
	case "auditor":
		perms = &safemembers.Permissions{
			ListAccounts:    true,
			ViewAuditLog:    true,
			ViewSafeMembers: true,
		}
	default:
		perms = safemembers.DefaultUserPermissions()
	}

	opts := safemembers.AddOptions{
		MemberName:  *member,
		Permissions: perms,
	}

	result, err := safemembers.Add(execCtx.Ctx, execCtx.Session, *safe, opts)
	if err != nil {
		return err
	}

	output.PrintSuccess("Member '%s' added to safe '%s' with %s permissions", *member, *safe, *role)
	return execCtx.Formatter.Format(result)
}

func (c *SafesCommand) removeMember(execCtx *ExecutionContext, args []string) error {
	fs := flag.NewFlagSet("safes remove-member", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	safe := fs.String("safe", "", "Safe name (required)")
	member := fs.String("member", "", "Member name (required)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *safe == "" {
		return fmt.Errorf("--safe is required")
	}
	if *member == "" {
		return fmt.Errorf("--member is required")
	}

	if err := safemembers.Remove(execCtx.Ctx, execCtx.Session, *safe, *member); err != nil {
		return err
	}

	output.PrintSuccess("Member '%s' removed from safe '%s'", *member, *safe)
	return nil
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func boolToStr(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
