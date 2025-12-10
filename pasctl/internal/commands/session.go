package commands

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/chrisranney/gopas"

	"pasctl/internal/output"
)

// ConnectCommand handles connecting to CyberArk.
type ConnectCommand struct{}

func (c *ConnectCommand) Name() string {
	return "connect"
}

func (c *ConnectCommand) Description() string {
	return "Connect to a CyberArk server"
}

func (c *ConnectCommand) Usage() string {
	return `connect [server-url] [options]

Connect and authenticate to a CyberArk PAS server.

Arguments:
  server-url          The CyberArk server URL (e.g., https://cyberark.example.com)

Options:
  --user=USERNAME     Username for authentication
  --auth=METHOD       Authentication method: cyberark, ldap, radius, windows (default: cyberark)
  --insecure          Skip TLS certificate verification
  --ccp               Use CCP (Central Credential Provider) for default login
                      Requires CCP to be configured (see 'ccp setup')

Examples:
  connect https://cyberark.example.com
  connect https://cyberark.example.com --user=admin --auth=ldap
  connect https://cyberark.example.com --insecure
  connect --ccp                          # Use CCP default login
  connect https://cyberark.example.com --ccp
`
}

func (c *ConnectCommand) Execute(execCtx *ExecutionContext, args []string) error {
	if execCtx.Session != nil && execCtx.Session.IsValid() {
		return fmt.Errorf("already connected - use 'disconnect' first")
	}

	// Parse arguments
	var serverURL, username, authMethod string
	var insecure, useCCP bool

	for _, arg := range args {
		if strings.HasPrefix(arg, "--user=") {
			username = strings.TrimPrefix(arg, "--user=")
		} else if strings.HasPrefix(arg, "--auth=") {
			authMethod = strings.ToLower(strings.TrimPrefix(arg, "--auth="))
		} else if arg == "--insecure" {
			insecure = true
		} else if arg == "--ccp" {
			useCCP = true
		} else if !strings.HasPrefix(arg, "-") && serverURL == "" {
			serverURL = arg
		}
	}

	var password string
	var err error

	// Handle CCP login
	if useCCP {
		if !execCtx.Config.IsCCPEnabled() {
			return fmt.Errorf("CCP is not configured - use 'ccp setup' to configure")
		}

		// Get CCP URL for credential retrieval (must be explicitly configured)
		ccpURL := execCtx.Config.GetCCPURL()
		if ccpURL == "" {
			return fmt.Errorf("no CCP URL configured - use 'ccp setup' to set the CCP server URL")
		}

		// Get PVWA URL for authentication (from command line, CCP config, or default server)
		// Note: CCP and PVWA are different services - CCP retrieves credentials,
		// PVWA handles authentication
		if serverURL == "" {
			serverURL = execCtx.Config.GetPVWAURL()
		}
		if serverURL == "" {
			return fmt.Errorf("no PVWA URL configured - use 'ccp setup' to set the PVWA server URL or provide it via command line")
		}

		output.PrintInfo("Retrieving credentials from CCP...")

		// Create CCP client
		ccpClient, err := gopas.NewCCPClient(gopas.CCPClientConfig{
			BaseURL:       ccpURL,
			SkipTLSVerify: insecure || execCtx.Config.InsecureSSL,
			Timeout:       time.Duration(execCtx.Config.Timeout) * time.Second,
			ClientCert:    execCtx.Config.CCP.ClientCert,
			ClientKey:     execCtx.Config.CCP.ClientKey,
		})
		if err != nil {
			return fmt.Errorf("failed to create CCP client: %w", err)
		}

		// Build CCP request
		ccpReq := gopas.CCPCredentialRequest{
			AppID:    execCtx.Config.CCP.AppID,
			Safe:     execCtx.Config.CCP.Safe,
			Object:   execCtx.Config.CCP.Object,
			Folder:   execCtx.Config.CCP.Folder,
			UserName: execCtx.Config.CCP.UserName,
			Address:  execCtx.Config.CCP.Address,
			Query:    execCtx.Config.CCP.Query,
		}

		// Retrieve credentials from CCP
		username, password, err = gopas.CCPGetLoginCredentials(execCtx.Ctx, ccpClient, ccpReq)
		if err != nil {
			return fmt.Errorf("failed to retrieve credentials from CCP: %w", err)
		}

		output.PrintSuccess("Retrieved credentials for user: %s", username)

		// Use CCP auth method if configured
		if authMethod == "" {
			authMethod = execCtx.Config.GetCCPAuthMethod()
		}
	} else {
		// Standard login flow
		if serverURL == "" {
			if execCtx.Config.DefaultServer != "" {
				serverURL = execCtx.Config.DefaultServer
			} else {
				return fmt.Errorf("server URL required")
			}
		}

		// Prompt for username if not provided
		if username == "" {
			username, err = prompt("Username: ")
			if err != nil {
				return err
			}
		}

		// Prompt for password
		password, err = promptPassword("Password: ")
		if err != nil {
			return err
		}

		// Prompt for auth method if not provided
		if authMethod == "" {
			if execCtx.Config.DefaultAuthType != "" {
				authMethod = execCtx.Config.DefaultAuthType
			} else {
				authMethod, err = prompt("Auth Method [cyberark/ldap/radius]: ")
				if err != nil {
					return err
				}
				if authMethod == "" {
					authMethod = "cyberark"
				}
			}
		}
	}

	// Ensure URL has scheme
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "https://" + serverURL
	}

	// Map auth method string to type
	var auth gopas.AuthMethod
	switch strings.ToLower(authMethod) {
	case "ldap":
		auth = gopas.AuthMethodLDAP
	case "radius":
		auth = gopas.AuthMethodRADIUS
	case "windows":
		auth = gopas.AuthMethodWindows
	default:
		auth = gopas.AuthMethodCyberArk
	}

	// Build session options
	opts := gopas.SessionOptions{
		BaseURL: serverURL,
		Credentials: gopas.Credentials{
			Username: username,
			Password: password,
		},
		AuthMethod: auth,
	}

	// Handle insecure mode
	if insecure || execCtx.Config.InsecureSSL {
		opts.CustomHTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: time.Duration(execCtx.Config.Timeout) * time.Second,
		}
	}

	// Attempt connection
	output.PrintInfo("Connecting to %s...", serverURL)

	sess, err := gopas.NewSession(execCtx.Ctx, opts)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Store session in context (this will be updated by the REPL)
	*execCtx.Session = *sess

	output.PrintSuccess("Connected to %s as %s", serverURL, username)

	return nil
}

// DisconnectCommand handles disconnecting from CyberArk.
type DisconnectCommand struct{}

func (c *DisconnectCommand) Name() string {
	return "disconnect"
}

func (c *DisconnectCommand) Description() string {
	return "Disconnect from the current CyberArk server"
}

func (c *DisconnectCommand) Usage() string {
	return `disconnect

Close the current session and disconnect from the CyberArk server.

Examples:
  disconnect
`
}

func (c *DisconnectCommand) Execute(execCtx *ExecutionContext, args []string) error {
	if err := RequireSession(execCtx); err != nil {
		return err
	}

	err := gopas.CloseSession(execCtx.Ctx, execCtx.Session)
	if err != nil {
		output.PrintWarning("Session close failed: %v", err)
	}

	// Clear the session
	*execCtx.Session = gopas.Session{}

	output.PrintSuccess("Session closed")
	return nil
}

// StatusCommand shows connection status.
type StatusCommand struct{}

func (c *StatusCommand) Name() string {
	return "status"
}

func (c *StatusCommand) Description() string {
	return "Show connection status"
}

func (c *StatusCommand) Usage() string {
	return `status

Display information about the current connection status.

Examples:
  status
`
}

func (c *StatusCommand) Execute(execCtx *ExecutionContext, args []string) error {
	fmt.Println()

	if execCtx.Session == nil || !execCtx.Session.IsValid() {
		fmt.Printf("  Connected:    %s\n", output.Error("No"))
		fmt.Println()
		return nil
	}

	sessionAge := time.Since(execCtx.Session.StartTime)

	fmt.Printf("  Connected:    %s\n", output.Success("Yes"))
	fmt.Printf("  Server:       %s\n", execCtx.Session.BaseURI)
	fmt.Printf("  User:         %s\n", execCtx.Session.User)
	fmt.Printf("  Auth Method:  %s\n", execCtx.Session.AuthMethod)
	fmt.Printf("  Session Age:  %s\n", formatDuration(sessionAge))
	if execCtx.Session.ExternalVersion != "" {
		fmt.Printf("  Version:      %s\n", execCtx.Session.ExternalVersion)
	}
	fmt.Println()

	return nil
}

// Helper functions

func prompt(message string) (string, error) {
	fmt.Print(message)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func promptPassword(message string) (string, error) {
	fmt.Print(message)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(password), nil
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
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
