// commandkit/help_factory_test.go
package commandkit

import (
	"testing"
)

func TestNewHelpFactory(t *testing.T) {
	factory := NewHelpFactory()
	if factory == nil {
		t.Error("Expected non-nil factory")
	}
}

func TestHelpFactory_DetectHelpRequest(t *testing.T) {
	factory := NewHelpFactory()
	
	tests := []struct {
		name           string
		args           []string
		expectedType   HelpType
		expectedCmd    string
		expectedSubcmd string
	}{
		{"No help", []string{"start", "server"}, HelpTypeNone, "", ""},
		{"Global help", []string{"--help"}, HelpTypeGlobal, "", ""},
		{"Command help", []string{"start", "--help"}, HelpTypeCommand, "start", ""},
		{"Subcommand help", []string{"start", "server", "--help"}, HelpTypeSubcommand, "start", "server"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := factory.DetectHelpRequest(tt.args)
			
			if request.Type != tt.expectedType {
				t.Errorf("DetectHelpRequest(%v) type = %v, expected %v", tt.args, request.Type, tt.expectedType)
			}
			
			if request.Command != tt.expectedCmd {
				t.Errorf("DetectHelpRequest(%v) command = %s, expected %s", tt.args, request.Command, tt.expectedCmd)
			}
			
			if request.Subcommand != tt.expectedSubcmd {
				t.Errorf("DetectHelpRequest(%v) subcommand = %s, expected %s", tt.args, request.Subcommand, tt.expectedSubcmd)
			}
		})
	}
}

func TestHelpFactory_IsHelpRequested(t *testing.T) {
	factory := NewHelpFactory()
	
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{"No help", []string{"start", "server"}, false},
		{"Help flag", []string{"--help"}, true},
		{"Short help", []string{"-h"}, true},
		{"Help command", []string{"help"}, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := factory.IsHelpRequested(tt.args)
			if result != tt.expected {
				t.Errorf("IsHelpRequested(%v) = %v, expected %v", tt.args, result, tt.expected)
			}
		})
	}
}

func TestHelpFactory_GetHelpType(t *testing.T) {
	factory := NewHelpFactory()
	
	tests := []struct {
		name     string
		args     []string
		expected HelpType
	}{
		{"No help", []string{"start", "server"}, HelpTypeNone},
		{"Global help", []string{"--help"}, HelpTypeGlobal},
		{"Command help", []string{"start", "--help"}, HelpTypeCommand},
		{"Subcommand help", []string{"start", "server", "--help"}, HelpTypeSubcommand},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := factory.GetHelpType(tt.args)
			if result != tt.expected {
				t.Errorf("GetHelpType(%v) = %v, expected %v", tt.args, result, tt.expected)
			}
		})
	}
}

func TestHelpFactory_CreateGlobalHelp(t *testing.T) {
	factory := NewHelpFactory()
	
	commands := map[string]*Command{
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
	
	globalHelp := factory.CreateGlobalHelp(commands, "testapp")
	
	if globalHelp.Executable != "testapp" {
		t.Errorf("Expected executable 'testapp', got '%s'", globalHelp.Executable)
	}
	
	if len(globalHelp.Commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(globalHelp.Commands))
	}
	
	if globalHelp.Commands[0].Name != "start" {
		t.Errorf("Expected first command 'start', got '%s'", globalHelp.Commands[0].Name)
	}
	
	if globalHelp.Template != DefaultGlobalTemplate {
		t.Error("Expected default global template")
	}
}

func TestHelpFactory_CreateCommandHelp(t *testing.T) {
	factory := NewHelpFactory()
	
	cmd := &Command{
		Name:      "deploy",
		ShortHelp: "Deploy the application",
		LongHelp:  "Deploy the application to production environment",
		Definitions: map[string]*Definition{
			"PORT": {
				key:         "PORT",
				flag:        "port",
				description: "HTTP server port",
				valueType:   TypeInt64,
				required:    true,
				defaultValue: int64(8080),
			},
		},
	}
	
	commandHelp := factory.CreateCommandHelp(cmd, "testapp")
	
	if commandHelp.Command.Name != "deploy" {
		t.Errorf("Expected command name 'deploy', got '%s'", commandHelp.Command.Name)
	}
	
	if commandHelp.Usage != "testapp deploy [options]" {
		t.Errorf("Expected usage 'testapp deploy [options]', got '%s'", commandHelp.Usage)
	}
	
	if commandHelp.Description != "Deploy the application to production environment" {
		t.Errorf("Expected long help description, got '%s'", commandHelp.Description)
	}
	
	if len(commandHelp.Flags) != 1 {
		t.Errorf("Expected 1 flag, got %d", len(commandHelp.Flags))
	}
	
	if commandHelp.Template != DefaultCommandTemplate {
		t.Error("Expected default command template")
	}
}

func TestHelpFactory_CreateSubcommandHelp(t *testing.T) {
	factory := NewHelpFactory()
	
	subcommands := map[string]*Command{
		"server": {
			Name:      "server",
			ShortHelp: "Start server component",
			Aliases:   []string{"srv"},
		},
		"worker": {
			Name:      "worker",
			ShortHelp: "Start worker component",
		},
	}
	
	subcommandHelp := factory.CreateSubcommandHelp("start", subcommands)
	
	if subcommandHelp.Parent != "start" {
		t.Errorf("Expected parent 'start', got '%s'", subcommandHelp.Parent)
	}
	
	if len(subcommandHelp.Subcommands) != 2 {
		t.Errorf("Expected 2 subcommands, got %d", len(subcommandHelp.Subcommands))
	}
	
	if subcommandHelp.Template != DefaultSubcommandTemplate {
		t.Error("Expected default subcommand template")
	}
}

func TestHelpFactory_CreateFlagHelp(t *testing.T) {
	factory := NewHelpFactory()
	
	defs := map[string]*Definition{
		"PORT": {
			key:         "PORT",
			flag:        "port",
			description: "HTTP server port",
			valueType:   TypeInt64,
			required:    true,
			defaultValue: int64(8080),
		},
		"API_KEY": {
			key:          "API_KEY",
			envVar:       "API_KEY",
			description:  "API authentication key",
			valueType:    TypeString,
			required:     true,
			secret:       true,
		},
	}
	
	flagHelp := factory.CreateFlagHelp("deploy", defs)
	
	if flagHelp.Command != "deploy" {
		t.Errorf("Expected command 'deploy', got '%s'", flagHelp.Command)
	}
	
	if len(flagHelp.Flags) != 2 {
		t.Errorf("Expected 2 flags, got %d", len(flagHelp.Flags))
	}
	
	if flagHelp.Template != DefaultFlagTemplate {
		t.Error("Expected default flag template")
	}
}

func TestHelpFactory_TemplateManagement(t *testing.T) {
	factory := NewHelpFactory()
	
	// Test getting default template
	template := factory.GetTemplate(TemplateGlobal)
	if template != DefaultGlobalTemplate {
		t.Error("Expected default global template")
	}
	
	// Test setting custom template
	customTemplate := "Custom global help template"
	factory.SetTemplate(TemplateGlobal, customTemplate)
	
	// Verify template was set
	template = factory.GetTemplate(TemplateGlobal)
	if template != customTemplate {
		t.Error("Expected custom template to be set")
	}
}
