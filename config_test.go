package commandkit

import (
	"os"
	"testing"
	"time"
)

func TestBasicConfigurationDefinition(t *testing.T) {
	cfg := New()

	// Test basic definition
	cfg.Define("PORT").
		Int64().
		Env("PORT").
		Flag("port").
		Default(int64(8080)).
		Range(1, 65535).
		Description("HTTP server port")

	cfg.Define("BASE_URL").
		String().
		Env("BASE_URL").
		Flag("base-url").
		Required().
		URL().
		Description("Public base URL of the service")

	cfg.Define("DEBUG").
		Bool().
		Env("DEBUG").
		Flag("debug").
		Default(false).
		Description("Enable debug mode")

	cfg.Define("TIMEOUT").
		Duration().
		Env("TIMEOUT").
		Default(30*time.Second).
		DurationRange(1*time.Second, 5*time.Minute).
		Description("Request timeout")

	cfg.Define("CORS_ORIGINS").
		StringSlice().
		Env("CORS_ORIGINS").
		Flag("cors-origins").
		Delimiter(",").
		Default([]string{"http://localhost:3000"}).
		Description("Allowed CORS origins")

	// Set environment variables for testing
	os.Setenv("BASE_URL", "https://api.example.com")
	os.Setenv("DEBUG", "true")
	defer func() {
		os.Unsetenv("BASE_URL")
		os.Unsetenv("DEBUG")
	}()

	// Process configuration
	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Configuration errors: %v", errs)
	}

	// Test values
	port := Get[int64](cfg, "PORT")
	if port != 8080 {
		t.Errorf("Expected PORT=8080, got %d", port)
	}

	baseURL := Get[string](cfg, "BASE_URL")
	if baseURL != "https://api.example.com" {
		t.Errorf("Expected BASE_URL=https://api.example.com, got %s", baseURL)
	}

	debug := Get[bool](cfg, "DEBUG")
	if !debug {
		t.Errorf("Expected DEBUG=true, got %v", debug)
	}

	timeout := Get[time.Duration](cfg, "TIMEOUT")
	if timeout != 30*time.Second {
		t.Errorf("Expected TIMEOUT=30s, got %v", timeout)
	}

	corsOrigins := Get[[]string](cfg, "CORS_ORIGINS")
	if len(corsOrigins) != 1 || corsOrigins[0] != "http://localhost:3000" {
		t.Errorf("Expected CORS_ORIGINS=[http://localhost:3000], got %v", corsOrigins)
	}

	// Test convenience methods
	if cfg.GetString("BASE_URL") != baseURL {
		t.Error("GetString() method failed")
	}

	if cfg.GetInt64("PORT") != port {
		t.Error("GetInt64() method failed")
	}

	if cfg.GetBool("DEBUG") != debug {
		t.Error("GetBool() method failed")
	}

	if cfg.GetDuration("TIMEOUT") != timeout {
		t.Error("GetDuration() method failed")
	}

	if cfg.GetStringSlice("CORS_ORIGINS")[0] != corsOrigins[0] {
		t.Error("GetStringSlice() method failed")
	}

	// Test Has method
	if !cfg.Has("PORT") {
		t.Error("Has() method failed for existing key")
	}

	if cfg.Has("NONEXISTENT") {
		t.Error("Has() method should return false for non-existent key")
	}

	// Test Keys method
	keys := cfg.Keys()
	if len(keys) != 5 {
		t.Errorf("Expected 5 keys, got %d", len(keys))
	}

	// Test Dump
	dump := cfg.Dump()
	if dump["PORT"] != "8080" {
		t.Errorf("Dump failed for PORT: %s", dump["PORT"])
	}

	if dump["BASE_URL"] != "https://api.example.com" {
		t.Errorf("Dump failed for BASE_URL: %s", dump["BASE_URL"])
	}
}

func TestValidation(t *testing.T) {
	cfg := New()

	cfg.Define("PORT").
		Int64().
		Env("PORT").
		Default(int64(8080)).
		Range(1, 65535)

	cfg.Define("RATE").
		Float64().
		Env("RATE").
		Default(100.0).
		Range(1.0, 1000.0)

	cfg.Define("API_KEY").
		String().
		Env("API_KEY").
		Required().
		MinLength(10)

	// Set invalid values
	os.Setenv("PORT", "99999") // Too high
	os.Setenv("RATE", "0.5")   // Too low
	// API_KEY not set (required)
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("RATE")
	}()

	errs := cfg.Process()
	if len(errs) != 3 {
		t.Fatalf("Expected 3 errors, got %d", len(errs))
	}

	// Check specific errors
	errorMessages := make(map[string]bool)
	for _, err := range errs {
		errorMessages[err.Message] = true
	}

	if !errorMessages["value 99999 is greater than maximum 65535"] {
		t.Error("Missing PORT validation error")
	}

	if !errorMessages["value 0.500000 is less than minimum 1"] {
		t.Error("Missing RATE validation error")
	}

	if !errorMessages["required value not provided (set in file, API_KEY or --)"] {
		t.Error("Missing API_KEY required error")
	}
}

func TestSecretHandling(t *testing.T) {
	cfg := New()

	cfg.Define("DATABASE_URL").
		String().
		Env("DATABASE_URL").
		Required().
		Secret()

	cfg.Define("API_KEY").
		String().
		Env("API_KEY").
		Default("secret123").
		Secret()

	os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost/db")
	defer os.Unsetenv("DATABASE_URL")

	errs := cfg.Process()
	if len(errs) > 0 {
		t.Fatalf("Configuration errors: %v", errs)
	}

	// Test secret access
	dbURL := cfg.GetSecret("DATABASE_URL")
	if !dbURL.IsSet() {
		t.Error("DATABASE_URL secret not set")
	}

	if dbURL.Size() != len("postgresql://user:pass@localhost/db") {
		t.Error("DATABASE_URL secret size incorrect")
	}

	apiKey := cfg.GetSecret("API_KEY")
	if apiKey.String() != "secret123" {
		t.Error("API_KEY secret value incorrect")
	}

	// Test that regular Get panics for secrets
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when accessing secret with Get()")
		}
	}()

	Get[string](cfg, "DATABASE_URL")
}
