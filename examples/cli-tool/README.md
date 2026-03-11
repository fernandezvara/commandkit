# CLI Tool Example

A comprehensive command-line application demonstrating CommandKit's command system, middleware pipeline, authentication, and advanced features.

## Features Demonstrated

- **Command system** - Multiple commands with subcommands and aliases
- **Middleware pipeline** - Global, command-specific, and custom middleware
- **Authentication** - Token-based auth for admin and API commands
- **Help customization** - Professional help system with long/short descriptions
- **Priority configuration** - Environment and flag source ordering
- **Error handling** - Unified cfg.Execute() API with professional error display
- **Rate limiting** - Built-in rate limiting for sensitive operations
- **Command aliases** - Multiple command names for the same functionality

## Usage

### Basic Commands
```bash
# Show help
go run main.go help
go run main.go ?

# Show system status
API_KEY=your-api-key go run main.go status
go run main.go status --detailed

# Show API status
API_KEY=your-api-key go run main.go status api
```

### Deploy Commands
```bash
# Deploy to staging
go run main.go deploy --env staging

# Deploy with specific branch
go run main.go deploy --env prod --branch feature/new-ui

# Dry run deployment
go run main.go deploy --env staging --dry-run

# Force deployment
go run main.go deploy --env prod --force --skip-tests

# Deploy subcommands
go run main.go deploy rollback --version v1.2.1
go run main.go deploy status --env prod
```

### Admin Commands (Authentication Required)
```bash
# List users
ADMIN_TOKEN=your-admin-token go run main.go admin users --action list

# Create user
ADMIN_TOKEN=your-admin-token go run main.go admin users --action create --username newuser --role admin

# Shutdown service
ADMIN_TOKEN=your-admin-token go run main.go admin shutdown --graceful --delay 60s
```

### Configuration Management
```bash
# Show current configuration
go run main.go config

# Show configuration with secrets
go run main.go config --show-secrets

# Validate configuration only
go run main.go config --validate-only
```

## Command Structure

### Global Options
These options apply to all commands:

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| --verbose | bool | false | Enable verbose logging |
| --log-level | string | info | Logging level (debug/info/warn/error) |
| --config | string | - | Configuration file path |
| --timeout | duration | 30s | Operation timeout |

### Deploy Command
```bash
cli-tool deploy [options] [subcommand]
```

**Flags:**
- `--env` (required): Target environment (dev/staging/prod)
- `--dry-run`: Perform dry run without changes
- `--skip-tests`: Skip running tests
- `--force`: Force deployment despite checks
- `--branch`: Git branch to deploy (default: main)
- `--tag`: Git tag to deploy

**Subcommands:**
- `rollback --version <version>`: Rollback to specific version
- `status --env <env>`: Show deployment status

### Admin Command
```bash
cli-tool admin [subcommand]
```

**Requires:** `ADMIN_TOKEN` environment variable

**Subcommands:**
- `users --action <action>`: Manage users
  - Actions: list, create, delete, update
  - Flags: `--username`, `--role`
- `shutdown`: Shutdown service
  - Flags: `--graceful`, `--delay`

### Status Command
```bash
cli-tool status [subcommand]
```

**Requires:** `API_KEY` environment variable

**Flags:**
- `--detailed`: Show detailed status
- `--format`: Output format (text/json/yaml)

**Subcommands:**
- `api --endpoint <path>`: API service status
- `database`: Database status
  - Flags: `--check-connection`, `--show-stats`

## Middleware Pipeline

The example demonstrates a comprehensive middleware pipeline:

### Global Middleware (applies to all commands)
1. **RecoveryMiddleware** - Catches panics and provides clean error messages
2. **TimingMiddleware** - Measures command execution time
3. **DefaultLoggingMiddleware** - Logs all command executions
4. **DefaultErrorHandlingMiddleware** - Consistent error formatting
5. **DefaultMetricsMiddleware** - Collects execution metrics

### Command-Specific Middleware
- **RateLimitMiddleware** - Applied to deploy commands (10 requests/minute)
- **TokenAuthMiddleware** - Applied to admin commands (ADMIN_TOKEN)
- **TokenAuthMiddleware** - Applied to status commands (API_KEY)

## Authentication Examples

### Admin Authentication
```bash
# Set admin token
export ADMIN_TOKEN="your-secure-admin-token"

# Run admin commands
go run main.go admin users --action list
go run main.go admin shutdown --graceful
```

### API Authentication
```bash
# Set API key
export API_KEY="your-api-key"

# Run status commands
go run main.go status --detailed
go run main.go status api --endpoint /health
```

## Command Aliases

Commands support multiple names for convenience:

```bash
# Deploy command aliases
go run main.go deploy
go run main.go dep
go run main.go release

# Help command aliases
go run main.go help
go run main.go ?
go run main.go --help
go run main.go -h
```

## Error Handling

The example uses CommandKit's unified error handling:

```bash
# Missing required option
go run main.go deploy
# Shows: --env string (required) -> value not provided

# Invalid option value
go run main.go deploy --env invalid
# Shows: --env string -> value invalid is not one of [dev staging prod]

# Authentication failure
go run main.go admin users --action list
# Shows: authentication failed: admin token not provided
```

## Help System

Professional help is automatically generated:

```bash
# Global help
go run main.go --help
# Shows: Available commands, global options

# Command help
go run main.go deploy --help
# Shows: Deploy command description, options, subcommands

# Subcommand help
go run main.go admin users --help
# Shows: Users subcommand options and usage
```

This example showcases CommandKit's complete command system capabilities for production CLI applications.
