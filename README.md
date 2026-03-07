# CommandKit

A production-ready Go CLI framework with type-safe configuration, powerful command system, and comprehensive middleware support.

## Features

- **Type-Safe Configuration**: Fluent API with generics for compile-time safety
- **Multiple Sources**: Config files → Flags → Environment → Defaults
- **Rich Validation**: Built-in validators with custom validation support
- **Secret Protection**: Memory-protected secrets with automatic cleanup
- **Command System**: Subcommands, aliases, and hierarchical organization
- **Middleware Pipeline**: Logging, auth, metrics, and custom middleware
- **Error Handling**: Unified error collection and helpful messages
- **Help Generation**: Automatic help with validation context

## Installation

```bash
go get github.com/fernandezvara/commandkit
```

## Quick Start

```go
package main

import (
    "fmt"
    "os"

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

    // Process configuration
    result := cfg.Process()
    if result.Error != nil {
        fmt.Fprintf(os.Stderr, "Configuration error: %v\n", result.Error)
        os.Exit(1)
    }
    defer cfg.Destroy()

    // Create command context
    ctx := commandkit.NewCommandContext([]string{}, cfg, "app", "")

    // Use configuration with type safety
    portResult := commandkit.Get[int64](ctx, "PORT")
    if portResult.Error != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", portResult.Error)
        os.Exit(1)
    }
    port := commandkit.GetValue[int64](portResult)

    logLevelResult := commandkit.Get[string](ctx, "LOG_LEVEL")
    if logLevelResult.Error != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", logLevelResult.Error)
        os.Exit(1)
    }
    logLevel := commandkit.GetValue[string](logLevelResult)

    // Access secrets safely
    dbURL := cfg.GetSecret("DATABASE_URL")
    if dbURL.IsSet() {
        fmt.Printf("Database connected (%d bytes)\n", dbURL.Size())
    }

    fmt.Printf("Server starting on port %d with log level %s\n", port, logLevel)
}
```

## Configuration Types

| Type         | Method           | Example                     |
| ------------ | ---------------- | --------------------------- |
| String       | `.String()`      | `"hello"`                   |
| Int64        | `.Int64()`       | `8080`                      |
| Float64      | `.Float64()`     | `3.14`                      |
| Bool         | `.Bool()`        | `true`, `false`             |
| Duration     | `.Duration()`    | `15m`, `1h`, `24h`          |
| URL          | `.URL()`         | `https://example.com`       |
| String Slice | `.StringSlice()` | `"a,b,c"` → `["a","b","c"]` |
| Int64 Slice  | `.Int64Slice()`  | `"1,2,3"` → `[1,2,3]`       |

## Configuration Sources

Sources are resolved in priority order (highest to lowest):

1. **Configuration files** (JSON, YAML, TOML)
2. **Command-line flags**
3. **Environment variables**
4. **Default values**

### Configuration Files

```go
cfg := commandkit.New()

// Load from file
cfg.LoadFile("config.yaml")

// Load multiple files (later files override earlier)
cfg.LoadFiles("base.yaml", "override.yaml")

// Load from environment variable
cfg.LoadFromEnv("CONFIG_FILE")

// Environment-specific configuration
cfg.SetEnvironment("production")
```

**config.yaml:**
```yaml
port: 8080
database_url: "postgresql://localhost/db"
log_level: "info"

environments:
  production:
    port: 80
    log_level: "warn"
```

## Validation

```go
// Numeric validation
cfg.Define("PORT").Int64().Range(1, 65535)

// String validation
cfg.Define("NAME").String().Required().MinLength(3).MaxLength(50)

// Pattern matching
cfg.Define("EMAIL").String().Regexp(`^[a-z]+@[a-z]+\.[a-z]+$`)

// Enum validation
cfg.Define("ENV").String().OneOf("dev", "staging", "prod")

// Duration validation
cfg.Define("TIMEOUT").Duration().DurationRange(1*time.Second, 5*time.Minute)

// Slice validation
cfg.Define("TAGS").StringSlice().MinItems(1).MaxItems(10)

// Custom validation
cfg.Define("CUSTOM").String().Custom("even-length", func(value any) error {
    if s, ok := value.(string); ok && len(s)%2 != 0 {
        return fmt.Errorf("string length must be even")
    }
    return nil
})
```

## Secrets

Secrets are protected in memory using memguard with automatic cleanup:

```go
cfg.Define("API_KEY").String().Required().Secret()

// Process configuration
result := cfg.Process()
if result.Error != nil {
    // Handle error
}
defer cfg.Destroy() // Clean up all secrets

// Access secrets safely
secret := cfg.GetSecret("API_KEY")
if secret.IsSet() {
    keyBytes := secret.Bytes() // Use the secret
    fmt.Printf("API key length: %d bytes\n", secret.Size())
}

// Note: Get[T]() will return an error for secret keys
// Use GetSecret() instead
```

### Secret Security Features

- **Memory Protection**: Secrets stored in encrypted memory buffers
- **Automatic Cleanup**: RAII pattern with finalizers
- **Access Control**: `Get[T]()` blocks secret access, use `GetSecret()`
- **Thread Safety**: All secret operations are thread-safe
- **Verification**: `VerifyDestroyed()` methods for cleanup confirmation

## Command System

```go
cfg := commandkit.New()

// Global configuration
cfg.Define("VERBOSE").Bool().Flag("verbose").Default(false)
cfg.Define("LOG_LEVEL").String().Flag("log-level").Default("info")

// Define commands
cfg.Command("start").
    Func(startCommand).
    ShortHelp("Start the service").
    LongHelp(`Start the service with all components initialized.
    
Usage: myapp start [options]

This will initialize the database, start HTTP server, and begin accepting connections.`).
    Aliases("s", "run", "up").
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
    }).
    SubCommand("server").
        Func(startServerCommand).
        ShortHelp("Start only the server").
        Aliases("srv").
        Config(func(cc *commandkit.CommandConfig) {
            cc.Define("WORKERS").
                Int64Slice().
                Flag("workers").
                Delimiter(",").
                Default([]int64{1}).
                Description("Number of worker processes")
        })

cfg.Command("stop").
    Func(stopCommand).
    ShortHelp("Stop the service gracefully").
    Aliases("quit", "exit").
    Config(func(cc *commandkit.CommandConfig) {
        cc.Define("TIMEOUT").
            Duration().
            Flag("timeout").
            Default(30*time.Second).
            Description("Graceful shutdown timeout")
    })

// Execute
if err := cfg.Execute(os.Args); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}

func startCommand(ctx *commandkit.CommandContext) error {
    // Process command-specific configuration
    var config *commandkit.Config
    if ctx.CommandConfig != nil {
        config = ctx.CommandConfig
    } else {
        config = ctx.GlobalConfig
    }
    if result := config.Process(); result.Error != nil {
        if result.Message != "" {
            fmt.Fprintln(os.Stderr, result.Message)
        }
        return fmt.Errorf("configuration errors")
    }

    // Get configuration values
    portResult := commandkit.Get[int64](ctx, "PORT")
    if portResult.Error != nil {
        return fmt.Errorf("failed to get PORT: %w", portResult.Error)
    }
    port := commandkit.GetValue[int64](portResult)

    daemonResult := commandkit.Get[bool](ctx, "DAEMON")
    if daemonResult.Error != nil {
        return fmt.Errorf("failed to get DAEMON: %w", daemonResult.Error)
    }
    daemon := commandkit.GetValue[bool](daemonResult)

    verboseResult := commandkit.Get[bool](ctx, "VERBOSE")
    if verboseResult.Error != nil {
        return fmt.Errorf("failed to get VERBOSE: %w", verboseResult.Error)
    }
    verbose := commandkit.GetValue[bool](verboseResult)

    fmt.Printf("Starting service on port %d\n", port)
    fmt.Printf("Daemon mode: %v\n", daemon)
    fmt.Printf("Verbose: %v\n", verbose)
    return nil
}
```

**Usage:**
```bash
myapp start --port 3000
myapp s --port 3000        # alias
myapp start server         # subcommand
myapp help start           # command help
```

## Builder Patterns

CommandKit provides powerful builder patterns for creating reusable configurations:

### DRY Configuration with Functions

```go
// Create reusable configuration function
baseServerConfig := func(cc *commandkit.CommandConfig) {
    cc.Define("PORT").
        Int64().
        Flag("port").
        Default(8080).
        Range(1, 65535).
        Description("Server port")

    cc.Define("HOST").
        String().
        Flag("host").
        Default("localhost").
        Description("Server host")

    cc.Define("VERBOSE").
        Bool().
        Flag("verbose").
        Default(false).
        Description("Enable verbose logging")
}

// Apply to multiple commands
cfg.Command("start").
    ShortHelp("Start the service").
    Config(func(cc *commandkit.CommandConfig) {
        baseServerConfig(cc)
        cc.Define("WORKERS").
            Int64().
            Flag("workers").
            Default(4).
            Description("Number of worker processes")
    })

cfg.Command("stop").
    ShortHelp("Stop the service").
    Config(func(cc *commandkit.CommandConfig) {
        baseServerConfig(cc)
        cc.Define("TIMEOUT").
            Duration().
            Flag("timeout").
            Default(30*time.Second).
            Description("Graceful shutdown timeout")
    })
```

### Builder Cloning for Variations

```go
// Create base configuration
basePortConfig := cfg.Define("PORT").
    Int64().
    Default(8080).
    Range(1, 65535).
    Description("Server port")

// Clone and customize for different environments
httpPortConfig := basePortConfig.Clone().
    Env("HTTP_PORT").
    Flag("http-port").
    Description("HTTP server port")

httpsPortConfig := basePortConfig.Clone().
    Env("HTTPS_PORT").
    Flag("https-port").
    Default(8443).
    Description("HTTPS server port")
```

## Middleware System

CommandKit provides a comprehensive middleware system for cross-cutting functionality:

### Built-in Middleware

#### Logging Middleware
```go
// Default logging with timing
cfg.UseMiddleware(commandkit.DefaultLoggingMiddleware())

// Custom logging
cfg.UseMiddleware(commandkit.LoggingMiddleware(func(ctx *commandkit.CommandContext, duration time.Duration) {
    log.Printf("Command %s completed in %v", ctx.Command, duration)
}))
```

#### Authentication Middleware
```go
// Token-based authentication for specific commands
cfg.UseMiddlewareForCommands([]string{"admin", "shutdown"},
    commandkit.TokenAuthMiddleware("ADMIN_TOKEN"))

// Custom authentication
cfg.UseMiddleware(commandkit.AuthMiddleware(func(ctx *commandkit.CommandContext) error {
    tokenResult := commandkit.Get[string](ctx, "AUTH_TOKEN")
    if tokenResult.Error != nil {
        return tokenResult.Error
    }
    token := commandkit.GetValue[string](tokenResult)
    if token != "secret-token" {
        return fmt.Errorf("invalid token")
    }
    return nil
}))
```

#### Error Handling Middleware
```go
// Default error handling with logging
cfg.UseMiddleware(commandkit.DefaultErrorHandlingMiddleware())

// Custom error handling with monitoring
cfg.UseMiddleware(commandkit.ErrorHandlingMiddleware(func(err error, ctx *commandkit.CommandContext) {
    // Send to monitoring system
    monitor.Error("command_failed", map[string]any{
        "command": ctx.Command,
        "error": err.Error(),
    })
}))
```

#### Recovery Middleware
```go
// Prevent panics from crashing the application
cfg.UseMiddleware(commandkit.RecoveryMiddleware())
```

#### Metrics Middleware
```go
// Collect command metrics
cfg.UseMiddleware(commandkit.DefaultMetricsMiddleware())

// Custom metrics collection
cfg.UseMiddleware(commandkit.MetricsMiddleware(func(ctx *commandkit.CommandContext, duration time.Duration, err error) {
    status := "success"
    if err != nil {
        status = "error"
    }
    metrics.Counter("command_executions", map[string]string{
        "command": ctx.Command,
        "status": status,
    }).Inc()
}))
```

### Middleware Patterns

#### Global Middleware
```go
cfg.UseMiddleware(commandkit.RecoveryMiddleware())
cfg.UseMiddleware(commandkit.DefaultLoggingMiddleware())
cfg.UseMiddleware(commandkit.DefaultErrorHandlingMiddleware())
```

#### Command-Specific Middleware
```go
cfg.UseMiddlewareForCommands([]string{"admin", "shutdown"},
    commandkit.TokenAuthMiddleware("ADMIN_TOKEN"))
```

#### Command-Level Middleware
```go
cfg.Command("deploy").
    Func(deployCommand).
    Middleware(commandkit.TimingMiddleware()).
    Middleware(commandkit.RecoveryMiddleware())
```

#### Conditional Middleware
```go
cfg.UseMiddleware(commandkit.ConditionalMiddleware(
    func(ctx *commandkit.CommandContext) bool {
        return ctx.Command == "admin"
    },
    commandkit.AuthMiddleware(adminAuthFunc),
))
```

### Custom Middleware

```go
type CommandMiddleware func(next CommandFunc) CommandFunc

// Custom middleware example
func CustomMiddleware(next CommandFunc) CommandFunc {
    return func(ctx *commandkit.CommandContext) error {
        // Pre-execution logic
        log.Printf("Starting command: %s", ctx.Command)

        // Execute next in chain
        err := next(ctx)

        // Post-execution logic
        log.Printf("Command %s completed with error: %v", ctx.Command, err)

        return err
    }
}
```

## Error Handling

CommandKit provides unified error handling throughout the framework:

### Configuration Access

```go
ctx := commandkit.NewCommandContext([]string{}, cfg, "app", "")

// Get configuration with error handling
portResult := commandkit.Get[int64](ctx, "PORT")
if portResult.Error != nil {
    return fmt.Errorf("failed to get PORT: %w", portResult.Error)
}
port := commandkit.GetValue[int64](portResult)

// Secret access (blocked for Get[T])
secretResult := commandkit.Get[string](ctx, "API_KEY")
if secretResult.Error != nil {
    // Expected: "validation error: configuration 'API_KEY' is secret, use GetSecret() instead"
}

// Correct secret access
secret := cfg.GetSecret("API_KEY")
if secret.IsSet() {
    // Use secret safely
}
```

### Error Messages

CommandKit provides clear, actionable error messages:

```
Configuration errors detected:
==================================================
ERROR: DATABASE_URL
  Error: required value not provided (set DATABASE_URL)
--------------------------------------------------
ERROR: PORT
  Source: env = 99999
  Error: value 99999 is greater than maximum 65535
==================================================
Total: 2 error(s)
```

### Configuration Processing

```go
result := cfg.Process()
if result.Error != nil {
    if result.Message != "" {
        fmt.Fprintln(os.Stderr, result.Message)
    }
    return fmt.Errorf("configuration processing failed")
}
```

## Help System

CommandKit provides a comprehensive help system with automatic generation and template-based formatting:

### Help Service Architecture

The help system is built around a clean, direct architecture:

```go
// Direct access to help functionality
helpService := commandkit.NewHelpService()

// Check if help is requested
if helpService.IsHelpRequested(args) {
    err := helpService.ShowHelp(args, commands)
    return err
}

// Generate help text (for testing or custom output)
helpText, err := helpService.GenerateHelp(args, commands)
```

### Automatic Help Detection

CommandKit automatically detects help requests in multiple formats:

```bash
myapp --help           # Global help
myapp -h              # Global help (short)
myapp help            # Global help (command)
myapp start --help    # Command-specific help
myapp start -h        # Command-specific help (short)
```

### Enhanced Flag Help

```bash
$ go run myapp start --help
Usage of start:
  -base-url string (required)
        Public base URL of the service
  
  -log-level string (env: LOG_LEVEL, required, default: 'info', oneOf: ['debug', 'info', 'warn', 'error'])
        Logging level
  
  -daemon bool (default: false)
        Run in background

  (no flag) string (env: DATABASE_URL, required, secret)
        Database connection string
```

### Command Help

```bash
$ go run myapp help start
Start the service with all components initialized.

Usage: myapp start [options]

This will initialize the database, start HTTP server, and begin accepting connections.
For production use, consider using the --daemon flag.

Options:
  -port int
        HTTP server port (default 8080, range 1-65535)
  -daemon
        Run in background (default false)
```

### Custom Help Templates

```go
helpService := commandkit.NewHelpService()
formatter := helpService.GetFormatter()

// Set custom templates for branding
if templateFormatter, ok := formatter.(*commandkit.TemplateHelpFormatter); ok {
    templateFormatter.SetTemplate(commandkit.TemplateGlobal, customTemplate)
    templateFormatter.SetTemplate(commandkit.TemplateCommand, commandTemplate)
}

// Add custom template functions
renderer := templateFormatter.GetRenderer()
renderer.AddFunction("reverse", func(s string) string {
    // Custom function implementation
})
```

### Help Output Control

```go
// Use string output for testing
stringOutput := commandkit.NewStringHelpOutput()
helpService.SetOutput(stringOutput)

// Generate help without displaying
text, err := helpService.GenerateHelp([]string{"--help"}, commands)
if err == nil {
    fmt.Println(text)
}

// Get accumulated output
output := stringOutput.Get()
```

## API Reference

### Config Methods

| Method              | Description                               |
| ------------------- | ----------------------------------------- |
| `New()`             | Create new Config instance                |
| `Define(key)`       | Start defining a configuration key        |
| `Process()`         | Process all definitions, returns CommandResult |
| `Destroy()`         | Clean up all secrets from memory          |
| `Dump()`            | Return all config values (secrets masked) |
| `Has(key)`          | Check if key exists (false for secrets)  |
| `HasSecret(key)`    | Check if secret key exists                |
| `GetSecret(key)`    | Get a secret value                        |
| `Command(name)`     | Define a new command                      |
| `UseMiddleware()`    | Add global middleware                      |

### Generic Getters

| Method                        | Description                             |
| ----------------------------- | --------------------------------------- |
| `Get[T](ctx, key)`            | Get value with type T (returns CommandResult) |
| `GetValue[T](result)`         | Extract value from CommandResult        |
| `GetOr[T](ctx, key, default)` | Get value or return default (returns CommandResult) |

### Definition Builder Methods

| Method                        | Description                         |
| ----------------------------- | ----------------------------------- |
| `.String()`, `.Int64()`, etc. | Set value type                      |
| `.Env(name)`                  | Set environment variable name       |
| `.Flag(name)`                 | Set command-line flag name          |
| `.Default(value)`             | Set default value                   |
| `.Required()`                 | Mark as required                    |
| `.Secret()`                   | Mark as secret (memguard protected) |
| `.Description(text)`          | Set description for help            |
| `.Clone()`                    | Create builder variation            |
| `.Delimiter(d)`               | Set delimiter for slices            |

### Validation Methods

| Method                                                           | Description         |
| ---------------------------------------------------------------- | ------------------- |
| `.Min(n)`, `.Max(n)`, `.Range(min, max)`                         | Numeric range       |
| `.MinLength(n)`, `.MaxLength(n)`, `.LengthRange(min, max)`       | String length       |
| `.Regexp(pattern)`                                               | Regex pattern match |
| `.OneOf(values...)`                                              | Enum validation     |
| `.URL()`                                                         | URL validation      |
| `.MinDuration(d)`, `.MaxDuration(d)`, `.DurationRange(min, max)` | Duration range      |
| `.MinItems(n)`, `.MaxItems(n)`, `.ItemsRange(min, max)`          | Slice item count    |
| `.Custom(name, func)`                                            | Custom validation   |

### Command Builder Methods

| Method                        | Description                         |
| ----------------------------- | ----------------------------------- |
| `.Func(fn)`                   | Set command function                |
| `.ShortHelp(text)`            | Set short help text                 |
| `.LongHelp(text)`             | Set long help text                  |
| `.Aliases(names...)`          | Set command aliases                 |
| `.Config(fn)`                 | Define command-specific config      |
| `.SubCommand(name)`           | Add subcommand                      |
| `.Middleware(mw)`             | Add command-level middleware         |
| `.Clone()`                    | Create builder variation            |

## Examples

The `examples/` directory contains comprehensive examples:

- `basic/` - Basic configuration usage
- `commands/` - Command system with subcommands
- `middleware/` - Middleware patterns
- `files/` - Configuration file loading
- `builder_clone/` - Builder patterns and DRY configuration

Run examples:
```bash
cd examples/basic
go run example_basic.go

cd examples/commands
go run example_commands.go start --port 3000
```

## License

MIT License
