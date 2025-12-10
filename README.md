# goPAS

[![Go Reference](https://pkg.go.dev/badge/github.com/chrisranney/gopas.svg)](https://pkg.go.dev/github.com/chrisranney/gopas)
[![Go Report Card](https://goreportcard.com/badge/github.com/chrisranney/gopas)](https://goreportcard.com/report/github.com/chrisranney/gopas)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/chrisranney/gopas)](https://go.dev/)
[![Coverage](https://img.shields.io/badge/coverage-46.5%25-yellow)](https://github.com/chrisranney/gopas)

**goPAS** is a Go SDK for the CyberArk Privileged Access Security (PAS) REST API.

Inspired by the popular [psPAS](https://github.com/pspete/psPAS) PowerShell module, goPAS brings the same comprehensive CyberArk functionality to Go applications with idiomatic APIs, strong typing, and full context support.

## Installation

```bash
go get github.com/chrisranney/gopas
```

**Requirements:** Go 1.21 or later

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/chrisranney/gopas"
)

func main() {
    ctx := context.Background()

    // Create an authenticated session
    sess, err := gopas.NewSession(ctx, gopas.SessionOptions{
        BaseURL: "https://cyberark.example.com",
        Credentials: gopas.Credentials{
            Username: "admin",
            Password: "password",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer gopas.CloseSession(ctx, sess)

    // List accounts
    accounts, err := gopas.ListAccounts(ctx, sess, gopas.ListAccountsOptions{
        Limit: 10,
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, acct := range accounts.Value {
        fmt.Printf("Account: %s@%s\n", acct.UserName, acct.Address)
    }
}
```

## Features

- **Full API Coverage** - Supports all major CyberArk operations: accounts, safes, users, platforms, requests, and more
- **Multiple Auth Methods** - CyberArk, LDAP, RADIUS, SAML, Windows, and CCP (Central Credential Provider)
- **Type-Safe** - Strongly typed request/response structures with full IDE support
- **Context Support** - First-class `context.Context` support for cancellation and timeouts
- **CyberArk v14.0 Compatible** - Tested against CyberArk versions up to v14.0

## Packages

| Package | Description |
|---------|-------------|
| `gopas` | Main entry point with convenience functions |
| `pkg/accounts` | Account CRUD, password retrieval, CPM operations |
| `pkg/safes` | Safe management |
| `pkg/safemembers` | Safe member permissions |
| `pkg/users` | User and group management |
| `pkg/platforms` | Platform configuration |
| `pkg/requests` | Access request workflows |
| `pkg/applications` | Application management |
| `pkg/authentication` | Session management |
| `pkg/ccp` | Central Credential Provider (CCP) credential retrieval |
| `pkg/monitoring` | PSM session monitoring |
| `pkg/connections` | PSM connections |
| `pkg/systemhealth` | Component health checks |
| `pkg/eventsecurity` | PTA events |
| `pkg/ldapdirectories` | LDAP configuration |
| `pkg/onboardingrules` | Automatic onboarding |
| `pkg/accountgroups` | Account groups |
| `pkg/accountacl` | Account ACLs |
| `pkg/policyacl` | Policy ACLs |
| `pkg/ipallowlist` | IP allow lists |

## Authentication

### CyberArk Native

```go
sess, err := gopas.NewSession(ctx, gopas.SessionOptions{
    BaseURL:    "https://cyberark.example.com",
    Credentials: gopas.Credentials{
        Username: "admin",
        Password: "password",
    },
    AuthMethod: gopas.AuthMethodCyberArk,
})
```

### LDAP

```go
sess, err := gopas.NewSession(ctx, gopas.SessionOptions{
    BaseURL:    "https://cyberark.example.com",
    Credentials: gopas.Credentials{
        Username: "user@domain.com",
        Password: "password",
    },
    AuthMethod: gopas.AuthMethodLDAP,
})
```

### RADIUS

```go
sess, err := gopas.NewSession(ctx, gopas.SessionOptions{
    BaseURL:    "https://cyberark.example.com",
    Credentials: gopas.Credentials{
        Username: "admin",
        Password: "password",
    },
    AuthMethod: gopas.AuthMethodRADIUS,
})
```

### CCP (Central Credential Provider)

CCP allows applications to retrieve credentials from CyberArk without storing passwords. This is ideal for automated systems and application-to-vault communication.

```go
import "github.com/chrisranney/gopas/pkg/ccp"

// Create a CCP client
ccpClient, err := ccp.NewClient(ccp.ClientConfig{
    BaseURL:       "https://cyberark.example.com",
    SkipTLSVerify: false,
})
if err != nil {
    log.Fatal(err)
}

// Retrieve credentials from CCP
username, password, err := ccpClient.GetLoginCredentials(ctx, ccp.CredentialRequest{
    AppID: "MyApplication",
    Safe:  "AdminCredentials",
    // Optional filters
    UserName: "admin",
    Address:  "server.example.com",
})
if err != nil {
    log.Fatal(err)
}

// Use retrieved credentials to create a session
sess, err := gopas.NewSession(ctx, gopas.SessionOptions{
    BaseURL: "https://cyberark.example.com",
    Credentials: gopas.Credentials{
        Username: username,
        Password: password,
    },
})
```

#### CCP with Mutual TLS

For enhanced security, CCP supports mutual TLS authentication:

```go
ccpClient, err := ccp.NewClient(ccp.ClientConfig{
    BaseURL:    "https://cyberark.example.com",
    ClientCert: "/path/to/client.crt",
    ClientKey:  "/path/to/client.key",
})
```

#### CCP Credential Request Options

```go
cred, err := ccpClient.GetCredential(ctx, ccp.CredentialRequest{
    AppID:       "MyApp",           // Required: Application ID registered in CyberArk
    Safe:        "MySafe",          // Required: Safe containing the credential
    Object:      "AccountName",     // Optional: Account object name
    Folder:      "Root\\Subfolder", // Optional: Folder path within safe
    UserName:    "admin",           // Optional: Filter by username
    Address:     "server.local",    // Optional: Filter by address
    Query:       "admin",           // Optional: Free-text search
    QueryFormat: "Exact",           // Optional: "Exact" or "Regexp"
    Reason:      "Automated login", // Optional: Audit reason
})
```

## Common Operations

### Accounts

```go
import "github.com/chrisranney/gopas/pkg/accounts"

// List accounts
resp, _ := accounts.List(ctx, sess, accounts.ListOptions{
    SafeName: "MySafe",
    Search:   "admin",
})

// Get account by ID
acct, _ := accounts.Get(ctx, sess, "12_34")

// Create account
newAcct, _ := accounts.Create(ctx, sess, accounts.CreateOptions{
    SafeName:   "MySafe",
    PlatformID: "WinServerLocal",
    Address:    "server.example.com",
    UserName:   "admin",
    Secret:     "password123",
})

// Retrieve password
password, _ := accounts.GetPassword(ctx, sess, "12_34", "Authorized access")

// CPM operations
accounts.ChangeCredentialsImmediately(ctx, sess, "12_34", accounts.ChangeCredentialsOptions{})
accounts.VerifyCredentials(ctx, sess, "12_34")
accounts.ReconcileCredentials(ctx, sess, "12_34")
```

### Safes

```go
import "github.com/chrisranney/gopas/pkg/safes"

// List safes
resp, _ := safes.List(ctx, sess, safes.ListOptions{
    Search: "Production",
})

// Get safe
safe, _ := safes.Get(ctx, sess, "MySafe")

// Create safe
newSafe, _ := safes.Create(ctx, sess, safes.CreateOptions{
    SafeName:    "NewSafe",
    Description: "Production accounts",
    ManagingCPM: "PasswordManager",
})

// Update safe
safes.Update(ctx, sess, "MySafe", safes.UpdateOptions{
    Description: "Updated description",
})

// Delete safe
safes.Delete(ctx, sess, "OldSafe")
```

### Safe Members

```go
import "github.com/chrisranney/gopas/pkg/safemembers"

// List members
members, _ := safemembers.List(ctx, sess, "MySafe", safemembers.ListOptions{})

// Add member with default permissions
safemembers.Add(ctx, sess, "MySafe", safemembers.AddOptions{
    MemberName:  "ServiceAccount",
    Permissions: safemembers.DefaultUserPermissions(),
})

// Update member permissions
safemembers.Update(ctx, sess, "MySafe", "ServiceAccount", safemembers.UpdateOptions{
    Permissions: safemembers.DefaultAdminPermissions(),
})

// Remove member
safemembers.Remove(ctx, sess, "MySafe", "OldMember")
```

### Users

```go
import "github.com/chrisranney/gopas/pkg/users"

// List users
usersResp, _ := users.List(ctx, sess, users.ListOptions{
    Search: "admin",
})

// Create user
newUser, _ := users.Create(ctx, sess, users.CreateOptions{
    Username:        "newuser",
    InitialPassword: "TempPassword123!",
    UserType:        "EPVUser",
})

// Reset password
users.ResetPassword(ctx, sess, 123, "NewPassword123!")

// Activate suspended user
users.ActivateUser(ctx, sess, 123)
```

### PSM Monitoring

```go
import "github.com/chrisranney/gopas/pkg/monitoring"

// List all sessions
sessions, _ := monitoring.ListSessions(ctx, sess, monitoring.ListOptions{
    FromTime: time.Now().Add(-24*time.Hour).Unix(),
})

// List live sessions
liveSessions, _ := monitoring.ListLiveSessions(ctx, sess, monitoring.ListOptions{})

// Terminate a live session
monitoring.TerminateSession(ctx, sess, "session-id")
```

## Error Handling

```go
import "github.com/chrisranney/gopas/internal/client"

account, err := accounts.Get(ctx, sess, "invalid-id")
if err != nil {
    if apiErr, ok := client.AsAPIError(err); ok {
        switch {
        case apiErr.IsNotFound():
            fmt.Println("Account not found")
        case apiErr.IsUnauthorized():
            fmt.Println("Not authorized")
        default:
            fmt.Printf("API Error: %s\n", apiErr.ErrorMsg)
        }
    }
}
```

## Testing

Run the test suite:

```bash
go test ./...
```

Run with coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Coverage

The SDK includes comprehensive tests for all major packages:

| Package | Coverage | Description |
|---------|----------|-------------|
| `internal/session` | 100.0% | Session management |
| `internal/helpers` | 98.5% | Utility functions and version parsing |
| `internal/client` | 97.3% | Core HTTP client and error handling |
| `pkg/ccp` | 85.0% | CCP credential retrieval |
| `pkg/safes` | 80.0% | Safe operations |
| `pkg/systemhealth` | 79.3% | Health checks |
| `pkg/connections` | 79.2% | PSM connections |
| `pkg/applications` | 78.4% | Application management |
| `pkg/safemembers` | 76.0% | Safe member management |
| `pkg/platforms` | 75.9% | Platform operations |
| `pkg/requests` | 71.4% | Access requests |
| `pkg/accounts` | 64.0% | Account operations |
| `pkg/authentication` | 60.4% | Authentication flows |
| `pkg/users` | 39.3% | User management |
| **Overall** | **46.5%** | All packages combined |

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Run `go test ./...` to ensure all tests pass
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [psPAS](https://github.com/pspete/psPAS) by Pete Maan - The original PowerShell module this SDK is based on
- [CyberArk](https://www.cyberark.com/) - For providing the REST API
