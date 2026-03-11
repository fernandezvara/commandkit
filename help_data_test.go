// commandkit/help_data_test.go
package commandkit

import (
	"testing"
)

func TestNewUnifiedExtractor(t *testing.T) {
	extractor := NewUnifiedExtractor()

	if extractor == nil {
		t.Fatal("Expected non-nil extractor")
	}
}

func TestUnifiedExtractor_ExtractHelpData_NilCommand(t *testing.T) {
	extractor := NewUnifiedExtractor()

	helpData := extractor.ExtractHelpData(nil, HelpModeEssential, []GetError{})

	if helpData == nil {
		t.Fatal("Expected non-nil help data")
	}

	if helpData.Command != nil {
		t.Error("Expected command to be nil when input is nil")
	}

	if helpData.HasErrors {
		t.Error("Expected HasErrors to be false when no errors provided")
	}
}

func TestUnifiedExtractor_ExtractHelpData_WithCommand(t *testing.T) {
	extractor := NewUnifiedExtractor()

	// Create a test command
	cmd := &Command{
		Name:     "test",
		LongHelp: "Test command description",
		Definitions: map[string]*Definition{
			"test-flag": {
				key:         "test-flag",
				flag:        "test-flag",
				valueType:   TypeString,
				required:    true,
				description: "A test flag",
			},
			"test-env": {
				key:         "test-env",
				envVar:      "TEST_ENV",
				valueType:   TypeString,
				required:    false,
				description: "A test environment variable",
			},
		},
		SubCommands: map[string]*Command{
			"subcmd": {
				Name:     "subcmd",
				LongHelp: "Subcommand description",
				Aliases:  []string{"alias1", "alias2"},
			},
		},
	}

	errors := []GetError{
		{
			Key:              "test-flag",
			Display:          "test-flag string",
			ErrorDescription: "Not provided",
		},
	}

	helpData := extractor.ExtractHelpData(cmd, HelpModeEssential, errors)

	if helpData == nil {
		t.Fatal("Expected non-nil help data")
	}

	if helpData.Command != cmd {
		t.Error("Expected command to be set correctly")
	}

	if helpData.Usage != "Usage: test [options]" {
		t.Errorf("Expected usage 'Usage: test [options]', got '%s'", helpData.Usage)
	}

	if helpData.Description != "Test command description" {
		t.Errorf("Expected description 'Test command description', got '%s'", helpData.Description)
	}

	if len(helpData.Flags) != 1 {
		t.Errorf("Expected 1 flag (test-flag only), got %d", len(helpData.Flags))
	}

	if len(helpData.EnvVars) != 0 {
		t.Errorf("Expected 0 env vars in essential mode for non-required env var, got %d", len(helpData.EnvVars))
	}

	if len(helpData.Subcommands) != 1 {
		t.Errorf("Expected 1 subcommand, got %d", len(helpData.Subcommands))
	}

	if !helpData.HasErrors {
		t.Error("Expected HasErrors to be true when errors provided")
	}

	if len(helpData.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(helpData.Errors))
	}
}

func TestUnifiedExtractor_ExtractHelpData_FullMode(t *testing.T) {
	extractor := NewUnifiedExtractor()

	// Create a test command with environment variables
	cmd := &Command{
		Name:     "test",
		LongHelp: "Test command description",
		Definitions: map[string]*Definition{
			"test-env": {
				key:         "test-env",
				envVar:      "TEST_ENV",
				valueType:   TypeString,
				required:    false,
				description: "A test environment variable",
			},
		},
	}

	helpData := extractor.ExtractHelpData(cmd, HelpModeFull, []GetError{})

	if helpData.Mode != HelpModeFull {
		t.Error("Expected mode to be HelpModeFull")
	}

	// In full mode, should include all env vars, not just required ones
	if len(helpData.EnvVars) != 1 {
		t.Errorf("Expected 1 env var in full mode, got %d", len(helpData.EnvVars))
	}
}

func TestUnifiedExtractor_ExtractFlags(t *testing.T) {
	extractor := NewUnifiedExtractor()

	defs := map[string]*Definition{
		"flag": {
			key:         "flag",
			flag:        "flag",
			valueType:   TypeString,
			required:    true,
			description: "A flag",
		},
		"env": {
			key:         "env",
			envVar:      "ENV",
			valueType:   TypeString,
			required:    false,
			description: "An env var",
		},
	}

	flags := extractor.ExtractFlags(defs)

	if len(flags) != 1 {
		t.Errorf("Expected 1 flag (only actual flags), got %d", len(flags))
	}

	// Check flag
	var flagInfo *FlagInfo

	for _, flag := range flags {
		if flag.Name == "flag" {
			flagInfo = &flag
		}
	}

	if flagInfo == nil {
		t.Error("Expected to find flag info")
	} else {
		if flagInfo.NoFlag {
			t.Error("Expected flag to not be marked as NoFlag")
		}
		if flagInfo.DisplayLine == "" {
			t.Error("Expected flag to have DisplayLine")
		}
	}

	// Environment variables should be extracted separately using ExtractEnvVars
	envVars := extractor.ExtractEnvVars(defs, HelpModeEssential)
	if len(envVars) != 1 {
		t.Errorf("Expected 1 env var, got %d", len(envVars))
	}
}

func TestUnifiedExtractor_FilterEnvVars(t *testing.T) {
	extractor := NewUnifiedExtractor()

	flags := []FlagInfo{
		{
			Key:      "flag",
			Name:     "flag",
			NoFlag:   false,
			EnvVar:   "",
			Required: true,
		},
		{
			Key:      "required-env",
			Name:     "",
			NoFlag:   true,
			EnvVar:   "REQUIRED_ENV",
			Required: true,
		},
		{
			Key:      "optional-env",
			Name:     "",
			NoFlag:   true,
			EnvVar:   "OPTIONAL_ENV",
			Required: false,
		},
	}

	// Test essential mode - only required env vars
	envVars := extractor.FilterEnvVars(flags, HelpModeEssential)

	if len(envVars) != 1 {
		t.Errorf("Expected 1 env var in essential mode, got %d", len(envVars))
	}

	if envVars[0].EnvVar != "REQUIRED_ENV" {
		t.Error("Expected to find only required env var in essential mode")
	}

	// Test full mode - all env vars
	envVars = extractor.FilterEnvVars(flags, HelpModeFull)

	if len(envVars) != 2 {
		t.Errorf("Expected 2 env vars in full mode, got %d", len(envVars))
	}
}

func TestUnifiedExtractor_ExtractSubcommands(t *testing.T) {
	extractor := NewUnifiedExtractor()

	cmd := &Command{
		SubCommands: map[string]*Command{
			"sub1": {
				Name:     "sub1",
				LongHelp: "Subcommand 1",
				Aliases:  []string{"alias1"},
			},
			"sub2": {
				Name:     "sub2",
				LongHelp: "Subcommand 2",
				Aliases:  []string{"alias2", "alias3"},
			},
		},
	}

	subcommands := extractor.ExtractSubcommands(cmd)

	if len(subcommands) != 2 {
		t.Errorf("Expected 2 subcommands, got %d", len(subcommands))
	}

	// Check that subcommands are sorted
	if subcommands[0].Name != "sub1" || subcommands[1].Name != "sub2" {
		t.Error("Expected subcommands to be sorted by name")
	}

	// Check subcommand 1
	if subcommands[0].Description != "Subcommand 1" {
		t.Error("Expected subcommand 1 to have correct description")
	}

	if len(subcommands[0].Aliases) != 1 || subcommands[0].Aliases[0] != "alias1" {
		t.Error("Expected subcommand 1 to have correct aliases")
	}
}

func TestUnifiedExtractor_buildUsage(t *testing.T) {
	extractor := NewUnifiedExtractor()

	// Test with named command
	cmd := &Command{Name: "test"}
	usage := extractor.buildUsage(cmd)
	expected := "Usage: test [options]"
	if usage != expected {
		t.Errorf("Expected usage '%s', got '%s'", expected, usage)
	}

	// Test with unnamed command
	cmd = &Command{}
	usage = extractor.buildUsage(cmd)
	expected = "Usage: [options]"
	if usage != expected {
		t.Errorf("Expected usage '%s', got '%s'", expected, usage)
	}

	// Test with nil command
	usage = extractor.buildUsage(nil)
	if usage != "" {
		t.Errorf("Expected empty usage for nil command, got '%s'", usage)
	}
}

func TestUnifiedExtractor_extractValidations(t *testing.T) {
	extractor := NewUnifiedExtractor()

	// Test with no validations
	validations := extractor.extractValidations([]Validation{})
	if len(validations) != 0 {
		t.Errorf("Expected 0 validations for empty input, got %d", len(validations))
	}

	// Test with some validations
	validations = extractor.extractValidations([]Validation{{}, {}})
	if len(validations) != 1 {
		t.Errorf("Expected 1 validation for non-empty input, got %d", len(validations))
	}

	if validations[0] != "validation" {
		t.Errorf("Expected validation to be 'validation', got '%s'", validations[0])
	}
}

func TestUnifiedExtractor_ExtractGlobalCommands(t *testing.T) {
	extractor := NewUnifiedExtractor()

	commands := map[string]*Command{
		"cmd1": {
			Name:     "cmd1",
			LongHelp: "Command 1",
			Aliases:  []string{"alias1"},
		},
		"cmd2": {
			Name:     "cmd2",
			LongHelp: "Command 2",
			Aliases:  []string{},
		},
	}

	summaries := extractor.ExtractGlobalCommands(commands)

	if len(summaries) != 2 {
		t.Errorf("Expected 2 command summaries, got %d", len(summaries))
	}

	// Check that commands are sorted
	if summaries[0].Name != "cmd1" || summaries[1].Name != "cmd2" {
		t.Error("Expected commands to be sorted by name")
	}

	// Check command 1
	if summaries[0].Description != "Command 1" {
		t.Error("Expected command 1 to have correct description")
	}

	if len(summaries[0].Aliases) != 1 || summaries[0].Aliases[0] != "alias1" {
		t.Error("Expected command 1 to have correct aliases")
	}
}

func TestFlagInfo_GetDisplayLine(t *testing.T) {
	// Test flag (not NoFlag)
	flag := FlagInfo{
		NoFlag:      false,
		DisplayLine: "--flag string",
	}

	if flag.GetDisplayLine() != "--flag string" {
		t.Errorf("Expected '--flag string', got '%s'", flag.GetDisplayLine())
	}

	// Test environment variable (NoFlag)
	env := FlagInfo{
		NoFlag:        true,
		EnvVarDisplay: "ENV_VAR string",
	}

	if env.GetDisplayLine() != "ENV_VAR string" {
		t.Errorf("Expected 'ENV_VAR string', got '%s'", env.GetDisplayLine())
	}
}
