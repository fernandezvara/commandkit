// commandkit/execution_context.go
package commandkit

import (
	"fmt"
	"os"
	"sort"
	"sync"
)

// ExecutionContext provides thread-safe error collection for command execution
type ExecutionContext struct {
	errors  []GetError
	command string
	mu      sync.Mutex
}

// NewExecutionContext creates a new execution context for the specified command
func NewExecutionContext(command string) *ExecutionContext {
	return &ExecutionContext{
		errors:  make([]GetError, 0),
		command: command,
	}
}

// CollectError adds an error to the execution context with thread safety
func (ctx *ExecutionContext) CollectError(key, expectedType, actualType, message string, isSecret bool) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// Extract flag and envVar info from config if available
	// Note: We'll need to pass config reference for complete error info
	err := GetError{
		Key:          key,
		ExpectedType: expectedType,
		ActualType:   actualType,
		Message:      message,
		IsSecret:     isSecret,
		Flag:         "",  // Will be populated by caller if needed
		EnvVar:       "",  // Will be populated by caller if needed
		config:       nil, // Will be populated by caller if needed
	}
	ctx.errors = append(ctx.errors, err)
}

// CollectErrorWithConfig adds an error to the execution context with full config information
func (ctx *ExecutionContext) CollectErrorWithConfig(c *Config, key, expectedType, actualType, message string, isSecret bool) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// Get definition to extract flag and envVar info
	flag := ""
	envVar := ""
	if def, hasDef := c.definitions[key]; hasDef {
		flag = def.flag
		envVar = def.envVar
	}

	err := GetError{
		Key:              key,
		ExpectedType:     expectedType,
		ActualType:       actualType,
		Message:          message,
		IsSecret:         isSecret,
		Flag:             flag,
		EnvVar:           envVar,
		Display:          getErrorDisplayName(GetError{Key: key, Flag: flag, EnvVar: envVar}, c),
		ErrorDescription: message,
		config:           c,
	}
	ctx.errors = append(ctx.errors, err)
}

// CollectConfigError adds a normalized config error to the execution context
func (ctx *ExecutionContext) CollectConfigError(c *Config, configErr ConfigError) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	flag := ""
	envVar := ""
	if c != nil {
		if def, hasDef := c.definitions[configErr.Key]; hasDef {
			flag = def.flag
			envVar = def.envVar
		}
	}

	err := GetError{
		Key:              configErr.Key,
		ExpectedType:     "validation",
		ActualType:       "",
		Message:          configErr.Message,
		IsSecret:         false,
		Flag:             flag,
		EnvVar:           envVar,
		Display:          configErr.Display,
		ErrorDescription: configErr.ErrorDescription,
		config:           c,
	}
	ctx.errors = append(ctx.errors, err)
}

// HasErrors returns true if there are collected errors
func (ctx *ExecutionContext) HasErrors() bool {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return len(ctx.errors) > 0
}

// GetErrors returns a copy of all collected errors
func (ctx *ExecutionContext) GetErrors() []GetError {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	errs := make([]GetError, len(ctx.errors))
	copy(errs, ctx.errors)
	return errs
}

// GetFormattedErrors returns all collected errors as a simplified fallback string
func (ctx *ExecutionContext) GetFormattedErrors() string {
	errs := ctx.GetErrors()
	if len(errs) == 0 {
		return ""
	}

	result := "Configuration errors:\n"
	for _, err := range errs {
		display := err.Display
		if display == "" && err.config != nil {
			display = getErrorDisplayName(err, err.config)
		}
		description := err.ErrorDescription
		if description == "" {
			description = err.Message
		}
		result += fmt.Sprintf("  %s -> %s\n", display, description)
	}

	return result
}

// GetCommand returns the command name for this context
func (ctx *ExecutionContext) GetCommand() string {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ctx.command
}

// SetCommand updates the command name for this context
func (ctx *ExecutionContext) SetCommand(command string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.command = command
}

// DisplayAndExit shows all collected errors and exits with non-zero code
func (ctx *ExecutionContext) DisplayAndExit() {
	errs := ctx.GetErrors()
	if len(errs) == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, "Configuration errors detected:\n\n")

	// Sort errors alphabetically by display name for consistency with help
	sort.Slice(errs, func(i, j int) bool {
		displayNameI := errs[i].Display
		if displayNameI == "" && errs[i].config != nil {
			displayNameI = getErrorDisplayName(errs[i], errs[i].config)
		}
		displayNameJ := errs[j].Display
		if displayNameJ == "" && errs[j].config != nil {
			displayNameJ = getErrorDisplayName(errs[j], errs[j].config)
		}
		return displayNameI < displayNameJ
	})

	for _, err := range errs {
		displayName := err.Display
		if displayName == "" && err.config != nil {
			displayName = getErrorDisplayName(err, err.config)
		}

		if err.IsSecret || err.ExpectedType == "secret" {
			fmt.Fprintf(os.Stderr, "  %s not defined\n", displayName)
		} else if err.ExpectedType == "validation" {
			fmt.Fprintf(os.Stderr, "  %s validation failed: %s\n", displayName, err.Message)
		} else if err.ExpectedType == "not found" {
			fmt.Fprintf(os.Stderr, "  %s not defined\n", displayName)
		} else if err.ErrorDescription != "" {
			fmt.Fprintf(os.Stderr, "  %s %s\n", displayName, err.ErrorDescription)
		} else {
			fmt.Fprintf(os.Stderr, "  %s: expected %s, got %s\n", displayName, err.ExpectedType, err.ActualType)
		}
	}

	fmt.Fprintf(os.Stderr, "\n")

	command := ctx.GetCommand()
	if command != "" {
		fmt.Fprintf(os.Stderr, "Use '%s --help' for more information.\n", command)
	}

	os.Exit(1)
}

// convertToConfigErrors converts GetError slice to ConfigError slice
func (ctx *ExecutionContext) convertToConfigErrors(getErrs []GetError) []ConfigError {
	configErrs := make([]ConfigError, len(getErrs))
	for i, err := range getErrs {
		configErrs[i] = ConfigError{
			Key:              err.Key,
			Source:           "validation",
			Value:            "",
			Message:          err.Message,
			Display:          err.Display,
			ErrorDescription: err.ErrorDescription,
		}
	}
	return configErrs
}

// Clear removes all collected errors (useful for testing)
func (ctx *ExecutionContext) Clear() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.errors = nil
}
