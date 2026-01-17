# CommandKit

A command and type-safe configuration library for Go with support for environment variables, command-line flags, configuration files, and a full command system.

## Features

- **Fluent/chainable API** for defining configuration
- **Type-safe** with generics for retrieval
- **Multiple sources**: Config Files → Flags → Environment Variables → Defaults
- **Rich validation**: Required, ranges, regex, oneOf, URL, min/max length
- **Secret protection**: Sensitive values protected with memguard
- **Clear error messages**: Beautiful formatted output for missing/invalid configs
- **Command system**: Define commands with subcommands, aliases, and help
- **Configuration files**: Support for JSON, YAML, TOML
- **Command middleware**: Add middleware for logging, authentication, and common functionality

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
    "time"

    "github.com/fernandezvara/commandkit"
)

func main() {
    cfg := commandkit.New()

    // Define configuration
    cfg.Define("PORT").Int64().Env("PORT").Flag("port").Default(int64(8080)).Range(1, 65535)
    cfg.Define("DATABASE_URL").String().Env("DATABASE_URL").Required().Secret()
    cfg.Define("LOG_LEVEL").String().Env("LOG_LEVEL").Default("info").OneOf("debug", "info", "warn", "error")
    cfg.Define("CORS_ORIGINS").StringSlice().Env("CORS_ORIGINS").Delimiter(",").Default([]string{"http://localhost:3000"})

    // Process configuration
    if errs := cfg.Process(); len(errs) > 0 {
        cfg.PrintErrors(errs)
        os.Exit(1)
    }
    defer cfg.Destroy()

    // Use with type safety
    port := commandkit.Get[int64](cfg, "PORT")
    logLevel := commandkit.Get[string](cfg, "LOG_LEVEL")

    fmt.Printf("Server starting on port %d with log level %s\n", port, logLevel)

    // Access secrets safely
    dbURL := cfg.GetSecret("DATABASE_URL")
    fmt.Printf("Database URL size: %d bytes\n", dbURL.Size())
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

Sources are resolved in this priority order (highest to lowest):

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
cfg.Define("PORT").Int64().Range(1, 65535)
cfg.Define("NAME").String().Required().MinLength(3).MaxLength(50)
cfg.Define("EMAIL").String().Regexp(`^[a-z]+@[a-z]+\.[a-z]+$`)
cfg.Define("ENV").String().OneOf("dev", "staging", "prod")
cfg.Define("TIMEOUT").Duration().DurationRange(1*time.Second, 5*time.Minute)
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

Secrets are protected in memory using memguard:

```go
cfg.Define("API_KEY").String().Required().Secret()

// Process configuration
cfg.Process()

// Access secrets (never use Get[] for secrets)
secret := cfg.GetSecret("API_KEY")
keyBytes := secret.Bytes()  // Use the secret
defer cfg.Destroy()         // Clean up all secrets
```

## Command System

```go
cfg := commandkit.New()

// Global configuration
cfg.Define("VERBOSE").Bool().Flag("verbose").Default(false)

// Define commands
cfg.Command("start").
    Func(startCommand).
    ShortHelp("Start the service").
    Aliases("s", "run").
    Config(func(cc *commandkit.CommandConfig) {
        cc.Define("PORT").Int64().Flag("port").Default(int64(8080))
    }).
    SubCommand("worker").
        Func(startWorkerCommand).
        ShortHelp("Start worker processes")

cfg.Command("stop").
    Func(stopCommand).
    ShortHelp("Stop the service")

// Execute
if err := cfg.Execute(os.Args); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}

func startCommand(ctx *commandkit.CommandContext) error {
    port := commandkit.Get[int64](ctx.Config, "PORT")
    fmt.Printf("Starting on port %d\n", port)
    return nil
}
```

**Usage:**

```bash
myapp start --port 3000
myapp s --port 3000        # alias
myapp start worker         # subcommand
myapp help start           # command help
```

## Middleware

CommandKit provides a powerful middleware system for adding cross-cutting functionality to commands like logging, authentication, error handling, and metrics.

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
// Token-based authentication
cfg.UseMiddlewareForCommands([]string{"admin", "shutdown"},
    commandkit.TokenAuthMiddleware("ADMIN_TOKEN"))

// Custom authentication
cfg.UseMiddleware(commandkit.AuthMiddleware(func(ctx *commandkit.CommandContext) error {
    token := ctx.Config.GetString("AUTH_TOKEN")
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

#### Timing Middleware

```go
// Measure execution time and store in context
cfg.UseMiddleware(commandkit.TimingMiddleware())
```

#### Rate Limiting Middleware

```go
// Limit command execution rate
cfg.UseMiddlewareForCommands([]string{"api", "status"},
    commandkit.RateLimitMiddleware(5, time.Minute))
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

Applied to all commands:

```go
cfg.UseMiddleware(commandkit.RecoveryMiddleware())
cfg.UseMiddleware(commandkit.DefaultLoggingMiddleware())
cfg.UseMiddleware(commandkit.DefaultErrorHandlingMiddleware())
```

#### Command-Specific Middleware

Applied only to specific commands:

```go
cfg.UseMiddlewareForCommands([]string{"admin", "shutdown"},
    commandkit.TokenAuthMiddleware("ADMIN_TOKEN"))
```

#### Subcommand-Specific Middleware

Applied only to specific subcommands:

```go
cfg.UseMiddlewareForSubcommands("admin", []string{"users", "shutdown"},
    commandkit.AdminOnlyMiddleware("ADMIN_TOKEN"))
```

#### Command-Level Middleware

Applied to a specific command during definition:

```go
cfg.Command("deploy").
    Func(deployCommand).
    Middleware(commandkit.TimingMiddleware()).
    Middleware(commandkit.RecoveryMiddleware())
```

#### Conditional Middleware

Applied based on conditions:

```go
cfg.UseMiddleware(commandkit.ConditionalMiddleware(
    func(ctx *commandkit.CommandContext) bool {
        return ctx.Command == "admin"
    },
    commandkit.AuthMiddleware(adminAuthFunc),
))
```

### Context Sharing

Middleware can share data through the command context:

```go
// Authentication middleware stores token
func TokenAuthMiddleware(tokenKey string) CommandMiddleware {
    return AuthMiddleware(func(ctx *CommandContext) error {
        token := ctx.Config.GetString(tokenKey)
        ctx.Set("auth_token", token) // Store in context
        return nil
    })
}

// Other middleware can access the token
func LoggingMiddleware(next CommandFunc) CommandFunc {
    return func(ctx *CommandContext) error {
        if token, exists := ctx.Get("auth_token"); exists {
            log.Printf("Command executed with token: %s", token)
        }
        return next(ctx)
    }
}
```

### Execution Order

Middleware executes in registration order:

```
1. Global Middleware (in order of registration)
2. Command-Specific Middleware (if applicable)
3. Command Function
4. Middleware unwinds in reverse order
```

### Custom Middleware

Create custom middleware by implementing the `CommandMiddleware` type:

```go
type CommandMiddleware func(next CommandFunc) CommandFunc

// Custom middleware example
func CustomMiddleware(next CommandFunc) CommandFunc {
    return func(ctx *CommandContext) error {
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

## Error Output

When configuration has errors, CommandKit displays them clearly:

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

## API Reference

### Config Methods

| Method              | Description                               |
| ------------------- | ----------------------------------------- |
| `New()`             | Create new Config instance                |
| `Define(key)`       | Start defining a configuration key        |
| `Process()`         | Process all definitions, returns errors   |
| `PrintErrors(errs)` | Print formatted errors to stderr          |
| `Destroy()`         | Clean up all secrets from memory          |
| `Dump()`            | Return all config values (secrets masked) |
| `GenerateHelp()`    | Generate help text for all options        |
| `Has(key)`          | Check if key exists and has value         |
| `Keys()`            | Return all defined keys                   |
| `GetSecret(key)`    | Get a secret value                        |

### Generic Getters

| Method                        | Description                             |
| ----------------------------- | --------------------------------------- |
| `Get[T](cfg, key)`            | Get value with type T (panics on error) |
| `GetOr[T](cfg, key, default)` | Get value or return default             |
| `MustGet[T](cfg, key)`        | Alias for Get                           |

### Convenience Getters

| Method                    | Description            |
| ------------------------- | ---------------------- |
| `cfg.GetString(key)`      | Get string value       |
| `cfg.GetInt64(key)`       | Get int64 value        |
| `cfg.GetFloat64(key)`     | Get float64 value      |
| `cfg.GetBool(key)`        | Get bool value         |
| `cfg.GetDuration(key)`    | Get duration value     |
| `cfg.GetStringSlice(key)` | Get string slice value |
| `cfg.GetInt64Slice(key)`  | Get int64 slice value  |

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

## License

MIT License
