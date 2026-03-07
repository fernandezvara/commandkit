// commandkit/get.go
package commandkit

import (
	"fmt"
	"log"
)

// GetError represents an error collected from Get functions
type GetError struct {
	Key              string
	ExpectedType     string
	ActualType       string
	Message          string
	IsSecret         bool
	Flag             string // Flag name (e.g., "port")
	EnvVar           string // Environment variable name (e.g., "PORT")
	Display          string
	ErrorDescription string
	config           *Config // Reference to config for definition lookup
}

// logWarningForDesigner logs warnings for CLI designers about configuration issues
func logWarningForDesigner(message string) {
	// Use log package with a distinct prefix for designer warnings
	log.Printf("[CONFIG WARNING] %s", message)
}

// getErrorDisplayName returns the display name matching help message format
func getErrorDisplayName(err GetError, c *Config) string {
	if c != nil {
		if def, hasDef := c.definitions[err.Key]; hasDef {
			return buildErrorDisplay(def)
		}
	}

	if err.Flag != "" {
		return fmt.Sprintf("--%s string", err.Flag)
	}
	if err.EnvVar != "" {
		return fmt.Sprintf("(no flag) string (env: %s)", err.EnvVar)
	}
	return fmt.Sprintf("%s string", err.Key)
}

// Get retrieves a configuration value and returns a CommandResult for unified error handling
// This function now performs early secret detection to prevent type assertion exposure
func Get[T any](ctx *CommandContext, key string) *CommandResult {
	// Check command config first (if it exists), then global config
	var c *Config
	if ctx.CommandConfig != nil {
		// Check if the key is defined in command config
		if _, hasDef := ctx.CommandConfig.definitions[key]; hasDef {
			c = ctx.CommandConfig
		} else {
			// Key not in command config, check global config
			c = ctx.GlobalConfig
		}
	} else {
		c = ctx.GlobalConfig
	}

	// Early secret detection - check definition before accessing values
	def, hasDef := c.definitions[key]
	if hasDef && def.secret {
		// Secure error handling - no secret exposure
		ctx.execution.CollectError(c, key, "secret", "", "use GetSecret() instead", true)
		return ValidationError(fmt.Sprintf("configuration '%s' is secret, use GetSecret() instead", key)).
			WithCommand(ctx.Command, ctx.SubCommand).
			WithContext("key", key).
			WithContext("expected_type", fmt.Sprintf("%T", (*new(T))))
	}

	value, exists := c.values[key]
	if !exists {
		// Check if this is required data - if so, return validation error
		if hasDef && def.required {
			// Log warning for designer and return validation error
			logWarningForDesigner(fmt.Sprintf("Required key '%s' is not provided", key))
			return ValidationError(fmt.Sprintf("required configuration '%s' not provided", key)).
				WithCommand(ctx.Command, ctx.SubCommand).
				WithContext("key", key).
				WithContext("expected_type", fmt.Sprintf("%T", (*new(T))))
		}
		// For non-required keys, collect error and return result
		ctx.execution.CollectError(c, key, "not found", "", "key not defined", false)
		return ErrorWithMessage(fmt.Errorf("configuration '%s' not found", key),
			fmt.Sprintf("configuration '%s' not found", key)).
			WithCommand(ctx.Command, ctx.SubCommand).
			WithContext("key", key)
	}

	// Additional safety check - ensure value is not a secret placeholder
	if strVal, ok := value.(string); ok && strVal == "[SECRET]" {
		// This should not happen with new implementation, but add safety check
		ctx.execution.CollectError(c, key, "secret", "", "use GetSecret() instead", true)
		return ValidationError(fmt.Sprintf("configuration '%s' is secret, use GetSecret() instead", key)).
			WithCommand(ctx.Command, ctx.SubCommand).
			WithContext("key", key).
			WithContext("is_secret", true)
	}

	result, ok := value.(T)
	if !ok {
		ctx.execution.CollectError(c, key, fmt.Sprintf("%T", (*new(T))), fmt.Sprintf("%T", value), "type mismatch", false)
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

// Has checks if a key exists and has a non-nil value
// Note: This function will return false for secret keys to prevent exposure
func (c *Config) Has(key string) bool {
	// Check if this is a secret key - if so, always return false to prevent exposure
	if def, hasDef := c.definitions[key]; hasDef && def.secret {
		return false
	}

	value, exists := c.values[key]
	return exists && value != nil
}

// GetSecret retrieves a secret value securely
func (c *Config) GetSecret(key string) *Secret {
	return c.secrets.Get(key)
}

// HasSecret checks if a secret exists and is set
func (c *Config) HasSecret(key string) bool {
	return c.secrets.Has(key)
}

// Keys returns all defined configuration keys
func (c *Config) Keys() []string {
	keys := make([]string, 0, len(c.definitions))
	for k := range c.definitions {
		keys = append(keys, k)
	}
	return keys
}
