package commandkit

import (
	"os"
	"testing"
	"time"
)

func TestGetOr(t *testing.T) {
	cfg := New()

	cfg.Define("PORT").Int64().Default(int64(8080))
	cfg.Define("HOST").String() // No default, not required

	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	// Test GetOr with existing value
	port := GetOr[int64](cfg, "PORT", 3000)
	if port != 8080 {
		t.Errorf("GetOr should return existing value 8080, got %d", port)
	}

	// Test GetOr with non-existent key
	timeout := GetOr[time.Duration](cfg, "TIMEOUT", 30*time.Second)
	if timeout != 30*time.Second {
		t.Errorf("GetOr should return default 30s for non-existent key, got %v", timeout)
	}

	// Test GetOr with nil value (HOST has no value)
	host := GetOr[string](cfg, "HOST", "localhost")
	if host != "localhost" {
		t.Errorf("GetOr should return default 'localhost' for nil value, got %s", host)
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

func TestGetPanicOnMissingKey(t *testing.T) {
	cfg := New()
	cfg.Process()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Get should panic for non-existent key")
		}
	}()

	Get[string](cfg, "NONEXISTENT")
}

func TestGetPanicOnWrongType(t *testing.T) {
	cfg := New()

	cfg.Define("PORT").Int64().Default(int64(8080))

	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Get should panic for wrong type")
		}
	}()

	Get[string](cfg, "PORT") // PORT is int64, not string
}

func TestGetPanicOnSecret(t *testing.T) {
	cfg := New()

	cfg.Define("API_KEY").String().Secret().Default("secret123")

	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Unexpected errors: %v", errs)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Get should panic for secret key")
		}
	}()

	Get[string](cfg, "API_KEY")
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
