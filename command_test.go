package commandkit

import (
	"fmt"
	"strings"
	"testing"
)

func TestCommandDefinition(t *testing.T) {
	cfg := New()

	// Define a simple command
	cfg.Command("start").
		Func(startCommand).
		ShortHelp("Start the service").
		LongHelp("Start the service with all components initialized").
		Aliases("s", "run").
		Config(func(cc *CommandConfig) {
			cc.Define("PORT").Int64().Flag("port").Default(8080).Range(1, 65535)
			cc.Define("DAEMON").Bool().Flag("daemon").Default(false)
		})

	// Check command was registered
	cmd, exists := cfg.commands["start"]
	if !exists {
		t.Fatal("Command not registered")
	}

	if cmd.Name != "start" {
		t.Errorf("Expected command name 'start', got '%s'", cmd.Name)
	}

	if cmd.ShortHelp != "Start the service" {
		t.Errorf("Expected short help 'Start the service', got '%s'", cmd.ShortHelp)
	}

	if cmd.LongHelp != "Start the service with all components initialized" {
		t.Errorf("Expected long help 'Start the service with all components initialized', got '%s'", cmd.LongHelp)
	}

	if len(cmd.Aliases) != 2 || cmd.Aliases[0] != "s" || cmd.Aliases[1] != "run" {
		t.Errorf("Expected aliases ['s', 'run'], got %v", cmd.Aliases)
	}

	// Check command-specific configuration
	if len(cmd.Definitions) != 2 {
		t.Errorf("Expected 2 command definitions, got %d", len(cmd.Definitions))
	}

	if portDef, exists := cmd.Definitions["PORT"]; !exists {
		t.Error("PORT definition not found in command")
	} else if portDef.defaultValue != 8080 {
		t.Errorf("Expected PORT default 8080, got %v", portDef.defaultValue)
	}
}

func TestCommandExecution(t *testing.T) {
	cfg := New()

	// Define a command
	cfg.Command("echo").
		Func(echoCommand).
		ShortHelp("Echo arguments").
		Config(func(cc *CommandConfig) {
			cc.Define("UPPERCASE").Bool().Flag("uppercase").Default(false)
		})

	// Test execution
	ctx := NewCommandContext([]string{"hello", "world"}, cfg, "echo", "")

	// Set up flag parsing for test
	cfg.flagSet.Bool("uppercase", false, "Convert to uppercase")
	cfg.flagSet.Parse([]string{"--uppercase=false"})
	ctx.Flags["uppercase"] = "false"

	err := cfg.commands["echo"].Execute(ctx)
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}
}

func TestSubCommands(t *testing.T) {
	cfg := New()

	// Define a command with subcommands
	cfg.Command("start").
		Func(startCommand).
		ShortHelp("Start the service").
		SubCommand("server").
		Func(startServerCommand).
		ShortHelp("Start server only").
		Aliases("srv").
		SubCommand("worker").
		Func(startWorkerCommand).
		ShortHelp("Start worker only").
		Aliases("wrk")

	// Test subcommand finding
	cmd := cfg.commands["start"]

	serverCmd := cmd.FindSubCommand("server")
	if serverCmd == nil {
		t.Fatal("Server subcommand not found")
	}

	if serverCmd.ShortHelp != "Start server only" {
		t.Errorf("Expected server short help 'Start server only', got '%s'", serverCmd.ShortHelp)
	}

	// Test alias finding
	serverCmd2 := cmd.FindSubCommand("srv")
	if serverCmd2 == nil {
		t.Fatal("Server subcommand not found by alias")
	}

	// Test non-existent subcommand
	nonExistent := cmd.FindSubCommand("nonexistent")
	if nonExistent != nil {
		t.Error("Should return nil for non-existent subcommand")
	}
}

func TestCommandHelp(t *testing.T) {
	cfg := New()

	// Define a command with help
	cfg.Command("start").
		Func(startCommand).
		ShortHelp("Start the service").
		LongHelp("This is a detailed help text for the start command.\nIt explains how to use the command.").
		Config(func(cc *CommandConfig) {
			cc.Define("PORT").Int64().Flag("port").Default(8080).Description("Server port")
			cc.Define("DAEMON").Bool().Flag("daemon").Default(false).Description("Run in background")
		})

	cmd := cfg.commands["start"]
	help := cmd.GetHelp()

	// Check that help contains expected content
	if !contains(help, "This is a detailed help text") {
		t.Error("Long help not found in help text")
	}

	if !contains(help, "Server port") {
		t.Error("PORT description not found in help text")
	}

	if !contains(help, "--port") {
		t.Error("Port flag not found in help text")
	}

	if !contains(help, "Run in background") {
		t.Error("DAEMON description not found in help text")
	}
}

func TestCommandSuggestions(t *testing.T) {
	cfg := New()

	// Define some commands
	cfg.Command("start").Func(startCommand).ShortHelp("Start")
	cfg.Command("stop").Func(stopCommand).ShortHelp("Stop")
	cfg.Command("restart").Func(restartCommand).ShortHelp("Restart")

	// Test suggestions
	suggestions := cfg.findSuggestions("stat")
	if !contains(suggestions, "start") {
		t.Error("Should suggest 'start' for 'stat'")
	}

	suggestions = cfg.findSuggestions("stp")
	if !contains(suggestions, "stop") {
		t.Error("Should suggest 'stop' for 'stp'")
	}

	suggestions = cfg.findSuggestions("restart")
	if !contains(suggestions, "restart") {
		t.Error("Should suggest 'restart' for exact match")
	}

	// Test non-matching
	suggestions = cfg.findSuggestions("xyz")
	if suggestions != "no similar commands found" {
		t.Errorf("Expected 'no similar commands found' for 'xyz', got '%s'", suggestions)
	}
}

func TestCommandMiddleware(t *testing.T) {
	cfg := New()

	// Define a command with middleware
	cfg.Command("test").
		Func(testCommand).
		ShortHelp("Test command").
		Middleware(loggingMiddleware).
		Middleware(authMiddleware)

	cmd := cfg.commands["test"]

	if len(cmd.Middleware) != 2 {
		t.Errorf("Expected 2 middleware, got %d", len(cmd.Middleware))
	}
}

// Test helper functions
func startCommand(ctx *CommandContext) error {
	fmt.Printf("Starting service\n")
	return nil
}

func stopCommand(ctx *CommandContext) error {
	fmt.Printf("Stopping service\n")
	return nil
}

func restartCommand(ctx *CommandContext) error {
	fmt.Printf("Restarting service\n")
	return nil
}

func startServerCommand(ctx *CommandContext) error {
	fmt.Printf("Starting server\n")
	return nil
}

func startWorkerCommand(ctx *CommandContext) error {
	fmt.Printf("Starting worker\n")
	return nil
}

func echoCommand(ctx *CommandContext) error {
	for _, arg := range ctx.Args {
		fmt.Printf("Echo: %s\n", arg)
	}
	return nil
}

func testCommand(ctx *CommandContext) error {
	fmt.Printf("Test command executed\n")
	return nil
}

func loggingMiddleware(next CommandFunc) CommandFunc {
	return func(ctx *CommandContext) error {
		fmt.Printf("Logging: Executing %s\n", ctx.Command)
		return next(ctx)
	}
}

func authMiddleware(next CommandFunc) CommandFunc {
	return func(ctx *CommandContext) error {
		fmt.Printf("Auth: Checking permissions for %s\n", ctx.Command)
		return next(ctx)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(strings.HasPrefix(s, substr) || strings.HasSuffix(s, substr) || strings.Contains(s, substr))))
}

func TestGlobalMiddleware(t *testing.T) {
	cfg := New()

	// Track middleware execution
	var executionOrder []string

	// Add global middleware
	cfg.UseMiddleware(func(next CommandFunc) CommandFunc {
		return func(ctx *CommandContext) error {
			executionOrder = append(executionOrder, "global1")
			return next(ctx)
		}
	})

	cfg.UseMiddleware(func(next CommandFunc) CommandFunc {
		return func(ctx *CommandContext) error {
			executionOrder = append(executionOrder, "global2")
			return next(ctx)
		}
	})

	// Define a command
	cfg.Command("test").
		Func(func(ctx *CommandContext) error {
			executionOrder = append(executionOrder, "command")
			return nil
		}).
		ShortHelp("Test command")

	// Execute
	err := cfg.Execute([]string{"app", "test"})
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	// Check execution order
	expected := []string{"global1", "global2", "command"}
	if len(executionOrder) != len(expected) {
		t.Fatalf("Expected %d executions, got %d: %v", len(expected), len(executionOrder), executionOrder)
	}

	for i, exp := range expected {
		if executionOrder[i] != exp {
			t.Errorf("Expected execution[%d] = %s, got %s", i, exp, executionOrder[i])
		}
	}
}

func TestUseMiddlewareForCommands(t *testing.T) {
	cfg := New()

	// Track middleware execution
	var executedFor []string

	// Add middleware only for specific commands
	cfg.UseMiddlewareForCommands([]string{"start", "stop"}, func(next CommandFunc) CommandFunc {
		return func(ctx *CommandContext) error {
			executedFor = append(executedFor, ctx.Command)
			return next(ctx)
		}
	})

	// Define commands
	cfg.Command("start").
		Func(func(ctx *CommandContext) error { return nil }).
		ShortHelp("Start")

	cfg.Command("stop").
		Func(func(ctx *CommandContext) error { return nil }).
		ShortHelp("Stop")

	cfg.Command("status").
		Func(func(ctx *CommandContext) error { return nil }).
		ShortHelp("Status")

	// Execute start - should trigger middleware
	executedFor = nil
	err := cfg.Execute([]string{"app", "start"})
	if err != nil {
		t.Fatalf("start execution failed: %v", err)
	}
	if len(executedFor) != 1 || executedFor[0] != "start" {
		t.Errorf("Expected middleware for 'start', got %v", executedFor)
	}

	// Execute stop - should trigger middleware
	executedFor = nil
	err = cfg.Execute([]string{"app", "stop"})
	if err != nil {
		t.Fatalf("stop execution failed: %v", err)
	}
	if len(executedFor) != 1 || executedFor[0] != "stop" {
		t.Errorf("Expected middleware for 'stop', got %v", executedFor)
	}

	// Execute status - should NOT trigger middleware
	executedFor = nil
	err = cfg.Execute([]string{"app", "status"})
	if err != nil {
		t.Fatalf("status execution failed: %v", err)
	}
	if len(executedFor) != 0 {
		t.Errorf("Expected no middleware for 'status', got %v", executedFor)
	}
}

func TestUseMiddlewareForSubcommands(t *testing.T) {
	cfg := New()

	// Track middleware execution
	var executedFor []string

	// Add middleware only for specific subcommands
	cfg.UseMiddlewareForSubcommands("start", []string{"worker"}, func(next CommandFunc) CommandFunc {
		return func(ctx *CommandContext) error {
			executedFor = append(executedFor, ctx.SubCommand)
			return next(ctx)
		}
	})

	// Define command with subcommands - need to add them separately to the same parent
	startCmd := cfg.Command("start").
		Func(func(ctx *CommandContext) error { return nil }).
		ShortHelp("Start")

	// Add server subcommand
	startCmd.SubCommand("server").
		Func(func(ctx *CommandContext) error { return nil }).
		ShortHelp("Start server")

	// Add worker subcommand (to start, not to server)
	startCmd.SubCommand("worker").
		Func(func(ctx *CommandContext) error { return nil }).
		ShortHelp("Start worker")

	// Execute start worker - should trigger middleware
	executedFor = nil
	err := cfg.Execute([]string{"app", "start", "worker"})
	if err != nil {
		t.Fatalf("start worker execution failed: %v", err)
	}
	if len(executedFor) != 1 || executedFor[0] != "worker" {
		t.Errorf("Expected middleware for 'worker', got %v", executedFor)
	}

	// Execute start server - should NOT trigger middleware
	executedFor = nil
	err = cfg.Execute([]string{"app", "start", "server"})
	if err != nil {
		t.Fatalf("start server execution failed: %v", err)
	}
	if len(executedFor) != 0 {
		t.Errorf("Expected no middleware for 'server', got %v", executedFor)
	}
}
