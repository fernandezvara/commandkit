# ConfigKit: Shared Configuration Library

## Universal Configuration for All MicroSaaS Services

---

# 1. Overview

**ConfigKit** is a shared Go library that provides a consistent, type-safe configuration interface across all services. It supports environment variables, command-line flags, validation, secure secret handling with memguard, and a powerful command system with subcommands.

## Features

- **Fluent/chainable API** for defining configuration
- **Type-safe** with generics for retrieval
- **Multiple sources**: Config Files → Flags (priority) → Environment Variables → Defaults
- **Rich validation**: Required, ranges, regex, oneOf, URL, min/max length
- **Secret protection**: Sensitive values protected with memguard
- **Clear error messages**: Beautiful formatted output for missing/invalid configs
- **Command system**: Define commands with subcommands, aliases, and help
- **Command-specific config**: Each command can have its own configuration
- **Global + command config**: Shared global config with command overrides
- **Smart help**: Auto-generated help with suggestions for unknown commands
- **Configuration files**: Support for JSON, YAML, TOML configuration files
- **Command middleware**: Add middleware for logging, authentication, and common functionality

---

# 2. Installation

```bash
go get github.com/fernandezvara/commandkit
```

Dependencies:

```bash
go get github.com/awnumar/memguard
```

---

# 3. Quick Start

```go
package main

import (
    "fmt"
    "os"
    "time"

    "github.com/fernandezvara/commandkit"
)

func main() {
    cfg := configkit.New()

    // Server
    cfg.Define("PORT").Int64().Env("PORT").Flag("port").Default(int64(8080)).Range(1, 65535)
    cfg.Define("BASE_URL").String().Env("BASE_URL").Flag("base-url").Required().URL()

    // Database
    cfg.Define("DATABASE_URL").String().Env("DATABASE_URL").Required().Secret()

    // JWT
    cfg.Define("JWT_SIGNING_KEY").String().Env("JWT_SIGNING_KEY").Required().Secret().MinLength(32)
    cfg.Define("ACCESS_TOKEN_TTL").Duration().Env("ACCESS_TOKEN_TTL").Default(15 * time.Minute)

    // OAuth (optional)
    cfg.Define("GOOGLE_CLIENT_ID").String().Env("GOOGLE_CLIENT_ID")
    cfg.Define("GOOGLE_CLIENT_SECRET").String().Env("GOOGLE_CLIENT_SECRET").Secret()

    // Arrays
    cfg.Define("CORS_ORIGINS").StringSlice().Env("CORS_ORIGINS").Flag("cors-origins").
        Delimiter(",").Default([]string{"http://localhost:3000"})

    // Process
    if errs := cfg.Process(); len(errs) > 0 {
        cfg.PrintErrors(errs)
        os.Exit(1)
    }

    // Use with generics
    port := configkit.Get[int64](cfg, "PORT")
    baseURL := configkit.Get[string](cfg, "BASE_URL")
    corsOrigins := configkit.Get[[]string](cfg, "CORS_ORIGINS")
    tokenTTL := configkit.Get[time.Duration](cfg, "ACCESS_TOKEN_TTL")

    fmt.Printf("Server starting on port %d\n", port)
    fmt.Printf("Base URL: %s\n", baseURL)
    fmt.Printf("CORS Origins: %v\n", corsOrigins)
    fmt.Printf("Token TTL: %s\n", tokenTTL)

    // Access secrets safely
    jwtKey := cfg.GetSecret("JWT_SIGNING_KEY")
    defer jwtKey.Destroy() // Clean up when done

    // Use the secret
    keyBytes := jwtKey.Bytes()
    fmt.Printf("JWT Key length: %d\n", len(keyBytes))
}
```

---

# 4. Configuration Files

ConfigKit supports loading configuration from JSON, YAML, and TOML files. Configuration files have the highest priority in the source chain.

## File Support

```go
// Load from config file (highest priority)
cfg := configkit.New()
cfg.LoadFile("config.yaml")  // or .json, .toml

// Or specify multiple files (later files override earlier ones)
cfg.LoadFiles("base.yaml", "override.yaml", "local.yaml")

// Or load from environment variable
cfg.LoadFromEnv("CONFIG_FILE") // Loads file path from CONFIG_FILE env var
```

## Configuration File Formats

### YAML Example (config.yaml)

```yaml
# Server configuration
port: 8080
base_url: "https://api.example.com"

# Database
database_url: "postgresql://user:pass@localhost/db"

# JWT settings
jwt_signing_key: "your-secret-key-here"
access_token_ttl: "15m"

# Arrays
cors_origins:
  - "http://localhost:3000"
  - "https://app.example.com"

# Environment-specific overrides
environments:
  development:
    port: 3000
    log_level: "debug"
  production:
    port: 80
    log_level: "info"
```

### JSON Example (config.json)

```json
{
  "port": 8080,
  "base_url": "https://api.example.com",
  "database_url": "postgresql://user:pass@localhost/db",
  "jwt_signing_key": "your-secret-key-here",
  "access_token_ttl": "15m",
  "cors_origins": ["http://localhost:3000"],
  "environments": {
    "development": {
      "port": 3000,
      "log_level": "debug"
    },
    "production": {
      "port": 80,
      "log_level": "info"
    }
  }
}
```

### TOML Example (config.toml)

```toml
port = 8080
base_url = "https://api.example.com"
database_url = "postgresql://user:pass@localhost/db"
jwt_signing_key = "your-secret-key-here"
access_token_ttl = "15m"

cors_origins = ["http://localhost:3000"]

[environments]
[environments.development]
port = 3000
log_level = "debug"

[environments.production]
port = 80
log_level = "info"
```

## Environment-Specific Configuration

```go
// Set environment (loads environment-specific overrides)
cfg.SetEnvironment("production") // Loads from environments.production.*

// Or set from environment variable
cfg.SetEnvironmentFromEnv("APP_ENV") // Uses APP_ENV env var
```

## Configuration Priority with Files

1. **Config file** (highest priority)
2. **Command flags**
3. **Command environment variables**
4. **Global flags**
5. **Global environment variables**
6. **Command defaults**
7. **Global defaults** (lowest priority)

---

# 5. Command Middleware

ConfigKit supports middleware for commands, allowing you to add common functionality like logging, authentication, and error handling.

## Middleware Definition

```go
type CommandMiddleware func(next CommandFunc) CommandFunc
type CommandFunc func(*CommandContext) error
```

## Built-in Middleware

### Logging Middleware

```go
cfg.UseMiddleware(configkit.LoggingMiddleware(func(ctx *configkit.CommandContext, duration time.Duration) {
    fmt.Printf("Command %s completed in %v\n", ctx.Command, duration)
}))
```

### Authentication Middleware

```go
cfg.UseMiddleware(configkit.AuthMiddleware(func(ctx *configkit.CommandContext) error {
    token := ctx.Config.GetString("AUTH_TOKEN")
    if token == "" {
        return fmt.Errorf("authentication required")
    }
    return nil
}))
```

### Error Handling Middleware

```go
cfg.UseMiddleware(configkit.ErrorHandlingMiddleware(func(err error, ctx *configkit.CommandContext) {
    fmt.Printf("Command %s failed: %v\n", ctx.Command, err)
    // Send to monitoring system
}))
```

## Custom Middleware

```go
// Timing middleware
func TimingMiddleware(next configkit.CommandFunc) configkit.CommandFunc {
    return func(ctx *configkit.CommandContext) error {
        start := time.Now()
        err := next(ctx)
        duration := time.Since(start)

        // Log timing
        fmt.Printf("Command %s took %v\n", ctx.Command, duration)

        // Add timing to context for other middleware
        ctx.Set("duration", duration)

        return err
    }
}

// Usage
cfg.UseMiddleware(TimingMiddleware)
```

## Middleware Chain

```go
// Middleware are executed in registration order
cfg.UseMiddleware(loggingMiddleware)      // 1st
cfg.UseMiddleware(authMiddleware)         // 2nd
cfg.UseMiddleware(errorHandlingMiddleware) // 3rd

// Execution flow:
// 1. loggingMiddleware starts
// 2. authMiddleware checks authentication
// 3. errorHandlingMiddleware wraps error handling
// 4. Command function executes
// 5. Middleware unwind in reverse order
```

## Conditional Middleware

```go
// Apply middleware only to specific commands
cfg.UseMiddlewareForCommands([]string{"start", "stop"}, authMiddleware)

// Apply middleware only to commands with specific subcommands
cfg.UseMiddlewareForSubcommands("start", []string{"worker", "server"}, timingMiddleware)
```

## Global vs Command-Specific Middleware

```go
// Global middleware (applies to all commands)
cfg.UseMiddleware(globalLoggingMiddleware)

// Command-specific middleware
cfg.Command("start").
    Func(startCommand).
    UseMiddleware(startSpecificMiddleware).  // Only for start command
    Config(func(cc *configkit.CommandConfig) {
        cc.Define("PORT").Int64().Flag("port").Default(8080)
    })
```

---

# 6. Project Structure

```
configkit/
├── go.mod
├── config.go           # Main Config struct and Process()
├── definition.go       # Definition builder with fluent API
├── types.go            # Type constants and parsing
├── validation.go       # Validation rules
├── errors.go           # Error types and formatting
├── get.go              # Generic Get function
├── secret.go           # memguard wrapper for secrets
└── config_test.go      # Tests
```

---

# 5. Implementation

## go.mod

```go
module github.com/fernandezvara/commandkit

go 1.21

require github.com/awnumar/memguard v0.22.5
```

## types.go

```go
// configkit/types.go
package configkit

import (
    "fmt"
    "net/url"
    "strconv"
    "strings"
    "time"
)

// ValueType represents the expected type of a configuration value
type ValueType int

const (
    TypeString ValueType = iota
    TypeInt64
    TypeFloat64
    TypeBool
    TypeDuration
    TypeURL
    TypeStringSlice
    TypeInt64Slice
)

func (t ValueType) String() string {
    switch t {
    case TypeString:
        return "string"
    case TypeInt64:
        return "int64"
    case TypeFloat64:
        return "float64"
    case TypeBool:
        return "bool"
    case TypeDuration:
        return "duration"
    case TypeURL:
        return "url"
    case TypeStringSlice:
        return "[]string"
    case TypeInt64Slice:
        return "[]int64"
    default:
        return "unknown"
    }
}

// parseValue parses a string value into the expected type
func parseValue(raw string, valueType ValueType, delimiter string) (any, error) {
    if raw == "" {
        return nil, nil
    }

    switch valueType {
    case TypeString:
        return raw, nil

    case TypeInt64:
        v, err := strconv.ParseInt(raw, 10, 64)
        if err != nil {
            return nil, fmt.Errorf("invalid int64: %s", raw)
        }
        return v, nil

    case TypeFloat64:
        v, err := strconv.ParseFloat(raw, 64)
        if err != nil {
            return nil, fmt.Errorf("invalid float64: %s", raw)
        }
        return v, nil

    case TypeBool:
        v, err := strconv.ParseBool(raw)
        if err != nil {
            return nil, fmt.Errorf("invalid bool: %s (use true/false, 1/0, yes/no)", raw)
        }
        return v, nil

    case TypeDuration:
        v, err := time.ParseDuration(raw)
        if err != nil {
            return nil, fmt.Errorf("invalid duration: %s (use format like 15m, 1h, 7d)", raw)
        }
        return v, nil

    case TypeURL:
        v, err := url.Parse(raw)
        if err != nil {
            return nil, fmt.Errorf("invalid URL: %s", raw)
        }
        if v.Scheme == "" || v.Host == "" {
            return nil, fmt.Errorf("invalid URL (missing scheme or host): %s", raw)
        }
        return raw, nil // Store as string, validated

    case TypeStringSlice:
        if raw == "" {
            return []string{}, nil
        }
        parts := strings.Split(raw, delimiter)
        result := make([]string, 0, len(parts))
        for _, p := range parts {
            trimmed := strings.TrimSpace(p)
            if trimmed != "" {
                result = append(result, trimmed)
            }
        }
        return result, nil

    case TypeInt64Slice:
        if raw == "" {
            return []int64{}, nil
        }
        parts := strings.Split(raw, delimiter)
        result := make([]int64, 0, len(parts))
        for _, p := range parts {
            trimmed := strings.TrimSpace(p)
            if trimmed == "" {
                continue
            }
            v, err := strconv.ParseInt(trimmed, 10, 64)
            if err != nil {
                return nil, fmt.Errorf("invalid int64 in array: %s", trimmed)
            }
            result = append(result, v)
        }
        return result, nil

    default:
        return nil, fmt.Errorf("unknown type: %v", valueType)
    }
}
```

## validation.go

```go
// configkit/validation.go
package configkit

import (
    "fmt"
    "regexp"
    "time"
)

// Validation represents a validation rule
type Validation struct {
    Name    string
    Check   func(value any) error
}

// Built-in validation constructors

func validateRequired() Validation {
    return Validation{
        Name: "required",
        Check: func(value any) error {
            if value == nil {
                return fmt.Errorf("value is required")
            }
            if s, ok := value.(string); ok && s == "" {
                return fmt.Errorf("value is required (empty string)")
            }
            return nil
        },
    }
}

func validateMin(min float64) Validation {
    return Validation{
        Name: fmt.Sprintf("min(%v)", min),
        Check: func(value any) error {
            switch v := value.(type) {
            case int64:
                if float64(v) < min {
                    return fmt.Errorf("value %d is less than minimum %v", v, min)
                }
            case float64:
                if v < min {
                    return fmt.Errorf("value %f is less than minimum %v", v, min)
                }
            }
            return nil
        },
    }
}

func validateMax(max float64) Validation {
    return Validation{
        Name: fmt.Sprintf("max(%v)", max),
        Check: func(value any) error {
            switch v := value.(type) {
            case int64:
                if float64(v) > max {
                    return fmt.Errorf("value %d is greater than maximum %v", v, max)
                }
            case float64:
                if v > max {
                    return fmt.Errorf("value %f is greater than maximum %v", v, max)
                }
            }
            return nil
        },
    }
}

func validateMinLength(min int) Validation {
    return Validation{
        Name: fmt.Sprintf("minLength(%d)", min),
        Check: func(value any) error {
            if s, ok := value.(string); ok {
                if len(s) < min {
                    return fmt.Errorf("value length %d is less than minimum %d", len(s), min)
                }
            }
            return nil
        },
    }
}

func validateMaxLength(max int) Validation {
    return Validation{
        Name: fmt.Sprintf("maxLength(%d)", max),
        Check: func(value any) error {
            if s, ok := value.(string); ok {
                if len(s) > max {
                    return fmt.Errorf("value length %d is greater than maximum %d", len(s), max)
                }
            }
            return nil
        },
    }
}

func validateRegexp(pattern string) Validation {
    re := regexp.MustCompile(pattern)
    return Validation{
        Name: fmt.Sprintf("regexp(%s)", pattern),
        Check: func(value any) error {
            if s, ok := value.(string); ok {
                if !re.MatchString(s) {
                    return fmt.Errorf("value does not match pattern %s", pattern)
                }
            }
            return nil
        },
    }
}

func validateOneOf(allowed []string) Validation {
    return Validation{
        Name: fmt.Sprintf("oneOf(%v)", allowed),
        Check: func(value any) error {
            if s, ok := value.(string); ok {
                for _, a := range allowed {
                    if s == a {
                        return nil
                    }
                }
                return fmt.Errorf("value '%s' is not one of: %v", s, allowed)
            }
            return nil
        },
    }
}

func validateMinDuration(min time.Duration) Validation {
    return Validation{
        Name: fmt.Sprintf("minDuration(%s)", min),
        Check: func(value any) error {
            if d, ok := value.(time.Duration); ok {
                if d < min {
                    return fmt.Errorf("duration %s is less than minimum %s", d, min)
                }
            }
            return nil
        },
    }
}

func validateMaxDuration(max time.Duration) Validation {
    return Validation{
        Name: fmt.Sprintf("maxDuration(%s)", max),
        Check: func(value any) error {
            if d, ok := value.(time.Duration); ok {
                if d > max {
                    return fmt.Errorf("duration %s is greater than maximum %s", d, max)
                }
            }
            return nil
        },
    }
}

func validateMinItems(min int) Validation {
    return Validation{
        Name: fmt.Sprintf("minItems(%d)", min),
        Check: func(value any) error {
            switch v := value.(type) {
            case []string:
                if len(v) < min {
                    return fmt.Errorf("array has %d items, minimum is %d", len(v), min)
                }
            case []int64:
                if len(v) < min {
                    return fmt.Errorf("array has %d items, minimum is %d", len(v), min)
                }
            }
            return nil
        },
    }
}

func validateMaxItems(max int) Validation {
    return Validation{
        Name: fmt.Sprintf("maxItems(%d)", max),
        Check: func(value any) error {
            switch v := value.(type) {
            case []string:
                if len(v) > max {
                    return fmt.Errorf("array has %d items, maximum is %d", len(v), max)
                }
            case []int64:
                if len(v) > max {
                    return fmt.Errorf("array has %d items, maximum is %d", len(v), max)
                }
            }
            return nil
        },
    }
}
```

## errors.go

```go
// configkit/errors.go
package configkit

import (
    "fmt"
    "strings"
)

// ConfigError represents a single configuration error
type ConfigError struct {
    Key     string
    Source  string // "env", "flag", "default", or "none"
    Value   string // Masked if secret
    Message string
}

func (e *ConfigError) Error() string {
    if e.Source == "none" {
        return fmt.Sprintf("%s: %s", e.Key, e.Message)
    }
    if e.Value != "" {
        return fmt.Sprintf("%s (%s=%s): %s", e.Key, e.Source, e.Value, e.Message)
    }
    return fmt.Sprintf("%s (%s): %s", e.Key, e.Source, e.Message)
}

// formatErrors creates a nicely formatted error output
func formatErrors(errs []ConfigError) string {
    if len(errs) == 0 {
        return ""
    }

    var sb strings.Builder

    sb.WriteString("\n")
    sb.WriteString("╔══════════════════════════════════════════════════════════════════╗\n")
    sb.WriteString("║                    CONFIGURATION ERRORS                          ║\n")
    sb.WriteString("╠══════════════════════════════════════════════════════════════════╣\n")

    for i, err := range errs {
        // Key line
        sb.WriteString(fmt.Sprintf("║  %-64s║\n", fmt.Sprintf("❌ %s", err.Key)))

        // Source and value
        if err.Source != "none" {
            sourceInfo := fmt.Sprintf("   Source: %s", err.Source)
            if err.Value != "" {
                sourceInfo += fmt.Sprintf(" = %s", err.Value)
            }
            sb.WriteString(fmt.Sprintf("║  %-64s║\n", sourceInfo))
        }

        // Error message
        sb.WriteString(fmt.Sprintf("║  %-64s║\n", fmt.Sprintf("   Error: %s", err.Message)))

        // Separator between errors
        if i < len(errs)-1 {
            sb.WriteString("║  ────────────────────────────────────────────────────────────    ║\n")
        }
    }

    sb.WriteString("╠══════════════════════════════════════════════════════════════════╣\n")
    sb.WriteString(fmt.Sprintf("║  %-64s║\n", fmt.Sprintf("Total: %d error(s)", len(errs))))
    sb.WriteString("╚══════════════════════════════════════════════════════════════════╝\n")

    return sb.String()
}

// maskSecret masks a secret value for display
func maskSecret(value string) string {
    if len(value) <= 4 {
        return "****"
    }
    return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}
```

## secret.go

```go
// configkit/secret.go
package configkit

import (
    "github.com/awnumar/memguard"
)

// Secret wraps a memguard LockedBuffer for secure secret storage
type Secret struct {
    buffer *memguard.LockedBuffer
}

// newSecret creates a new Secret from a string value
func newSecret(value string) *Secret {
    if value == "" {
        return &Secret{}
    }

    buf := memguard.NewBufferFromBytes([]byte(value))
    return &Secret{buffer: buf}
}

// Bytes returns the secret value as bytes
// The returned slice is only valid until Destroy() is called
func (s *Secret) Bytes() []byte {
    if s.buffer == nil {
        return nil
    }
    return s.buffer.Bytes()
}

// String returns the secret value as a string
// The returned string is only valid until Destroy() is called
func (s *Secret) String() string {
    if s.buffer == nil {
        return ""
    }
    return string(s.buffer.Bytes())
}

// Destroy securely wipes the secret from memory
func (s *Secret) Destroy() {
    if s.buffer != nil {
        s.buffer.Destroy()
        s.buffer = nil
    }
}

// IsSet returns true if the secret has a value
func (s *Secret) IsSet() bool {
    return s.buffer != nil && s.buffer.Size() > 0
}

// Size returns the length of the secret
func (s *Secret) Size() int {
    if s.buffer == nil {
        return 0
    }
    return s.buffer.Size()
}

// SecretStore holds all secrets for cleanup
type SecretStore struct {
    secrets map[string]*Secret
}

func newSecretStore() *SecretStore {
    return &SecretStore{
        secrets: make(map[string]*Secret),
    }
}

func (ss *SecretStore) Store(key, value string) {
    ss.secrets[key] = newSecret(value)
}

func (ss *SecretStore) Get(key string) *Secret {
    if s, ok := ss.secrets[key]; ok {
        return s
    }
    return &Secret{}
}

func (ss *SecretStore) DestroyAll() {
    for _, s := range ss.secrets {
        s.Destroy()
    }
    ss.secrets = make(map[string]*Secret)
}
```

## definition.go

```go
// configkit/definition.go
package configkit

import (
    "time"
)

// Definition represents a configuration definition with fluent builder
type Definition struct {
    key          string
    valueType    ValueType
    envVar       string
    flag         string
    defaultValue any
    required     bool
    secret       bool
    delimiter    string
    validations  []Validation
    description  string
}

// DefinitionBuilder provides a fluent API for building definitions
type DefinitionBuilder struct {
    def    *Definition
    config *Config
}

// newDefinitionBuilder creates a new builder
func newDefinitionBuilder(cfg *Config, key string) *DefinitionBuilder {
    return &DefinitionBuilder{
        def: &Definition{
            key:       key,
            valueType: TypeString, // default
            delimiter: ",",        // default delimiter
        },
        config: cfg,
    }
}

// Type setters

func (b *DefinitionBuilder) String() *DefinitionBuilder {
    b.def.valueType = TypeString
    return b
}

func (b *DefinitionBuilder) Int64() *DefinitionBuilder {
    b.def.valueType = TypeInt64
    return b
}

func (b *DefinitionBuilder) Float64() *DefinitionBuilder {
    b.def.valueType = TypeFloat64
    return b
}

func (b *DefinitionBuilder) Bool() *DefinitionBuilder {
    b.def.valueType = TypeBool
    return b
}

func (b *DefinitionBuilder) Duration() *DefinitionBuilder {
    b.def.valueType = TypeDuration
    return b
}

func (b *DefinitionBuilder) URL() *DefinitionBuilder {
    b.def.valueType = TypeURL
    return b
}

func (b *DefinitionBuilder) StringSlice() *DefinitionBuilder {
    b.def.valueType = TypeStringSlice
    return b
}

func (b *DefinitionBuilder) Int64Slice() *DefinitionBuilder {
    b.def.valueType = TypeInt64Slice
    return b
}

// Source setters

func (b *DefinitionBuilder) Env(envVar string) *DefinitionBuilder {
    b.def.envVar = envVar
    return b
}

func (b *DefinitionBuilder) Flag(flag string) *DefinitionBuilder {
    b.def.flag = flag
    return b
}

// Behavior setters

func (b *DefinitionBuilder) Required() *DefinitionBuilder {
    b.def.required = true
    b.def.validations = append(b.def.validations, validateRequired())
    return b
}

func (b *DefinitionBuilder) Secret() *DefinitionBuilder {
    b.def.secret = true
    return b
}

func (b *DefinitionBuilder) Default(value any) *DefinitionBuilder {
    b.def.defaultValue = value
    return b
}

func (b *DefinitionBuilder) Delimiter(d string) *DefinitionBuilder {
    b.def.delimiter = d
    return b
}

func (b *DefinitionBuilder) Description(desc string) *DefinitionBuilder {
    b.def.description = desc
    return b
}

// Validation setters

func (b *DefinitionBuilder) Min(min float64) *DefinitionBuilder {
    b.def.validations = append(b.def.validations, validateMin(min))
    return b
}

func (b *DefinitionBuilder) Max(max float64) *DefinitionBuilder {
    b.def.validations = append(b.def.validations, validateMax(max))
    return b
}

func (b *DefinitionBuilder) Range(min, max float64) *DefinitionBuilder {
    b.def.validations = append(b.def.validations, validateMin(min))
    b.def.validations = append(b.def.validations, validateMax(max))
    return b
}

func (b *DefinitionBuilder) MinLength(min int) *DefinitionBuilder {
    b.def.validations = append(b.def.validations, validateMinLength(min))
    return b
}

func (b *DefinitionBuilder) MaxLength(max int) *DefinitionBuilder {
    b.def.validations = append(b.def.validations, validateMaxLength(max))
    return b
}

func (b *DefinitionBuilder) LengthRange(min, max int) *DefinitionBuilder {
    b.def.validations = append(b.def.validations, validateMinLength(min))
    b.def.validations = append(b.def.validations, validateMaxLength(max))
    return b
}

func (b *DefinitionBuilder) Regexp(pattern string) *DefinitionBuilder {
    b.def.validations = append(b.def.validations, validateRegexp(pattern))
    return b
}

func (b *DefinitionBuilder) OneOf(allowed ...string) *DefinitionBuilder {
    b.def.validations = append(b.def.validations, validateOneOf(allowed))
    return b
}

func (b *DefinitionBuilder) MinDuration(min time.Duration) *DefinitionBuilder {
    b.def.validations = append(b.def.validations, validateMinDuration(min))
    return b
}

func (b *DefinitionBuilder) MaxDuration(max time.Duration) *DefinitionBuilder {
    b.def.validations = append(b.def.validations, validateMaxDuration(max))
    return b
}

func (b *DefinitionBuilder) DurationRange(min, max time.Duration) *DefinitionBuilder {
    b.def.validations = append(b.def.validations, validateMinDuration(min))
    b.def.validations = append(b.def.validations, validateMaxDuration(max))
    return b
}

func (b *DefinitionBuilder) MinItems(min int) *DefinitionBuilder {
    b.def.validations = append(b.def.validations, validateMinItems(min))
    return b
}

func (b *DefinitionBuilder) MaxItems(max int) *DefinitionBuilder {
    b.def.validations = append(b.def.validations, validateMaxItems(max))
    return b
}

func (b *DefinitionBuilder) ItemsRange(min, max int) *DefinitionBuilder {
    b.def.validations = append(b.def.validations, validateMinItems(min))
    b.def.validations = append(b.def.validations, validateMaxItems(max))
    return b
}

// Custom adds a custom validation function
func (b *DefinitionBuilder) Custom(name string, check func(value any) error) *DefinitionBuilder {
    b.def.validations = append(b.def.validations, Validation{Name: name, Check: check})
    return b
}

// Build finalizes the definition and adds it to the config
// This is called automatically; you don't need to call it explicitly
func (b *DefinitionBuilder) build() *Definition {
    return b.def
}
```

## get.go

```go
// configkit/get.go
package configkit

import (
    "fmt"
    "time"
)

// Get retrieves a configuration value with type safety using generics
func Get[T any](c *Config, key string) T {
    value, exists := c.values[key]
    if !exists {
        panic(fmt.Sprintf("configkit: key '%s' not found (did you define it?)", key))
    }

    // Check if it's a secret (stored as string, needs special handling)
    def, hasDef := c.definitions[key]
    if hasDef && def.secret {
        panic(fmt.Sprintf("configkit: key '%s' is a secret, use GetSecret() instead", key))
    }

    result, ok := value.(T)
    if !ok {
        panic(fmt.Sprintf("configkit: key '%s' has type %T, not %T", key, value, result))
    }

    return result
}

// MustGet is an alias for Get (both panic on error)
func MustGet[T any](c *Config, key string) T {
    return Get[T](c, key)
}

// GetOr retrieves a configuration value or returns a default if not set
func GetOr[T any](c *Config, key string, defaultValue T) T {
    value, exists := c.values[key]
    if !exists || value == nil {
        return defaultValue
    }

    result, ok := value.(T)
    if !ok {
        return defaultValue
    }

    return result
}

// Has checks if a key exists and has a non-nil value
func (c *Config) Has(key string) bool {
    value, exists := c.values[key]
    return exists && value != nil
}

// GetSecret retrieves a secret value
func (c *Config) GetSecret(key string) *Secret {
    return c.secrets.Get(key)
}

// Keys returns all defined configuration keys
func (c *Config) Keys() []string {
    keys := make([]string, 0, len(c.definitions))
    for k := range c.definitions {
        keys = append(keys, k)
    }
    return keys
}

// Convenience typed getters (non-generic alternative)

func (c *Config) GetString(key string) string {
    return Get[string](c, key)
}

func (c *Config) GetInt64(key string) int64 {
    return Get[int64](c, key)
}

func (c *Config) GetFloat64(key string) float64 {
    return Get[float64](c, key)
}

func (c *Config) GetBool(key string) bool {
    return Get[bool](c, key)
}

func (c *Config) GetDuration(key string) time.Duration {
    return Get[time.Duration](c, key)
}

func (c *Config) GetStringSlice(key string) []string {
    return Get[[]string](c, key)
}

func (c *Config) GetInt64Slice(key string) []int64 {
    return Get[[]int64](c, key)
}
```

## config.go

```go
// configkit/config.go
package configkit

import (
    "flag"
    "fmt"
    "os"
    "strings"
)

// Config holds configuration definitions and values
type Config struct {
    definitions map[string]*Definition
    values      map[string]any
    secrets     *SecretStore
    flagSet     *flag.FlagSet
    flagValues  map[string]*string
    processed   bool
}

// New creates a new Config instance
func New() *Config {
    return &Config{
        definitions: make(map[string]*Definition),
        values:      make(map[string]any),
        secrets:     newSecretStore(),
        flagSet:     flag.NewFlagSet(os.Args[0], flag.ContinueOnError),
        flagValues:  make(map[string]*string),
    }
}

// Define starts a new configuration definition
func (c *Config) Define(key string) *DefinitionBuilder {
    builder := newDefinitionBuilder(c, key)
    c.definitions[key] = builder.def
    return builder
}

// Process parses flags and environment variables, validates all definitions,
// and populates the values map. Returns any configuration errors.
func (c *Config) Process() []ConfigError {
    if c.processed {
        return nil
    }
    c.processed = true

    var errs []ConfigError

    // Register all flags first
    for key, def := range c.definitions {
        if def.flag != "" {
            c.flagValues[key] = c.flagSet.String(def.flag, "", def.description)
        }
    }

    // Parse command line flags
    // Ignore errors from unknown flags to allow partial parsing
    c.flagSet.Parse(os.Args[1:])

    // Process each definition
    for key, def := range c.definitions {
        value, source, err := c.resolveValue(key, def)
        if err != nil {
            displayValue := ""
            if value != nil && !def.secret {
                displayValue = fmt.Sprintf("%v", value)
            } else if value != nil && def.secret {
                displayValue = maskSecret(fmt.Sprintf("%v", value))
            }
            errs = append(errs, ConfigError{
                Key:     key,
                Source:  source,
                Value:   displayValue,
                Message: err.Error(),
            })
            continue
        }

        // Store the value
        if def.secret && value != nil {
            // Store secrets in memguard
            strValue := fmt.Sprintf("%v", value)
            c.secrets.Store(key, strValue)
            // Also store a placeholder in values for Has() checks
            c.values[key] = "[SECRET]"
        } else {
            c.values[key] = value
        }
    }

    return errs
}

// resolveValue determines the value from flags, env, or default
func (c *Config) resolveValue(key string, def *Definition) (any, string, error) {
    var rawValue string
    var source string

    // Priority 1: Command line flags
    if def.flag != "" {
        if flagVal, ok := c.flagValues[key]; ok && flagVal != nil && *flagVal != "" {
            rawValue = *flagVal
            source = "flag"
        }
    }

    // Priority 2: Environment variables
    if rawValue == "" && def.envVar != "" {
        if envVal := os.Getenv(def.envVar); envVal != "" {
            rawValue = envVal
            source = "env"
        }
    }

    // Priority 3: Default value
    if rawValue == "" && def.defaultValue != nil {
        source = "default"
        // Default is already the correct type, validate and return
        for _, v := range def.validations {
            if v.Name == "required" {
                continue // Skip required check for defaults
            }
            if err := v.Check(def.defaultValue); err != nil {
                return def.defaultValue, source, err
            }
        }
        return def.defaultValue, source, nil
    }

    // No value found
    if rawValue == "" {
        source = "none"
        if def.required {
            return nil, source, fmt.Errorf("required value not provided (set %s or --%s)", def.envVar, def.flag)
        }
        return nil, source, nil
    }

    // Parse the raw string value into the expected type
    parsedValue, err := parseValue(rawValue, def.valueType, def.delimiter)
    if err != nil {
        return rawValue, source, err
    }

    // Run validations
    for _, validation := range def.validations {
        if err := validation.Check(parsedValue); err != nil {
            return parsedValue, source, err
        }
    }

    return parsedValue, source, nil
}

// PrintErrors prints formatted error messages to stderr
func (c *Config) PrintErrors(errs []ConfigError) {
    fmt.Fprint(os.Stderr, formatErrors(errs))
}

// Destroy cleans up all secrets from memory
func (c *Config) Destroy() {
    c.secrets.DestroyAll()
}

// Dump returns a map of all configuration values (secrets masked)
func (c *Config) Dump() map[string]string {
    result := make(map[string]string)
    for key, def := range c.definitions {
        if def.secret {
            if c.secrets.Get(key).IsSet() {
                result[key] = "[SECRET:" + fmt.Sprintf("%d", c.secrets.Get(key).Size()) + " bytes]"
            } else {
                result[key] = "[SECRET:not set]"
            }
        } else if val, ok := c.values[key]; ok && val != nil {
            result[key] = fmt.Sprintf("%v", val)
        } else {
            result[key] = "[not set]"
        }
    }
    return result
}

// GenerateHelp creates a help message with all configuration options
func (c *Config) GenerateHelp() string {
    var sb strings.Builder

    sb.WriteString("Configuration Options:\n\n")

    for key, def := range c.definitions {
        sb.WriteString(fmt.Sprintf("  %s\n", key))
        sb.WriteString(fmt.Sprintf("    Type: %s\n", def.valueType))

        if def.envVar != "" {
            sb.WriteString(fmt.Sprintf("    Env:  %s\n", def.envVar))
        }
        if def.flag != "" {
            sb.WriteString(fmt.Sprintf("    Flag: --%s\n", def.flag))
        }
        if def.required {
            sb.WriteString("    Required: yes\n")
        }
        if def.secret {
            sb.WriteString("    Secret: yes (protected in memory)\n")
        }
        if def.defaultValue != nil {
            if def.secret {
                sb.WriteString("    Default: [hidden]\n")
            } else {
                sb.WriteString(fmt.Sprintf("    Default: %v\n", def.defaultValue))
            }
        }
        if def.description != "" {
            sb.WriteString(fmt.Sprintf("    Description: %s\n", def.description))
        }

        // List validations
        if len(def.validations) > 0 {
            var valNames []string
            for _, v := range def.validations {
                valNames = append(valNames, v.Name)
            }
            sb.WriteString(fmt.Sprintf("    Validations: %s\n", strings.Join(valNames, ", ")))
        }

        sb.WriteString("\n")
    }

    return sb.String()
}
```

---

# 6. Complete Usage Example

```go
// Example: AuthForge using ConfigKit
package main

import (
    "fmt"
    "os"
    "time"

    "github.com/fernandezvara/commandkit"
)

func main() {
    cfg := configkit.New()

    // ═══════════════════════════════════════════════════════════════════════
    // SERVER
    // ═══════════════════════════════════════════════════════════════════════
    cfg.Define("PORT").
        Int64().
        Env("PORT").
        Flag("port").
        Default(int64(8080)).
        Range(1, 65535).
        Description("HTTP server port")

    cfg.Define("BASE_URL").
        String().
        Env("BASE_URL").
        Flag("base-url").
        Required().
        URL().
        Description("Public base URL of the service")

    cfg.Define("LOG_LEVEL").
        String().
        Env("LOG_LEVEL").
        Flag("log-level").
        Default("info").
        OneOf("debug", "info", "warn", "error").
        Description("Logging verbosity")

    // ═══════════════════════════════════════════════════════════════════════
    // DATABASE
    // ═══════════════════════════════════════════════════════════════════════
    cfg.Define("DATABASE_URL").
        String().
        Env("DATABASE_URL").
        Required().
        Secret().
        MinLength(10).
        Description("PostgreSQL connection string")

    // ═══════════════════════════════════════════════════════════════════════
    // JWT / SECURITY
    // ═══════════════════════════════════════════════════════════════════════
    cfg.Define("JWT_SIGNING_KEY").
        String().
        Env("JWT_SIGNING_KEY").
        Required().
        Secret().
        MinLength(32).
        Description("Secret key for signing access tokens (min 32 chars)")

    cfg.Define("JWT_REFRESH_KEY").
        String().
        Env("JWT_REFRESH_KEY").
        Required().
        Secret().
        MinLength(32).
        Description("Secret key for signing refresh tokens (min 32 chars)")

    cfg.Define("ENCRYPTION_KEY").
        String().
        Env("ENCRYPTION_KEY").
        Required().
        Secret().
        LengthRange(32, 32).
        Description("AES-256 encryption key (exactly 32 chars)")

    cfg.Define("ACCESS_TOKEN_TTL").
        Duration().
        Env("ACCESS_TOKEN_TTL").
        Flag("access-token-ttl").
        Default(15 * time.Minute).
        DurationRange(1*time.Minute, 24*time.Hour).
        Description("Access token lifetime")

    cfg.Define("REFRESH_TOKEN_TTL").
        Duration().
        Env("REFRESH_TOKEN_TTL").
        Flag("refresh-token-ttl").
        Default(7 * 24 * time.Hour).
        DurationRange(1*time.Hour, 90*24*time.Hour).
        Description("Refresh token lifetime")

    // ═══════════════════════════════════════════════════════════════════════
    // CORS
    // ═══════════════════════════════════════════════════════════════════════
    cfg.Define("CORS_ORIGINS").
        StringSlice().
        Env("CORS_ORIGINS").
        Flag("cors-origins").
        Delimiter(",").
        Default([]string{}).
        Description("Allowed CORS origins (comma-separated)")

    // ═══════════════════════════════════════════════════════════════════════
    // OAUTH PROVIDERS (optional)
    // ═══════════════════════════════════════════════════════════════════════
    cfg.Define("GOOGLE_CLIENT_ID").
        String().
        Env("GOOGLE_CLIENT_ID").
        Description("Google OAuth client ID")

    cfg.Define("GOOGLE_CLIENT_SECRET").
        String().
        Env("GOOGLE_CLIENT_SECRET").
        Secret().
        Description("Google OAuth client secret")

    cfg.Define("MICROSOFT_CLIENT_ID").
        String().
        Env("MICROSOFT_CLIENT_ID").
        Description("Microsoft OAuth client ID")

    cfg.Define("MICROSOFT_CLIENT_SECRET").
        String().
        Env("MICROSOFT_CLIENT_SECRET").
        Secret().
        Description("Microsoft OAuth client secret")

    cfg.Define("MICROSOFT_TENANT_ID").
        String().
        Env("MICROSOFT_TENANT_ID").
        Default("common").
        Description("Microsoft tenant ID")

    // ═══════════════════════════════════════════════════════════════════════
    // RATE LIMITING
    // ═══════════════════════════════════════════════════════════════════════
    cfg.Define("RATE_LIMIT_RPS").
        Float64().
        Env("RATE_LIMIT_RPS").
        Default(float64(100)).
        Range(1, 10000).
        Description("Rate limit requests per second")

    cfg.Define("RATE_LIMIT_BURST").
        Int64().
        Env("RATE_LIMIT_BURST").
        Default(int64(200)).
        Range(1, 10000).
        Description("Rate limit burst size")

    // ═══════════════════════════════════════════════════════════════════════
    // PROCESS & VALIDATE
    // ═══════════════════════════════════════════════════════════════════════
    if errs := cfg.Process(); len(errs) > 0 {
        cfg.PrintErrors(errs)
        os.Exit(1)
    }

    // Ensure secrets are cleaned up on exit
    defer cfg.Destroy()

    // ═══════════════════════════════════════════════════════════════════════
    // USE CONFIGURATION
    // ═══════════════════════════════════════════════════════════════════════

    // Non-secrets with generics
    port := configkit.Get[int64](cfg, "PORT")
    baseURL := configkit.Get[string](cfg, "BASE_URL")
    logLevel := configkit.Get[string](cfg, "LOG_LEVEL")
    accessTTL := configkit.Get[time.Duration](cfg, "ACCESS_TOKEN_TTL")
    corsOrigins := configkit.Get[[]string](cfg, "CORS_ORIGINS")
    rateLimitRPS := configkit.Get[float64](cfg, "RATE_LIMIT_RPS")

    fmt.Printf("Starting AuthForge on port %d\n", port)
    fmt.Printf("Base URL: %s\n", baseURL)
    fmt.Printf("Log Level: %s\n", logLevel)
    fmt.Printf("Access Token TTL: %s\n", accessTTL)
    fmt.Printf("CORS Origins: %v\n", corsOrigins)
    fmt.Printf("Rate Limit: %.0f rps\n", rateLimitRPS)

    // Secrets (protected with memguard)
    jwtKey := cfg.GetSecret("JWT_SIGNING_KEY")
    dbURL := cfg.GetSecret("DATABASE_URL")

    fmt.Printf("JWT Key size: %d bytes\n", jwtKey.Size())
    fmt.Printf("Database URL size: %d bytes\n", dbURL.Size())

    // Check optional OAuth
    if cfg.Has("GOOGLE_CLIENT_ID") {
        googleID := configkit.Get[string](cfg, "GOOGLE_CLIENT_ID")
        fmt.Printf("Google OAuth enabled: %s\n", googleID)
    }

    // Print all config (secrets masked)
    fmt.Println("\nFull configuration:")
    for k, v := range cfg.Dump() {
        fmt.Printf("  %s = %s\n", k, v)
    }

    // Use secrets carefully
    // keyBytes := jwtKey.Bytes() // Use for actual signing
}
```

---

# 7. Error Output Example

When configuration has errors:

```
╔══════════════════════════════════════════════════════════════════╗
║                    CONFIGURATION ERRORS                          ║
╠══════════════════════════════════════════════════════════════════╣
║  ❌ BASE_URL                                                     ║
║     Source: none                                                 ║
║     Error: required value not provided (set BASE_URL or --base-url)║
║  ────────────────────────────────────────────────────────────    ║
║  ❌ JWT_SIGNING_KEY                                              ║
║     Source: env = ab****ef                                       ║
║     Error: value length 8 is less than minimum 32                ║
║  ────────────────────────────────────────────────────────────    ║
║  ❌ PORT                                                         ║
║     Source: env = 99999                                          ║
║     Error: value 99999 is greater than maximum 65535             ║
╠══════════════════════════════════════════════════════════════════╣
║  Total: 3 error(s)                                               ║
╚══════════════════════════════════════════════════════════════════╝
```

---

# 8. Summary

| Feature      | Implementation                                                      |
| ------------ | ------------------------------------------------------------------- |
| Fluent API   | Chainable methods on `DefinitionBuilder`                            |
| Generics     | `configkit.Get[T](cfg, key)`                                        |
| Secrets      | memguard `LockedBuffer` with `cfg.GetSecret()`                      |
| Types        | String, Int64, Float64, Bool, Duration, URL, []String, []Int64      |
| Sources      | Config Files → Flags (priority) → Env Vars → Defaults               |
| Validations  | Required, Range, Length, Regexp, OneOf, Duration range, Items range |
| Arrays       | Configurable delimiter (default: comma)                             |
| Errors       | Collected, formatted, secrets masked                                |
| Commands     | Fluent command definition with subcommands and aliases              |
| Help         | Auto-generated help with suggestions                                |
| Config Merge | Global + command config with override warnings                      |
| Config Files | JSON, YAML, TOML support with environment-specific overrides        |
| Middleware   | Command middleware for logging, auth, error handling                |

## Key Design Decisions

1. **Secrets never in `values` map** — Only in memguard-protected `SecretStore`
2. **Generic Get panics for secrets** — Forces use of `GetSecret()`
3. **All errors collected** — Process() returns ALL errors, not just first
4. **Secrets masked in errors** — Shows `ab****ef` not full value
5. **Destroy() for cleanup** — Explicitly wipe secrets from memory
6. **Commands use same fluent API** — Consistent with configuration definition
7. **Command flags override globals** — With clear warnings for conflicts
8. **Help is auto-generated** — From definitions and descriptions
9. **Subcommands are nested naturally** — Same API as parent commands
10. **Config files have highest priority** — But can be overridden by flags
11. **Middleware executes in registration order** — With proper unwinding

---

# 9. Command System

ConfigKit includes a powerful command system that allows you to define CLI commands with subcommands, aliases, and command-specific configuration.

## Features

- **Fluent command definition** with same API as configuration
- **Subcommands** with natural nesting
- **Aliases** for short commands
- **Command-specific config** that can override global settings
- **Auto-generated help** with `--help` flag
- **Smart suggestions** for unknown commands
- **Global + command config merging** with override warnings

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
    cfg := configkit.New()

    // Global configuration (available to all commands)
    cfg.Define("VERBOSE").Bool().Flag("verbose").Default(false)
    cfg.Define("LOG_LEVEL").String().Env("LOG_LEVEL").Default("info")

    // Start command
    cfg.Command("start").
        Func(startCommand).
        ShortHelp("Start the service").
        LongHelp(`Start the service with all components initialized.
        Usage: myapp start [options]
        This will initialize the database, start HTTP server, and begin accepting connections.`).
        Aliases("s", "run", "up").
        Config(func(cc *configkit.CommandConfig) {
            cc.Define("PORT").Int64().Flag("port").Default(8080).Range(1, 65535)
            cc.Define("DAEMON").Bool().Flag("daemon").Default(false)
            cc.Define("WORKERS").Int64Slice().Flag("workers").Delimiter(",").Default([]int64{1})
        }).
        SubCommand("worker").
            Func(startWorkerCommand).
            ShortHelp("Start only worker processes").
            Aliases("w").
            Config(func(cc *configkit.CommandConfig) {
                cc.Define("COUNT").Int64().Flag("count").Default(1)
            })

    // Stop command
    cfg.Command("stop").
        Func(stopCommand).
        ShortHelp("Stop the service gracefully").
        Aliases("quit", "exit").
        Config(func(cc *configkit.CommandConfig) {
            cc.Define("TIMEOUT").Duration().Flag("timeout").Default(30 * time.Second)
        })

    // Execute
    if err := cfg.Execute(os.Args); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func startCommand(ctx *configkit.CommandContext) error {
    // Command-specific config
    port := configkit.Get[int64](ctx.Config, "PORT")
    daemon := configkit.Get[bool](ctx.Config, "DAEMON")
    workers := configkit.Get[[]int64](ctx.Config, "WORKERS")

    // Global config
    verbose := configkit.Get[bool](ctx.Config, "VERBOSE")
    logLevel := configkit.Get[string](ctx.Config, "LOG_LEVEL")

    fmt.Printf("Starting on port %d (daemon: %v, verbose: %v, log: %s)\n", port, daemon, verbose, logLevel)
    fmt.Printf("Workers: %v\n", workers)
    return nil
}

func startWorkerCommand(ctx *configkit.CommandContext) error {
    count := configkit.Get[int64](ctx.Config, "COUNT")
    fmt.Printf("Starting %d worker processes\n", count)
    return nil
}

func stopCommand(ctx *configkit.CommandContext) error {
    timeout := configkit.Get[time.Duration](ctx.Config, "TIMEOUT")
    fmt.Printf("Stopping gracefully with timeout: %v\n", timeout)
    return nil
}
```

## Usage Examples

```bash
# List all commands
$ myapp help
Available commands:
  start       Start the service (aliases: s, run, up)
  stop        Stop the service (aliases: quit, exit)

# Command help
$ myapp start --help
Usage: myapp start [options]

Start the service with all components initialized.
Usage: myapp start [options]
This will initialize the database, start HTTP server, and begin accepting connections.

Options:
  --port         Port to listen on (default: 8080)
  --daemon       Run in background (default: false)
  --workers      Worker processes (default: [1])

Subcommands:
  worker         Start only worker processes (aliases: w)

# Using commands
$ myapp start --port 3000 --daemon
$ myapp s --port 3000                    # alias
$ myapp start worker --count 5
$ myapp start worker w                   # subcommand alias
$ myapp stop --timeout 1m

# Global flags work with all commands
$ myapp --verbose start --port 3000
$ myapp start --port 3000 --verbose      # same effect
```

## API Reference

### Command Definition

```go
// Define a new command
func (c *Config) Command(name string) *CommandBuilder

// Command builder methods
func (b *CommandBuilder) Func(fn func(*CommandContext) error) *CommandBuilder
func (b *CommandBuilder) ShortHelp(help string) *CommandBuilder
func (b *CommandBuilder) LongHelp(help string) *CommandBuilder
func (b *CommandBuilder) Aliases(aliases ...string) *CommandBuilder
func (b *CommandBuilder) Config(fn func(*CommandConfig)) *CommandBuilder
func (b *CommandBuilder) SubCommand(name string) *CommandBuilder
```

### Command Context

```go
type CommandContext struct {
    Args       []string    // Raw command arguments
    Config     *Config     // Merged global + command config
    Command    string      // Command name
    SubCommand string      // Subcommand name if any
    Flags      map[string]string // Parsed flags
}
```

### Command Config

```go
type CommandConfig struct {
    *Config
    commandName string
}

// Define command-specific configuration (same API as global)
func (cc *CommandConfig) Define(key string) *DefinitionBuilder
```

### Execution

```go
// Execute commands from command line
func (c *Config) Execute(args []string) error

// Show help
func (c *Config) ShowGlobalHelp() error
func (c *Config) ShowCommandHelp(command string) error

// List commands
func (c *Config) ListCommands() map[string]string
```

## Configuration Priority

Command configuration follows this priority (highest to lowest):

1. **Command flag** (e.g., `myapp start --port 3000`)
2. **Command environment variable** (e.g., `START_PORT=3000`)
3. **Global flag** (e.g., `myapp --port 3000 start`)
4. **Global environment variable** (e.g., `PORT=3000`)
5. **Command default value**
6. **Global default value**

## Flag Override Warning

⚠️ **WARNING**: Command flags with the same name as global flags will override global values.

```go
// Global definition
cfg.Define("VERBOSE").Bool().Flag("verbose").Default(false)

// Command definition (overrides global)
cfg.Command("start").
    Config(func(cc *configkit.CommandConfig) {
        cc.Define("VERBOSE").Bool().Flag("verbose").Default(true)  // Overrides global
    })

// This will print a warning:
// ⚠️  Warning: Command flag "VERBOSE" overrides global flag
```

Be careful when using flag names that might conflict with global configuration.

## Error Handling

### Unknown Commands

```bash
$ myapp stat
❌ Unknown command: "stat"
Did you mean: "start"?
```

### Missing Required Flags

```bash
$ myapp start
❌ Required flag --port not provided
Usage: myapp start [options]
Options:
  --port      Port to listen on (required)
  --daemon    Run in background (default: false)
```

### Configuration Errors

Configuration errors in commands use the same beautiful formatting as global configuration:

```
╔══════════════════════════════════════════════════════════════════╗
║                    COMMAND CONFIGURATION ERRORS                  ║
╠══════════════════════════════════════════════════════════════════╣
║  ❌ PORT                                                         ║
║     Source: flag = 99999                                         ║
║     Error: value 99999 is greater than maximum 65535             ║
╠══════════════════════════════════════════════════════════════════╣
║  Total: 1 error(s)                                               ║
╚══════════════════════════════════════════════════════════════════╝
```

## Implementation Structure

```
configkit/
├── go.mod
├── config.go           # Main Config struct and Process()
├── definition.go       # Definition builder with fluent API
├── types.go            # Type constants and parsing
├── validation.go       # Validation rules
├── errors.go           # Error types and formatting
├── get.go              # Generic Get function
├── secret.go           # memguard wrapper for secrets
├── files.go            # Configuration file loading (JSON, YAML, TOML)
├── middleware.go       # Command middleware system
├── command.go          # Command system implementation
├── command_builder.go  # Command builder with fluent API
├── command_context.go  # Command context struct
├── help.go             # Help generation
├── suggestions.go      # Command suggestion algorithm
└── config_test.go      # Tests
```

---

# 11. Future Enhancements

The following features are planned for future versions of ConfigKit:

## Environment Variable Prefixing

```go
cfg.SetEnvPrefix("MYAPP") // MYAPP_PORT, MYAPP_DATABASE_URL
```

## Configuration Profiles/Environments

```go
cfg.SetProfile("production") // Loads production-specific overrides
```

## Configuration Validation Groups

```go
cfg.ValidationGroup("database", func(cfg *Config) error {
    if cfg.GetString("DB_TYPE") == "postgres" && !cfg.Has("DATABASE_URL") {
        return errors.New("DATABASE_URL required for PostgreSQL")
    }
    return nil
})
```

## Configuration Export/Import

```go
cfg.ExportJSON() // Export current config as JSON
cfg.ImportYAML(file) // Import from YAML file
```

## Command Completion

```go
cfg.GenerateCompletion("bash") // Generate bash completion script
```

## Configuration Templates

```go
cfg.GenerateTemplate("yaml", os.Stdout) // Generate YAML template
```

## Enhanced Error Context

```go
// Instead of just "value 99999 is greater than maximum 65535"
// Show: "PORT=99999 is greater than maximum 65535 (allowed range: 1-65535)"
```

## Performance Optimization

- Configuration parsing speed improvements
- Memory usage optimization
- Secret handling performance

## Migration Support

- Migration path from existing configuration systems (viper, envconfig, etc.)
- Configuration versioning and migration support

## Internationalization

- i18n support for help and error messages

## Plugin System

- Extensibility beyond custom validation
- Custom sources, validators, and formatters

## Configuration Schema

- JSON schema generation for validation
- Configuration documentation generation

## Debugging Support

- Configuration tracing and debugging tools
- Configuration change tracking

## Security Enhancements

- Secret access logging
- Configuration change auditing
- Security scanning integration

## Contributing

Future enhancements are prioritized based on community feedback and practical needs. If you'd like to contribute or request a specific feature, please open an issue on the project repository.
