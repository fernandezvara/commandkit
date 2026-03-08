package commandkit

import (
	"os"
	"testing"
)

func TestSourcePriorityTypes(t *testing.T) {
	// Test SourceType String() method
	tests := []struct {
		source   SourceType
		expected string
	}{
		{SourceDefault, "default"},
		{SourceFlag, "flag"},
		{SourceEnv, "environment"},
		{SourceFile, "file"},
		{SourceType(999), "unknown"},
	}

	for _, test := range tests {
		if test.source.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.source.String())
		}
	}
}

func TestPriorityPresets(t *testing.T) {
	// Test that presets are correctly defined
	if len(PriorityFlagEnvDefault) != 3 {
		t.Error("PriorityFlagEnvDefault should have 3 elements")
	}
	if PriorityFlagEnvDefault[0] != SourceFlag || PriorityFlagEnvDefault[1] != SourceEnv || PriorityFlagEnvDefault[2] != SourceDefault {
		t.Error("PriorityFlagEnvDefault should be [Flag, Env, Default]")
	}

	if len(PriorityEnvFlagDefault) != 3 {
		t.Error("PriorityEnvFlagDefault should have 3 elements")
	}
	if PriorityEnvFlagDefault[0] != SourceEnv || PriorityEnvFlagDefault[1] != SourceFlag || PriorityEnvFlagDefault[2] != SourceDefault {
		t.Error("PriorityEnvFlagDefault should be [Env, Flag, Default]")
	}

	if len(PriorityFileEnvFlagDefault) != 4 {
		t.Error("PriorityFileEnvFlagDefault should have 4 elements")
	}
	if PriorityFileEnvFlagDefault[0] != SourceFile || PriorityFileEnvFlagDefault[1] != SourceEnv || PriorityFileEnvFlagDefault[2] != SourceFlag || PriorityFileEnvFlagDefault[3] != SourceDefault {
		t.Error("PriorityFileEnvFlagDefault should be [File, Env, Flag, Default]")
	}
}

func TestConfigDefaultPriority(t *testing.T) {
	cfg := New()

	// Test default priority
	defaultPriority := cfg.GetDefaultPriority()
	if len(defaultPriority) != 3 {
		t.Error("Default priority should have 3 elements")
	}
	if defaultPriority[0] != SourceFlag || defaultPriority[1] != SourceEnv || defaultPriority[2] != SourceDefault {
		t.Error("Default priority should be [Flag, Env, Default]")
	}

	// Test setting custom priority
	customPriority := SourcePriority{SourceEnv, SourceDefault}
	cfg.SetDefaultPriority(customPriority)

	newPriority := cfg.GetDefaultPriority()
	if len(newPriority) != 2 {
		t.Error("Custom priority should have 2 elements")
	}
	if newPriority[0] != SourceEnv || newPriority[1] != SourceDefault {
		t.Error("Custom priority should be [Env, Default]")
	}
}

func TestDefinitionPriorityMethods(t *testing.T) {
	cfg := New()

	// Test Sources() method
	def := cfg.Define("TEST").String().Sources(SourceFlag, SourceEnv)
	if len(def.def.sources) != 2 {
		t.Error("Sources should set 2 sources")
	}
	if def.def.sources[0] != SourceFlag || def.def.sources[1] != SourceEnv {
		t.Error("Sources should be [Flag, Env]")
	}

	// Test Priority() method
	def = cfg.Define("TEST2").String().Priority(PriorityEnvFlagDefault)
	if len(def.def.priority) != 3 {
		t.Error("Priority should set 3 elements")
	}
	if def.def.priority[0] != SourceEnv || def.def.priority[1] != SourceFlag || def.def.priority[2] != SourceDefault {
		t.Error("Priority should be [Env, Flag, Default]")
	}

	// Test preset methods
	def = cfg.Define("TEST3").String().PriorityFlagEnvDefault()
	if len(def.def.priority) != 3 {
		t.Error("PriorityFlagEnvDefault should set 3 elements")
	}
	if def.def.priority[0] != SourceFlag || def.def.priority[1] != SourceEnv || def.def.priority[2] != SourceDefault {
		t.Error("PriorityFlagEnvDefault should be [Flag, Env, Default]")
	}
}

func TestEffectivePriority(t *testing.T) {
	cfg := New()

	// Definition with explicit priority
	def1 := &Definition{
		priority: PriorityEnvFlagDefault,
	}
	effective1 := def1.getEffectivePriority(cfg.defaultPriority)
	if len(effective1) != 3 || effective1[0] != SourceEnv {
		t.Error("Should use definition's priority when set")
	}

	// Definition without explicit priority
	def2 := &Definition{
		priority: nil,
	}
	effective2 := def2.getEffectivePriority(cfg.defaultPriority)
	if len(effective2) != 3 || effective2[0] != SourceFlag {
		t.Error("Should use config default when definition priority is nil")
	}
}

func TestInferAvailableSources(t *testing.T) {
	// Definition with all sources
	def1 := &Definition{
		envVar:       "TEST_ENV",
		flag:         "test-flag",
		defaultValue: "default",
	}
	sources1 := def1.inferAvailableSources()
	if len(sources1) != 4 {
		t.Error("Should infer 4 sources when all are available")
	}

	// Definition with only env
	def2 := &Definition{
		envVar: "TEST_ENV",
	}
	sources2 := def2.inferAvailableSources()
	if len(sources2) != 2 {
		t.Error("Should infer 2 sources (env + file) when only env is set")
	}

	// Definition with only default
	def3 := &Definition{
		defaultValue: "default",
	}
	sources3 := def3.inferAvailableSources()
	if len(sources3) != 2 {
		t.Error("Should infer 2 sources (default + file) when only default is set")
	}
}

func TestPriorityResolution(t *testing.T) {
	// Set environment variable
	os.Setenv("PRIORITY_TEST", "env_value")
	defer os.Unsetenv("PRIORITY_TEST")

	cfg := New()

	// Test Flag > Env > Default priority
	cfg.Define("TEST_FLAG_PRIORITY").
		String().
		Env("PRIORITY_TEST").
		Flag("test-flag").
		Default("default_value").
		PriorityFlagEnvDefault()

	// Save original os.Args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set flag value
	os.Args = []string{"test", "--test-flag", "flag_value"}

	if err := cfg.Execute(os.Args); err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Flag should win with Flag > Env > Default priority
	value := cfg.values["TEST_FLAG_PRIORITY"]
	if value != "flag_value" {
		t.Errorf("Expected flag_value, got %s", value)
	}
}

func TestEnvOverFlagPriority(t *testing.T) {
	// Set environment variable
	os.Setenv("PRIORITY_TEST2", "env_value")
	defer os.Unsetenv("PRIORITY_TEST2")

	cfg := New()

	// Test Env > Flag > Default priority
	cfg.Define("TEST_ENV_PRIORITY").
		String().
		Env("PRIORITY_TEST2").
		Flag("test-flag2").
		Default("default_value").
		PriorityEnvFlagDefault()

	// Save original os.Args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set flag value
	os.Args = []string{"test", "--test-flag2", "flag_value"}

	if err := cfg.Execute(os.Args); err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Env should win with Env > Flag > Default priority
	value := cfg.values["TEST_ENV_PRIORITY"]
	if value != "env_value" {
		t.Errorf("Expected env_value, got %s", value)
	}
}

func TestConfigLevelPriority(t *testing.T) {
	// Set environment variable
	os.Setenv("CONFIG_PRIORITY_TEST", "env_value")
	defer os.Unsetenv("CONFIG_PRIORITY_TEST")

	cfg := New()

	// Set config-level priority to Env > Flag > Default
	cfg.SetDefaultPriority(PriorityEnvFlagDefault)

	// Define config without explicit priority (should use config default)
	cfg.Define("TEST_CONFIG_PRIORITY").
		String().
		Env("CONFIG_PRIORITY_TEST").
		Flag("test-config-flag").
		Default("default_value")

	// Save original os.Args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set flag value
	os.Args = []string{"test", "--test-config-flag", "flag_value"}

	if err := cfg.Execute(os.Args); err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Env should win with config-level Env > Flag > Default priority
	value := cfg.values["TEST_CONFIG_PRIORITY"]
	if value != "env_value" {
		t.Errorf("Expected env_value, got %s", value)
	}
}

func TestOverrideDetectionWithCustomPriority(t *testing.T) {
	// Set environment variable
	os.Setenv("OVERRIDE_TEST", "env_value")
	defer os.Unsetenv("OVERRIDE_TEST")

	cfg := New()

	// Define with custom priority where env > flag
	cfg.Define("TEST_OVERRIDE").
		String().
		Env("OVERRIDE_TEST").
		Flag("override-flag").
		Default("default_value").
		PriorityEnvFlagDefault()

	// Save original os.Args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set flag value
	os.Args = []string{"test", "--override-flag", "flag_value"}

	if err := cfg.Execute(os.Args); err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Check for override warnings
	if !cfg.HasOverrideWarnings() {
		t.Error("Expected override warnings")
	}

	warnings := cfg.GetOverrideWarnings()
	if len(warnings.GetWarnings()) == 0 {
		t.Error("Expected at least one override warning")
	}

	// Should have env overriding flag (since env has higher priority in this case)
	foundEnvOverride := false
	for _, warning := range warnings.GetWarnings() {
		if warning.Key == "TEST_OVERRIDE" && warning.OverrideBy == "environment" && warning.Source == "flag" {
			foundEnvOverride = true
			break
		}
	}

	if !foundEnvOverride {
		t.Error("Expected to find env overriding flag warning")
	}

	// Env should win with Env > Flag > Default priority
	value := cfg.values["TEST_OVERRIDE"]
	if value != "env_value" {
		t.Errorf("Expected env_value, got %s", value)
	}
}
