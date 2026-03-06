package commandkit

import (
	"testing"
)

// Simple test to verify basic functionality without dependencies on other test files
func TestBasicExecutionContext(t *testing.T) {
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
