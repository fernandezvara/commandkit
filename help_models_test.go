// commandkit/help_models_test.go
package commandkit

import (
	"testing"
)

func TestHelpType(t *testing.T) {
	tests := []struct {
		name     string
		helpType helpType
		expected int
	}{
		{"helpTypeNone", helpTypeNone, 0},
		{"helpTypeGlobal", helpTypeGlobal, 1},
		{"helpTypeCommand", helpTypeCommand, 2},
		{"helpTypeSubcommand", helpTypeSubcommand, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.helpType) != tt.expected {
				t.Errorf("helpType %s = %d, expected %d", tt.name, int(tt.helpType), tt.expected)
			}
		})
	}
}

func TestCommandSummary(t *testing.T) {
	summary := commandSummary{
		Name:    "deploy",
		Aliases: []string{"dep", "deploy-app"},
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

func TestFlagInfo(t *testing.T) {
	flag := flagInfo{
		Name:     "port",
		Required: true,
		Default:  8080,
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
	info := subcommandInfo{
		Name: "server",
	}

	if info.Name != "server" {
		t.Errorf("Expected subcommand name 'server', got '%s'", info.Name)
	}
}
