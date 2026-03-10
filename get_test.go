package commandkit

import (
	"os"
	"os/exec"
	"testing"
)

func TestMustGet(t *testing.T) {
	cfg := New()

	cfg.Define("PORT").Int64().Default(8080)
	cfg.Define("HOST").String() // No default, not required

	if err := cfg.Execute([]string{"test"}); err != nil {
		t.Fatalf("Unexpected errors: %v", err)
	}

	// Create context for new API
	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Test MustGet with existing value
	port := MustGet[int64](ctx, "PORT")
	if port != 8080 {
		t.Errorf("MustGet should return 8080, got %d", port)
	}

	// Test MustGet with missing key (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected MustGet to panic for missing key")
		}
	}()

	MustGet[string](ctx, "MISSING_KEY")
}

func TestGetWithMissingKey(t *testing.T) {
	cfg := New()

	cfg.Define("MISSING_KEY").String()

	if err := cfg.Execute([]string{"test"}); err != nil {
		t.Fatalf("Unexpected errors: %v", err)
	}

	// Create context for new API
	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Test Get with missing key
	_, err := Get[string](ctx, "MISSING_KEY")
	if err == nil {
		t.Error("Expected error for missing key")
	}

	// Verify error was collected
	if !ctx.execution.HasErrors() {
		t.Error("Expected error to be collected for missing key")
	}

	collected := ctx.execution.GetErrors()
	if len(collected) == 0 {
		t.Error("Expected error to be collected for missing key")
	}

	if collected[0].Key != "MISSING_KEY" {
		t.Errorf("Expected key 'MISSING_KEY', got '%s'", collected[0].Key)
	}
}

func TestGetWithTypeConversion(t *testing.T) {
	cfg := New()

	cfg.Define("PORT").String().Default("8080") // String should convert to int64

	if err := cfg.Execute([]string{"test"}); err != nil {
		t.Fatalf("Unexpected errors: %v", err)
	}

	// Create context for new API
	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Test Get with type conversion (should now work)
	value, err := Get[int64](ctx, "PORT")
	if err != nil {
		t.Errorf("Expected successful conversion, got error: %v", err)
	}

	// Verify the converted value
	if value != 8080 {
		t.Errorf("Expected value 8080, got %d", value)
	}

	// Verify no errors were collected (conversion succeeded)
	if ctx.execution.HasErrors() {
		t.Error("Expected no errors for successful conversion")
	}
}

func TestGetWithSecret(t *testing.T) {
	cfg := New()

	cfg.Define("API_KEY").String().Secret()

	if err := cfg.Execute([]string{"test"}); err != nil {
		t.Fatalf("Unexpected errors: %v", err)
	}

	// Create context for new API
	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Test Get with secret (should return error)
	_, err := Get[string](ctx, "API_KEY")
	if err == nil {
		t.Error("Expected error for secret access")
	}

	if err.Error() != "validation error: configuration 'API_KEY' is secret, use GetSecret() instead" {
		t.Errorf("Expected secret access error, got: %v", err)
	}

	// Verify error was collected
	if !ctx.execution.HasErrors() {
		t.Error("Expected error to be collected for secret access")
	}

	collected := ctx.execution.GetErrors()
	if len(collected) == 0 {
		t.Error("Expected error to be collected for secret access")
	}

	if collected[0].Key != "API_KEY" {
		t.Errorf("Expected key 'API_KEY', got '%s'", collected[0].Key)
	}

	if !collected[0].IsSecret {
		t.Error("Expected error to be marked as secret")
	}
}

func TestGetErrorDisplayName(t *testing.T) {
	cfg := New()

	cfg.Define("PORT").Int64().Flag("port").Env("PORT")
	cfg.Define("DATABASE_URL").String().Env("DATABASE_URL").Secret()
	cfg.Define("DEBUG").Bool().Env("DEBUG")

	// Don't process config since we're just testing display name formatting
	// Create context for new API
	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Collect some errors to test display name formatting
	ctx.execution.CollectError(cfg, "PORT", "not found", "", "key not defined", false)
	ctx.execution.CollectError(cfg, "DATABASE_URL", "secret", "", "use GetSecret() instead", true)
	ctx.execution.CollectError(cfg, "DEBUG", "not found", "", "key not defined", false)

	collected := ctx.execution.GetErrors()
	if len(collected) != 3 {
		t.Errorf("Expected 3 collected errors, got %d", len(collected))
	}

	// Test display name formatting
	portDisplayName := getErrorDisplayName(collected[0], cfg)
	if portDisplayName != "--port int64 (env: PORT)" {
		t.Errorf("Expected '--port int64 (env: PORT)', got '%s'", portDisplayName)
	}

	dbDisplayName := getErrorDisplayName(collected[1], cfg)
	if dbDisplayName != "(no flag) string (env: DATABASE_URL)" {
		t.Errorf("Expected '(no flag) string (env: DATABASE_URL)', got '%s'", dbDisplayName)
	}

	debugDisplayName := getErrorDisplayName(collected[2], cfg)
	if debugDisplayName != "(no flag) bool (env: DEBUG)" {
		t.Errorf("Expected '(no flag) bool (env: DEBUG)', got '%s'", debugDisplayName)
	}
}

func TestDisplayGetErrorsAndExit(t *testing.T) {
	if os.Getenv("COMMANDKIT_TEST_DISPLAY_ERRORS") == "1" {
		cfg := New()
		cfg.Define("DATABASE_URL").String().Env("DATABASE_URL").Required().Secret().Description("Database connection string")
		cfg.Define("PORT").Int64().Flag("port").Range(1, 65535).Description("HTTP server port")

		ctx := NewCommandContext([]string{}, cfg, "start", "")

		ctx.execution.CollectError(cfg, "DATABASE_URL", "not found", "", "key not defined", false)
		ctx.execution.CollectError(cfg, "PORT", "validation", "", "value 99999 is greater than maximum 65535", false)
		ctx.execution.DisplayAndExit()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestDisplayGetErrorsAndExit")
	cmd.Env = append(os.Environ(), "COMMANDKIT_TEST_DISPLAY_ERRORS=1")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("Expected command to exit with error")
	}

	outputStr := string(output)
	if !contains(outputStr, "Usage: start [options]") {
		t.Error("Expected templated output to contain command usage")
	}
	if !contains(outputStr, "Configuration errors:") {
		t.Error("Expected templated output to contain configuration errors section")
	}
	if !contains(outputStr, "(no flag) string (env: DATABASE_URL, required) -> key not defined") {
		t.Error("Expected templated output to contain DATABASE_URL error")
	}
	if !contains(outputStr, "--port int64 -> value 99999 is greater than maximum 65535") {
		t.Error("Expected templated output to contain PORT validation error")
	}
}
