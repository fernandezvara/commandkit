// commandkit/help_service_unified_test.go
package commandkit

import (
	"testing"
)

func TestHelpService_ShowHelpUnified(t *testing.T) {
	service := newHelpService()

	// Test global help
	err := service.ShowHelpUnified("", "", false, []GetError{}, nil)
	if err != nil {
		t.Errorf("Unexpected error showing global help: %v", err)
	}

	// Test command help with non-existent command
	err = service.ShowHelpUnified("nonexistent", "", false, []GetError{}, nil)
	if err != nil {
		t.Errorf("Unexpected error for non-existent command: %v", err)
	}
}

func TestHelpService_TriggerHelpUnified(t *testing.T) {
	service := newHelpService()

	// Test with nil context
	err := service.TriggerHelpUnified(nil, []GetError{})
	if err == nil {
		t.Error("Expected error for nil context")
	}

	// Test with valid context
	config := New()
	ctx := NewCommandContext([]string{"--help"}, config, "", "")

	err = service.TriggerHelpUnified(ctx, []GetError{})
	if err != nil {
		t.Errorf("Unexpected error with valid context: %v", err)
	}
}
