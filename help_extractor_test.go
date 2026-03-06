// commandkit/help_extractor_test.go
package commandkit

import (
	"testing"
)

func TestNewHelpExtractor(t *testing.T) {
	extractor := NewHelpExtractor()
	if extractor == nil {
		t.Error("Expected non-nil extractor")
	}
}

func TestHelpExtractor_ExtractGlobalSummary(t *testing.T) {
	extractor := NewHelpExtractor()

	commands := map[string]*Command{
		"deploy": {
			Name:      "deploy",
			ShortHelp: "Deploy the application",
			Aliases:   []string{"dep"},
		},
		"start": {
			Name:      "start",
			ShortHelp: "Start the service",
			Aliases:   []string{"run", "up"},
		},
		"stop": {
			Name:      "stop",
			ShortHelp: "Stop the service",
		},
	}

	summaries := extractor.ExtractGlobalSummary(commands)

	if len(summaries) != 3 {
		t.Errorf("Expected 3 summaries, got %d", len(summaries))
	}

	// Check alphabetical sorting
	if summaries[0].Name != "deploy" {
		t.Errorf("Expected first command 'deploy', got '%s'", summaries[0].Name)
	}

	if summaries[1].Name != "start" {
		t.Errorf("Expected second command 'start', got '%s'", summaries[1].Name)
	}

	if summaries[2].Name != "stop" {
		t.Errorf("Expected third command 'stop', got '%s'", summaries[2].Name)
	}

	// Check deploy summary
	deploy := summaries[0]
	if deploy.Description != "Deploy the application" {
		t.Errorf("Expected description 'Deploy the application', got '%s'", deploy.Description)
	}

	if len(deploy.Aliases) != 1 || deploy.Aliases[0] != "dep" {
		t.Errorf("Expected alias 'dep', got %v", deploy.Aliases)
	}
}

func TestHelpExtractor_ExtractCommandInfo(t *testing.T) {
	extractor := NewHelpExtractor()

	cmd := &Command{
		Name:      "deploy",
		ShortHelp: "Deploy the application",
		LongHelp:  "Deploy the application to production environment with all components",
		Definitions: map[string]*Definition{
			"PORT": {
				key:          "PORT",
				flag:         "port",
				description:  "HTTP server port",
				valueType:    TypeInt64,
				required:     true,
				defaultValue: int64(8080),
			},
		},
		SubCommands: map[string]*Command{
			"server": {
				Name:      "server",
				ShortHelp: "Deploy server component",
			},
		},
	}

	commandInfo := extractor.ExtractCommandInfo(cmd, "testapp")

	// Check basic info
	if commandInfo.Usage != "testapp deploy [options]" {
		t.Errorf("Expected usage 'testapp deploy [options]', got '%s'", commandInfo.Usage)
	}

	if commandInfo.Description != "Deploy the application to production environment with all components" {
		t.Errorf("Expected long help description, got '%s'", commandInfo.Description)
	}

	// Check flags
	if len(commandInfo.Flags) != 1 {
		t.Errorf("Expected 1 flag, got %d", len(commandInfo.Flags))
	}

	flag := commandInfo.Flags[0]
	if flag.Name != "port" {
		t.Errorf("Expected flag name 'port', got '%s'", flag.Name)
	}

	if flag.Type != "int64" {
		t.Errorf("Expected flag type 'int64', got '%s'", flag.Type)
	}

	if !flag.Required {
		t.Error("Expected flag to be required")
	}

	// Check subcommands
	if len(commandInfo.Subcommands) != 1 {
		t.Errorf("Expected 1 subcommand, got %d", len(commandInfo.Subcommands))
	}

	subcmd := commandInfo.Subcommands[0]
	if subcmd.Name != "server" {
		t.Errorf("Expected subcommand name 'server', got '%s'", subcmd.Name)
	}
}

func TestHelpExtractor_ExtractFlagInfo(t *testing.T) {
	extractor := NewHelpExtractor()

	defs := map[string]*Definition{
		"PORT": {
			key:          "PORT",
			flag:         "port",
			description:  "HTTP server port",
			valueType:    TypeInt64,
			required:     true,
			defaultValue: int64(8080),
			envVar:       "PORT",
		},
		"API_KEY": {
			key:          "API_KEY",
			envVar:       "API_KEY",
			description:  "API authentication key",
			valueType:    TypeString,
			required:     true,
			secret:       true,
			defaultValue: "secret123",
		},
		"WORKERS": {
			key:          "WORKERS",
			flag:         "workers",
			description:  "Number of worker processes",
			valueType:    TypeInt64Slice,
			defaultValue: []int64{1, 2, 3},
		},
	}

	flags := extractor.ExtractFlagInfo(defs)

	if len(flags) != 3 {
		t.Errorf("Expected 3 flags, got %d", len(flags))
	}

	// Check PORT flag
	portFlag := flags[0] // Should be first alphabetically
	if portFlag.Name != "API_KEY" {
		t.Errorf("Expected first flag 'API_KEY', got '%s'", portFlag.Name)
	}

	// Check API_KEY flag (secret)
	apiKeyFlag := flags[0]
	if apiKeyFlag.Secret != true {
		t.Error("Expected API_KEY flag to be secret")
	}

	if apiKeyFlag.Default != "[hidden]" {
		t.Errorf("Expected hidden default for secret, got %v", apiKeyFlag.Default)
	}

	if apiKeyFlag.NoFlag != true {
		t.Error("Expected API_KEY to have no flag")
	}

	// Check WORKERS flag
	workersFlag := flags[2]
	if workersFlag.Name != "workers" {
		t.Errorf("Expected flag name 'workers', got '%s'", workersFlag.Name)
	}

	if workersFlag.Type != "[]int64" {
		t.Errorf("Expected type '[]int64', got '%s'", workersFlag.Type)
	}
}

func TestHelpExtractor_ExtractSubcommandInfo(t *testing.T) {
	extractor := NewHelpExtractor()

	subcommands := map[string]*Command{
		"server": {
			Name:      "server",
			ShortHelp: "Start server component",
			Aliases:   []string{"srv"},
		},
		"worker": {
			Name:      "worker",
			ShortHelp: "Start worker component",
			Aliases:   []string{"wrk", "worker-proc"},
		},
		"database": {
			Name:      "database",
			ShortHelp: "Initialize database",
		},
	}

	subcommandInfo := extractor.ExtractSubcommandInfo(subcommands)

	if len(subcommandInfo) != 3 {
		t.Errorf("Expected 3 subcommands, got %d", len(subcommandInfo))
	}

	// Check alphabetical sorting
	if subcommandInfo[0].Name != "database" {
		t.Errorf("Expected first subcommand 'database', got '%s'", subcommandInfo[0].Name)
	}

	// Check worker subcommand
	var workerInfo *SubcommandInfo
	for _, info := range subcommandInfo {
		if info.Name == "worker" {
			workerInfo = &info
			break
		}
	}

	if workerInfo == nil {
		t.Error("Expected to find worker subcommand")
		return
	}

	if workerInfo.Description != "Start worker component" {
		t.Errorf("Expected description 'Start worker component', got '%s'", workerInfo.Description)
	}

	if len(workerInfo.Aliases) != 2 {
		t.Errorf("Expected 2 aliases, got %d", len(workerInfo.Aliases))
	}
}

func TestHelpExtractor_ExtractFlagInfo_WithValidations(t *testing.T) {
	extractor := NewHelpExtractor()

	// Create definitions with validations
	portDef := &Definition{
		key:          "PORT",
		flag:         "port",
		description:  "HTTP server port",
		valueType:    TypeInt64,
		required:     true,
		defaultValue: int64(8080),
		envVar:       "PORT",
	}

	// Add validations manually (through the public API we can't add them directly)
	// This test will focus on the basic extraction functionality
	defs := map[string]*Definition{
		"PORT": portDef,
	}

	flags := extractor.ExtractFlagInfo(defs)

	if len(flags) != 1 {
		t.Errorf("Expected 1 flag, got %d", len(flags))
	}

	flag := flags[0]
	if flag.Name != "port" {
		t.Errorf("Expected flag name 'port', got '%s'", flag.Name)
	}

	if flag.Type != "int64" {
		t.Errorf("Expected flag type 'int64', got '%s'", flag.Type)
	}

	if !flag.Required {
		t.Error("Expected flag to be required")
	}

	if flag.Default != int64(8080) {
		t.Errorf("Expected default 8080, got %v", flag.Default)
	}
}
