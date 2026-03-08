# CommandKit

A production-ready Go CLI framework with type-safe configuration, powerful command system, and comprehensive middleware support.

## Features

- **Type-Safe Configuration**: Fluent API with generics for compile-time safety
- **Multiple Sources**: Environment variables, flags, defaults with flexible priority ordering
- **Rich Validation**: Built-in validators with custom validation support
- **Secret Protection**: Memory-protected secrets with automatic cleanup
- **Command System**: Subcommands, aliases, and hierarchical organization
- **Middleware Pipeline**: Logging, auth, metrics, and custom middleware
- **Professional Help**: Auto-generated help with validation context
- **Error Handling**: Clear, actionable error messages with templated display

## Installation

```bash
go get github.com/fernandezvara/commandkit
```

## Quick Start

### Configuration-Only Applications

```go
package main

import (
    "fmt"
    "os"
    "time"

    "github.com/fernandezvara/commandkit"
)

func main() {
    cfg := commandkit.New()

    // Define configuration
    cfg.Define("PORT").
        Int64().
        Env("PORT").
        Flag("port").
        Default(int64(8080)).
        Range(1, 65535).
        Description("HTTP server port")

    cfg.Define("DATABASE_URL").
        String().
        Env("DATABASE_URL").
        Required().
        Secret().
        Description("Database connection string")

    cfg.Define("LOG_LEVEL").
        String().
        Env("LOG_LEVEL").
        Flag("log-level").
        Default("info").
        OneOf("debug", "info", "warn", "error").
        Description("Logging level")

    // Execute configuration
    if err := cfg.Execute(os.Args); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    defer cfg.Destroy()

    // Create command context for accessing configuration
    ctx := commandkit.NewCommandContext([]string{}, cfg, "app", "")

    // Use configuration with type safety
    port, err := commandkit.Get[int64](ctx, "PORT")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    logLevel, err := commandkit.Get[string](ctx, "LOG_LEVEL")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    // Or use MustGet for critical values (panics on error)
    debug := commandkit.MustGet[bool](ctx, "DEBUG")

    // Access secrets safely
    dbURL := cfg.GetSecret("DATABASE_URL")
    if dbURL.IsSet() {
        fmt.Printf("Database connected (%d bytes)\n", dbURL.Size())
    }

    fmt.Printf("Server starting on port %d with log level %s\n", port, logLevel)
}
```

### Command-Based Applications

```go
package main

import (
    "fmt"
    "os"

    "github.com/fernandezvara/commandkit"
)

func main() {
    cfg := commandkit.New()

    // Global configuration
    cfg.Define("VERBOSE").
        Bool().
        Env("VERBOSE").
        Flag("verbose").
        Default(false).
        Description("Enable verbose logging")

    // Start command
    cfg.Command("start").
        Func(startCommand).
        ShortHelp("Start the service").
        LongHelp("Start the service with all components initialized.").
        Config(func(cc *commandkit.CommandConfig) {
            cc.Define("PORT").
                Int64().
                Flag("port").
                Default(int64(8080)).
                Range(1, 65535).
                Description("HTTP server port")

            cc.Define("DAEMON").
                Bool().
                Flag("daemon").
                Default(false).
                Description("Run in background")
        })

    // Execute commands
    if err := cfg.Execute(os.Args); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func startCommand(ctx *commandkit.CommandContext) error {
    // Get configuration values
    port, err := commandkit.Get[int64](ctx, "PORT")
    if err != nil {
        return fmt.Errorf("failed to get PORT: %w", err)
    }

    daemon, err := commandkit.Get[bool](ctx, "DAEMON")
    if err != nil {
        return fmt.Errorf("failed to get DAEMON: %w", err)
    }

    fmt.Printf("=== Starting Service ===\n")
    fmt.Printf("Port: %d\n", port)
    fmt.Printf("Daemon mode: %v\n", daemon)
    
    return nil
}
```

## File Configuration

CommandKit supports loading configuration from JSON, YAML, and TOML files with flexible key mapping.

### Basic File Loading

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/fernandezvara/commandkit"
)

func main() {
    cfg := commandkit.New()
    
    // Define configuration with file key mapping
    cfg.Define("PORT").
        Int64().
        Env("PORT").
        Flag("port").
        File("port_in_file").  // Look for "port_in_file" key in files
        Default(8080).
        Description("HTTP server port")
    
    cfg.Define("DATABASE_URL").
        String().
        File("db_connection_string").
        Required().
        Secret().
        Description("Database connection string")
    
    // Load configuration file
    cfg.LoadFile("config.json")
    
    // Execute configuration
    if err := cfg.Execute(os.Args); err != nil {
        os.Exit(1)
    }
    
    // Use configuration
    ctx := commandkit.NewCommandContext([]string{}, cfg, "app", "")
    port, _ := commandkit.Get[int64](ctx, "PORT")
    fmt.Printf("Server will start on port %d\n", port)
}
```

### Configuration File Format

```json
{
  "port_in_file": 3000,
  "db_connection_string": "postgres://localhost/mydb",
  "log_level": "debug"
}
```

### Loading Files from Environment Variables

```go
// Load file from environment variable containing the path
cfg.LoadFileFromEnv("CONFIG_FILE")
```

Environment variable should contain the file path:
```bash
export CONFIG_FILE="/etc/myapp/config.json"
```

### Multiple Files

```go
// Load multiple files (later files override earlier ones)
cfg.LoadFiles("config.json", "secrets.json")
```

### Priority System

File configuration integrates seamlessly with the existing priority system:

```go
cfg.Define("PORT").
    Int64().
    Env("PORT").
    Flag("port").
    File("port_in_file").  // File source
    Default(8080)         // Default source

// Priority: Flag > Env > File > Default (configurable)
```

## Configuration Types

### Basic Types

```go
cfg.Define("PORT").Int64().Default(8080)
cfg.Define("ENABLED").Bool().Default(true)
cfg.Define("NAME").String().Default("app")
cfg.Define("RATE").Float64().Default(100.0)
cfg.Define("TAGS").StringSlice().Default([]string{"v1", "api"})
cfg.Define("TIMEOUT").Duration().Default(30 * time.Second)
```

### Validation

```go
cfg.Define("PORT").
    Int64().
    Range(1, 65535).           // Value range
    Required()                   // Required field

cfg.Define("EMAIL").
    String().
    Regex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`). // Email format
    MinLength(5).               // Minimum length
    MaxLength(100).             // Maximum length

cfg.Define("LOG_LEVEL").
    String().
    OneOf("debug", "info", "warn", "error"). // Enum validation
    Default("info")
```

### Environment and Flags

```go
cfg.Define("DATABASE_URL").
    String().
    Env("DATABASE_URL").        // Environment variable
    Flag("db-url").             // Command-line flag
    Required().
    Secret().                   // Memory-protected secret
    Description("Database connection")

cfg.Define("PORT").
    Int64().
    Env("PORT").
    Flag("port").
    Default(8080).
    Description("Server port")
```

## Accessing Configuration

### Get[T] - Safe Access

```go
// Returns (T, error) - handle errors appropriately
port, err := commandkit.Get[int64](ctx, "PORT")
if err != nil {
    return fmt.Errorf("configuration error: %w", err)
}
fmt.Printf("Port: %d\n", port)
```

### MustGet[T] - Critical Values

```go
// Panics on error - use for values that must exist
port := commandkit.MustGet[int64](ctx, "PORT")
fmt.Printf("Port: %d\n", port)
```

### Secret Access

```go
// Access secrets safely with masking
secret := cfg.GetSecret("API_KEY")
if secret.IsSet() {
    fmt.Printf("API key configured (%d bytes)\n", secret.Size())
    // Use secret.Value() for the actual value when needed
}
```

## Command System

### Basic Commands

```go
cfg.Command("start").
    Func(startCommand).
    ShortHelp("Start the service").
    LongHelp("Detailed description of start command...")

func startCommand(ctx *commandkit.CommandContext) error {
    // Command implementation
    return nil
}
```

### Subcommands

```go
cfg.Command("docker").
    ShortHelp("Docker operations").
    Config(func(cc *commandkit.CommandConfig) {
        cc.Command("run").
            Func(dockerRunCommand).
            ShortHelp("Run Docker container")
        
        cc.Command("stop").
            Func(dockerStopCommand).
            ShortHelp("Stop Docker container")
    })
```

### Aliases

```go
cfg.Command("start").
    Func(startCommand).
    ShortHelp("Start the service").
    Aliases("run", "up")  // Multiple aliases
```

## Middleware

### Global Middleware

```go
// Apply to all commands
cfg.UseMiddleware(commandkit.RecoveryMiddleware())
cfg.UseMiddleware(commandkit.DefaultLoggingMiddleware())
cfg.UseMiddleware(commandkit.DefaultMetricsMiddleware())
```

### Command-Specific Middleware

```go
// Apply only to specific commands
cfg.UseMiddlewareForCommands([]string{"admin", "shutdown"},
    commandkit.TokenAuthMiddleware("ADMIN_TOKEN"))
```

### Custom Middleware

```go
cfg.UseMiddleware(func(ctx *commandkit.CommandContext, next commandkit.HandlerFunc) error {
    start := time.Now()
    err := next(ctx)
    duration := time.Since(start)
    
    fmt.Printf("Command %s took %v\n", ctx.Command, duration)
    return err
})
```

## Error Handling

CommandKit provides clear, actionable error messages with templated display:

```
Usage: myapp [options]

Configuration errors:
  --port int64 (default: 8080) -> value 99999 is greater than maximum 65535

Options:
  --port int64 (default: 8080) (valid: 1-65535)
        HTTP server port
  --database-url string (required)
        Database connection string
```

## Help System

CommandKit automatically generates professional help:

### Global Help
```bash
$ go run myapp --help
Usage: myapp <command> [options]

Available commands:
  start        Start the service (aliases: run, up)
  stop         Stop the service
  config       Show configuration

Use 'myapp <command> --help' for command-specific help
```

### Command Help
```bash
$ go run myapp start --help
Usage: start [options]

Start the service with all components initialized.

Options:
  --port int64 (default: 8080) (valid: 1-65535)
        HTTP server port
  --daemon bool
        Run in background
```

## Examples

CommandKit includes two comprehensive examples that demonstrate all major features:

### Web Server Example
**Location:** `examples/web-server/`

A complete production-ready web server demonstrating configuration-only mode:

```bash
cd examples/web-server

# Run with defaults
go run main.go

# Set environment variables
DATABASE_URL="postgres://user:pass@localhost/db" \
JWT_SIGNING_KEY="your-32-character-secret-key-here" \
go run main.go

# Use configuration file
ENVIRONMENT=production go run main.go
```

**Features demonstrated:**
- Configuration-only mode (no commands)
- Comprehensive validation (ports, URLs, secrets, durations)
- Multiple sources (env, flags, files, defaults)
- Source priority and builder patterns
- Secret protection and file-based configuration
- Professional error handling

### CLI Tool Example  
**Location:** `examples/cli-tool/`

A full-featured command-line application with middleware and commands:

```bash
cd examples/cli-tool

# Show help
go run main.go help

# Deploy application
go run main.go deploy --env staging --dry-run=true

# Show system status
go run main.go status --detailed=true

# Manage configuration
go run main.go config --show-secrets=true
```

**Features demonstrated:**
- Command system with aliases and subcommands
- Middleware pipeline (logging, timing, metrics, recovery)
- Command-specific configuration and validation
- Professional help system
- Rate limiting and error handling
- Unified cfg.Execute() API

Both examples use the current unified API and demonstrate production-ready patterns for building applications with CommandKit.

## API Reference

### Configuration Methods

| Method | Description |
| ------ | ----------- |
| `Define(key)` | Start defining a configuration key |
| `String()`, `Int64()`, `Bool()`, etc. | Set value type |
| `Env(name)` | Set environment variable name |
| `Flag(name)` | Set command-line flag name |
| `Default(value)` | Set default value |
| `Required()` | Mark as required |
| `Secret()` | Mark as secret (memguard protected) |
| `Description(text)` | Set description for help |
| `Range(min, max)` | Set numeric range validation |
| `OneOf(values...)` | Set enum validation |
| `Regex(pattern)` | Set regex validation |
| `MinLength(n)` | Set minimum string length |
| `MaxLength(n)` | Set maximum string length |

### Access Methods

| Method | Description |
| ------ | ----------- |
| `Get[T](ctx, key)` | Get value with type T (returns T, error) |
| `MustGet[T](ctx, key)` | Get value or panic on error |
| `GetSecret(key)` | Get a secret value |
| `Execute(args)` | Execute with command routing |

### Command Methods

| Method | Description |
| ------ | ----------- |
| `Command(name)` | Define a new command |
| `Func(fn)` | Set command function |
| `ShortHelp(text)` | Set short help text |
| `LongHelp(text)` | Set long help text |
| `Aliases(names...)` | Set command aliases |
| `Config(fn)` | Define command-specific config |
| `UseMiddleware(fn)` | Add middleware |

## License

MIT License
