// commandkit/command_executor_test.go
package commandkit

import (
	"errors"
	"strings"
	"testing"
)

func TestCommandExecutor_Execute_Success(t *testing.T) {
	executor := NewCommandExecutor()

	// Create a command with a simple function
	cmd := &Command{
		Name: "test",
		Func: func(ctx *CommandContext) error {
			ctx.Set("executed", true)
			return nil
		},
	}

	ctx := NewCommandContext([]string{}, New(), "test", "")
	services := NewCommandServices()

	// Execute the command
	result := executor.Execute(cmd, ctx, services)

	// Check that execution succeeded
	if result.Error != nil {
		t.Errorf("Execute() returned error: %v", result.Error)
	}

	// Check that command function was executed
	if val, exists := ctx.GetData("executed"); !exists || val != true {
		t.Error("Command function was not executed")
	}
}

func TestCommandExecutor_Execute_HelpRequest(t *testing.T) {
	executor := NewCommandExecutor()

	// Create a command
	cmd := &Command{
		Name: "test",
		Func: func(ctx *CommandContext) error {
			ctx.Set("executed", true)
			return nil
		},
	}

	// Create context with help request
	ctx := NewCommandContext([]string{"--help"}, New(), "test", "")
	services := NewCommandServices()

	// Execute the command
	result := executor.Execute(cmd, ctx, services)

	// Check that execution succeeded (help was shown)
	if result.Error != nil {
		t.Errorf("Execute() with help request returned error: %v", result.Error)
	}

	// Check that command function was NOT executed
	if _, exists := ctx.GetData("executed"); exists {
		t.Error("Command function should not have been executed when help was requested")
	}
}

func TestCommandExecutor_Execute_SubcommandsOnly(t *testing.T) {
	executor := NewCommandExecutor()

	// Create a command with subcommands but no function
	cmd := &Command{
		Name: "parent",
		SubCommands: map[string]*Command{
			"child": {
				Name:      "child",
				ShortHelp: "Child command",
			},
		},
	}

	ctx := NewCommandContext([]string{}, New(), "parent", "")
	services := NewCommandServices()

	// Execute the command
	result := executor.Execute(cmd, ctx, services)

	// Check that execution failed with subcommand help
	if result.Error == nil {
		t.Error("Execute() should have returned error for command with subcommands but no function")
	}

	// Check that error message contains subcommand help
	if !contains(result.Error.Error(), "Subcommands for") {
		t.Error("Error should contain subcommand help")
	}
}

func TestCommandExecutor_Execute_NoImplementation(t *testing.T) {
	executor := NewCommandExecutor()

	// Create a command with no function and no subcommands
	cmd := &Command{
		Name: "empty",
	}

	ctx := NewCommandContext([]string{}, New(), "empty", "")
	services := NewCommandServices()

	// Execute the command
	result := executor.Execute(cmd, ctx, services)

	// Check that execution failed
	if result.Error == nil {
		t.Error("Execute() should have returned error for command with no implementation")
	}

	expectedError := "command 'empty' has no implementation"
	if result.Error.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, result.Error.Error())
	}
}

func TestCommandExecutor_Execute_ConfigProcessing(t *testing.T) {
	executor := NewCommandExecutor()

	// Create a command with configuration
	cmd := &Command{
		Name: "configured",
		Func: func(ctx *CommandContext) error {
			// Check that config was processed
			portResult := Get[int64](ctx, "PORT")
			if portResult.Error != nil {
				return portResult.Error
			}
			if GetValue[int64](portResult) != 9000 {
				return errors.New("unexpected port value")
			}
			return nil
		},
		Definitions: map[string]*Definition{
			"PORT": {
				key:          "PORT",
				valueType:    TypeInt64,
				flag:         "port",
				description:  "Server port",
				defaultValue: 8080,
			},
		},
	}

	// Create context with flag
	ctx := NewCommandContext([]string{"--port", "9000"}, New(), "configured", "")
	services := NewCommandServices()

	// Execute the command
	result := executor.Execute(cmd, ctx, services)

	// Check that execution succeeded
	if result.Error != nil {
		t.Errorf("Execute() with config processing returned error: %v", result.Error)
	}
}

func TestCommandExecutor_Execute_ConfigError(t *testing.T) {
	executor := NewCommandExecutor()

	// Create a command with required configuration
	cmd := &Command{
		Name: "required",
		Func: func(ctx *CommandContext) error {
			return nil
		},
		Definitions: map[string]*Definition{
			"REQUIRED": {
				key:         "REQUIRED",
				valueType:   TypeString,
				flag:        "required",
				description: "Required flag",
				required:    true,
			},
		},
	}

	// Create context without required flag
	ctx := NewCommandContext([]string{}, New(), "required", "")
	services := NewCommandServices()

	// Execute the command
	result := executor.Execute(cmd, ctx, services)

	// Check that execution failed
	if result.Error == nil {
		t.Error("Execute() should have returned error for missing required configuration")
	}
}

func TestCommandExecutor_Execute_Middleware(t *testing.T) {
	executor := NewCommandExecutor()

	// Create middleware that tracks execution
	var middlewareExecuted bool

	middleware := func(next CommandFunc) CommandFunc {
		return func(ctx *CommandContext) error {
			middlewareExecuted = true
			return next(ctx)
		}
	}

	// Create a command with middleware
	cmd := &Command{
		Name: "middleware",
		Func: func(ctx *CommandContext) error {
			ctx.Set("executed", true)
			return nil
		},
		Middleware: []CommandMiddleware{middleware},
	}

	ctx := NewCommandContext([]string{}, New(), "middleware", "")
	services := NewCommandServices()

	// Execute the command
	result := executor.Execute(cmd, ctx, services)

	// Check that execution succeeded
	if result.Error != nil {
		t.Errorf("Execute() with middleware returned error: %v", result.Error)
	}

	// Check that middleware was executed
	if !middlewareExecuted {
		t.Error("Middleware was not executed")
	}

	// Check that command function was executed
	if val, exists := ctx.GetData("executed"); !exists || val != true {
		t.Error("Command function was not executed")
	}
}

func TestCommandExecutor_Execute_MiddlewareError(t *testing.T) {
	executor := NewCommandExecutor()

	// Create middleware that returns an error
	errorMiddleware := func(next CommandFunc) CommandFunc {
		return func(ctx *CommandContext) error {
			return errors.New("middleware error")
		}
	}

	// Create a command with error middleware
	cmd := &Command{
		Name: "error",
		Func: func(ctx *CommandContext) error {
			ctx.Set("executed", true)
			return nil
		},
		Middleware: []CommandMiddleware{errorMiddleware},
	}

	ctx := NewCommandContext([]string{}, New(), "error", "")
	services := NewCommandServices()

	// Execute the command
	result := executor.Execute(cmd, ctx, services)

	// Check that execution failed
	if result.Error == nil {
		t.Error("Execute() should have returned error when middleware failed")
	}

	if result.Error.Error() != "middleware error" {
		t.Errorf("Expected middleware error, got %v", result.Error)
	}

	// Check that command function was not executed
	if _, exists := ctx.GetData("executed"); exists {
		t.Error("Command function should not have been executed when middleware failed")
	}
}

func TestCommandExecutor_Execute_CommandError(t *testing.T) {
	executor := NewCommandExecutor()

	// Create a command that returns an error
	cmd := &Command{
		Name: "error",
		Func: func(ctx *CommandContext) error {
			return errors.New("command error")
		},
	}

	ctx := NewCommandContext([]string{}, New(), "error", "")
	services := NewCommandServices()

	// Execute the command
	result := executor.Execute(cmd, ctx, services)

	// Check that execution failed
	if result.Error == nil {
		t.Error("Execute() should have returned error when command failed")
	}

	if result.Error.Error() != "command error" {
		t.Errorf("Expected command error, got %v", result.Error)
	}
}

func TestCommandExecutor_Execute_CollectedErrors(t *testing.T) {
	executor := NewCommandExecutor()

	// Create a command that collects errors
	cmd := &Command{
		Name: "collect",
		Func: func(ctx *CommandContext) error {
			// Simulate collected errors with proper config reference
			config := getConfig(ctx)
			ctx.execution.CollectErrorWithConfig(config, "TEST", "string", "", "test error", false)
			return nil
		},
	}

	ctx := NewCommandContext([]string{}, New(), "collect", "")
	services := NewCommandServices()

	// Execute the command
	result := executor.Execute(cmd, ctx, services)

	// Check that execution returned appropriate result for collected errors
	// Note: ErrorWithExit sets Error=nil but ShouldExit=true with message
	if result.ShouldExit {
		// This is the expected behavior for collected errors
		if result.Message == "" {
			t.Error("Result should have message when ShouldExit is true")
		}
	} else {
		t.Error("Result should have ShouldExit flag set for collected errors")
	}

	// Check that error message contains the collected error
	if !strings.Contains(result.Message, "test error") {
		t.Errorf("Error message should contain the collected error, got: %s", result.Message)
	}
}

func TestCommandExecutor_Execute_NilParameters(t *testing.T) {
	executor := NewCommandExecutor()

	ctx := NewCommandContext([]string{}, New(), "test", "")
	services := NewCommandServices()

	// Test with nil command
	result := executor.Execute(nil, ctx, services)
	if result.Error == nil {
		t.Error("Execute() should return error for nil command")
	}

	// Test with nil context
	cmd := &Command{Name: "test"}
	result = executor.Execute(cmd, nil, services)
	if result.Error == nil {
		t.Error("Execute() should return error for nil context")
	}

	// Test with nil services
	result = executor.Execute(cmd, ctx, nil)
	if result.Error == nil {
		t.Error("Execute() should return error for nil services")
	}
}

func TestCommandExecutor_Execute_ExecutionContextInitialization(t *testing.T) {
	executor := NewCommandExecutor()

	// Create a command
	cmd := &Command{
		Name: "context",
		Func: func(ctx *CommandContext) error {
			// Check that execution context was initialized
			if ctx.execution == nil {
				return errors.New("execution context not initialized")
			}
			return nil
		},
	}

	// Create context without execution context
	ctx := NewCommandContext([]string{}, New(), "context", "")
	ctx.execution = nil // Ensure it's nil

	services := NewCommandServices()

	// Execute the command
	result := executor.Execute(cmd, ctx, services)

	// Check that execution succeeded
	if result.Error != nil {
		t.Errorf("Execute() failed: %v", result.Error)
	}
}

func TestCommandExecutor_Integration(t *testing.T) {
	// Test that CommandExecutor works correctly with the service factory
	services := NewCommandServices()
	executor := services.Executor

	// Create a comprehensive command
	cmd := &Command{
		Name: "integration",
		Func: func(ctx *CommandContext) error {
			// Check configuration
			portResult := Get[int64](ctx, "PORT")
			if portResult.Error != nil {
				return portResult.Error
			}

			// Check middleware context
			if _, exists := ctx.GetData("middleware_ran"); !exists {
				return errors.New("middleware did not run")
			}

			return nil
		},
		Definitions: map[string]*Definition{
			"PORT": {
				key:          "PORT",
				valueType:    TypeInt64,
				flag:         "port",
				description:  "Server port",
				defaultValue: 3000,
			},
		},
		Middleware: []CommandMiddleware{
			func(next CommandFunc) CommandFunc {
				return func(ctx *CommandContext) error {
					ctx.Set("middleware_ran", true)
					return next(ctx)
				}
			},
		},
	}

	ctx := NewCommandContext([]string{"--port", "5000"}, New(), "integration", "")

	// Execute the command
	result := executor.Execute(cmd, ctx, services)

	// Check that execution succeeded
	if result.Error != nil {
		t.Errorf("Integrated Execute failed: %v", result.Error)
	}

	// Verify configuration was processed
	portResult := Get[int64](ctx, "PORT")
	if GetValue[int64](portResult) != 5000 {
		t.Errorf("Expected PORT=5000, got %d", GetValue[int64](portResult))
	}
}
