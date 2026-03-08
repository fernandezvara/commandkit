# Web Server Example

A complete production-ready web server demonstrating CommandKit's configuration-only mode with comprehensive features.

## Features Demonstrated

- **Configuration-only mode** - No commands, pure configuration processing
- **Comprehensive validation** - Ports, URLs, secrets, durations, ranges
- **Multiple sources** - Environment variables, flags, files, defaults
- **Source priority** - Configurable priority ordering
- **Secret protection** - Memory-protected sensitive values
- **Builder patterns** - Fluent API with cloning for DRY code
- **File-based configuration** - Environment-specific JSON configs
- **Professional error handling** - Unified cfg.Execute() API

## Usage

### Basic Usage
```bash
# Run with defaults
go run main.go

# Set environment variables
PORT=3000 BASE_URL=http://localhost:3000 go run main.go

# Use command-line flags
go run main.go --port 3000 --base-url http://localhost:3000

# Use configuration file
ENVIRONMENT=production go run main.go
```

### Environment-Specific Configuration

The example automatically loads configuration from `config/{ENVIRONMENT}.json`:

```bash
# Development (default)
go run main.go

# Production
ENVIRONMENT=production go run main.go

# Test
ENVIRONMENT=test go run main.go
```

### Required Configuration

These must be provided via environment variables or flags:

```bash
# Required secrets and URLs
DATABASE_URL="postgres://user:pass@localhost/db" \
JWT_SIGNING_KEY="your-32-character-secret-key-here" \
go run main.go
```

## Configuration Options

| Option | Type | Sources | Default | Description |
|--------|------|---------|---------|-------------|
| PORT | int64 | flag, env, file, default | 8080 | HTTP server port (1-65535) |
| HOST | string | flag, env, file, default | localhost | Server host |
| BASE_URL | string | flag, env, file | required | Public base URL |
| DATABASE_URL | string | env | required, secret | Database connection |
| REDIS_URL | string | env | optional, secret | Redis connection |
| LOG_LEVEL | string | flag, env, file, default | info | Logging level |
| ACCESS_TOKEN_TTL | duration | env, file, default | 15m0s | Token lifetime |
| CORS_ORIGINS | []string | flag, env, file, default | localhost:3000 | Allowed origins |
| JWT_SIGNING_KEY | string | env | required, secret | JWT signing key |
| ENVIRONMENT | string | flag, env, file, default | development | App environment |
| MAX_CONNECTIONS | int64 | env, file, default | 100 | Max DB connections |
| ENABLE_METRICS | bool | flag, env, file, default | true | Enable metrics |

## Builder Pattern Examples

The example demonstrates CommandKit's builder pattern with cloning:

```go
// Base timeout configuration
baseTimeoutConfig := cfg.Define("READ_TIMEOUT").
    Duration().
    Default(30 * time.Second).
    MinDuration(1 * time.Second).
    Description("Read timeout")

// Clone and customize for similar configurations
baseTimeoutConfig.Clone().
    Env("WRITE_TIMEOUT").
    Default(60 * time.Second).
    Description("Write timeout")

baseTimeoutConfig.Clone().
    Env("IDLE_TIMEOUT").
    Default(120 * time.Second).
    Description("Idle timeout")
```

## File Configuration

Environment-specific JSON files in `config/` directory:

- `config/development.json` - Development settings
- `config/production.json` - Production settings  
- `config/test.json` - Test settings

Files are automatically loaded based on the `ENVIRONMENT` variable.

## Error Handling

The example uses the unified `cfg.Execute()` API with professional error display:

```bash
# Missing required field
go run main.go
# Shows: Configuration errors with detailed help

# Validation errors
PORT=99999 go run main.go  
# Shows: --port int64 -> value 99999 is greater than maximum 65535
```

## Secret Protection

Sensitive values are automatically memory-protected:

```go
// Access secrets safely
dbURL := cfg.GetSecret("DATABASE_URL")
if dbURL.IsSet() {
    fmt.Printf("Database configured (%d bytes)\n", dbURL.Size())
}
```

This example showcases CommandKit's complete feature set for production applications.
