# pasctl - CyberArk PAS Interactive Shell

A fully interactive REPL (Read-Eval-Print Loop) shell for managing CyberArk Privileged Access Security infrastructure. Built on top of the [goPAS](https://github.com/chrisranney/gopas) SDK.

## Features

- **Interactive Shell**: Modern shell experience with command history, tab completion, and colored output
- **Multiple Output Formats**: Table, JSON, and YAML output formats
- **Session Management**: Secure authentication with support for CyberArk, LDAP, RADIUS, and Windows authentication
- **CCP Integration**: Automatic credential retrieval from CyberArk Central Credential Provider (CCP)
- **Comprehensive Commands**: Manage accounts, safes, users, platforms, PSM sessions, and system health
- **Script Mode**: Execute commands from files or stdin for automation
- **Configuration**: Persistent configuration with sensible defaults

## Installation

### Build from Source

```bash
cd pasctl
go build -o pasctl ./cmd/pasctl
```

### Run

```bash
./pasctl
```

## Usage

### Interactive Mode

```bash
$ ./pasctl
pasctl - CyberArk PAS Interactive Shell
Type 'help' for available commands, 'exit' to quit

pasctl> connect https://cyberark.example.com
Username: admin
Password: ********
Auth Method [cyberark/ldap/radius]: ldap
✓ Connected to cyberark.example.com as admin

pasctl> safes list --limit=5
┌─────────────────┬────────────────────────────┬───────────────────┬──────┐
│ NAME            │ DESCRIPTION                │ CPM               │ OLAC │
├─────────────────┼────────────────────────────┼───────────────────┼──────┤
│ Production      │ Production servers         │ PasswordManager   │ No   │
│ Development     │ Dev environment accounts   │ PasswordManager   │ No   │
└─────────────────┴────────────────────────────┴───────────────────┴──────┘

pasctl> exit
✓ Session closed
Goodbye!
```

### Single Command Mode

```bash
# Run a single command and exit
./pasctl -c "safes list --limit=10"

# Pipe commands
echo "accounts list --safe=Production" | ./pasctl
```

### Script Mode

```bash
# Execute commands from a script file
./pasctl --script=commands.txt
```

## Commands

### Session Commands

| Command | Description |
|---------|-------------|
| `connect <url>` | Connect to a CyberArk server |
| `connect --ccp` | Connect using CCP credentials |
| `disconnect` | Disconnect from the current server |
| `status` | Show connection status |

### CCP Commands

| Command | Description |
|---------|-------------|
| `ccp setup` | Interactive setup wizard for CCP configuration |
| `ccp show` | Show current CCP configuration |
| `ccp enable` | Enable CCP default login |
| `ccp disable` | Disable CCP default login |
| `ccp clear` | Clear all CCP configuration |

### Account Commands

| Command | Description |
|---------|-------------|
| `accounts list` | List accounts |
| `accounts get <id>` | Get account details |
| `accounts create` | Create a new account |
| `accounts delete <id>` | Delete an account |
| `accounts password <id>` | Retrieve account password |
| `accounts change <id>` | Trigger immediate password change |
| `accounts verify <id>` | Trigger credential verification |
| `accounts reconcile <id>` | Trigger credential reconciliation |

### Safe Commands

| Command | Description |
|---------|-------------|
| `safes list` | List safes |
| `safes get <name>` | Get safe details |
| `safes create <name>` | Create a new safe |
| `safes update <name>` | Update a safe |
| `safes delete <name>` | Delete a safe |
| `safes members <name>` | List safe members |
| `safes add-member` | Add a member to a safe |
| `safes remove-member` | Remove a member from a safe |

### User Commands

| Command | Description |
|---------|-------------|
| `users list` | List users |
| `users get <id>` | Get user details |
| `users create <username>` | Create a new user |
| `users delete <id>` | Delete a user |
| `users activate <id>` | Activate a suspended user |
| `users reset-password` | Reset a user's password |

### Platform Commands

| Command | Description |
|---------|-------------|
| `platforms list` | List platforms |
| `platforms get <id>` | Get platform details |
| `platforms activate <id>` | Activate a platform |
| `platforms deactivate <id>` | Deactivate a platform |
| `platforms duplicate` | Duplicate a platform |
| `platforms export <id>` | Export a platform |
| `platforms delete <id>` | Delete a platform |

### PSM Commands

| Command | Description |
|---------|-------------|
| `psm sessions` | List recorded PSM sessions |
| `psm live` | List active PSM sessions |
| `psm get <id>` | Get session details |
| `psm terminate <id>` | Terminate a live session |
| `psm suspend <id>` | Suspend a live session |
| `psm resume <id>` | Resume a suspended session |
| `psm activities <id>` | View session activities |

### Health Commands

| Command | Description |
|---------|-------------|
| `health check` | Quick system health check |
| `health components` | List all component health |
| `health summary` | Overall system health summary |
| `health detail <id>` | Get detailed component info |

### Settings Commands

| Command | Description |
|---------|-------------|
| `set output <format>` | Set output format (table/json/yaml) |
| `config` | View or modify configuration |
| `history` | Show command history |
| `clear` | Clear the screen |
| `help [command]` | Show help |

## Configuration

Configuration is stored in `~/.pasctl/config.json`:

```json
{
  "default_server": "https://cyberark.example.com",
  "default_auth_type": "ldap",
  "output_format": "table",
  "history_size": 1000,
  "insecure_ssl": false,
  "timeout_seconds": 30
}
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `default_server` | Default CyberArk server URL | (none) |
| `default_auth_type` | Default authentication method | cyberark |
| `output_format` | Default output format | table |
| `history_size` | Command history size | 1000 |
| `insecure_ssl` | Skip TLS verification | false |
| `timeout_seconds` | Request timeout | 30 |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `PASCTL_SERVER` | Default server URL |
| `PASCTL_USER` | Default username |
| `PASCTL_AUTH` | Default auth method |

## Examples

### Connect and List Accounts

```
pasctl> connect https://cyberark.example.com --auth=ldap
pasctl> accounts list --safe=Production --limit=10
pasctl> accounts password 12_34 --reason="Maintenance"
pasctl> disconnect
```

### Change Output Format

```
pasctl> set output json
pasctl> accounts get 12_34
{
  "id": "12_34",
  "name": "admin",
  "address": "server1.example.com",
  ...
}
pasctl> set output table
```

### Safe Member Management

```
pasctl> safes members Production
pasctl> safes add-member --safe=Production --member=newuser --role=admin
pasctl> safes remove-member --safe=Production --member=olduser
```

### System Health Check

```
pasctl> health check
  Vault Status: Healthy
  Components:   10 total, 10 healthy, 0 unhealthy

pasctl> health components
┌────────────────────┬────────┬────────────────────┬────────┐
│ COMPONENT          │ TYPE   │ STATUS             │ LAST   │
├────────────────────┼────────┼────────────────────┼────────┤
│ Vault              │ Vault  │ Online             │ ...    │
│ CPM01              │ CPM    │ Online             │ ...    │
└────────────────────┴────────┴────────────────────┴────────┘
```

## CCP Authentication

CCP (Central Credential Provider) allows pasctl to retrieve login credentials from a vaulted account, eliminating the need to enter passwords manually. This is ideal for automation scenarios or environments where you want to avoid storing credentials locally.

**Security Note:** Passwords are NEVER stored in the configuration file - they are retrieved from CCP at runtime only.

### Prerequisites

Before using CCP authentication, ensure:

1. **CCP Service**: A CCP web service is deployed and accessible
2. **Application ID**: An application ID is registered in CyberArk for CCP access
3. **Vaulted Account**: The login credentials are stored in a CyberArk safe
4. **Permissions**: The application ID has permission to retrieve the credential

### CCP Setup

#### Interactive Setup (Recommended)

Run the interactive setup wizard:

```
pasctl> ccp setup

CCP Setup Wizard
================
Configure automatic credential retrieval from CyberArk CCP.
Passwords are NEVER stored - they are retrieved at runtime only.

Application ID (required): MyApp
Safe name (required): AdminCredentials
Object name (optional):
Folder path (optional): Root
Username filter (optional): admin
Address filter (optional):
CCP Server URL [https://cyberark.example.com]:
Auth method [ldap]:

✓ CCP configuration saved and enabled

Use 'connect --ccp' to login with CCP credentials.
```

#### Command-Line Setup

You can also configure CCP with command-line options:

```bash
# Basic setup with required options
pasctl> ccp setup --app-id=MyApp --safe=AdminCredentials

# Full setup with all options
pasctl> ccp setup --app-id=MyApp --safe=AdminCredentials --username=admin --auth-method=ldap

# Setup with client certificate for mutual TLS
pasctl> ccp setup --app-id=MyApp --safe=AdminCredentials --client-cert=/path/to/cert.pem --client-key=/path/to/key.pem
```

#### Setup Options

| Option | Description | Required |
|--------|-------------|----------|
| `--app-id=ID` | Application ID registered in CyberArk | Yes |
| `--safe=SAFE` | Safe containing the login credential | Yes |
| `--object=NAME` | Account object name | No |
| `--folder=PATH` | Folder path within the safe | No |
| `--username=USER` | Filter by username | No |
| `--address=ADDR` | Filter by address/hostname | No |
| `--query=QUERY` | Free-text search query | No |
| `--auth-method=METHOD` | Auth method after login (cyberark/ldap/radius) | No |
| `--ccp-url=URL` | CCP server URL (defaults to server URL) | No |
| `--client-cert=PATH` | Client certificate for mutual TLS | No |
| `--client-key=PATH` | Client key for mutual TLS | No |

### Using CCP to Connect

Once CCP is configured, use the `--ccp` flag to connect:

```
pasctl> connect --ccp
ℹ Retrieving credentials from CCP...
✓ Retrieved credentials for user: admin
ℹ Connecting to https://cyberark.example.com...
✓ Connected to https://cyberark.example.com as admin
```

You can also specify a different server URL:

```
pasctl> connect https://other-server.example.com --ccp
```

### Managing CCP Configuration

```
# View current CCP configuration
pasctl> ccp show

CCP Configuration
-----------------
  Status:        Enabled
  App ID:        MyApp
  Safe:          AdminCredentials
  Username:      admin
  Auth Method:   ldap

# Temporarily disable CCP (keeps configuration)
pasctl> ccp disable
✓ CCP default login disabled

# Re-enable CCP
pasctl> ccp enable
✓ CCP default login enabled

# Clear all CCP configuration
pasctl> ccp clear
✓ CCP configuration cleared
```

### CCP Configuration File

CCP settings are stored in `~/.pasctl/config.json`:

```json
{
  "default_server": "https://cyberark.example.com",
  "default_auth_type": "ldap",
  "ccp": {
    "enabled": true,
    "app_id": "MyApp",
    "safe": "AdminCredentials",
    "username": "admin",
    "auth_method": "ldap"
  }
}
```

### Troubleshooting CCP

| Issue | Solution |
|-------|----------|
| "CCP is not configured" | Run `ccp setup` to configure CCP settings |
| "Failed to retrieve credentials" | Verify the application ID has access to the safe |
| Certificate errors | Use `--insecure` flag or configure proper certificates |
| Wrong credentials retrieved | Use additional filters (`--username`, `--address`, `--object`) to narrow down the account |

## Tab Completion

pasctl supports intelligent tab completion:

- Top-level commands: `accounts`, `safes`, `users`, etc.
- Subcommands: `accounts <TAB>` → `list`, `get`, `create`, etc.
- Flags: `accounts list --<TAB>` → `--safe`, `--search`, `--limit`

## License

See the main [goPAS](https://github.com/chrisranney/gopas) repository for license information.
