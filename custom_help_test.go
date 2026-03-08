// commandkit/custom_help_test.go
package commandkit

import (
	"testing"
)

func TestCustomHelpFunctionality(t *testing.T) {
	cfg := New()

	// Command with custom help
	cfg.Command("custom").
		Func(func(ctx *CommandContext) error {
			return nil
		}).
		ShortHelp("Custom command").
		LongHelp("This is custom long help text that should be displayed instead of the default template.").
		CustomHelp()

	// Command without custom help (default behavior)
	cfg.Command("default").
		Func(func(ctx *CommandContext) error {
			return nil
		}).
		ShortHelp("Default command").
		LongHelp("This is default long help text.")

	// Test custom help display
	args := []string{"custom", "--help"}
	err := cfg.Execute(args)
	if err != nil {
		t.Errorf("Custom help should execute without error, got: %v", err)
	}

	// Test default help display
	args = []string{"default", "--help"}
	err = cfg.Execute(args)
	if err != nil {
		t.Errorf("Default help should execute without error, got: %v", err)
	}
}

func TestCustomHelpWithValidationErrors(t *testing.T) {
	cfg := New()

	// Command with custom help and validation
	cfg.Command("custom-with-validation").
		Func(func(ctx *CommandContext) error {
			return nil
		}).
		ShortHelp("Custom command with validation").
		LongHelp("This is custom long help text that should be displayed with validation errors.").
		CustomHelp().
		Config(func(cc *CommandConfig) {
			cc.Define("REQUIRED_FIELD").
				String().
				Flag("required").
				Required().
				Description("Required field for testing")
		})

	// Test custom help with validation errors - provide a flag to trigger validation processing
	args := []string{"custom-with-validation", "--required", ""}
	err := cfg.Execute(args)
	// When there are validation errors, the system shows custom help and returns an error
	// This is the expected behavior - validation errors show custom help then fail
	if err == nil {
		t.Error("Command should fail due to empty required field")
	}
	// The test passes if we get an error, which means the custom help was shown with validation errors
}

func TestCustomHelpField(t *testing.T) {
	cfg := New()

	builder := cfg.Command("test")

	// Initially customHelp should be false
	if builder.cmd.customHelp != false {
		t.Error("customHelp should be false by default")
	}

	// After calling CustomHelp(), it should be true
	builder.CustomHelp()
	if builder.cmd.customHelp != true {
		t.Error("customHelp should be true after calling CustomHelp()")
	}

	// Test cloning preserves customHelp
	cloned := builder.Clone()
	if cloned.cmd.customHelp != true {
		t.Error("cloned command should preserve customHelp setting")
	}
}

func TestCustomHelpTemplates(t *testing.T) {
	factory := NewHelpFactory()

	// Test that custom help templates are available
	customTemplate := factory.GetTemplate(TemplateCustomHelp)
	if customTemplate == "" {
		t.Error("TemplateCustomHelp should not be empty")
	}

	customErrorTemplate := factory.GetTemplate(TemplateCustomHelpError)
	if customErrorTemplate == "" {
		t.Error("TemplateCustomHelpError should not be empty")
	}

	// Verify templates contain expected content
	if !stringContains(customTemplate, "{{.Command.LongHelp}}") {
		t.Error("Custom help template should contain Command.LongHelp")
	}

	if !stringContains(customErrorTemplate, "{{.Command.LongHelp}}") {
		t.Error("Custom help error template should contain Command.LongHelp")
	}

	// Check for errors section (should be .HasErrors)
	if !stringContains(customErrorTemplate, "HasErrors") {
		t.Error("Custom help error template should contain HasErrors")
	}
}

func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) != -1
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
