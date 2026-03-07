package commandkit

import (
	"fmt"
)

// ErrorCategory represents the category of an error
type ErrorCategory int

const (
	ErrorCategoryValidation ErrorCategory = iota
	ErrorCategoryConfiguration
	ErrorCategoryRuntime
	ErrorCategorySystem
	ErrorCategoryUser
)

// String returns the string representation of the error category
func (c ErrorCategory) String() string {
	switch c {
	case ErrorCategoryValidation:
		return "validation"
	case ErrorCategoryConfiguration:
		return "configuration"
	case ErrorCategoryRuntime:
		return "runtime"
	case ErrorCategorySystem:
		return "system"
	case ErrorCategoryUser:
		return "user"
	default:
		return "unknown"
	}
}

// CommandError represents a unified error type for all command operations
type CommandError struct {
	Category   ErrorCategory
	Key        string
	Source     string // "env", "flag", "default", "runtime", "system", "user"
	Value      string // Masked if secret
	Message    string
	Command    string         // Command that generated this error
	SubCommand string         // Subcommand if applicable
	Context    map[string]any // Additional context
}

// Error implements the error interface
func (e *CommandError) Error() string {
	if e.Source == "" || e.Source == "none" {
		return fmt.Sprintf("%s: %s", e.Key, e.Message)
	}
	if e.Value == "" {
		return fmt.Sprintf("%s (%s): %s", e.Key, e.Source, e.Message)
	}
	return fmt.Sprintf("%s (%s=%s): %s", e.Key, e.Source, e.Value, e.Message)
}

// IsSecret returns true if the error is related to a secret value
func (e *CommandError) IsSecret() bool {
	return e.Source == "secret" || e.Category == ErrorCategoryConfiguration
}

// WithContext adds context to the error
func (e *CommandError) WithContext(key string, value any) *CommandError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// WithCommand sets the command context
func (e *CommandError) WithCommand(command, subcommand string) *CommandError {
	e.Command = command
	e.SubCommand = subcommand
	return e
}

// NewValidationError creates a new validation error
func NewValidationError(key, message string) *CommandError {
	return &CommandError{
		Category: ErrorCategoryValidation,
		Key:      key,
		Source:   "validation",
		Message:  message,
	}
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(key, source, value, message string) *CommandError {
	return &CommandError{
		Category: ErrorCategoryConfiguration,
		Key:      key,
		Source:   source,
		Value:    value,
		Message:  message,
	}
}

// NewRuntimeError creates a new runtime error
func NewRuntimeError(message string) *CommandError {
	return &CommandError{
		Category: ErrorCategoryRuntime,
		Source:   "runtime",
		Message:  message,
	}
}

// NewSystemError creates a new system error
func NewSystemError(message string) *CommandError {
	return &CommandError{
		Category: ErrorCategorySystem,
		Source:   "system",
		Message:  message,
	}
}

// NewUserError creates a new user error
func NewUserError(message string) *CommandError {
	return &CommandError{
		Category: ErrorCategoryUser,
		Source:   "user",
		Message:  message,
	}
}

// ConvertFromConfigError converts a legacy ConfigError to CommandError
func ConvertFromConfigError(configErr ConfigError, command string) *CommandError {
	return &CommandError{
		Category: ErrorCategoryConfiguration,
		Key:      configErr.Key,
		Source:   configErr.Source,
		Value:    configErr.Value,
		Message:  configErr.ErrorDescription,
		Command:  command,
	}
}

// ConvertFromGetError converts a legacy GetError to CommandError
func ConvertFromGetError(getErr GetError, command string) *CommandError {
	return &CommandError{
		Category: ErrorCategoryConfiguration,
		Key:      getErr.Key,
		Source:   "get",
		Value:    getErr.ActualType,
		Message:  getErr.Message,
		Command:  command,
		Context: map[string]any{
			"expected_type": getErr.ExpectedType,
			"is_secret":     getErr.IsSecret,
		},
	}
}
