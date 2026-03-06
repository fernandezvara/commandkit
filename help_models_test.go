// commandkit/help_models_test.go
package commandkit

import (
	"testing"
)

func TestHelpType(t *testing.T) {
	tests := []struct {
		name     string
		helpType HelpType
		expected int
	}{
		{"HelpTypeNone", HelpTypeNone, 0},
		{"HelpTypeGlobal", HelpTypeGlobal, 1},
		{"HelpTypeCommand", HelpTypeCommand, 2},
		{"HelpTypeSubcommand", HelpTypeSubcommand, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.helpType) != tt.expected {
				t.Errorf("HelpType %s = %d, expected %d", tt.name, int(tt.helpType), tt.expected)
			}
		})
	}
}

func TestGlobalHelp(t *testing.T) {
	help := &GlobalHelp{
		Executable: "testapp",
		Commands: []CommandSummary{
			{Name: "start", Description: "Start the service"},
			{Name: "stop", Description: "Stop the service"},
		},
		Template: "test template",
	}

	if help.Executable != "testapp" {
		t.Errorf("Expected executable 'testapp', got '%s'", help.Executable)
	}

	if len(help.Commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(help.Commands))
	}

	if help.Commands[0].Name != "start" {
		t.Errorf("Expected first command 'start', got '%s'", help.Commands[0].Name)
	}
}

func TestCommandSummary(t *testing.T) {
	summary := CommandSummary{
		Name:        "deploy",
		Description: "Deploy the application",
		Aliases:     []string{"dep", "deploy-app"},
	}

	if summary.Name != "deploy" {
		t.Errorf("Expected name 'deploy', got '%s'", summary.Name)
	}

	if len(summary.Aliases) != 2 {
		t.Errorf("Expected 2 aliases, got %d", len(summary.Aliases))
	}

	if summary.Aliases[0] != "dep" {
		t.Errorf("Expected first alias 'dep', got '%s'", summary.Aliases[0])
	}
}

func TestCommandHelpModel(t *testing.T) {
	cmd := &Command{
		Name:      "start",
		ShortHelp: "Start the service",
		LongHelp:  "Start the service with all components",
	}

	help := &CommandHelp{
		Command:     cmd,
		Usage:       "testapp start [options]",
		Description: "Start the service with all components",
		Flags:       []FlagInfo{},
		Subcommands: []SubcommandInfo{},
		Template:    "test template",
	}

	if help.Command.Name != "start" {
		t.Errorf("Expected command name 'start', got '%s'", help.Command.Name)
	}

	if help.Usage != "testapp start [options]" {
		t.Errorf("Expected usage 'testapp start [options]', got '%s'", help.Usage)
	}
}

func TestFlagInfo(t *testing.T) {
	flag := FlagInfo{
		Name:        "port",
		Description: "HTTP server port",
		Type:        "int64",
		Required:    true,
		Default:     8080,
		EnvVar:      "PORT",
		Validations: []string{"range: 1-65535"},
		Secret:      false,
		NoFlag:      false,
	}

	if flag.Name != "port" {
		t.Errorf("Expected flag name 'port', got '%s'", flag.Name)
	}

	if !flag.Required {
		t.Error("Expected flag to be required")
	}

	if flag.Default != 8080 {
		t.Errorf("Expected default 8080, got %v", flag.Default)
	}
}

func TestSubcommandInfo(t *testing.T) {
	info := SubcommandInfo{
		Name:        "server",
		Description: "Start server component",
		Aliases:     []string{"srv"},
	}

	if info.Name != "server" {
		t.Errorf("Expected subcommand name 'server', got '%s'", info.Name)
	}

	if len(info.Aliases) != 1 {
		t.Errorf("Expected 1 alias, got %d", len(info.Aliases))
	}
}

func TestHelpRequest(t *testing.T) {
	request := &HelpRequest{
		Type:       HelpTypeCommand,
		Command:    "start",
		Subcommand: "",
		Args:       []string{"start", "--help"},
		Original:   []string{"start", "--help"},
	}

	if request.Type != HelpTypeCommand {
		t.Errorf("Expected HelpTypeCommand, got %v", request.Type)
	}

	if request.Command != "start" {
		t.Errorf("Expected command 'start', got '%s'", request.Command)
	}

	if len(request.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(request.Args))
	}
}
