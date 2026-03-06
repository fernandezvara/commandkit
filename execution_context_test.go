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
	ctx.CollectError("test-key", "string", "", "test error", false)

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
	result := Get[string](ctx, "MISSING_KEY")
	if result.Error == nil {
		t.Error("Expected error for missing required key")
	}

	// Note: Required keys don't collect errors, they just return errors directly
	// The warning is logged but not collected in execution context
	// This is the new behavior - required data validation is separate from error collection
}

func TestGetOrWithExecutionContext(t *testing.T) {
	cfg := New()
	cfg.Define("OPTIONAL_KEY").String().Default("default")

	// Process the configuration to set the default value
	result := cfg.Process()
	if result.Error != nil {
		t.Fatalf("Configuration errors: %v", result.Error)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Test GetOr with existing key
	value := GetOr[string](ctx, "OPTIONAL_KEY", "fallback")
	if value != "default" {
		t.Errorf("Expected 'default', got '%s'", value)
	}

	// Test GetOr with missing key
	value = GetOr[string](ctx, "MISSING_KEY", "fallback")
	if value != "fallback" {
		t.Errorf("Expected 'fallback', got '%s'", value)
	}

	// Test GetOr with existing key and fallback
	value = GetOr[string](ctx, "OPTIONAL_KEY", "new-fallback")
	if value != "default" {
		t.Errorf("Expected 'default', got '%s'", value)
	}

	// Test GetOr with missing key and fallback
	value = GetOr[string](ctx, "MISSING_KEY", "new-fallback")
	if value != "new-fallback" {
		t.Errorf("Expected 'new-fallback', got '%s'", value)
	}
}

func TestMustGetWithExecutionContext(t *testing.T) {
	cfg := New()
	cfg.Define("TEST_VALUE").String().Default("test")

	// Process the configuration to set the default value
	result := cfg.Process()
	if result.Error != nil {
		t.Fatalf("Configuration errors: %v", result.Error)
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
