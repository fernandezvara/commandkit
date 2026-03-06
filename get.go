// commandkit/get.go
package commandkit

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
)

type configProvider interface {
	getConfig() *Config
}

func (c *Config) getConfig() *Config {
	return c
}

func (ctx *CommandContext) getConfig() *Config {
	return ctx.Config
}

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

var (
	getErrors      []GetError
	getErrorsMutex sync.Mutex
	currentCommand string // Track current command for help display
)

// collectGetError adds an error to the global error collection
func collectGetError(c *Config, key, expectedType, actualType, message string, isSecret bool) {
	getErrorsMutex.Lock()
	defer getErrorsMutex.Unlock()

	// Get definition to extract flag and envVar info
	flag := ""
	envVar := ""
	if def, hasDef := c.definitions[key]; hasDef {
		flag = def.flag
		envVar = def.envVar
	}

	getErrors = append(getErrors, GetError{
		Key:          key,
		ExpectedType: expectedType,
		ActualType:   actualType,
		Message:      message,
		IsSecret:     isSecret,
		Flag:         flag,
		EnvVar:       envVar,
		config:       c,
	})
}

// displayGetErrorsAndExit shows all collected Get errors and exits with non-zero code
func displayGetErrorsAndExit() {
	getErrorsMutex.Lock()
	errs := make([]GetError, len(getErrors))
	command := currentCommand
	copy(errs, getErrors)
	getErrorsMutex.Unlock()

	fmt.Fprintf(os.Stderr, "Configuration errors detected:\n\n")

	// Sort errors alphabetically by display name for consistency with help
	sort.Slice(errs, func(i, j int) bool {
		displayNameI := getErrorDisplayName(errs[i], errs[i].config)
		displayNameJ := getErrorDisplayName(errs[j], errs[j].config)
		return displayNameI < displayNameJ
	})

	for _, err := range errs {
		displayName := getErrorDisplayName(err, err.config)

		if err.IsSecret || err.ExpectedType == "secret" {
			// Secret errors show simple "not defined" message
			fmt.Fprintf(os.Stderr, "  %s not defined\n", displayName)
		} else if err.ExpectedType == "validation" {
			// Validation errors show the validation message
			fmt.Fprintf(os.Stderr, "  %s validation failed: %s\n", displayName, err.Message)
		} else if err.ExpectedType == "not found" {
			fmt.Fprintf(os.Stderr, "  %s not defined\n", displayName)
		} else {
			fmt.Fprintf(os.Stderr, "  %s: expected %s, got %s\n", displayName, err.ExpectedType, err.ActualType)
		}
	}

	fmt.Fprintf(os.Stderr, "\n")

	// Show command help if we have a current command context
	if command != "" {
		fmt.Fprintf(os.Stderr, "Use '%s --help' for more information.\n", command)
	}

	os.Exit(1)
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

// hasCollectedGetErrors checks if there are any collected Get errors
func hasCollectedGetErrors() bool {
	getErrorsMutex.Lock()
	defer getErrorsMutex.Unlock()
	return len(getErrors) > 0
}

// logWarningForDesigner logs warnings for CLI designers about configuration issues
func logWarningForDesigner(message string) {
	// Use log package with a distinct prefix for designer warnings
	log.Printf("[CONFIG WARNING] %s", message)
}

// Get retrieves a configuration value with type safety using generics
// Returns nil for missing required data (designer responsibility to check)
func Get[T any](provider configProvider, key string) T {
	c := provider.getConfig()
	value, exists := c.values[key]
	if !exists {
		// Check if this is required data - if so, return nil for designer to handle
		if def, hasDef := c.definitions[key]; hasDef && def.required {
			// Log warning for designer but don't collect error
			logWarningForDesigner(fmt.Sprintf("Required key '%s' is not provided", key))
			var zero T
			return zero
		}
		// For non-required keys, collect error as before
		collectGetError(c, key, "not found", "", "key not defined", false)
		var zero T
		return zero
	}

	// Check if it's a secret (stored as string, needs special handling)
	def, hasDef := c.definitions[key]
	if hasDef && def.secret {
		collectGetError(c, key, "secret", "", "use GetSecret() instead", true)
		var zero T
		return zero
	}

	result, ok := value.(T)
	if !ok {
		collectGetError(c, key, fmt.Sprintf("%T", (*new(T))), fmt.Sprintf("%T", value), "type mismatch", false)
		var zero T
		return zero
	}

	return result
}

// MustGet is an alias for Get (both collect errors and exit)
func MustGet[T any](provider configProvider, key string) T {
	return Get[T](provider, key)
}

// GetOr retrieves a configuration value or returns a default if not set
// Note: This function now also collects errors and exits for consistency
func GetOr[T any](c *Config, key string, defaultValue T) T {
	value, exists := c.values[key]
	if !exists || value == nil {
		collectGetError(c, key, "not found", "", "key not defined", false)
		return defaultValue
	}

	result, ok := value.(T)
	if !ok {
		collectGetError(c, key, fmt.Sprintf("%T", defaultValue), fmt.Sprintf("%T", value), "type mismatch", false)
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
