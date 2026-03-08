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

// success creates a successful command result
func success() *CommandResult {
	return &CommandResult{
		ExitCode:   0,
		ShouldExit: false,
	}
}

// successWithData creates a successful command result with data
func successWithData(data any) *CommandResult {
	return &CommandResult{
		ExitCode:   0,
		ShouldExit: false,
		Data:       data,
	}
}

// successWithMessage creates a successful command result with a message
func successWithMessage(message string) *CommandResult {
	return &CommandResult{
		ExitCode:   0,
		ShouldExit: false,
		Message:    message,
	}
}

// errorResult creates an error command result
func errorResult(err error) *CommandResult {
	return &CommandResult{
		Error:      err,
		ExitCode:   1,
		ShouldExit: false,
	}
}

// errorWithMessage creates an error command result with a custom message
func errorWithMessage(err error, message string) *CommandResult {
	return &CommandResult{
		Error:      err,
		ExitCode:   1,
		ShouldExit: false,
		Message:    message,
	}
}

// errorWithExit creates an error command result that should exit
func errorWithExit(err error, message string) *CommandResult {
	return &CommandResult{
		Error:      err,
		ExitCode:   1,
		ShouldExit: true,
		Message:    message,
	}
}

// validationError creates a validation error result that should exit
func validationError(message string) *CommandResult {
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

// configErrorResult creates a configuration error result that should exit
func configErrorResult(message string) *CommandResult {
	return &CommandResult{
		Error:      fmt.Errorf("configuration error: %s", message),
		ExitCode:   1,
		ShouldExit: true,
		Message:    message,
	}
}
