package commands

import (
	"fmt"
	"strings"

	"pasctl/internal/config"
	"pasctl/internal/output"
)

// CCPCommand handles CCP configuration.
type CCPCommand struct{}

func (c *CCPCommand) Name() string {
	return "ccp"
}

func (c *CCPCommand) Description() string {
	return "Configure CCP (Central Credential Provider) default login"
}

func (c *CCPCommand) Usage() string {
	return `ccp <subcommand> [options]

Configure CCP (Central Credential Provider) for automatic credential retrieval.
CCP allows pasctl to retrieve login credentials from a vaulted account,
eliminating the need to enter passwords manually.

NOTE: Passwords are NEVER stored in configuration - they are retrieved
from CCP at runtime only.

Subcommands:
  setup                 Interactive setup wizard for CCP configuration
  show                  Show current CCP configuration
  enable                Enable CCP default login
  disable               Disable CCP default login
  clear                 Clear all CCP configuration

Setup Options (can also be set via 'set' command):
  --app-id=ID           Application ID registered in CyberArk (required)
  --safe=SAFE           Safe containing the login credential (required)
  --object=NAME         Account object name (optional)
  --folder=PATH         Folder path within the safe (optional)
  --username=USER       Filter by username (optional)
  --address=ADDR        Filter by address/hostname (optional)
  --query=QUERY         Free-text search query (optional)
  --auth-method=METHOD  Auth method to use after retrieving creds (optional)
  --ccp-url=URL         CCP server URL for credential retrieval (required)
  --pvwa-url=URL        PVWA server URL for authentication (defaults to server URL)
  --client-cert=PATH    Client certificate for mutual TLS (optional)
  --client-key=PATH     Client key for mutual TLS (optional)

Examples:
  ccp setup                           # Interactive setup wizard
  ccp setup --app-id=MyApp --safe=AdminCreds --username=admin
  ccp show                            # Show current configuration
  ccp enable                          # Enable CCP login
  ccp disable                         # Disable CCP login
  ccp clear                           # Clear all CCP settings

After setup, use 'connect --ccp' to login with CCP credentials.
`
}

func (c *CCPCommand) Execute(execCtx *ExecutionContext, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("subcommand required (setup, show, enable, disable, clear)")
	}

	subcommand := strings.ToLower(args[0])
	subArgs := args[1:]

	switch subcommand {
	case "setup":
		return c.setup(execCtx, subArgs)
	case "show":
		return c.show(execCtx)
	case "enable":
		return c.enable(execCtx)
	case "disable":
		return c.disable(execCtx)
	case "clear":
		return c.clear(execCtx)
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func (c *CCPCommand) setup(execCtx *ExecutionContext, args []string) error {
	// Initialize CCP config if nil
	if execCtx.Config.CCP == nil {
		execCtx.Config.CCP = &config.CCPConfig{}
	}

	// Parse command line options first
	hasOptions := false
	for _, arg := range args {
		if strings.HasPrefix(arg, "--app-id=") {
			execCtx.Config.CCP.AppID = strings.TrimPrefix(arg, "--app-id=")
			hasOptions = true
		} else if strings.HasPrefix(arg, "--safe=") {
			execCtx.Config.CCP.Safe = strings.TrimPrefix(arg, "--safe=")
			hasOptions = true
		} else if strings.HasPrefix(arg, "--object=") {
			execCtx.Config.CCP.Object = strings.TrimPrefix(arg, "--object=")
			hasOptions = true
		} else if strings.HasPrefix(arg, "--folder=") {
			execCtx.Config.CCP.Folder = strings.TrimPrefix(arg, "--folder=")
			hasOptions = true
		} else if strings.HasPrefix(arg, "--username=") {
			execCtx.Config.CCP.UserName = strings.TrimPrefix(arg, "--username=")
			hasOptions = true
		} else if strings.HasPrefix(arg, "--address=") {
			execCtx.Config.CCP.Address = strings.TrimPrefix(arg, "--address=")
			hasOptions = true
		} else if strings.HasPrefix(arg, "--query=") {
			execCtx.Config.CCP.Query = strings.TrimPrefix(arg, "--query=")
			hasOptions = true
		} else if strings.HasPrefix(arg, "--auth-method=") {
			execCtx.Config.CCP.AuthMethod = strings.TrimPrefix(arg, "--auth-method=")
			hasOptions = true
		} else if strings.HasPrefix(arg, "--ccp-url=") {
			execCtx.Config.CCP.CCPURL = strings.TrimPrefix(arg, "--ccp-url=")
			hasOptions = true
		} else if strings.HasPrefix(arg, "--pvwa-url=") {
			execCtx.Config.CCP.PVWAURL = strings.TrimPrefix(arg, "--pvwa-url=")
			hasOptions = true
		} else if strings.HasPrefix(arg, "--client-cert=") {
			execCtx.Config.CCP.ClientCert = strings.TrimPrefix(arg, "--client-cert=")
			hasOptions = true
		} else if strings.HasPrefix(arg, "--client-key=") {
			execCtx.Config.CCP.ClientKey = strings.TrimPrefix(arg, "--client-key=")
			hasOptions = true
		}
	}

	// If no options provided, run interactive setup
	if !hasOptions {
		return c.interactiveSetup(execCtx)
	}

	// Validate required fields
	if execCtx.Config.CCP.AppID == "" {
		return fmt.Errorf("--app-id is required")
	}
	if execCtx.Config.CCP.Safe == "" {
		return fmt.Errorf("--safe is required")
	}
	if execCtx.Config.CCP.CCPURL == "" {
		return fmt.Errorf("--ccp-url is required")
	}

	// Enable CCP
	execCtx.Config.CCP.Enabled = true

	// Save configuration
	if err := execCtx.Config.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	output.PrintSuccess("CCP configuration saved and enabled")
	return c.show(execCtx)
}

func (c *CCPCommand) interactiveSetup(execCtx *ExecutionContext) error {
	fmt.Println()
	fmt.Println("CCP Setup Wizard")
	fmt.Println("================")
	fmt.Println("Configure automatic credential retrieval from CyberArk CCP.")
	fmt.Println("Passwords are NEVER stored - they are retrieved at runtime only.")
	fmt.Println()

	// Initialize CCP config if nil
	if execCtx.Config.CCP == nil {
		execCtx.Config.CCP = &config.CCPConfig{}
	}

	var err error

	// AppID (required)
	defaultAppID := execCtx.Config.CCP.AppID
	promptText := "Application ID (required)"
	if defaultAppID != "" {
		promptText = fmt.Sprintf("Application ID [%s]", defaultAppID)
	}
	appID, err := prompt(promptText + ": ")
	if err != nil {
		return err
	}
	if appID != "" {
		execCtx.Config.CCP.AppID = appID
	}
	if execCtx.Config.CCP.AppID == "" {
		return fmt.Errorf("Application ID is required")
	}

	// Safe (required)
	defaultSafe := execCtx.Config.CCP.Safe
	promptText = "Safe name (required)"
	if defaultSafe != "" {
		promptText = fmt.Sprintf("Safe name [%s]", defaultSafe)
	}
	safe, err := prompt(promptText + ": ")
	if err != nil {
		return err
	}
	if safe != "" {
		execCtx.Config.CCP.Safe = safe
	}
	if execCtx.Config.CCP.Safe == "" {
		return fmt.Errorf("Safe name is required")
	}

	// Object (optional)
	defaultObject := execCtx.Config.CCP.Object
	promptText = "Object name (optional)"
	if defaultObject != "" {
		promptText = fmt.Sprintf("Object name [%s]", defaultObject)
	}
	object, err := prompt(promptText + ": ")
	if err != nil {
		return err
	}
	if object != "" {
		execCtx.Config.CCP.Object = object
	}

	// Folder (optional)
	defaultFolder := execCtx.Config.CCP.Folder
	promptText = "Folder path (optional)"
	if defaultFolder != "" {
		promptText = fmt.Sprintf("Folder path [%s]", defaultFolder)
	}
	folder, err := prompt(promptText + ": ")
	if err != nil {
		return err
	}
	if folder != "" {
		execCtx.Config.CCP.Folder = folder
	}

	// Username filter (optional)
	defaultUsername := execCtx.Config.CCP.UserName
	promptText = "Username filter (optional)"
	if defaultUsername != "" {
		promptText = fmt.Sprintf("Username filter [%s]", defaultUsername)
	}
	username, err := prompt(promptText + ": ")
	if err != nil {
		return err
	}
	if username != "" {
		execCtx.Config.CCP.UserName = username
	}

	// Address filter (optional)
	defaultAddress := execCtx.Config.CCP.Address
	promptText = "Address filter (optional)"
	if defaultAddress != "" {
		promptText = fmt.Sprintf("Address filter [%s]", defaultAddress)
	}
	address, err := prompt(promptText + ": ")
	if err != nil {
		return err
	}
	if address != "" {
		execCtx.Config.CCP.Address = address
	}

	// CCP URL (required - for credential retrieval)
	fmt.Println()
	fmt.Println("Server URLs:")
	fmt.Println("  CCP = Central Credential Provider (retrieves credentials)")
	fmt.Println("  PVWA = Privileged Vault Web Access (authenticates to CyberArk)")
	fmt.Println()
	defaultCCPURL := execCtx.Config.CCP.CCPURL
	promptText = "CCP Server URL (required)"
	if defaultCCPURL != "" {
		promptText = fmt.Sprintf("CCP Server URL [%s]", defaultCCPURL)
	}
	ccpURL, err := prompt(promptText + ": ")
	if err != nil {
		return err
	}
	if ccpURL != "" {
		execCtx.Config.CCP.CCPURL = ccpURL
	}
	if execCtx.Config.CCP.CCPURL == "" {
		return fmt.Errorf("CCP Server URL is required")
	}

	// PVWA URL (optional - for authentication)
	defaultPVWAURL := execCtx.Config.CCP.PVWAURL
	if defaultPVWAURL == "" && execCtx.Config.DefaultServer != "" {
		defaultPVWAURL = execCtx.Config.DefaultServer
	}
	promptText = "PVWA Server URL (optional, defaults to server URL)"
	if defaultPVWAURL != "" {
		promptText = fmt.Sprintf("PVWA Server URL [%s]", defaultPVWAURL)
	}
	pvwaURL, err := prompt(promptText + ": ")
	if err != nil {
		return err
	}
	if pvwaURL != "" {
		execCtx.Config.CCP.PVWAURL = pvwaURL
	}

	// Auth method (optional)
	defaultAuthMethod := execCtx.Config.CCP.AuthMethod
	if defaultAuthMethod == "" {
		defaultAuthMethod = execCtx.Config.DefaultAuthType
	}
	promptText = "Auth method after login (cyberark/ldap/radius)"
	if defaultAuthMethod != "" {
		promptText = fmt.Sprintf("Auth method [%s]", defaultAuthMethod)
	}
	authMethod, err := prompt(promptText + ": ")
	if err != nil {
		return err
	}
	if authMethod != "" {
		execCtx.Config.CCP.AuthMethod = authMethod
	}

	// Enable CCP
	execCtx.Config.CCP.Enabled = true

	// Save configuration
	if err := execCtx.Config.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println()
	output.PrintSuccess("CCP configuration saved and enabled")
	fmt.Println()
	fmt.Println("Use 'connect --ccp' to login with CCP credentials.")
	fmt.Println()

	return nil
}

func (c *CCPCommand) show(execCtx *ExecutionContext) error {
	fmt.Println()
	fmt.Println("CCP Configuration")
	fmt.Println("-----------------")

	if execCtx.Config.CCP == nil {
		fmt.Println("  Not configured")
		fmt.Println()
		return nil
	}

	ccp := execCtx.Config.CCP

	if ccp.Enabled {
		fmt.Printf("  Status:        %s\n", output.Success("Enabled"))
	} else {
		fmt.Printf("  Status:        %s\n", output.Warning("Disabled"))
	}

	if ccp.AppID != "" {
		fmt.Printf("  App ID:        %s\n", ccp.AppID)
	}
	if ccp.Safe != "" {
		fmt.Printf("  Safe:          %s\n", ccp.Safe)
	}
	if ccp.Object != "" {
		fmt.Printf("  Object:        %s\n", ccp.Object)
	}
	if ccp.Folder != "" {
		fmt.Printf("  Folder:        %s\n", ccp.Folder)
	}
	if ccp.UserName != "" {
		fmt.Printf("  Username:      %s\n", ccp.UserName)
	}
	if ccp.Address != "" {
		fmt.Printf("  Address:       %s\n", ccp.Address)
	}
	if ccp.Query != "" {
		fmt.Printf("  Query:         %s\n", ccp.Query)
	}
	if ccp.CCPURL != "" {
		fmt.Printf("  CCP URL:       %s\n", ccp.CCPURL)
	}
	if ccp.PVWAURL != "" {
		fmt.Printf("  PVWA URL:      %s\n", ccp.PVWAURL)
	}
	if ccp.AuthMethod != "" {
		fmt.Printf("  Auth Method:   %s\n", ccp.AuthMethod)
	}
	if ccp.ClientCert != "" {
		fmt.Printf("  Client Cert:   %s\n", ccp.ClientCert)
	}
	if ccp.ClientKey != "" {
		fmt.Printf("  Client Key:    %s\n", ccp.ClientKey)
	}

	fmt.Println()
	return nil
}

func (c *CCPCommand) enable(execCtx *ExecutionContext) error {
	if execCtx.Config.CCP == nil {
		return fmt.Errorf("CCP is not configured - use 'ccp setup' first")
	}

	if execCtx.Config.CCP.AppID == "" || execCtx.Config.CCP.Safe == "" {
		return fmt.Errorf("CCP is not fully configured - use 'ccp setup' to configure required fields")
	}

	execCtx.Config.CCP.Enabled = true

	if err := execCtx.Config.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	output.PrintSuccess("CCP default login enabled")
	return nil
}

func (c *CCPCommand) disable(execCtx *ExecutionContext) error {
	if execCtx.Config.CCP == nil {
		output.PrintWarning("CCP is not configured")
		return nil
	}

	execCtx.Config.CCP.Enabled = false

	if err := execCtx.Config.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	output.PrintSuccess("CCP default login disabled")
	return nil
}

func (c *CCPCommand) clear(execCtx *ExecutionContext) error {
	execCtx.Config.CCP = nil

	if err := execCtx.Config.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	output.PrintSuccess("CCP configuration cleared")
	return nil
}
