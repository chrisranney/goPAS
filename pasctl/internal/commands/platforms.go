package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/chrisranney/gopas/pkg/platforms"

	"pasctl/internal/output"
)

// PlatformsCommand handles platform operations.
type PlatformsCommand struct{}

func (c *PlatformsCommand) Name() string {
	return "platforms"
}

func (c *PlatformsCommand) Description() string {
	return "Manage CyberArk platforms"
}

func (c *PlatformsCommand) Usage() string {
	return `platforms <subcommand> [options]

Subcommands:
  list                    List platforms
  get <platform-id>       Get platform details
  activate <platform-id>  Activate a platform
  deactivate <platform-id> Deactivate a platform
  duplicate               Duplicate a platform
  export <platform-id>    Export a platform definition
  delete <platform-id>    Delete a platform

Options for 'list':
  --search=TERM           Search term
  --active                Show only active platforms
  --inactive              Show only inactive platforms
  --type=TYPE             Filter by platform type
  --system=TYPE           Filter by system type

Options for 'duplicate':
  --id=PLATFORM_ID        Source platform ID (required)
  --name=NAME             New platform name (required)
  --description=DESC      New platform description

Options for 'export':
  --output=FILE           Output file path (default: stdout)

Examples:
  platforms list --search=Windows
  platforms list --active
  platforms get WinServerLocal
  platforms activate WinServerLocal
  platforms deactivate WinServerLocal
  platforms duplicate --id=WinServerLocal --name=MyWinPlatform
  platforms export WinServerLocal --output=platform.zip
  platforms delete MyWinPlatform
`
}

func (c *PlatformsCommand) Subcommands() []string {
	return []string{"list", "get", "activate", "deactivate", "duplicate", "export", "delete"}
}

func (c *PlatformsCommand) Execute(execCtx *ExecutionContext, args []string) error {
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
	case "activate":
		return c.activate(execCtx, args[1:])
	case "deactivate":
		return c.deactivate(execCtx, args[1:])
	case "duplicate":
		return c.duplicate(execCtx, args[1:])
	case "export":
		return c.export(execCtx, args[1:])
	case "delete":
		return c.delete(execCtx, args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", args[0])
	}
}

func (c *PlatformsCommand) list(execCtx *ExecutionContext, args []string) error {
	fs := flag.NewFlagSet("platforms list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	search := fs.String("search", "", "Search term")
	active := fs.Bool("active", false, "Show only active platforms")
	inactive := fs.Bool("inactive", false, "Show only inactive platforms")
	platformType := fs.String("type", "", "Filter by platform type")
	systemType := fs.String("system", "", "Filter by system type")

	if err := fs.Parse(args); err != nil {
		return err
	}

	opts := platforms.ListOptions{}
	if *search != "" {
		opts.Search = *search
	}
	if *active {
		t := true
		opts.Active = &t
	}
	if *inactive {
		f := false
		opts.Active = &f
	}
	if *platformType != "" {
		opts.PlatformType = *platformType
	}
	if *systemType != "" {
		opts.SystemType = *systemType
	}

	result, err := platforms.List(execCtx.Ctx, execCtx.Session, opts)
	if err != nil {
		return err
	}

	if len(result.Platforms) == 0 {
		output.PrintInfo("No platforms found")
		return nil
	}

	if execCtx.Formatter.GetFormat() == output.FormatTable {
		table := output.NewTable("ID", "NAME", "TYPE", "SYSTEM", "ACTIVE")
		for _, p := range result.Platforms {
			id := p.PlatformID.String()
			if id == "" {
				id = p.ID.String()
			}
			table.AddRow(
				id,
				p.Name,
				p.PlatformType,
				p.SystemType,
				boolToStr(p.Active),
			)
		}
		table.Render()
		fmt.Printf("\nTotal: %d platforms\n", len(result.Platforms))
	} else {
		return execCtx.Formatter.Format(result)
	}

	return nil
}

func (c *PlatformsCommand) get(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("platform ID required")
	}

	platformID := args[0]
	platform, err := platforms.Get(execCtx.Ctx, execCtx.Session, platformID)
	if err != nil {
		return err
	}

	return execCtx.Formatter.Format(platform)
}

func (c *PlatformsCommand) activate(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("platform ID required")
	}

	platformID := args[0]
	if err := platforms.Activate(execCtx.Ctx, execCtx.Session, platformID); err != nil {
		return err
	}

	output.PrintSuccess("Platform '%s' activated", platformID)
	return nil
}

func (c *PlatformsCommand) deactivate(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("platform ID required")
	}

	platformID := args[0]
	if err := platforms.Deactivate(execCtx.Ctx, execCtx.Session, platformID); err != nil {
		return err
	}

	output.PrintSuccess("Platform '%s' deactivated", platformID)
	return nil
}

func (c *PlatformsCommand) duplicate(execCtx *ExecutionContext, args []string) error {
	fs := flag.NewFlagSet("platforms duplicate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	id := fs.String("id", "", "Source platform ID (required)")
	name := fs.String("name", "", "New platform name (required)")
	description := fs.String("description", "", "New platform description")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == "" {
		return fmt.Errorf("--id is required")
	}
	if *name == "" {
		return fmt.Errorf("--name is required")
	}

	opts := platforms.DuplicateOptions{
		Name:        *name,
		Description: *description,
	}

	platform, err := platforms.Duplicate(execCtx.Ctx, execCtx.Session, *id, opts)
	if err != nil {
		return err
	}

	output.PrintSuccess("Platform '%s' duplicated as '%s'", *id, platform.Name)
	return execCtx.Formatter.Format(platform)
}

func (c *PlatformsCommand) export(execCtx *ExecutionContext, args []string) error {
	fs := flag.NewFlagSet("platforms export", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	outputFile := fs.String("output", "", "Output file path")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("platform ID required")
	}

	platformID := fs.Arg(0)
	data, err := platforms.ExportPlatform(execCtx.Ctx, execCtx.Session, platformID)
	if err != nil {
		return err
	}

	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		output.PrintSuccess("Platform exported to %s", *outputFile)
	} else {
		fmt.Printf("%s\n", string(data))
	}

	return nil
}

func (c *PlatformsCommand) delete(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("platform ID required")
	}

	platformID := args[0]
	if err := platforms.Delete(execCtx.Ctx, execCtx.Session, platformID); err != nil {
		return err
	}

	output.PrintSuccess("Platform '%s' deleted", platformID)
	return nil
}
