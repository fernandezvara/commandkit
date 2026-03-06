// commandkit/command_router_test.go
package commandkit

import (
	"testing"
)

func TestCommandRouter_RouteCommand_Success(t *testing.T) {
	router := NewCommandRouter()

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
	router := NewCommandRouter()

	config := New()

	// Route with no command
	cmd, _, err := router.RouteCommand([]string{"app"}, config)

	// Check that routing returned special help case
	if err == nil {
		t.Error("RouteCommand() should have returned error for no command")
	}

	if err.Error() != "show global help" {
		t.Errorf("Expected 'show global help' error, got %v", err)
	}

	if cmd != nil {
		t.Error("RouteCommand() should return nil command for help case")
	}
}

func TestCommandRouter_RouteCommand_HelpCommand(t *testing.T) {
	router := NewCommandRouter()

	config := New()
	config.Command("start").Func(func(ctx *CommandContext) error {
		return nil
	}).ShortHelp("Start the service")

	// Test various help commands
	helpCommands := []string{"help", "--help", "-h"}

	for _, helpCmd := range helpCommands {
		cmd, _, err := router.RouteCommand([]string{"app", helpCmd}, config)

		// Check that routing returned special help case
		if err == nil {
			t.Errorf("RouteCommand() should have returned error for help command '%s'", helpCmd)
		}

		if err.Error() != "show global help" {
			t.Errorf("Expected 'show global help' error for '%s', got %v", helpCmd, err)
		}

		if cmd != nil {
			t.Error("RouteCommand() should return nil command for help case")
		}
	}
}

func TestCommandRouter_RouteCommand_UnknownCommand(t *testing.T) {
	router := NewCommandRouter()

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
	router := NewCommandRouter()

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
	router := NewCommandRouter()

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
	router := NewCommandRouter()

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
	router := NewCommandRouter()

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
	router := NewCommandRouter()

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

func TestCommandRouter_RouteWithErrorHandling_Error(t *testing.T) {
	router := NewCommandRouter()

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
	services := NewCommandServices()
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
