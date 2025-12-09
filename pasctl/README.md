# pasctl - CyberArk PAS Interactive Shell

A fully interactive REPL (Read-Eval-Print Loop) shell for managing CyberArk Privileged Access Security infrastructure. Built on top of the [goPAS](https://github.com/chrisranney/gopas) SDK.

## Features

- **Interactive Shell**: Modern shell experience with command history, tab completion, and colored output
- **Multiple Output Formats**: Table, JSON, and YAML output formats
- **Session Management**: Secure authentication with support for CyberArk, LDAP, RADIUS, and Windows authentication
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
| `disconnect` | Disconnect from the current server |
| `status` | Show connection status |

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

## Tab Completion

pasctl supports intelligent tab completion:

- Top-level commands: `accounts`, `safes`, `users`, etc.
- Subcommands: `accounts <TAB>` → `list`, `get`, `create`, etc.
- Flags: `accounts list --<TAB>` → `--safe`, `--search`, `--limit`

## License

See the main [goPAS](https://github.com/chrisranney/gopas) repository for license information.
