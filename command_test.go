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

	result := cfg.commands["echo"].Execute(ctx)
	if result.Error != nil {
		t.Fatalf("Command execution failed: %v", result.Error)
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

	// Use new help system to get command help
	helpIntegration := NewHelpIntegration()
	help, err := helpIntegration.GenerateHelp([]string{"start", "--help"}, cfg.commands)
	if err != nil {
		t.Fatalf("Failed to generate help: %v", err)
	}

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

func TestShowGlobalHelp(t *testing.T) {
	cfg := New()
	cfg.Command("start").Func(startCommand).ShortHelp("Start the service").Aliases("run")
	cfg.Command("stop").Func(stopCommand).ShortHelp("Stop the service")

	output := captureStdout(t, func() {
		if err := cfg.ShowGlobalHelp(); err != nil {
			t.Fatalf("ShowGlobalHelp failed: %v", err)
		}
	})

	if !strings.Contains(output, "Usage:") {
		t.Fatalf("expected usage in output, got: %s", output)
	}
	if !strings.Contains(output, "Available commands:") {
		t.Fatalf("expected commands heading in output, got: %s", output)
	}
	if !strings.Contains(output, "start") || !strings.Contains(output, "Start the service") {
		t.Fatalf("expected start command in output, got: %s", output)
	}
	if !strings.Contains(output, "aliases: run") {
		t.Fatalf("expected aliases in output, got: %s", output)
	}
	if !strings.Contains(output, "<command> --help") {
		t.Fatalf("expected command help hint in output, got: %s", output)
	}
}

func TestShowCommandHelp(t *testing.T) {
	cfg := New()
	cfg.Command("deploy").
		Func(testCommand).
		ShortHelp("Deploy the service").
		LongHelp("Deploys the current release.").
		Config(func(cc *CommandConfig) {
			cc.Define("ENV").String().Flag("env").Required().Description("Target environment")
			cc.Define("FORCE").Bool().Flag("force").Default(false).Description("Force deployment")
		})

	output := captureStdout(t, func() {
		if err := cfg.ShowCommandHelp("deploy"); err != nil {
			t.Fatalf("ShowCommandHelp failed: %v", err)
		}
	})

	if !strings.Contains(output, "Usage:") || !strings.Contains(output, "deploy [options]") {
		t.Fatalf("expected deploy usage in output, got: %s", output)
	}
	if !strings.Contains(output, "Deploys the current release.") {
		t.Fatalf("expected long help in output, got: %s", output)
	}
	if !strings.Contains(output, "--env") || !strings.Contains(output, "Target environment") {
		t.Fatalf("expected env option in output, got: %s", output)
	}
	if !strings.Contains(output, "--force") || !strings.Contains(output, "Force deployment") {
		t.Fatalf("expected force option in output, got: %s", output)
	}
}

func TestShowCommandHelpUnknownCommand(t *testing.T) {
	cfg := New()

	err := cfg.ShowCommandHelp("missing")
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if err.Error() != "unknown command: missing" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigExecuteHelpPaths(t *testing.T) {
	cfg := New()
	cfg.Command("deploy").Func(testCommand).ShortHelp("Deploy service")

	globalOutput := captureStdout(t, func() {
		if err := cfg.Execute([]string{"app"}); err != nil {
			t.Fatalf("Execute without command failed: %v", err)
		}
	})

	if !strings.Contains(globalOutput, "Available commands:") || !strings.Contains(globalOutput, "deploy") {
		t.Fatalf("expected global help output, got: %s", globalOutput)
	}

	commandOutput := captureStdout(t, func() {
		if err := cfg.Execute([]string{"app", "help", "deploy"}); err != nil {
			t.Fatalf("Execute help deploy failed: %v", err)
		}
	})

	if !strings.Contains(commandOutput, "Deploy service") || !strings.Contains(commandOutput, "Usage:") {
		t.Fatalf("expected command help output, got: %s", commandOutput)
	}
}

func TestConfigExecuteUnknownCommand(t *testing.T) {
	cfg := New()
	cfg.Command("start").Func(startCommand).ShortHelp("Start")

	err := cfg.Execute([]string{"app", "stat"})
	if err == nil {
		t.Fatal("expected unknown command error")
	}
	if !strings.Contains(err.Error(), `unknown command: "stat"`) {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "Did you mean: start?") {
		t.Fatalf("expected suggestion in error: %v", err)
	}
}

func TestValidateRequiredFlagsLogsWarning(t *testing.T) {
	cmd := NewCommand("deploy")
	cmd.Definitions["BASE_URL"] = &Definition{
		key:       "BASE_URL",
		valueType: TypeString,
		envVar:    "BASE_URL",
		flag:      "base-url",
		required:  true,
	}

	ctx := NewCommandContext(nil, New(), "deploy", "")
	ctx.GlobalConfig.flagValues = make(map[string]*string)

	logs := captureLogs(t, func() {
		validateRequiredFlags(cmd, ctx)
	})

	if !strings.Contains(logs, "[CONFIG WARNING]") {
		t.Fatalf("expected config warning log, got: %s", logs)
	}
	if !strings.Contains(logs, "--base-url (env: BASE_URL)") {
		t.Fatalf("expected display name in warning, got: %s", logs)
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

func TestCommandWithNilFuncAndSubcommands(t *testing.T) {
	cfg := New()

	// Define a command with subcommands but no Func
	cfg.Command("user").
		ShortHelp("User management commands").
		SubCommand("create").
		Func(func(ctx *CommandContext) error { return nil }).
		ShortHelp("Create a new user").
		SubCommand("update").
		Func(func(ctx *CommandContext) error { return nil }).
		ShortHelp("Update an existing user").
		SubCommand("list").
		Func(func(ctx *CommandContext) error { return nil }).
		ShortHelp("List all users").
		SubCommand("show").
		Func(func(ctx *CommandContext) error { return nil }).
		ShortHelp("Show details of a user").
		SubCommand("delete").
		Func(func(ctx *CommandContext) error { return nil }).
		ShortHelp("Delete a user")

	// Test executing parent command without subcommand - should show help and succeed
	err := cfg.Execute([]string{"app", "user"})
	if err != nil {
		t.Fatalf("Expected no error when executing command with subcommands (help should be shown), got %v", err)
	}

	// With new help system, help is shown directly and no error is returned
	// The test passes if we get here without an error
}

func TestCommandWithNilFuncAndNoSubcommands(t *testing.T) {
	cfg := New()

	// Define a command with no Func and no subcommands
	cfg.Command("broken").
		ShortHelp("This command has no implementation")

	// Test executing command with no Func and no subcommands
	err := cfg.Execute([]string{"app", "broken"})
	if err == nil {
		t.Fatal("Expected error when executing command with nil Func and no subcommands")
	}

	expectedErr := "command 'broken' has no implementation"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestCommandWithFuncAndSubcommands(t *testing.T) {
	cfg := New()
	executed := false

	// Define a command with both Func and subcommands (should work normally)
	cfg.Command("start").
		Func(func(ctx *CommandContext) error {
			executed = true
			return nil
		}).
		ShortHelp("Start the service").
		SubCommand("server").
		Func(func(ctx *CommandContext) error { return nil }).
		ShortHelp("Start server only")

	// Test executing parent command - should execute Func normally
	err := cfg.Execute([]string{"app", "start"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !executed {
		t.Error("Expected command Func to be executed")
	}
}

func TestGetSubcommandHelp(t *testing.T) {
	cmd := NewCommand("user")
	cmd.ShortHelp = "User management commands"

	// Add subcommands
	cmd.SubCommands["create"] = &Command{
		Name:      "create",
		ShortHelp: "Create a new user",
	}
	cmd.SubCommands["update"] = &Command{
		Name:      "update",
		ShortHelp: "Update an existing user",
	}
	cmd.SubCommands["list"] = &Command{
		Name:      "list",
		ShortHelp: "List all users",
	}
	cmd.SubCommands["show"] = &Command{
		Name:      "show",
		ShortHelp: "Show details of a user",
	}
	cmd.SubCommands["delete"] = &Command{
		Name:      "delete",
		ShortHelp: "Delete a user",
	}

	// Use new help system to get subcommand help
	helpIntegration := NewHelpIntegration()
	commands := map[string]*Command{"user": cmd}
	help, err := helpIntegration.GenerateHelp([]string{"user", "--help"}, commands)
	if err != nil {
		t.Fatalf("Failed to generate help: %v", err)
	}

	// Check help content - updated for new help system format
	expectedParts := []string{
		"Usage: user [options]",
		"User management commands",
		"Subcommands:",
		"create       Create a new user",
		"update       Update an existing user",
		"list         List all users",
		"show         Show details of a user",
		"delete       Delete a user",
	}

	for _, part := range expectedParts {
		if !strings.Contains(help, part) {
			t.Errorf("Expected '%s' in help, got: %s", part, help)
		}
	}

	// Print the actual help output to verify format
	t.Logf("Actual help output:\n%s", help)
}
