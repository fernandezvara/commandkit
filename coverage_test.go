package commandkit

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestAddSubCommand tests the AddSubCommand functionality
func TestAddSubCommand(t *testing.T) {
	cmd := NewCommand("parent")
	cmd.Func = func(ctx *CommandContext) error {
		return nil
	}

	subCmd := NewCommand("child")
	subCmd.Func = func(ctx *CommandContext) error {
		return nil
	}

	// Test adding subcommand
	cmd.AddSubCommand("child", subCmd)

	// Verify subcommand was added
	if len(cmd.SubCommands) != 1 {
		t.Errorf("Expected 1 subcommand, got %d", len(cmd.SubCommands))
	}

	if found, exists := cmd.SubCommands["child"]; !exists {
		t.Error("Subcommand 'child' not found")
	} else if found != subCmd {
		t.Error("Wrong subcommand stored")
	}
}

// TestCommandContextHelpers tests CommandContext helper methods
func TestCommandContextHelpers(t *testing.T) {
	ctx := NewCommandContext([]string{}, New(), "test", "")

	// Test Set and Get
	ctx.Set("key", "value")
	if val, exists := ctx.GetData("key"); !exists {
		t.Error("Key not found after Set")
	} else if val != "value" {
		t.Errorf("Expected 'value', got %v", val)
	}

	// Test ContextGet string
	ctx.Set("str_key", "string_value")
	if str := ContextGet[string](ctx, "str_key"); str != "string_value" {
		t.Errorf("Expected 'string_value', got %s", str)
	}

	// Test ContextGet int
	ctx.Set("int_key", 42)
	if i := ContextGet[int](ctx, "int_key"); i != 42 {
		t.Errorf("Expected 42, got %d", i)
	}

	// Test ContextGet bool
	ctx.Set("bool_key", true)
	if b := ContextGet[bool](ctx, "bool_key"); !b {
		t.Error("Expected true, got false")
	}

	// Test non-existent keys
	if _, exists := ctx.GetData("nonexistent"); exists {
		t.Error("Non-existent key should not exist")
	}
	if str := ContextGet[string](ctx, "nonexistent"); str != "" {
		t.Errorf("Expected empty string for non-existent key, got %s", str)
	}
	if i := ContextGet[int](ctx, "nonexistent"); i != 0 {
		t.Errorf("Expected 0 for non-existent key, got %d", i)
	}
	if b := ContextGet[bool](ctx, "nonexistent"); b {
		t.Error("Expected false for non-existent key, got true")
	}
}

// TestConfigPrintErrors tests the PrintErrors functionality
func TestConfigPrintErrors(t *testing.T) {
	cfg := New()

	// Create some config errors
	errs := []ConfigError{
		{Key: "TEST_KEY", Source: "none", Message: "required value not provided"},
		{Key: "ANOTHER_KEY", Source: "env", Message: "invalid format"},
	}

	// Capture stderr output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cfg.PrintErrors(errs)

	// Close pipe and read output
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stderr = oldStderr

	output := buf.String()
	if !strings.Contains(output, "TEST_KEY") {
		t.Error("PrintErrors output should contain TEST_KEY")
	}
	if !strings.Contains(output, "ANOTHER_KEY") {
		t.Error("PrintErrors output should contain ANOTHER_KEY")
	}
}

// TestConfigDestroy tests the Destroy functionality
func TestConfigDestroy(t *testing.T) {
	cfg := New()

	// Add a secret
	cfg.Define("SECRET").String().Secret()
	cfg.Process()
	secret := cfg.GetSecret("SECRET")
	cfg.secrets.Store("SECRET", "test_secret_value")
	secret = cfg.GetSecret("SECRET")

	// Verify secret exists
	if !secret.IsSet() {
		t.Error("Secret should be set before destroy")
	}

	// Destroy config
	cfg.Destroy()

	// Verify secret is destroyed
	if secret.IsSet() {
		t.Error("Secret should be destroyed after config.Destroy()")
	}
}

// TestConfigIsSecret tests the IsSecret functionality
func TestConfigIsSecret(t *testing.T) {
	cfg := New()

	// Define a secret and a regular key
	cfg.Define("SECRET_KEY").String().Secret()
	cfg.Define("REGULAR_KEY").String()
	cfg.Process()

	// Test secret detection
	if !cfg.IsSecret("SECRET_KEY") {
		t.Error("SECRET_KEY should be detected as secret")
	}

	if cfg.IsSecret("REGULAR_KEY") {
		t.Error("REGULAR_KEY should not be detected as secret")
	}

	if cfg.IsSecret("NONEXISTENT_KEY") {
		t.Error("Non-existent key should not be detected as secret")
	}
}

// TestConfigPrintOverrideWarnings tests the PrintOverrideWarnings functionality
func TestConfigPrintOverrideWarnings(t *testing.T) {
	cfg := New()

	// Create some override warnings
	cfg.overrideWarnings = NewOverrideWarnings()
	cfg.overrideWarnings.warnings = []OverrideWarning{
		{
			Key:        "TEST_KEY",
			Source:     "default",
			OverrideBy: "environment",
			Message:    "Environment variable overrides default value",
		},
	}

	// Capture stderr output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cfg.PrintOverrideWarnings()

	// Close pipe and read output
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stderr = oldStderr

	output := buf.String()
	if !strings.Contains(output, "TEST_KEY") {
		t.Error("PrintOverrideWarnings output should contain TEST_KEY")
	}
	if !strings.Contains(output, "Environment variable overrides default value") {
		t.Error("PrintOverrideWarnings output should contain override message")
	}
}

// TestConfigGenerateHelp tests the GenerateHelp functionality
func TestConfigGenerateHelp(t *testing.T) {
	cfg := New()

	// Add some configuration
	cfg.Define("PORT").Int64().Default(8080).Description("Server port")
	cfg.Define("HOST").String().Default("localhost").Description("Server host")
	cfg.Process()

	// Generate help
	help := cfg.GenerateHelp()

	// Verify help content
	if !strings.Contains(help, "PORT") {
		t.Error("Help should contain PORT")
	}
	if !strings.Contains(help, "HOST") {
		t.Error("Help should contain HOST")
	}
	if !strings.Contains(help, "Server port") {
		t.Error("Help should contain PORT description")
	}
	if !strings.Contains(help, "Server host") {
		t.Error("Help should contain HOST description")
	}
}

// TestCommandExecuteEdgeCases tests edge cases in Command.Execute
func TestCommandExecuteEdgeCases(t *testing.T) {
	// Test command with no function but with subcommands
	parentCmd := &Command{
		Name: "parent",
		SubCommands: map[string]*Command{
			"child": {
				Name: "child",
				Func: func(ctx *CommandContext) error {
					return nil
				},
			},
		},
	}

	ctx := NewCommandContext([]string{}, New(), "parent", "")
	err := parentCmd.Execute(ctx)
	if err == nil {
		t.Error("Expected error for command with no function but subcommands")
	}

	// Test command with no function and no subcommands
	emptyCmd := &Command{
		Name: "empty",
	}

	ctx = NewCommandContext([]string{}, New(), "empty", "")
	err = emptyCmd.Execute(ctx)
	if err == nil {
		t.Error("Expected error for command with no function and no subcommands")
	}
	if !strings.Contains(err.Error(), "no implementation") {
		t.Errorf("Expected 'no implementation' error, got: %s", err.Error())
	}
}

// TestConfigDump tests the Dump functionality
func TestConfigDump(t *testing.T) {
	cfg := New()

	// Add some configuration
	cfg.Define("PORT").Int64().Default(8080)
	cfg.Define("HOST").String().Default("localhost")
	cfg.Define("SECRET").String().Secret()
	cfg.Process()

	// Set some values
	cfg.values["PORT"] = int64(3000)
	cfg.values["HOST"] = "example.com"
	cfg.secrets.Store("SECRET", "secret_value")

	// Dump configuration
	output := cfg.Dump()

	// Convert map to string for checking
	dumpStr := fmt.Sprintf("%v", output)

	// Verify dump content
	if !strings.Contains(dumpStr, "PORT") {
		t.Error("Dump should contain PORT")
	}
	if !strings.Contains(dumpStr, "HOST") {
		t.Error("Dump should contain HOST")
	}
	if !strings.Contains(dumpStr, "3000") {
		t.Error("Dump should contain PORT value")
	}
	if !strings.Contains(dumpStr, "example.com") {
		t.Error("Dump should contain HOST value")
	}
	// Secret should be masked with bytes format
	if strings.Contains(dumpStr, "secret_value") {
		t.Error("Dump should not contain actual secret value")
	}
	if !strings.Contains(dumpStr, "[SECRET:") && !strings.Contains(dumpStr, "bytes]") {
		t.Error("Dump should contain masked secret in format [SECRET:X bytes]")
	}
}

// TestGetErrorCollectionIntegration tests the complete error collection flow
func TestGetErrorCollectionIntegration(t *testing.T) {
	cfg := New()

	// Define required configuration with flags and env vars
	cfg.Define("PORT").Int64().Flag("port").Env("PORT").Required()
	cfg.Define("HOST").String().Flag("host").Env("HOST").Required()
	cfg.Define("API_KEY").String().Flag("api-key").Env("API_KEY").Secret().Required()

	cfg.Process()

	// Clear any previous errors
	ctx := NewCommandContext([]string{}, cfg, "test", "")

	// Execute - this should collect errors and exit
	// We can't test the os.Exit directly, but we can verify error collection
	// by calling the Get functions directly
	// Note: Get functions now return error for missing data
	_, err := Get[int64](ctx, "PORT")
	if err == nil {
		t.Errorf("Expected error for missing PORT, got nil")
	}
	_, err = Get[string](ctx, "HOST")
	if err == nil {
		t.Errorf("Expected error for missing HOST, got nil")
	}

	_, err = Get[string](ctx, "API_KEY")
	if err == nil {
		t.Errorf("Expected error for missing API_KEY, got nil")
	}

	// Note: Required keys don't collect errors, they return errors directly
	// The new behavior separates required validation from error collection
	// Only non-required keys collect errors in the execution context

	// Test non-required data still collects errors
	_, err = Get[string](ctx, "NONEXISTENT_KEY")
	if err == nil {
		t.Error("Expected error for non-required key")
	}

	// Now we should have collected errors for the non-required key
	if !ctx.execution.HasErrors() {
		t.Error("Expected errors to be collected for non-required data")
	}

	collected := ctx.execution.GetErrors()
	if len(collected) == 0 {
		t.Errorf("Expected collected errors for non-required data, got %d", len(collected))
	}

	// Test that we can add more errors
	ctx.execution.CollectErrorWithConfig(cfg, "ANOTHER_KEY", "not found", "", "test error", false)
	collectedAfter := ctx.execution.GetErrors()
	if len(collectedAfter) != len(collected)+1 {
		t.Errorf("Expected %d collected errors, got %d", len(collected)+1, len(collectedAfter))
	}
}
