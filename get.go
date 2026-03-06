// commandkit/get.go
package commandkit

import (
	"fmt"
	"log"
	"strings"
)

// GetError represents an error collected from Get functions
type GetError struct {
	Key          string
	ExpectedType string
	ActualType   string
	Message      string
	IsSecret     bool
	Flag         string  // Flag name (e.g., "port")
	EnvVar       string  // Environment variable name (e.g., "PORT")
	config       *Config // Reference to config for definition lookup
}

// logWarningForDesigner logs warnings for CLI designers about configuration issues
func logWarningForDesigner(message string) {
	// Use log package with a distinct prefix for designer warnings
	log.Printf("[CONFIG WARNING] %s", message)
}

// getErrorDisplayName returns the display name matching help message format
func getErrorDisplayName(err GetError, c *Config) string {
	// Get the definition to extract type and other info
	var valueType string
	var indicators []string

	if def, hasDef := c.definitions[err.Key]; hasDef {
		// Add value type
		valueType = fmt.Sprintf("%s", def.valueType)

		// Add environment variable context
		if def.envVar != "" {
			indicators = append(indicators, fmt.Sprintf("env: %s", def.envVar))
		}

		// Add required indicator
		if def.required {
			indicators = append(indicators, "required")
		}

		// Add secret indicator
		if def.secret {
			indicators = append(indicators, "secret")
		}
	} else {
		valueType = "string" // fallback type
	}

	// Build the display name
	var displayName string
	if err.Flag != "" {
		displayName = fmt.Sprintf("-%s %s", err.Flag, valueType)
	} else if def, hasDef := c.definitions[err.Key]; hasDef && def.flag == "" && def.envVar != "" {
		// Environment-only configuration
		displayName = fmt.Sprintf("(no flag) %s", valueType)
		// env var already added in indicators above
	} else if err.EnvVar != "" {
		displayName = fmt.Sprintf("(no flag) %s", valueType)
		indicators = append([]string{fmt.Sprintf("env: %s", err.EnvVar)}, indicators...)
	} else {
		displayName = fmt.Sprintf("%s %s", err.Key, valueType)
	}

	// Add indicators
	if len(indicators) > 0 {
		return fmt.Sprintf("%s (%s)", displayName, strings.Join(indicators, ", "))
	}

	return displayName
}

// Get retrieves a configuration value and returns a CommandResult for unified error handling
// This is a breaking change from the previous (T, error) return pattern
func Get[T any](ctx *CommandContext, key string) *CommandResult {
	c := ctx.Config
	value, exists := c.values[key]
	if !exists {
		// Check if this is required data - if so, return validation error
		if def, hasDef := c.definitions[key]; hasDef && def.required {
			// Log warning for designer and return validation error
			logWarningForDesigner(fmt.Sprintf("Required key '%s' is not provided", key))
			return ValidationError(fmt.Sprintf("required configuration '%s' not provided", key)).
				WithCommand(ctx.Command, ctx.SubCommand).
				WithContext("key", key).
				WithContext("expected_type", fmt.Sprintf("%T", (*new(T))))
		}
		// For non-required keys, collect error and return result
		ctx.execution.CollectErrorWithConfig(c, key, "not found", "", "key not defined", false)
		return ErrorWithMessage(fmt.Errorf("configuration '%s' not found", key),
			fmt.Sprintf("configuration '%s' not found", key)).
			WithCommand(ctx.Command, ctx.SubCommand).
			WithContext("key", key)
	}

	// Check if it's a secret (stored as string, needs special handling)
	def, hasDef := c.definitions[key]
	if hasDef && def.secret {
		ctx.execution.CollectErrorWithConfig(c, key, "secret", "", "use GetSecret() instead", true)
		return ErrorWithMessage(fmt.Errorf("configuration '%s' is secret, use GetSecret()", key),
			fmt.Sprintf("configuration '%s' is secret, use GetSecret()", key)).
			WithCommand(ctx.Command, ctx.SubCommand).
			WithContext("key", key).
			WithContext("is_secret", true)
	}

	result, ok := value.(T)
	if !ok {
		ctx.execution.CollectErrorWithConfig(c, key, fmt.Sprintf("%T", (*new(T))), fmt.Sprintf("%T", value), "type mismatch", false)
		return ErrorWithMessage(fmt.Errorf("configuration '%s' type mismatch: expected %T, got %T", key, (*new(T)), value),
			fmt.Sprintf("configuration '%s' type mismatch: expected %s, got %s", key, fmt.Sprintf("%T", (*new(T))), fmt.Sprintf("%T", value))).
			WithCommand(ctx.Command, ctx.SubCommand).
			WithContext("key", key).
			WithContext("expected_type", fmt.Sprintf("%T", (*new(T)))).
			WithContext("actual_type", fmt.Sprintf("%T", value))
	}

	return SuccessWithData(result)
}

// MustGet retrieves a configuration value and panics on error
// Use when you expect the configuration to be valid and want to fail fast
func MustGet[T any](ctx *CommandContext, key string) T {
	result := Get[T](ctx, key)
	if result.Error != nil {
		panic(result.Error)
	}
	return GetValue[T](result)
}

// GetOr retrieves a configuration value or returns a default if not set
// Note: This function now also collects errors and returns the default
func GetOr[T any](ctx *CommandContext, key string, defaultValue T) T {
	result := Get[T](ctx, key)
	if result.Error != nil {
		// Error already collected in Get, return default value
		return defaultValue
	}
	return GetValue[T](result)
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
