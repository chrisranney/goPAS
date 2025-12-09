package commands

import (
	"fmt"
	"time"

	"github.com/chrisranney/gopas/pkg/systemhealth"

	"pasctl/internal/output"
)

// HealthCommand handles health check operations.
type HealthCommand struct{}

func (c *HealthCommand) Name() string {
	return "health"
}

func (c *HealthCommand) Description() string {
	return "Check CyberArk system health"
}

func (c *HealthCommand) Usage() string {
	return `health <subcommand>

Subcommands:
  check                 Quick system health check
  components            List all component health status
  detail <component-id> Get detailed component information
  summary               Overall system health summary

Examples:
  health check
  health components
  health detail CPM01
  health summary
`
}

func (c *HealthCommand) Subcommands() []string {
	return []string{"check", "components", "detail", "summary"}
}

func (c *HealthCommand) Execute(execCtx *ExecutionContext, args []string) error {
	if err := RequireSession(execCtx); err != nil {
		return err
	}

	if len(args) == 0 {
		// Default to check
		return c.check(execCtx, args)
	}

	switch args[0] {
	case "check":
		return c.check(execCtx, args[1:])
	case "components":
		return c.components(execCtx, args[1:])
	case "detail":
		return c.detail(execCtx, args[1:])
	case "summary":
		return c.summary(execCtx, args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", args[0])
	}
}

func (c *HealthCommand) check(execCtx *ExecutionContext, args []string) error {
	// Get vault health
	health, err := systemhealth.GetVaultHealth(execCtx.Ctx, execCtx.Session)
	if err != nil {
		return err
	}

	fmt.Println()
	if health.IsHealthy {
		fmt.Printf("  Vault Status: %s\n", output.Success("Healthy"))
	} else {
		fmt.Printf("  Vault Status: %s\n", output.Error("Unhealthy"))
		if health.HealthDetails != "" {
			fmt.Printf("  Details:      %s\n", health.HealthDetails)
		}
	}
	fmt.Println()

	// Also show component summary
	components, err := systemhealth.ListComponentSummary(execCtx.Ctx, execCtx.Session)
	if err != nil {
		// Non-fatal, just skip components
		return nil
	}

	if len(components) > 0 {
		healthy := 0
		unhealthy := 0
		for _, comp := range components {
			if comp.IsLoggedOn {
				healthy++
			} else {
				unhealthy++
			}
		}

		fmt.Printf("  Components:   %d total, %s healthy, %s unhealthy\n",
			len(components),
			output.Success(fmt.Sprintf("%d", healthy)),
			output.Error(fmt.Sprintf("%d", unhealthy)),
		)
		fmt.Println()
	}

	return nil
}

func (c *HealthCommand) components(execCtx *ExecutionContext, args []string) error {
	components, err := systemhealth.ListComponentSummary(execCtx.Ctx, execCtx.Session)
	if err != nil {
		return err
	}

	if len(components) == 0 {
		output.PrintInfo("No components found")
		return nil
	}

	if execCtx.Formatter.GetFormat() == output.FormatTable {
		table := output.NewTable("COMPONENT", "TYPE", "STATUS", "LAST SEEN")
		for _, comp := range components {
			status := output.Success("Online")
			if !comp.IsLoggedOn {
				status = output.Error("Offline")
			}

			lastSeen := "-"
			if comp.LastLogonDate > 0 {
				lastSeen = time.Unix(comp.LastLogonDate, 0).Format("2006-01-02 15:04:05")
			}

			table.AddRow(
				comp.ComponentName,
				comp.ComponentType,
				status,
				lastSeen,
			)
		}
		table.Render()
	} else {
		return execCtx.Formatter.Format(components)
	}

	return nil
}

func (c *HealthCommand) detail(execCtx *ExecutionContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("component ID required")
	}

	componentID := args[0]
	detail, err := systemhealth.GetComponentDetail(execCtx.Ctx, execCtx.Session, componentID)
	if err != nil {
		return err
	}

	return execCtx.Formatter.Format(detail)
}

func (c *HealthCommand) summary(execCtx *ExecutionContext, args []string) error {
	// Get vault health
	health, err := systemhealth.GetVaultHealth(execCtx.Ctx, execCtx.Session)
	if err != nil {
		return err
	}

	// Get all components
	components, err := systemhealth.ListComponentSummary(execCtx.Ctx, execCtx.Session)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("  %s\n", output.Header("CyberArk System Health Summary"))
	fmt.Println()

	// Overall status
	if health.IsHealthy {
		fmt.Printf("  Overall Status: %s\n", output.SuccessBold("HEALTHY"))
	} else {
		fmt.Printf("  Overall Status: %s\n", output.ErrorBold("UNHEALTHY"))
		if health.HealthDetails != "" {
			fmt.Printf("  Details: %s\n", health.HealthDetails)
		}
	}
	fmt.Println()

	// Component breakdown by type
	if len(components) > 0 {
		fmt.Printf("  %s\n", output.Header("Component Status"))
		fmt.Println()

		typeStatus := make(map[string]struct {
			online  int
			offline int
		})

		for _, comp := range components {
			status := typeStatus[comp.ComponentType]
			if comp.IsLoggedOn {
				status.online++
			} else {
				status.offline++
			}
			typeStatus[comp.ComponentType] = status
		}

		for compType, status := range typeStatus {
			statusIcon := output.Success("✓")
			if status.offline > 0 {
				statusIcon = output.Warning("!")
			}
			if status.online == 0 && status.offline > 0 {
				statusIcon = output.Error("✗")
			}

			fmt.Printf("  %s %-15s %d online, %d offline\n",
				statusIcon,
				compType+":",
				status.online,
				status.offline,
			)
		}
		fmt.Println()
	}

	return nil
}
