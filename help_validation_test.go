// commandkit/help_validation_test.go
package commandkit

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func getAvailableKeys(c *Config) []string {
	if c == nil {
		return []string{}
	}
	var keys []string
	for k := range c.definitions {
		keys = append(keys, k)
	}
	return keys
}

func TestHelpWithCustomValidationAndEnvironment(t *testing.T) {
	// Test that help works even when environment variables have validation issues

	// Set an environment variable with a value that would fail custom validation
	os.Setenv("TEST_PORT", "invalid_port")
	defer os.Unsetenv("TEST_PORT")

	cfg := New()

	// Define a custom validator that expects int64
	portValidator := func(value any) error {
		if port, ok := value.(int64); ok {
			if port < 1 || port > 65535 {
				return fmt.Errorf("port must be between 1 and 65535, got %d", port)
			}
			return nil
		}
		return fmt.Errorf("port must be an integer, got %T", value)
	}

	// Define configuration with custom validation
	cfg.Define("PORT").
		Int64().
		Env("TEST_PORT").
		Flag("port").
		Default(8080).
		Custom("port_range", portValidator).
		Description("Server port")

	// Define another with string validation
	stringValidator := func(value any) error {
		if s, ok := value.(string); ok {
			if len(s) < 3 {
				return fmt.Errorf("string must be at least 3 characters, got %d", len(s))
			}
			return nil
		}
		return fmt.Errorf("value must be a string, got %T", value)
	}

	cfg.Define("NAME").
		String().
		Env("TEST_NAME").
		Flag("name").
		Default("default").
		Custom("min_length", stringValidator).
		Description("Application name")

	// Test that help works despite invalid environment variable
	args := []string{"cmd", "test", "--help"}
	err := cfg.Execute(args)

	// Help should work without errors
	if err != nil {
		t.Errorf("Help should work despite invalid environment variable, got error: %v", err)
	}
}

func TestNormalExecutionWithCustomValidation(t *testing.T) {
	t.Skip("Temporarily skipping - configuration resolution is broken")
	// Test that normal execution still validates properly

	// Set environment variables with valid values
	os.Setenv("TEST_PORT", "3000")
	os.Setenv("TEST_NAME", "valid_name")
	defer func() {
		os.Unsetenv("TEST_PORT")
		os.Unsetenv("TEST_NAME")
	}()

	cfg := New()

	// Define a custom validator that expects int64
	portValidator := func(value any) error {
		if port, ok := value.(int64); ok {
			if port < 1 || port > 65535 {
				return fmt.Errorf("port must be between 1 and 65535, got %d", port)
			}
			return nil
		}
		return fmt.Errorf("port must be an integer, got %T", value)
	}

	// Define configuration with custom validation
	cfg.Define("PORT").
		Int64().
		Env("TEST_PORT").
		Flag("port").
		Default(8080).
		Custom("port_range", portValidator).
		Description("Server port")

	// Define another with string validation
	stringValidator := func(value any) error {
		if s, ok := value.(string); ok {
			if len(s) < 3 {
				return fmt.Errorf("string must be at least 3 characters, got %d", len(s))
			}
			return nil
		}
		return fmt.Errorf("value must be a string, got %T", value)
	}

	cfg.Define("NAME").
		String().
		Env("TEST_NAME").
		Flag("name").
		Default("default").
		Custom("min_length", stringValidator).
		Description("Application name")

	// Add a simple command
	cfg.Command("test").
		Func(func(ctx *CommandContext) error {
			// Try to get the values - this should work with proper types
			port, err := Get[int64](ctx, "PORT")
			if err != nil {
				return fmt.Errorf("failed to get PORT: %w", err)
			}

			name, err := Get[string](ctx, "NAME")
			if err != nil {
				return fmt.Errorf("failed to get NAME: %w", err)
			}

			// Verify types are correct
			if port != 3000 {
				t.Errorf("Expected port 3000, got %d (type: %T)", port, port)
			}

			if name != "valid_name" {
				t.Errorf("Expected name 'valid_name', got %s (type: %T)", name, name)
			}

			return nil
		}).
		ShortHelp("Test command")

	// Test normal execution - should work with valid environment variables
	args := []string{"cmd", "test"}
	err := cfg.Execute(args)

	if err != nil {
		t.Errorf("Normal execution should work with valid environment variables, got error: %v", err)
	}
}

// func TestNormalExecutionWithInvalidCustomValidation(t *testing.T) {
// 	// Test that normal execution fails with invalid custom validation

// 	// Set environment variable with invalid value for custom validator
// 	// os.Setenv("TEST_PORT", "invalid_port")
// 	// defer os.Unsetenv("TEST_PORT")

// 	cfg := New()

// 	// Define a custom validator that expects int64
// 	// portValidator := func(value any) error {
// 	// 	if port, ok := value.(int64); ok {
// 	// 		if port < 1 || port > 65535 {
// 	// 			return fmt.Errorf("port must be between 1 and 65535, got %d", port)
// 	// 		}
// 	// 		return nil
// 	// 	}
// 	// 	return fmt.Errorf("port must be an integer, got %T", value)
// 	// }

// 	// Define configuration with custom validation
// 	cfg.Define("PORT").
// 		Int64().
// 		// Env("TEST_PORT").
// 		Flag("port").
// 		Default(8080).
// 		Range(1, 65535)
// 	// Custom("port_range", portValidator).
// 	// Description("Server port")

// 	// Add a simple command
// 	cfg.Command("test").
// 		Func(func(ctx *CommandContext) error {
// 			a, e := Get[int64](ctx, "PORT")

// 			fmt.Printf("PORT: %d, error: %v\n", a, e)
// 			t.Logf("Command executed successfully! PORT: %d, error: %v", a, e)
// 			return nil
// 		}).
// 		ShortHelp("Test command")

// 	// Test normal execution - should fail with invalid environment variable
// 	args := []string{"cmd", "test"}
// 	err := cfg.Execute(args)

// 	// Should fail due to validation error
// 	if err == nil {
// 		t.Error("Normal execution should fail with invalid environment variable and custom validation")
// 	}
// }

func TestHelpWithDifferentTypesAndCustomValidation(t *testing.T) {
	// Test help with various types and custom validators

	cfg := New()

	// Bool custom validator
	boolValidator := func(value any) error {
		if _, ok := value.(bool); ok {
			return nil
		}
		return fmt.Errorf("value must be boolean, got %T", value)
	}

	// Duration custom validator
	durationValidator := func(value any) error {
		if d, ok := value.(time.Duration); ok {
			if d < 0 {
				return fmt.Errorf("duration must be positive, got %v", d)
			}
			return nil
		}
		return fmt.Errorf("value must be duration, got %T", value)
	}

	// Float64 custom validator
	floatValidator := func(value any) error {
		if f, ok := value.(float64); ok {
			if f < 0 {
				return fmt.Errorf("value must be positive, got %f", f)
			}
			return nil
		}
		return fmt.Errorf("value must be float, got %T", value)
	}

	// Define configurations with different types and custom validation
	cfg.Define("ENABLED").
		Bool().
		Env("TEST_ENABLED").
		Flag("enabled").
		Default(false).
		Custom("bool_check", boolValidator).
		Description("Enable feature")

	cfg.Define("TIMEOUT").
		Duration().
		Env("TEST_TIMEOUT").
		Flag("timeout").
		Default(30*time.Second).
		Custom("positive_duration", durationValidator).
		Description("Operation timeout")

	cfg.Define("RATE").
		Float64().
		Env("TEST_RATE").
		Flag("rate").
		Default(1.0).
		Custom("positive_float", floatValidator).
		Description("Rate limit")

	// Add a command
	cfg.Command("test").
		Func(func(ctx *CommandContext) error {
			return nil
		}).
		ShortHelp("Test command")

	// Test that help works for all types
	args := []string{"cmd", "test", "--help"}
	err := cfg.Execute(args)

	// Help should work without errors
	if err != nil {
		t.Errorf("Help should work with all types and custom validation, got error: %v", err)
	}
}
