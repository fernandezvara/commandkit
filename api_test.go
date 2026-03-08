package commandkit

import (
	"testing"
)

// TestAPIChanges verifies that our API surface reduction works correctly
func TestAPIChanges(t *testing.T) {
	// Test that new Get[T] API works
	cfg := New()
	cfg.Define("PORT").Int64().Default(int64(8080))

	result := cfg.Process()
	if result.Error != nil {
		t.Fatalf("Config processing failed: %v", result.Error)
	}

	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Test new Get[T] signature
	port, err := Get[int64](ctx, "PORT")
	if err != nil {
		t.Errorf("Get[T] failed: %v", err)
	}
	if port != 8080 {
		t.Errorf("Expected 8080, got %d", port)
	}

	// Test getting non-existent key returns error
	_, err = Get[bool](ctx, "DEBUG")
	if err == nil {
		t.Error("Expected error for non-existent key")
	}

	t.Log("✅ API changes working correctly")
}
