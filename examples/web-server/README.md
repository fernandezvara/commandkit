# Web Server Example

A production-ready web server demonstrating CommandKit's **empty string command** approach for configuration-only applications.

## Features Demonstrated

- **Empty string command** - Default action when no command provided
- **Configuration-only mode** - No subcommands, pure configuration processing
- **Type-safe configuration** - Compile-time guarantees with generics
- **Multiple sources** - Environment variables, flags, defaults
- **Secret management** - Memory-protected sensitive values
- **Help generation** - Automatic help for configuration options
- **Validation** - Type checking, ranges, required fields
- **Professional error handling** - Clear error messages and suggestions

## 🎯 Empty String Command Magic

This example uses `cfg.Command("")` to create a **default command** that executes when no command is provided:

```bash
# These all execute the empty string command:
./web-server                    # Runs with defaults
./web-server --port 9000        # Runs with port 9000
./web-server --help             # Shows help for default command
DATABASE_URL=... ./web-server  # Runs with environment variable
```

The empty string command has full access to:
- ✅ Type-safe configuration access
- ✅ Environment variable support
- ✅ Flag parsing and validation
- ✅ Secret management
- ✅ Help generation
- ✅ Error handling

## Usage

### Basic Usage
```bash
# Run with defaults (empty string command executes)
go run main.go

# Set environment variables
PORT=3000 LOG_LEVEL=debug go run main.go

# Use command-line flags
go run main.go --port 3000 --host 0.0.0.0 --log-level debug

# Get help for the default command
go run main.go --help

# Get full help with all options
go run main.go --full-help
```

### With Secrets
```bash
# Required secrets and URLs
DATABASE_URL="postgres://user:pass@localhost/db" \
JWT_SIGNING_KEY="your-32-character-secret-key-here" \
go run main.go
```

## Configuration Options

| Option | Type | Sources | Default | Description |
|--------|------|---------|---------|-------------|
| PORT | int64 | flag, env, default | 8080 | HTTP server port |
| HOST | string | flag, env, default | localhost | Server host |
| LOG_LEVEL | string | flag, env, default | info | Logging level (debug, info, warn, error) |
| DATABASE_URL | string | env | optional, secret | Database connection URL |
| JWT_SIGNING_KEY | string | env | optional, secret | JWT signing key |

## Code Structure

The empty string command pattern:

```go
// Add empty string command for config-only mode
cfg.Command("").
    Func(func(ctx *commandkit.CommandContext) error {
        // Type-safe access to configuration
        port, _ := commandkit.Get[int64](ctx, "PORT")
        host, _ := commandkit.Get[string](ctx, "HOST")
        logLevel, _ := commandkit.Get[string](ctx, "LOG_LEVEL")
        
        fmt.Printf("🚀 Web Server Starting!\n")
        fmt.Printf("   Port: %d\n", port)
        fmt.Printf("   Host: %s\n", host)
        fmt.Printf("   Log Level: %s\n", logLevel)
        
        // Check for secrets
        if dbSecret := ctx.GlobalConfig.GetSecret("DATABASE_URL"); dbSecret.IsSet() {
            fmt.Printf("   Database: %s\n", maskSecret(dbSecret.String()))
        }
        
        return nil
    }).
    ShortHelp("Start the web server").
    LongHelp("Starts the web server with the specified configuration.").
    Config(func(cc *commandkit.CommandConfig) {
        // Add configuration to the default command
        cc.Define("PORT").Int64().Env("PORT").Flag("port").Default(8080)
        cc.Define("HOST").String().Env("HOST").Flag("host").Default("localhost")
        // ... more configuration
    })
```

## Error Handling

The example uses the unified `cfg.Execute()` API with professional error display:

```bash
# Missing required field (if any were required)
go run main.go
# Shows: Configuration errors with detailed help

# Validation errors
PORT=99999 go run main.go  
# Shows: --port int64 -> value 99999 is greater than maximum 65535
```

## Secret Management

Sensitive values are automatically memory-protected:

```go
// Access secrets safely in the command function
if dbSecret := ctx.GlobalConfig.GetSecret("DATABASE_URL"); dbSecret.IsSet() {
    fmt.Printf("Database: %s\n", maskSecret(dbSecret.String()))
}

// Helper function to mask secrets in output
func maskSecret(secret string) string {
    if len(secret) <= 8 {
        return strings.Repeat("*", len(secret))
    }
    return secret[:4] + strings.Repeat("*", len(secret)-8) + secret[len(secret)-4:]
}
```

## Help System

Automatic help generation for the empty string command:

```bash
# Basic help
go run main.go --help
# Shows: Usage, description, flags, environment variables

# Full help with all options
go run main.go --full-help  
# Shows: All available options including optional ones
```

## Key Benefits of Empty String Command

✅ **Zero Boilerplate** - No separate APIs needed for config-only apps  
✅ **Type Safety** - Compile-time guarantees with generics  
✅ **Natural Usage** - `./app` just works, `./app --port 9000` works too  
✅ **Full Features** - Help, validation, secrets, everything works  
✅ **Consistent** - Same patterns as command-based CLIs  

This example showcases how CommandKit's empty string command creates elegant configuration-only applications with zero compromise on features or developer experience.
