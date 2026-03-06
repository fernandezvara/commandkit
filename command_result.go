package commandkit

import (
	"fmt"
	"os"
)

// CommandResult represents the result of command execution with unified error handling
type CommandResult struct {
	Error      error
	ExitCode   int
	ShouldExit bool
	Message    string
	Data       any            // Optional data for successful results
	Context    map[string]any // Context information
}

// Handle processes the command result according to its state
func (r *CommandResult) Handle() {
	if r.ShouldExit {
		r.displayAndExit()
	}
}

// displayAndExit displays the result message and exits with the appropriate code
func (r *CommandResult) displayAndExit() {
	if r.Message != "" {
		fmt.Fprintln(os.Stderr, r.Message)
	}
	os.Exit(r.ExitCode)
}

// GetValue extracts the value from a successful CommandResult
func (r *CommandResult) GetValue() any {
	if r.Error != nil {
		panic("cannot get value from failed CommandResult")
	}
	return r.Data
}

// GetValueTyped extracts a typed value from a successful CommandResult
func GetValue[T any](r *CommandResult) T {
	if r.Error != nil {
		panic("cannot get value from failed CommandResult")
	}
	if r.Data == nil {
		var zero T
		return zero
	}
	return r.Data.(T)
}

// Success creates a successful command result
func Success() *CommandResult {
	return &CommandResult{
		ExitCode:   0,
		ShouldExit: false,
	}
}

// SuccessWithData creates a successful command result with data
func SuccessWithData(data any) *CommandResult {
	return &CommandResult{
		ExitCode:   0,
		ShouldExit: false,
		Data:       data,
	}
}

// SuccessWithMessage creates a successful command result with a message
func SuccessWithMessage(message string) *CommandResult {
	return &CommandResult{
		ExitCode:   0,
		ShouldExit: false,
		Message:    message,
	}
}

// Error creates an error command result
func Error(err error) *CommandResult {
	return &CommandResult{
		Error:      err,
		ExitCode:   1,
		ShouldExit: false,
	}
}

// ErrorWithMessage creates an error command result with a custom message
func ErrorWithMessage(err error, message string) *CommandResult {
	return &CommandResult{
		Error:      err,
		ExitCode:   1,
		ShouldExit: false,
		Message:    message,
	}
}

// ErrorWithExit creates an error command result that should exit
func ErrorWithExit(err error, message string) *CommandResult {
	return &CommandResult{
		Error:      err,
		ExitCode:   1,
		ShouldExit: true,
		Message:    message,
	}
}

// ValidationError creates a validation error result that should exit
func ValidationError(message string) *CommandResult {
	return &CommandResult{
		Error:      fmt.Errorf("validation error: %s", message),
		ExitCode:   1,
		ShouldExit: true,
		Message:    message,
	}
}

// WithContext adds context to the CommandResult
func (r *CommandResult) WithContext(key string, value any) *CommandResult {
	if r.Context == nil {
		r.Context = make(map[string]any)
	}
	r.Context[key] = value
	return r
}

// WithCommand sets the command context in the CommandResult
func (r *CommandResult) WithCommand(command, subcommand string) *CommandResult {
	if r.Context == nil {
		r.Context = make(map[string]any)
	}
	r.Context["command"] = command
	r.Context["subcommand"] = subcommand
	return r
}

// ConfigErrorResult creates a configuration error result that should exit
func ConfigErrorResult(message string) *CommandResult {
	return &CommandResult{
		Error:      fmt.Errorf("configuration error: %s", message),
		ExitCode:   1,
		ShouldExit: true,
		Message:    message,
	}
}
