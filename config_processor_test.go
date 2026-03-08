// commandkit/config_processor_test.go
package commandkit

import (
	"os"
	"testing"
)

func TestConfigProcessor_ProcessCommandConfig(t *testing.T) {
	processor := newConfigProcessor()

	// Create a command with definitions
	cmd := &Command{
		Name: "start",
		Definitions: map[string]*Definition{
			"PORT": {
				key:          "PORT",
				valueType:    TypeInt64,
				flag:         "port",
				description:  "Server port",
				defaultValue: 8080,
			},
			"ENV": {
				key:         "ENV",
				valueType:   TypeString,
				flag:        "env",
				description: "Environment",
				required:    true,
			},
		},
	}

	// Create a base config
	baseConfig := New()

	// Create command context with args
	ctx := NewCommandContext([]string{"--port", "9000", "--env", "prod"}, baseConfig, "start", "")

	// Process command config
	result := processor.ProcessCommandConfig(cmd, ctx)

	// Check that processing succeeded
	if result.Error != nil {
		t.Errorf("ProcessCommandConfig() returned error: %v", result.Error)
	}

	// Check that command config was set (not global config)
	if ctx.CommandConfig == nil {
		t.Error("CommandConfig should have been set with temporary config")
	}
	if ctx.CommandConfig == baseConfig {
		t.Error("CommandConfig should be different from global config")
	}
	// Global config should remain unchanged
	if ctx.GlobalConfig != baseConfig {
		t.Error("GlobalConfig should remain unchanged")
	}

	// Check that flag values were parsed
	port, err := Get[int64](ctx, "PORT")
	if err != nil {
		t.Errorf("Failed to get PORT value: %v", err)
	}
	if port != 9000 {
		t.Errorf("Expected PORT=9000, got %d", port)
	}

	env, err := Get[string](ctx, "ENV")
	if err != nil {
		t.Errorf("Failed to get ENV value: %v", err)
	}
	if env != "prod" {
		t.Errorf("Expected ENV=prod, got %s", env)
	}
}

func TestConfigProcessor_ProcessCommandConfig_WitherrorResult(t *testing.T) {
	processor := newConfigProcessor()

	// Create a command with validation
	cmd := &Command{
		Name: "start",
		Definitions: map[string]*Definition{
			"PORT": {
				key:         "PORT",
				valueType:   TypeInt64,
				flag:        "port",
				description: "Server port",
				required:    true, // Make it required to trigger validation error
			},
		},
	}

	// Create a base config
	baseConfig := New()

	// Create command context with missing required flag
	ctx := NewCommandContext([]string{}, baseConfig, "start", "")

	// Process command config
	result := processor.ProcessCommandConfig(cmd, ctx)

	// Check that processing failed
	if result.Error == nil {
		t.Error("ProcessCommandConfig() should have returned an error for missing required flag")
	}

	if result.Message != "" {
		t.Error("ProcessCommandConfig() should not return a pre-rendered message block for missing required flag")
	}

	if !ctx.execution.HasErrors() {
		t.Error("Execution context should retain collected config errors for templated rendering")
	}
}

func TestConfigProcessor_ValidateRequiredFlags(t *testing.T) {
	processor := newConfigProcessor()

	// Create a command with required flags
	cmd := &Command{
		Name: "deploy",
		Definitions: map[string]*Definition{
			"ENV": {
				key:         "ENV",
				valueType:   TypeString,
				flag:        "env",
				description: "Environment",
				required:    true,
			},
			"SECRET": {
				key:         "SECRET",
				valueType:   TypeString,
				envVar:      "API_SECRET",
				description: "API secret",
				required:    true,
			},
			"HOST": {
				key:          "HOST",
				valueType:    TypeString,
				flag:         "host",
				description:  "Host address",
				defaultValue: "localhost",
				required:     true,
			},
		},
	}

	// Create a config with some values set
	config := New()

	// Set flag value for ENV
	config.flagValues["ENV"] = new(string)
	*config.flagValues["ENV"] = "prod"

	// Set environment variable for SECRET
	os.Setenv("API_SECRET", "test-secret")
	defer os.Unsetenv("API_SECRET")

	// Create command context
	ctx := NewCommandContext([]string{}, config, "deploy", "")

	// Validate required flags
	result := processor.ValidateRequiredFlags(cmd, ctx)

	// Check that validation succeeded (returns Success)
	if result.Error != nil {
		t.Errorf("ValidateRequiredFlags() returned error: %v", result.Error)
	}
}

func TestConfigProcessor_ValidateRequiredFlags_Missing(t *testing.T) {
	processor := newConfigProcessor()

	// Create a command with required flags
	cmd := &Command{
		Name: "deploy",
		Definitions: map[string]*Definition{
			"ENV": {
				key:         "ENV",
				valueType:   TypeString,
				flag:        "env",
				description: "Environment",
				required:    true,
			},
			"SECRET": {
				key:         "SECRET",
				valueType:   TypeString,
				envVar:      "API_SECRET",
				description: "API secret",
				required:    true,
			},
		},
	}

	// Create a config without required values
	config := New()

	// Create command context
	ctx := NewCommandContext([]string{}, config, "deploy", "")

	// Validate required flags (this will log warnings but should not error)
	result := processor.ValidateRequiredFlags(cmd, ctx)

	// Check that validation succeeded (warnings are logged, not returned as errors)
	if result.Error != nil {
		t.Errorf("ValidateRequiredFlags() should not return error for missing flags: %v", result.Error)
	}
}

func TestConfigProcessor_ValidateRequiredFlags_NoDefinitions(t *testing.T) {
	processor := newConfigProcessor()

	// Create a command without definitions
	cmd := &Command{
		Name: "simple",
	}

	// Create command context
	ctx := NewCommandContext([]string{}, New(), "simple", "")

	// Validate required flags should succeed
	result := processor.ValidateRequiredFlags(cmd, ctx)

	if result.Error != nil {
		t.Errorf("ValidateRequiredFlags() returned error for command without definitions: %v", result.Error)
	}
}

func TestConfigProcessor_Integration(t *testing.T) {
	// Test that ConfigProcessor works correctly with the service factory
	services := newCommandServices()
	processor := services.ConfigProcessor

	// Create a simple command
	cmd := &Command{
		Name: "test",
		Definitions: map[string]*Definition{
			"DEBUG": {
				key:          "DEBUG",
				valueType:    TypeBool,
				flag:         "debug",
				description:  "Enable debug mode",
				defaultValue: false,
			},
		},
	}

	ctx := NewCommandContext([]string{"--debug", "true"}, New(), "test", "")

	// Test ProcessCommandConfig
	result := processor.ProcessCommandConfig(cmd, ctx)
	if result.Error != nil {
		t.Errorf("Integrated ProcessCommandConfig failed: %v", result.Error)
	}

	// Test ValidateRequiredFlags
	result = processor.ValidateRequiredFlags(cmd, ctx)
	if result.Error != nil {
		t.Errorf("Integrated ValidateRequiredFlags failed: %v", result.Error)
	}
}

func TestConfigProcessor_ErrorHandling(t *testing.T) {
	processor := newConfigProcessor()

	// Test with nil command (should not panic)
	ctx := NewCommandContext([]string{}, New(), "test", "")

	// This should handle the nil case gracefully - just return success
	result := processor.ValidateRequiredFlags(nil, ctx)
	if result.Error != nil {
		t.Errorf("ValidateRequiredFlags with nil command should succeed: %v", result.Error)
	}

	// Test ProcessCommandConfig with nil command
	result = processor.ProcessCommandConfig(nil, ctx)
	if result.Error == nil {
		t.Error("ProcessCommandConfig with nil command should return an error")
	}
}
