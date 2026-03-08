package commandkit

import (
	"testing"
)

func TestExecutionContext(t *testing.T) {
	// Test basic execution context creation
	ctx := NewExecutionContext("test-command")

	if ctx.GetCommand() != "test-command" {
		t.Errorf("Expected command 'test-command', got '%s'", ctx.GetCommand())
	}

	// Test error collection
	if ctx.HasErrors() {
		t.Error("Expected no errors initially")
	}

	// Test collecting an error
	ctx.CollectError(nil, "test-key", "string", "", "test error", false)

	if !ctx.HasErrors() {
		t.Error("Expected errors after collecting one")
	}

	errors := ctx.GetErrors()
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}

	if errors[0].Key != "test-key" {
		t.Errorf("Expected key 'test-key', got '%s'", errors[0].Key)
	}

	// Test clearing errors
	ctx.Clear()

	if ctx.HasErrors() {
		t.Error("Expected no errors after clearing")
	}
}

func TestCommandContextWithExecution(t *testing.T) {
	cfg := New()
	cfg.Define("TEST_VALUE").String().Default("test")

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Test that execution context is initialized
	if ctx.execution == nil {
		t.Error("Expected execution context to be initialized")
	}

	if ctx.execution.GetCommand() != "test" {
		t.Errorf("Expected command 'test', got '%s'", ctx.execution.GetCommand())
	}
}

func TestGetWithErrorCollection(t *testing.T) {
	cfg := New()
	cfg.Define("MISSING_KEY").String().Required()

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Test that Get returns error for missing key
	_, err := Get[string](ctx, "MISSING_KEY")
	if err == nil {
		t.Error("Expected error for missing required key")
	}

	// Note: Required keys don't collect errors, they just return errors directly
	// The warning is logged but not collected in execution context
	// This is the new behavior - required data validation is separate from error collection
}

func TestMustGetWithExecutionContext(t *testing.T) {
	cfg := New()
	cfg.Define("TEST_VALUE").String().Default("test")

	if err := cfg.Execute([]string{"test"}); err != nil {
		t.Fatalf("Configuration errors: %v", err)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Test MustGet with existing key
	value := MustGet[string](ctx, "TEST_VALUE")
	if value != "test" {
		t.Errorf("Expected 'test', got '%s'", value)
	}

	// Test MustGet with missing key (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected MustGet to panic for missing key")
		}
	}()

	MustGet[string](ctx, "MISSING_KEY")
}
