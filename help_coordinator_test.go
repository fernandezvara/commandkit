// commandkit/help_coordinator_test.go
package commandkit

import (
	"strings"
	"testing"
)

func TestNewHelpCoordinator(t *testing.T) {
	coordinator := newHelpCoordinator()

	if coordinator == nil {
		t.Fatal("Expected non-nil coordinator")
	}

	if coordinator.templates == nil {
		t.Error("Expected non-nil templates")
	}

	if coordinator.extractor == nil {
		t.Error("Expected non-nil extractor")
	}

	if coordinator.output == nil {
		t.Error("Expected non-nil output")
	}
}

func TestHelpCoordinator_SetOutput(t *testing.T) {
	coordinator := newHelpCoordinator()

	stringOutput := &StringHelpOutput{}
	coordinator.SetOutput(stringOutput)

	if coordinator.output != stringOutput {
		t.Error("Expected output to be set to stringOutput")
	}
}

func TestArgsContainFullHelp(t *testing.T) {
	tests := []struct {
		args     []string
		expected bool
	}{
		{[]string{}, false},
		{[]string{"--help"}, false},
		{[]string{"--full-help"}, true},
		{[]string{"command", "--full-help"}, true},
		{[]string{"command", "subcommand", "--full-help"}, true},
		{[]string{"--full-help", "other"}, true},
	}

	for _, test := range tests {
		result := argsContainFullHelp(test.args)
		if result != test.expected {
			t.Errorf("argsContainFullHelp(%v) = %v, expected %v", test.args, result, test.expected)
		}
	}
}

func TestHelpCoordinator_getExecutableName(t *testing.T) {
	// This test depends on os.Args, so we just verify it returns a non-empty string
	name := getExecutableName()
	if name == "" {
		t.Error("Expected non-empty executable name")
	}
}

func TestHelpCoordinator_ShowHelp_Global(t *testing.T) {
	coordinator := newHelpCoordinator()
	stringOutput := &StringHelpOutput{}
	coordinator.SetOutput(stringOutput)

	err := coordinator.ShowHelp("", "", false, []GetError{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := stringOutput.Get()
	if !strings.Contains(output, "Usage:") {
		t.Error("Expected output to contain usage information")
	}

	// Global help should show usage and command help info
	if !strings.Contains(output, "<command>") {
		t.Error("Expected output to contain command placeholder")
	}
}

func TestHelpCoordinator_ShowHelp_WithCommand(t *testing.T) {
	coordinator := newHelpCoordinator()
	stringOutput := &StringHelpOutput{}
	coordinator.SetOutput(stringOutput)

	// Test with a non-existent command should return simple help (no commands map provided)
	err := coordinator.ShowHelp("nonexistent", "", false, []GetError{})
	if err != nil {
		t.Errorf("Unexpected error for non-existent command: %v", err)
	}

	output := stringOutput.Get()
	if !strings.Contains(output, "Usage: nonexistent [options]") {
		t.Error("Expected simple help output for non-existent command")
	}
}

func TestHelpCoordinator_TriggerHelp(t *testing.T) {
	coordinator := newHelpCoordinator()
	stringOutput := &StringHelpOutput{}
	coordinator.SetOutput(stringOutput)

	// Test with nil context
	err := coordinator.TriggerHelp(nil, []GetError{})
	if err == nil {
		t.Error("Expected error for nil context")
	}

	if !strings.Contains(err.Error(), "command context cannot be nil") {
		t.Errorf("Expected error message to mention nil context, got: %v", err)
	}

	// Test with valid context but no command
	config := New()
	ctx := NewCommandContext([]string{"--help"}, config, "", "")

	err = coordinator.TriggerHelp(ctx, []GetError{})
	if err != nil {
		t.Errorf("Unexpected error with valid context: %v", err)
	}

	output := stringOutput.Get()
	if !strings.Contains(output, "Usage:") {
		t.Error("Expected output to contain usage information")
	}
}

func TestConsoleHelpOutput_Print(t *testing.T) {
	output := &ConsoleHelpOutput{}

	err := output.Print("test output")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestStringHelpOutput(t *testing.T) {
	output := &StringHelpOutput{}

	// Test Print
	err := output.Print("test output")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test Get
	result := output.Get()
	if result != "test output" {
		t.Errorf("Expected 'test output', got '%s'", result)
	}

	// Test multiple prints
	err = output.Print(" more")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	result = output.Get()
	if result != "test output more" {
		t.Errorf("Expected 'test output more', got '%s'", result)
	}
}
