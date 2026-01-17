package commandkit

import (
	"os"
	"testing"
)

func TestOverrideWarnings(t *testing.T) {
	ow := NewOverrideWarnings()

	// Test empty warnings
	if ow.HasWarnings() {
		t.Error("Expected no warnings initially")
	}

	if len(ow.GetWarnings()) != 0 {
		t.Error("Expected empty warnings list")
	}

	// Add a warning
	ow.Add(OverrideWarning{
		Key:        "TEST_KEY",
		Source:     "default",
		OverrideBy: "flag",
		Message:    "Test override",
	})

	if !ow.HasWarnings() {
		t.Error("Expected to have warnings after adding one")
	}

	warnings := ow.GetWarnings()
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}

	if warnings[0].Key != "TEST_KEY" {
		t.Errorf("Expected key 'TEST_KEY', got %s", warnings[0].Key)
	}
}

func TestCommandOverrideDetection(t *testing.T) {
	cfg := New()

	// Define global config
	cfg.Define("PORT").Int64().Env("PORT").Flag("port").Default(8080)
	cfg.Define("DEBUG").Bool().Env("DEBUG").Flag("debug").Default(false)

	// Create command with override
	cfg.Command("server").
		Config(func(cc *CommandConfig) {
			cc.Define("PORT").Int64().Flag("port").Default(3000)        // Override global default
			cc.Define("HOST").String().Env("HOST").Default("localhost") // New command-specific
		})

	// Check that warnings were generated
	if !cfg.HasOverrideWarnings() {
		t.Error("Expected override warnings for command config")
	}

	warnings := cfg.GetOverrideWarnings()
	if len(warnings.GetWarnings()) == 0 {
		t.Error("Expected at least one override warning")
	}

	// Check for PORT override warning
	foundPortWarning := false
	for _, warning := range warnings.GetWarnings() {
		if warning.Key == "PORT" && warning.Command == "server" {
			foundPortWarning = true
			if warning.Source != "global config" || warning.OverrideBy != "command config" {
				t.Errorf("Expected global->command override, got %s->%s", warning.Source, warning.OverrideBy)
			}
		}
	}

	if !foundPortWarning {
		t.Error("Expected to find PORT override warning")
	}
}

func TestSourceOverrideDetection(t *testing.T) {
	// Set environment variable
	os.Setenv("TEST_PORT", "9000")
	defer os.Unsetenv("TEST_PORT")

	cfg := New()

	// Define config with flag, env, and default
	cfg.Define("TEST_PORT").Int64().Env("TEST_PORT").Flag("port").Default(8080)

	// Simulate flag value (normally set by flag parsing)
	cfg.flagValues["TEST_PORT"] = &[]string{"3000"}[0]

	// Process config to trigger override detection
	cfg.Process()

	// Check for warnings
	if !cfg.HasOverrideWarnings() {
		t.Error("Expected override warnings for source overrides")
	}

	warnings := cfg.GetOverrideWarnings()
	if len(warnings.GetWarnings()) == 0 {
		t.Error("Expected at least one override warning")
	}

	// Should have flag overriding env and default
	foundFlagOverride := false
	overrideCount := 0
	for _, warning := range warnings.GetWarnings() {
		if warning.Key == "TEST_PORT" && warning.OverrideBy == "flag" {
			foundFlagOverride = true
			overrideCount++

			// Check that it's overriding either env or default
			if warning.Source != "environment" && warning.Source != "default" {
				t.Errorf("Expected flag to override env or default, got %s", warning.Source)
			}
		}
	}

	if !foundFlagOverride {
		t.Error("Expected to find flag override warning")
	}

	if overrideCount < 2 {
		t.Errorf("Expected flag to override both env and default, got %d overrides", overrideCount)
	}
}

func TestOverrideWarningFormatting(t *testing.T) {
	ow := NewOverrideWarnings()

	ow.Add(OverrideWarning{
		Key:        "PORT",
		Command:    "server",
		Source:     "global config",
		OverrideBy: "command config",
		OldValue:   "8080",
		NewValue:   "3000",
		Message:    "Command-specific configuration overrides global configuration",
	})

	ow.Add(OverrideWarning{
		Key:        "DEBUG",
		Source:     "environment",
		OverrideBy: "flag",
		OldValue:   "false",
		NewValue:   "true",
		Message:    "Command-line flag overrides environment variable",
	})

	formatted := ow.FormatWarnings()

	// Check that formatted output contains expected elements
	if !containsString(formatted, "Warning: Configuration overrides detected") {
		t.Error("Expected formatted output to contain 'Warning: Configuration overrides detected'")
	}

	if !containsString(formatted, "PORT") {
		t.Error("Expected formatted output to contain 'PORT'")
	}

	if !containsString(formatted, "server") {
		t.Error("Expected formatted output to contain 'server'")
	}

	if !containsString(formatted, "DEBUG") {
		t.Error("Expected formatted output to contain 'DEBUG'")
	}

	if !containsString(formatted, "Total: 2 override(s)") {
		t.Error("Expected formatted output to contain total count")
	}

	// Check for simple separator lines
	if !containsString(formatted, "==================================================") {
		t.Error("Expected formatted output to contain separator lines")
	}
}

func TestSecretOverrideWarnings(t *testing.T) {
	cfg := New()

	// Define secret in global config
	cfg.Define("API_KEY").String().Env("API_KEY").Secret().Default("default-key")

	// Create command that overrides secret
	cfg.Command("api").
		Config(func(cc *CommandConfig) {
			cc.Define("API_KEY").String().Env("API_KEY").Secret() // No default, different definition
		})

	// Check warnings
	if !cfg.HasOverrideWarnings() {
		t.Error("Expected override warnings for secret config")
	}

	warnings := cfg.GetOverrideWarnings()
	foundSecretWarning := false
	for _, warning := range warnings.GetWarnings() {
		if warning.Key == "API_KEY" && warning.Command == "api" {
			foundSecretWarning = true
			// Secret values should be masked in warnings
			if warning.OldValue != "" && !containsString(warning.OldValue, "[SECRET:") {
				t.Error("Expected secret value to be masked in warning")
			}
		}
	}

	if !foundSecretWarning {
		t.Error("Expected to find API_KEY override warning")
	}
}

func TestNoOverrideWarnings(t *testing.T) {
	cfg := New()

	// Define global config
	cfg.Define("PORT").Int64().Env("PORT").Flag("port").Default(8080)

	// Create command with identical config (no overrides)
	cfg.Command("server").
		Config(func(cc *CommandConfig) {
			cc.Define("PORT").Int64().Env("PORT").Flag("port").Default(8080) // Same as global
		})

	// Should have no warnings for identical definitions
	if cfg.HasOverrideWarnings() {
		t.Error("Expected no override warnings for identical definitions")
	}
}

func TestDefinitionsEqual(t *testing.T) {
	cfg := New()

	// Create identical definitions
	def1 := &Definition{
		key:          "PORT",
		valueType:    TypeInt64,
		flag:         "port",
		envVar:       "PORT",
		defaultValue: int64(8080),
		required:     false,
		secret:       false,
	}

	def2 := &Definition{
		key:          "PORT",
		valueType:    TypeInt64,
		flag:         "port",
		envVar:       "PORT",
		defaultValue: int64(8080),
		required:     false,
		secret:       false,
	}

	if !cfg.definitionsEqual(def1, def2) {
		t.Error("Expected identical definitions to be equal")
	}

	// Different default value
	def3 := &Definition{
		key:          "PORT",
		valueType:    TypeInt64,
		flag:         "port",
		envVar:       "PORT",
		defaultValue: int64(3000),
		required:     false,
		secret:       false,
	}

	if cfg.definitionsEqual(def1, def3) {
		t.Error("Expected definitions with different defaults to be unequal")
	}
}

func TestMaskValueIfNeeded(t *testing.T) {
	cfg := New()

	// Define a secret and a regular key
	cfg.Define("SECRET_KEY").String().Secret()
	cfg.Define("REGULAR_KEY").String()

	// Test masking secret
	masked := cfg.maskValueIfNeeded("SECRET_KEY", "my-secret-value")
	if masked == "my-secret-value" {
		t.Error("Expected secret value to be masked")
	}

	if !containsString(masked, "****") {
		t.Error("Expected masked value to contain asterisks")
	}

	// Test non-secret value
	regular := cfg.maskValueIfNeeded("REGULAR_KEY", "regular-value")
	if regular != "regular-value" {
		t.Error("Expected regular value to not be masked")
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
