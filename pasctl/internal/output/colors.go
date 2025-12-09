// Package output provides output formatting and colorization for pasctl.
package output

import (
	"github.com/fatih/color"
)

// Color functions for consistent styling across the application.
var (
	// Success colors
	SuccessColor = color.New(color.FgGreen)
	Success      = SuccessColor.SprintFunc()
	SuccessBold  = color.New(color.FgGreen, color.Bold).SprintFunc()

	// Error colors
	ErrorColor = color.New(color.FgRed)
	Error      = ErrorColor.SprintFunc()
	ErrorBold  = color.New(color.FgRed, color.Bold).SprintFunc()

	// Warning colors
	WarningColor = color.New(color.FgYellow)
	Warning      = WarningColor.SprintFunc()
	WarnBold     = color.New(color.FgYellow, color.Bold).SprintFunc()

	// Info colors
	InfoColor = color.New(color.FgCyan)
	Info      = InfoColor.SprintFunc()
	InfoBold  = color.New(color.FgCyan, color.Bold).SprintFunc()

	// Header colors
	HeaderColor = color.New(color.FgWhite, color.Bold)
	Header      = HeaderColor.SprintFunc()

	// Dim colors for secondary information
	DimColor = color.New(color.FgHiBlack)
	Dim      = DimColor.SprintFunc()

	// Bold
	Bold = color.New(color.Bold).SprintFunc()
)

// Status returns a colored status indicator.
func Status(ok bool) string {
	if ok {
		return Success("✓")
	}
	return Error("✗")
}

// StatusText returns a colored status text.
func StatusText(ok bool) string {
	if ok {
		return Success("OK")
	}
	return Error("FAILED")
}

// BoolStatus returns a colored boolean status.
func BoolStatus(b bool) string {
	if b {
		return Success("Yes")
	}
	return Dim("No")
}
