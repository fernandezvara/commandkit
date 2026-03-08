// commandkit/flag_parser_test.go
package commandkit

import (
	"strings"
	"testing"
)

func TestFlagParser_ParseCommand(t *testing.T) {
	flagParser := newFlagParser()

	// Create test definitions
	defs := make(map[string]*Definition)
	defs["port"] = &Definition{
		key:         "port",
		valueType:   TypeInt64,
		flag:        "port",
		description: "HTTP server port",
	}
	defs["verbose"] = &Definition{
		key:         "verbose",
		valueType:   TypeBool,
		flag:        "verbose",
		description: "Enable verbose logging",
	}

	tests := []struct {
		name     string
		args     []string
		expected map[string]string
	}{
		{
			name: "no flags",
			args: []string{},
			expected: map[string]string{
				"port":    "",
				"verbose": "",
			},
		},
		{
			name: "single flag",
			args: []string{"--port", "8080"},
			expected: map[string]string{
				"port":    "8080",
				"verbose": "",
			},
		},
		{
			name: "multiple flags",
			args: []string{"--port", "3000", "--verbose", "true"},
			expected: map[string]string{
				"port":    "3000",
				"verbose": "true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedFlags, err := flagParser.ParseCommand(tt.args, defs)

			if err != nil {
				t.Errorf("ParseCommand() returned error: %v", err)
			}

			if parsedFlags == nil {
				t.Fatal("ParseCommand() returned nil")
			}

			// Check parsed values
			for key, expectedValue := range tt.expected {
				if actualValue, exists := parsedFlags.Values[key]; !exists {
					t.Errorf("ParseCommand() missing value for key %s", key)
				} else if *actualValue != expectedValue {
					t.Errorf("ParseCommand() for key %s: expected %s, got %s", key, expectedValue, *actualValue)
				}
			}
		})
	}
}

func TestFlagParser_ParseGlobal(t *testing.T) {
	flagParser := newFlagParser()

	// Create test definitions
	defs := make(map[string]*Definition)
	defs["config"] = &Definition{
		key:         "config",
		valueType:   TypeString,
		flag:        "config",
		description: "Configuration file path",
	}

	tests := []struct {
		name     string
		args     []string
		expected map[string]string
	}{
		{
			name: "no flags",
			args: []string{},
			expected: map[string]string{
				"config": "",
			},
		},
		{
			name: "with config flag",
			args: []string{"--config", "/path/to/config.yaml"},
			expected: map[string]string{
				"config": "/path/to/config.yaml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedFlags, err := flagParser.ParseGlobal(tt.args, defs)

			if err != nil {
				t.Errorf("ParseGlobal() returned error: %v", err)
			}

			if parsedFlags == nil {
				t.Fatal("ParseGlobal() returned nil")
			}

			// Check parsed values
			for key, expectedValue := range tt.expected {
				if actualValue, exists := parsedFlags.Values[key]; !exists {
					t.Errorf("ParseGlobal() missing value for key %s", key)
				} else if *actualValue != expectedValue {
					t.Errorf("ParseGlobal() for key %s: expected %s, got %s", key, expectedValue, *actualValue)
				}
			}
		})
	}
}

func TestFlagParser_ParseGlobalFiltersTestFlags(t *testing.T) {
	flagParser := newFlagParser()

	// Create test definitions
	defs := make(map[string]*Definition)
	defs["port"] = &Definition{
		key:         "port",
		valueType:   TypeInt64,
		flag:        "port",
		description: "HTTP server port",
	}

	// Test that test flags are filtered out
	args := []string{"--port", "8080", "-test.timeout", "30s", "-test.v", "true"}

	parsedFlags, err := flagParser.ParseGlobal(args, defs)

	if err != nil {
		t.Errorf("ParseGlobal() returned error: %v", err)
	}

	if parsedFlags == nil {
		t.Fatal("ParseGlobal() returned nil")
	}

	// Check that port was parsed
	if portValue, exists := parsedFlags.Values["port"]; !exists {
		t.Error("ParseGlobal() missing value for key port")
	} else if *portValue != "8080" {
		t.Errorf("ParseGlobal() for port: expected 8080, got %s", *portValue)
	}

	// Check that test flags were not included in parsed values
	if _, exists := parsedFlags.Values["test.timeout"]; exists {
		t.Error("ParseGlobal() should not parse test flags")
	}
	if _, exists := parsedFlags.Values["test.v"]; exists {
		t.Error("ParseGlobal() should not parse test flags")
	}
}

func TestFlagParser_GenerateHelp(t *testing.T) {
	flagParser := newFlagParser()

	// Create test definitions with various properties
	defs := make(map[string]*Definition)
	defs["port"] = &Definition{
		key:          "port",
		valueType:    TypeInt64,
		flag:         "port",
		description:  "HTTP server port",
		defaultValue: int64(8080),
	}
	defs["verbose"] = &Definition{
		key:          "verbose",
		valueType:    TypeBool,
		flag:         "verbose",
		description:  "Enable verbose logging",
		defaultValue: false,
	}
	defs["database_url"] = &Definition{
		key:         "database_url",
		valueType:   TypeString,
		envVar:      "DATABASE_URL",
		description: "Database connection string",
		required:    true,
		secret:      true,
	}
	defs["log_level"] = &Definition{
		key:          "log_level",
		valueType:    TypeString,
		flag:         "log-level",
		envVar:       "LOG_LEVEL",
		description:  "Logging level",
		defaultValue: "info",
		validations:  []Validation{{Name: "oneOf(['debug','info','warn','error'])"}},
	}

	helpText := flagParser.GenerateHelp(defs)

	if helpText == "" {
		t.Error("GenerateHelp() returned empty string")
	}

	// Check that all flags are mentioned in help
	expectedFlags := []string{"port", "verbose", "log-level"}
	for _, flag := range expectedFlags {
		if !strings.Contains(helpText, flag) {
			t.Errorf("GenerateHelp() should contain flag %s", flag)
		}
	}

	// Check that environment-only config is mentioned
	if !strings.Contains(helpText, "DATABASE_URL") {
		t.Error("GenerateHelp() should contain environment variable DATABASE_URL")
	}

	// Check that enhanced descriptions are included
	if !strings.Contains(helpText, "default:") {
		t.Error("GenerateHelp() should contain default values")
	}
	if !strings.Contains(helpText, "required") {
		t.Error("GenerateHelp() should contain required indicator")
	}
	if !strings.Contains(helpText, "secret") {
		t.Error("GenerateHelp() should contain secret indicator")
	}
}

func TestFlagParser_GenerateHelpWithNoFlags(t *testing.T) {
	flagParser := newFlagParser()

	// Create definitions with only environment variables
	defs := make(map[string]*Definition)
	defs["database_url"] = &Definition{
		key:         "database_url",
		valueType:   TypeString,
		envVar:      "DATABASE_URL",
		description: "Database connection string",
		required:    true,
	}

	helpText := flagParser.GenerateHelp(defs)

	if helpText == "" {
		t.Error("GenerateHelp() returned empty string")
	}

	// Check that environment-only config is mentioned
	if !strings.Contains(helpText, "DATABASE_URL") {
		t.Error("GenerateHelp() should contain environment variable DATABASE_URL")
	}
	if !strings.Contains(helpText, "(no flag)") {
		t.Error("GenerateHelp() should contain (no flag) indicator for environment-only configs")
	}
}

func TestFlagParser_GenerateHelpWithEmptyDefinitions(t *testing.T) {
	flagParser := newFlagParser()

	helpText := flagParser.GenerateHelp(make(map[string]*Definition))

	// Empty definitions should return empty help text
	if helpText != "" {
		t.Errorf("GenerateHelp() should return empty string for empty definitions, got: %s", helpText)
	}
}

func TestParsedFlags_Structure(t *testing.T) {
	flagParser := newFlagParser()

	defs := make(map[string]*Definition)
	defs["port"] = &Definition{
		key:         "port",
		valueType:   TypeInt64,
		flag:        "port",
		description: "HTTP server port",
	}

	args := []string{"--port", "8080"}
	parsedFlags, err := flagParser.ParseCommand(args, defs)

	if err != nil {
		t.Errorf("ParseCommand() returned error: %v", err)
	}

	if parsedFlags == nil {
		t.Fatal("ParseCommand() returned nil")
	}

	// Check ParsedFlags structure
	if parsedFlags.Values == nil {
		t.Error("ParsedFlags.Values should not be nil")
	}
	if parsedFlags.FlagSet == nil {
		t.Error("ParsedFlags.FlagSet should not be nil")
	}
	if parsedFlags.Args == nil {
		t.Error("ParsedFlags.Args should not be nil")
	}
}
