package commandkit

import (
	"os"
	"os/exec"
	"testing"
)

func TestMustGet(t *testing.T) {
	cfg := New()

	cfg.Define("PORT").Int64().Default(int64(8080))
	cfg.Define("HOST").String() // No default, not required

	result := cfg.Process()
	if result.Error != nil {
		t.Fatalf("Unexpected errors: %v", result.Error)
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

	result := cfg.Process()
	if result.Error != nil {
		t.Fatalf("Unexpected errors: %v", result.Error)
	}

	// Create context for new API
	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Test Get with missing key
	getResult := Get[string](ctx, "MISSING_KEY")
	if getResult.Error == nil {
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

func TestGetWithTypeMismatch(t *testing.T) {
	cfg := New()

	cfg.Define("PORT").String().Default("8080") // String but we'll try to get as int64

	result := cfg.Process()
	if result.Error != nil {
		t.Fatalf("Unexpected errors: %v", result.Error)
	}

	// Create context for new API
	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Test Get with type mismatch
	typeResult := Get[int64](ctx, "PORT")
	if typeResult.Error == nil {
		t.Error("Expected error for type mismatch")
	}

	// Verify error was collected
	if !ctx.execution.HasErrors() {
		t.Error("Expected error to be collected for type mismatch")
	}

	collected := ctx.execution.GetErrors()
	if len(collected) == 0 {
		t.Error("Expected error to be collected for type mismatch")
	}

	if collected[0].Key != "PORT" {
		t.Errorf("Expected key 'PORT', got '%s'", collected[0].Key)
	}

	if collected[0].ExpectedType != "int64" {
		t.Errorf("Expected expected type 'int64', got '%s'", collected[0].ExpectedType)
	}

	if collected[0].ActualType != "string" {
		t.Errorf("Expected actual type 'string', got '%s'", collected[0].ActualType)
	}
}

func TestGetWithSecret(t *testing.T) {
	cfg := New()

	cfg.Define("API_KEY").String().Secret()

	result := cfg.Process()
	if result.Error != nil {
		t.Fatalf("Unexpected errors: %v", result.Error)
	}

	// Create context for new API
	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Test Get with secret (should return error)
	secretResult := Get[string](ctx, "API_KEY")
	if secretResult.Error == nil {
		t.Error("Expected error for secret access")
	}

	if secretResult.Error.Error() != "validation error: configuration 'API_KEY' is secret, use GetSecret() instead" {
		t.Errorf("Expected secret access error, got: %v", secretResult.Error)
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
	ctx.execution.CollectErrorWithConfig(cfg, "PORT", "not found", "", "key not defined", false)
	ctx.execution.CollectErrorWithConfig(cfg, "DATABASE_URL", "secret", "", "use GetSecret() instead", true)
	ctx.execution.CollectErrorWithConfig(cfg, "DEBUG", "not found", "", "key not defined", false)

	collected := ctx.execution.GetErrors()
	if len(collected) != 3 {
		t.Errorf("Expected 3 collected errors, got %d", len(collected))
	}

	// Test display name formatting
	portDisplayName := getErrorDisplayName(collected[0], cfg)
	if portDisplayName != "-port int64 (env: PORT)" {
		t.Errorf("Expected '-port int64 (env: PORT)', got '%s'", portDisplayName)
	}

	dbDisplayName := getErrorDisplayName(collected[1], cfg)
	if dbDisplayName != "(no flag) string (env: DATABASE_URL, secret)" {
		t.Errorf("Expected '(no flag) string (env: DATABASE_URL, secret)', got '%s'", dbDisplayName)
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

		ctx.execution.CollectErrorWithConfig(cfg, "DATABASE_URL", "not found", "", "key not defined", false)
		ctx.execution.CollectErrorWithConfig(cfg, "PORT", "validation", "", "value 99999 is greater than maximum 65535", false)
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
	if !contains(outputStr, "Configuration errors detected") {
		t.Error("Expected error message to contain 'Configuration errors detected'")
	}
	if !contains(outputStr, "(no flag) string (env: DATABASE_URL, required, secret) not defined") {
		t.Error("Expected error message to contain DATABASE_URL not defined")
	}
	if !contains(outputStr, "validation failed: value 99999 is greater than maximum 65535") {
		t.Error("Expected error message to contain PORT validation failed")
	}
	if !contains(outputStr, "Use 'start --help' for more information") {
		t.Error("Expected help message to contain command context")
	}
}
