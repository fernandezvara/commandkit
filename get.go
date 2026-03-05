// commandkit/get.go
package commandkit

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// GetError represents an error collected from Get functions
type GetError struct {
	Key          string
	ExpectedType string
	ActualType   string
	Message      string
	IsSecret     bool
}

var (
	getErrors      []GetError
	getErrorsMutex sync.Mutex
	currentCommand string // Track current command for help display
)

// collectGetError adds an error to the global error collection
func collectGetError(key, expectedType, actualType, message string, isSecret bool) {
	getErrorsMutex.Lock()
	defer getErrorsMutex.Unlock()

	getErrors = append(getErrors, GetError{
		Key:          key,
		ExpectedType: expectedType,
		ActualType:   actualType,
		Message:      message,
		IsSecret:     isSecret,
	})
}

// triggerExitWithErrors displays collected errors and exits
func triggerExitWithErrors() {
	getErrorsMutex.Lock()
	hasErrors := len(getErrors) > 0
	getErrorsMutex.Unlock()

	if hasErrors {
		displayGetErrorsAndExit()
	}
}

// displayGetErrorsAndExit shows all collected Get errors and exits with non-zero code
func displayGetErrorsAndExit() {
	getErrorsMutex.Lock()
	errs := make([]GetError, len(getErrors))
	command := currentCommand
	copy(errs, getErrors)
	getErrorsMutex.Unlock()

	fmt.Fprintf(os.Stderr, "Configuration errors detected:\n\n")

	for _, err := range errs {
		if err.IsSecret {
			fmt.Fprintf(os.Stderr, "  %s: %s (use GetSecret() for secrets)\n", err.Key, err.Message)
		} else if err.ExpectedType == "not found" {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", err.Key, err.Message)
		} else {
			fmt.Fprintf(os.Stderr, "  %s: expected %s, got %s\n", err.Key, err.ExpectedType, err.ActualType)
		}
	}

	fmt.Fprintf(os.Stderr, "\n")

	// Show command help if we have a current command context
	if command != "" {
		// Try to get command help from available sources
		if command == "help" {
			fmt.Fprintf(os.Stderr, "Use '%s --help' for more information.\n", command)
		} else {
			fmt.Fprintf(os.Stderr, "Use '%s --help' for more information.\n", command)
		}
	}

	os.Exit(1)
}

// SetCurrentCommand sets the current command context for error display
func SetCurrentCommand(command string) {
	getErrorsMutex.Lock()
	defer getErrorsMutex.Unlock()
	currentCommand = command
}

// ClearGetErrors clears the collected Get errors (for testing)
func ClearGetErrors() {
	getErrorsMutex.Lock()
	defer getErrorsMutex.Unlock()
	getErrors = nil
	currentCommand = ""
}

// GetCollectedErrors returns a copy of collected Get errors (for testing)
func GetCollectedErrors() []GetError {
	getErrorsMutex.Lock()
	defer getErrorsMutex.Unlock()
	errs := make([]GetError, len(getErrors))
	copy(errs, getErrors)
	return errs
}

// Get retrieves a configuration value with type safety using generics
func Get[T any](c *Config, key string) T {
	value, exists := c.values[key]
	if !exists {
		collectGetError(key, "not found", "", "key not defined", false)
		triggerExitWithErrors()
		var zero T
		return zero
	}

	// Check if it's a secret (stored as string, needs special handling)
	def, hasDef := c.definitions[key]
	if hasDef && def.secret {
		collectGetError(key, "secret", "", "use GetSecret() instead", true)
		triggerExitWithErrors()
		var zero T
		return zero
	}

	result, ok := value.(T)
	if !ok {
		collectGetError(key, fmt.Sprintf("%T", (*new(T))), fmt.Sprintf("%T", value), "type mismatch", false)
		triggerExitWithErrors()
		var zero T
		return zero
	}

	return result
}

// MustGet is an alias for Get (both collect errors and exit)
func MustGet[T any](c *Config, key string) T {
	return Get[T](c, key)
}

// GetOr retrieves a configuration value or returns a default if not set
// Note: This function now also collects errors and exits for consistency
func GetOr[T any](c *Config, key string, defaultValue T) T {
	value, exists := c.values[key]
	if !exists || value == nil {
		collectGetError(key, "not found", "", "key not defined", false)
		triggerExitWithErrors()
		return defaultValue
	}

	result, ok := value.(T)
	if !ok {
		collectGetError(key, fmt.Sprintf("%T", defaultValue), fmt.Sprintf("%T", value), "type mismatch", false)
		triggerExitWithErrors()
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
