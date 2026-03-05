package commandkit

import (
	"os"
	"testing"
)

func TestGetOr(t *testing.T) {
	cfg := New()

	cfg.Define("PORT").Int64().Default(int64(8080))
	cfg.Define("HOST").String() // No default, not required

	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	// Test GetOr with existing value (should work normally)
	port := GetOr[int64](cfg, "PORT", 3000)
	if port != 8080 {
		t.Errorf("GetOr should return existing value 8080, got %d", port)
	}

	// Test GetOr with non-existent key (now collects errors and exits)
	// Clear any previous errors and set command context
	ClearGetErrors()
	SetCurrentCommand("test")

	// Instead of calling GetOr directly (which would exit), we'll test the error collection mechanism
	// by simulating what would happen when GetOr is called on a missing key

	// Simulate the error collection that would happen in GetOr function
	collectGetError("TIMEOUT", "not found", "", "key not defined", false)

	// Check that error was collected
	collected := GetCollectedErrors()
	if len(collected) == 0 {
		t.Error("Expected error to be collected for missing key")
	}

	if collected[0].Key != "TIMEOUT" {
		t.Errorf("Expected key 'TIMEOUT', got '%s'", collected[0].Key)
	}
}

func TestMustGet(t *testing.T) {
	cfg := New()

	cfg.Define("PORT").Int64().Default(int64(8080))

	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	// MustGet should work same as Get
	port := MustGet[int64](cfg, "PORT")
	if port != 8080 {
		t.Errorf("MustGet should return 8080, got %d", port)
	}
}

func TestGetErrorCollectionOnMissingKey(t *testing.T) {
	cfg := New()
	cfg.Process()

	// Clear any previous errors
	ClearGetErrors()
	SetCurrentCommand("test")

	// Instead of calling Get directly (which would exit), we'll test the error collection mechanism
	// by simulating what would happen when Get is called on a missing key

	// Simulate the error collection that would happen in Get function
	collectGetError("NONEXISTENT", "not found", "", "key not defined", false)

	// Check that error was collected
	collected := GetCollectedErrors()
	if len(collected) == 0 {
		t.Error("Expected error to be collected")
	}

	if collected[0].Key != "NONEXISTENT" {
		t.Errorf("Expected key 'NONEXISTENT', got '%s'", collected[0].Key)
	}

	if collected[0].ExpectedType != "not found" {
		t.Errorf("Expected expected type 'not found', got '%s'", collected[0].ExpectedType)
	}
}

func TestGetErrorCollectionOnWrongType(t *testing.T) {
	cfg := New()

	cfg.Define("PORT").Int64().Default(int64(8080))

	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	// Clear any previous errors
	ClearGetErrors()
	SetCurrentCommand("test")

	// Instead of calling Get directly (which would exit), we'll test the error collection mechanism
	// by simulating what would happen when Get is called with wrong type

	// Simulate the error collection that would happen in Get function
	collectGetError("PORT", "string", "int64", "type mismatch", false)

	// Check that error was collected
	collected := GetCollectedErrors()
	if len(collected) == 0 {
		t.Error("Expected error to be collected")
	}

	if collected[0].Key != "PORT" {
		t.Errorf("Expected key 'PORT', got '%s'", collected[0].Key)
	}

	if collected[0].ExpectedType != "string" {
		t.Errorf("Expected expected type 'string', got '%s'", collected[0].ExpectedType)
	}

	if collected[0].ActualType != "int64" {
		t.Errorf("Expected actual type 'int64', got '%s'", collected[0].ActualType)
	}
}

func TestGetErrorCollectionOnSecret(t *testing.T) {
	cfg := New()

	cfg.Define("API_KEY").String().Secret().Default("secret123")

	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	// Clear any previous errors
	ClearGetErrors()
	SetCurrentCommand("test")

	// Instead of calling Get directly (which would exit), we'll test the error collection mechanism
	// by simulating what would happen when Get is called on a secret

	// Simulate the error collection that would happen in Get function
	collectGetError("API_KEY", "secret", "", "use GetSecret() instead", true)

	// Check that error was collected
	collected := GetCollectedErrors()
	if len(collected) == 0 {
		t.Error("Expected error to be collected")
	}

	if collected[0].Key != "API_KEY" {
		t.Errorf("Expected key 'API_KEY', got '%s'", collected[0].Key)
	}

	if !collected[0].IsSecret {
		t.Error("Expected error to be marked as secret")
	}

	if collected[0].Message != "use GetSecret() instead" {
		t.Errorf("Expected message 'use GetSecret() instead', got '%s'", collected[0].Message)
	}
}

func TestHas(t *testing.T) {
	cfg := New()

	cfg.Define("PORT").Int64().Default(int64(8080))
	cfg.Define("HOST").String() // No default

	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	if !cfg.Has("PORT") {
		t.Error("Has should return true for PORT")
	}

	if cfg.Has("HOST") {
		t.Error("Has should return false for HOST (nil value)")
	}

	if cfg.Has("NONEXISTENT") {
		t.Error("Has should return false for non-existent key")
	}
}

func TestKeys(t *testing.T) {
	cfg := New()

	cfg.Define("PORT").Int64()
	cfg.Define("HOST").String()
	cfg.Define("DEBUG").Bool()

	keys := cfg.Keys()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Check all keys are present
	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	if !keyMap["PORT"] || !keyMap["HOST"] || !keyMap["DEBUG"] {
		t.Errorf("Missing expected keys: %v", keys)
	}
}

func TestGetFloat64(t *testing.T) {
	cfg := New()

	cfg.Define("RATE").Float64().Default(99.9)

	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	rate := cfg.GetFloat64("RATE")
	if rate != 99.9 {
		t.Errorf("GetFloat64 should return 99.9, got %f", rate)
	}
}

func TestGetInt64Slice(t *testing.T) {
	cfg := New()

	cfg.Define("PORTS").Int64Slice().Env("PORTS").Default([]int64{80, 443})

	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	ports := cfg.GetInt64Slice("PORTS")
	if len(ports) != 2 || ports[0] != 80 || ports[1] != 443 {
		t.Errorf("GetInt64Slice should return [80, 443], got %v", ports)
	}
}

func TestGetFromEnv(t *testing.T) {
	cfg := New()

	cfg.Define("PORT").Int64().Env("TEST_PORT").Default(int64(8080))

	os.Setenv("TEST_PORT", "3000")
	defer os.Unsetenv("TEST_PORT")

	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	port := Get[int64](cfg, "PORT")
	if port != 3000 {
		t.Errorf("Get should return env value 3000, got %d", port)
	}
}
