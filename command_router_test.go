// commandkit/command_router_test.go
package commandkit

import (
	"testing"
)

func TestCommandRouter_RouteCommand_success(t *testing.T) {
	router := newCommandRouter()

	// Create a config with commands
	config := New()
	config.Command("start").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Start the service")

	// Route to existing command
	cmd, routedCtx, err := router.RouteCommand([]string{"app", "start", "--port", "8080"}, config)

	// Check that routing succeeded
	if err != nil {
		t.Errorf("RouteCommand() returned error: %v", err)
	}

	if cmd == nil {
		t.Error("RouteCommand() returned nil command")
	}

	if routedCtx == nil {
		t.Error("RouteCommand() returned nil context")
	}

	// Check context properties
	if routedCtx.Command != "start" {
		t.Errorf("Expected command name 'start', got '%s'", routedCtx.Command)
	}

	if len(routedCtx.Args) != 2 || routedCtx.Args[0] != "--port" || routedCtx.Args[1] != "8080" {
		t.Errorf("Expected args [--port, 8080], got %v", routedCtx.Args)
	}
}

func TestCommandRouter_RouteCommand_NoCommand(t *testing.T) {
	router := newCommandRouter()

	config := New()

	// Route with no command - should show help and return nil
	cmd, _, err := router.RouteCommand([]string{"app"}, config)

	// With new help system, help should be shown directly and no error should be returned
	if err != nil {
		t.Errorf("RouteCommand() should not return error for no command, got %v", err)
	}

	if cmd != nil {
		t.Error("RouteCommand() should return nil command for no command")
	}
}

func TestCommandRouter_RouteCommand_HelpCommand(t *testing.T) {
	router := newCommandRouter()

	config := New()
	config.Command("start").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Start the service")

	// Test various help commands
	helpCommands := []string{"help", "--help", "-h"}

	for _, helpCmd := range helpCommands {
		cmd, _, err := router.RouteCommand([]string{"app", helpCmd}, config)

		// With new help system, help should be shown directly and no error should be returned
		// The help system handles the display and returns nil for both cmd and err
		if err != nil {
			t.Errorf("RouteCommand() should not return error for help command '%s', got %v", helpCmd, err)
		}

		if cmd != nil {
			t.Error("RouteCommand() should return nil command for help case")
		}
	}
}

func TestCommandRouter_RouteCommand_UnknownCommand(t *testing.T) {
	router := newCommandRouter()

	config := New()
	config.Command("start").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Start the service")

	// Route to unknown command
	cmd, _, err := router.RouteCommand([]string{"app", "unknown"}, config)

	// Check that routing failed
	if err == nil {
		t.Error("RouteCommand() should have returned error for unknown command")
	}

	if cmd != nil {
		t.Error("RouteCommand() should return nil command for unknown command")
	}

	if !contains(err.Error(), "unknown command") {
		t.Error("Error should mention unknown command")
	}

	if !contains(err.Error(), "Did you mean") {
		t.Error("Error should provide suggestions")
	}
}

func TestCommandRouter_RouteCommand_NilConfig(t *testing.T) {
	router := newCommandRouter()

	// Route with nil config
	_, _, err := router.RouteCommand([]string{"app", "start"}, nil)

	// Check that routing failed
	if err == nil {
		t.Error("RouteCommand() should have returned error for nil config")
	}

	if !contains(err.Error(), "config cannot be nil") {
		t.Error("Error should mention config cannot be nil")
	}
}

func TestCommandRouter_HandleSubcommands_WithSubcommand(t *testing.T) {
	router := newCommandRouter()

	// Create a command with subcommands
	parentCmd := &Command{
		Name: "parent",
		SubCommands: map[string]*Command{
			"child": {
				Name:      "child",
				ShortHelp: "Child command",
				Func: func(ctx *CommandContext) error {
					return nil
				},
			},
		},
	}

	// Create context with subcommand args
	ctx := NewCommandContext([]string{"child", "--flag"}, New(), "parent", "")

	// Handle subcommands
	finalCmd, finalCtx, err := router.HandleSubcommands(parentCmd, ctx)

	// Check that subcommand was handled
	if err != nil {
		t.Errorf("HandleSubcommands() returned error: %v", err)
	}

	if finalCmd == nil {
		t.Error("HandleSubcommands() returned nil command")
	}

	if finalCmd.Name != "child" {
		t.Errorf("Expected subcommand name 'child', got '%s'", finalCmd.Name)
	}

	if finalCtx.SubCommand != "child" {
		t.Errorf("Expected subcommand 'child', got '%s'", finalCtx.SubCommand)
	}

	if len(finalCtx.Args) != 1 || finalCtx.Args[0] != "--flag" {
		t.Errorf("Expected args [--flag], got %v", finalCtx.Args)
	}
}

func TestCommandRouter_HandleSubcommands_NoSubcommand(t *testing.T) {
	router := newCommandRouter()

	// Create a command with subcommands
	parentCmd := &Command{
		Name: "parent",
		SubCommands: map[string]*Command{
			"child": {
				Name:      "child",
				ShortHelp: "Child command",
			},
		},
	}

	// Create context with non-subcommand args
	ctx := NewCommandContext([]string{"--flag"}, New(), "parent", "")

	// Handle subcommands
	finalCmd, finalCtx, err := router.HandleSubcommands(parentCmd, ctx)

	// Check that no subcommand was found, parent command returned
	if err != nil {
		t.Errorf("HandleSubcommands() returned error: %v", err)
	}

	if finalCmd != parentCmd {
		t.Error("HandleSubcommands() should return parent command when no subcommand found")
	}

	if finalCtx.SubCommand != "" {
		t.Errorf("Expected empty subcommand, got '%s'", finalCtx.SubCommand)
	}

	if len(finalCtx.Args) != 1 || finalCtx.Args[0] != "--flag" {
		t.Errorf("Expected args [--flag], got %v", finalCtx.Args)
	}
}

func TestCommandRouter_HandleSubcommands_NilCommand(t *testing.T) {
	router := newCommandRouter()

	ctx := NewCommandContext([]string{"child"}, New(), "parent", "")

	// Handle subcommands with nil command
	_, _, err := router.HandleSubcommands(nil, ctx)

	// Check that handling failed
	if err == nil {
		t.Error("HandleSubcommands() should have returned error for nil command")
	}

	if !contains(err.Error(), "command cannot be nil") {
		t.Error("Error should mention command cannot be nil")
	}
}

func TestCommandRouter_HandleSubcommands_NilContext(t *testing.T) {
	router := newCommandRouter()

	parentCmd := &Command{Name: "parent"}

	// Handle subcommands with nil context
	_, _, err := router.HandleSubcommands(parentCmd, nil)

	// Check that handling failed
	if err == nil {
		t.Error("HandleSubcommands() should have returned error for nil context")
	}

	if !contains(err.Error(), "context cannot be nil") {
		t.Error("Error should mention context cannot be nil")
	}
}

func TestCommandRouter_RouteWithErrorHandling_errorResult(t *testing.T) {
	router := newCommandRouter()

	config := New()
	config.Command("start").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Start the service")

	// Route to unknown command
	cmd, _, err := router.RouteWithErrorHandling([]string{"app", "unknown"}, config)

	// Check that error was returned
	if err == nil {
		t.Error("RouteWithErrorHandling() should have returned error for unknown command")
	}

	if cmd != nil {
		t.Error("RouteWithErrorHandling() should return nil command for error")
	}

	if !contains(err.Error(), "unknown command") {
		t.Error("Error should mention unknown command")
	}
}

func TestCommandRouter_Integration(t *testing.T) {
	// Test that CommandRouter works correctly with the service factory
	services := newCommandServices()
	router := services.CommandRouter

	// Create a comprehensive config
	config := New()
	config.Command("start").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Start the service")

	config.Command("deploy").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Deploy the service")

	parentCmd := config.Command("parent").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Parent command")

	parentCmd.SubCommand("child").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Child command")

	// Test routing to regular command
	cmd, _, err := router.RouteCommand([]string{"app", "start", "--port", "8080"}, config)
	if err != nil {
		t.Errorf("Integrated routing failed: %v", err)
	}

	if cmd.Name != "start" {
		t.Errorf("Expected command 'start', got '%s'", cmd.Name)
	}

	// Test routing to subcommand
	parent, parentCtx, err := router.RouteCommand([]string{"app", "parent", "child"}, config)
	if err != nil {
		t.Errorf("Integrated subcommand routing failed: %v", err)
	}

	finalCmd, finalCtx, err := router.HandleSubcommands(parent, parentCtx)
	if err != nil {
		t.Errorf("Integrated subcommand handling failed: %v", err)
	}

	if finalCmd.Name != "child" {
		t.Errorf("Expected subcommand 'child', got '%s'", finalCmd.Name)
	}

	if finalCtx.SubCommand != "child" {
		t.Errorf("Expected subcommand 'child', got '%s'", finalCtx.SubCommand)
	}
}

func TestCommandRouter_RouteWithHelpHandling_SubcommandHelp(t *testing.T) {
	router := newCommandRouter()

	// Create a config with commands and subcommands
	config := New()
	parentCmd := config.Command("start").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Start the service")

	parentCmd.SubCommand("server").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Start only the server")

	// Test subcommand help - should show server help
	cmd, ctx, err := router.RouteWithHelpHandling([]string{"app", "start", "server", "--help"}, config)

	// Help should be shown, so cmd should be nil and no error
	if cmd != nil {
		t.Error("RouteWithHelpHandling() should return nil command when help is shown")
	}

	if ctx != nil {
		t.Error("RouteWithHelpHandling() should return nil context when help is shown")
	}

	if err != nil {
		t.Errorf("RouteWithHelpHandling() should not return error when showing help, got: %v", err)
	}
}

func TestCommandRouter_RouteWithHelpHandling_CommandHelp(t *testing.T) {
	router := newCommandRouter()

	// Create a config with commands
	config := New()
	config.Command("start").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Start the service")

	// Test command help - should show start help
	cmd, ctx, err := router.RouteWithHelpHandling([]string{"app", "start", "--help"}, config)

	// Help should be shown, so cmd should be nil and no error
	if cmd != nil {
		t.Error("RouteWithHelpHandling() should return nil command when help is shown")
	}

	if ctx != nil {
		t.Error("RouteWithHelpHandling() should return nil context when help is shown")
	}

	if err != nil {
		t.Errorf("RouteWithHelpHandling() should not return error when showing help, got: %v", err)
	}
}

func TestCommandRouter_RouteWithHelpHandling_GlobalHelp(t *testing.T) {
	router := newCommandRouter()

	// Create a config with commands
	config := New()
	config.Command("start").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Start the service")

	// Test global help
	cmd, ctx, err := router.RouteWithHelpHandling([]string{"app", "--help"}, config)

	// Help should be shown, so cmd should be nil and no error
	if cmd != nil {
		t.Error("RouteWithHelpHandling() should return nil command when help is shown")
	}

	if ctx != nil {
		t.Error("RouteWithHelpHandling() should return nil context when help is shown")
	}

	if err != nil {
		t.Errorf("RouteWithHelpHandling() should not return error when showing help, got: %v", err)
	}
}

func TestCommandRouter_RouteWithHelpHandling_NoHelp(t *testing.T) {
	router := newCommandRouter()

	// Create a config with commands and subcommands
	config := New()
	parentCmd := config.Command("start").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Start the service")

	parentCmd.SubCommand("server").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Start only the server")

	// Test normal execution without help - should route to subcommand
	cmd, ctx, err := router.RouteWithHelpHandling([]string{"app", "start", "server", "--port", "8080"}, config)

	// Should route normally
	if err != nil {
		t.Errorf("RouteWithHelpHandling() should not return error for normal execution, got: %v", err)
	}

	if cmd == nil {
		t.Error("RouteWithHelpHandling() should return command for normal execution")
	}

	if ctx == nil {
		t.Error("RouteWithHelpHandling() should return context for normal execution")
	}

	// Check that it routed to the subcommand
	if cmd.Name != "server" {
		t.Errorf("Expected subcommand 'server', got '%s'", cmd.Name)
	}

	if ctx.SubCommand != "server" {
		t.Errorf("Expected subcommand 'server' in context, got '%s'", ctx.SubCommand)
	}
}

func TestCommandRouter_RouteWithHelpHandling_ShortHelpFlag(t *testing.T) {
	router := newCommandRouter()

	// Create a config with commands and subcommands
	config := New()
	parentCmd := config.Command("start").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Start the service")

	parentCmd.SubCommand("server").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Start only the server")

	// Test subcommand help with short flag
	cmd, ctx, err := router.RouteWithHelpHandling([]string{"app", "start", "server", "-h"}, config)

	// Help should be shown, so cmd should be nil and no error
	if cmd != nil {
		t.Error("RouteWithHelpHandling() should return nil command when help is shown")
	}

	if ctx != nil {
		t.Error("RouteWithHelpHandling() should return nil context when help is shown")
	}

	if err != nil {
		t.Errorf("RouteWithHelpHandling() should not return error when showing help, got: %v", err)
	}
}

func TestCommandRouter_EmptyStringCommand(t *testing.T) {
	router := newCommandRouter()
	cfg := New()

	// Add empty string command
	cfg.Command("").
		Func(func(ctx *CommandContext) error {
			return nil
		}).
		ShortHelp("Default command")

	// Test routing to empty string command
	cmd, ctx, err := router.RouteCommand([]string{"app"}, cfg)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cmd == nil {
		t.Fatal("Expected command to be routed to empty string command")
	}

	if ctx.Command != "" {
		t.Fatalf("Expected empty command name, got: %s", ctx.Command)
	}
}

func TestCommandRouter_EmptyStringCommandWithFlags(t *testing.T) {
	router := newCommandRouter()
	cfg := New()

	// Add empty string command
	cfg.Command("").
		Func(func(ctx *CommandContext) error {
			return nil
		}).
		ShortHelp("Default command")

	// Test routing with flags - this should work because all args after program name are treated as flags
	cmd, ctx, err := router.RouteCommand([]string{"app", "--verbose", "--debug"}, cfg)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cmd == nil {
		t.Fatal("Expected command to be routed to empty string command")
	}

	// The empty string command should receive all args as flags
	if len(ctx.Args) != 2 {
		t.Fatalf("Expected 2 args, got: %d", len(ctx.Args))
	}

	if ctx.Args[0] != "--verbose" || ctx.Args[1] != "--debug" {
		t.Fatalf("Expected args to contain --verbose and --debug, got: %v", ctx.Args)
	}
}

func TestCommandRouter_EmptyStringCommandFallback(t *testing.T) {
	router := newCommandRouter()
	cfg := New()

	// No empty string command defined
	cfg.Command("test").
		Func(func(ctx *CommandContext) error {
			return nil
		}).
		ShortHelp("Test command")

	// Test routing falls back to global help
	cmd, ctx, err := router.RouteCommand([]string{"app"}, cfg)

	// Should return nil for command and context, and help error
	if cmd != nil || ctx != nil {
		t.Fatal("Expected nil command and context when no default command exists")
	}

	// err should be nil because help was shown successfully
	if err != nil {
		t.Fatalf("Expected no error when help is shown, got: %v", err)
	}
}
