// commandkit/help_handler_test.go
package commandkit

import (
	"strings"
	"testing"
)

func TestHelpHandler_IsHelpRequested(t *testing.T) {
	handler := NewHelpHandler()

	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "no help flags",
			args:     []string{"start", "--port", "8080"},
			expected: false,
		},
		{
			name:     "long help flag",
			args:     []string{"start", "--help"},
			expected: true,
		},
		{
			name:     "short help flag",
			args:     []string{"start", "-h"},
			expected: true,
		},
		{
			name:     "help command",
			args:     []string{"help"},
			expected: true,
		},
		{
			name:     "help as subcommand",
			args:     []string{"user", "help"},
			expected: true,
		},
		{
			name:     "mixed with other flags",
			args:     []string{"start", "--port", "8080", "--help"},
			expected: true,
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.IsHelpRequested(tt.args)
			if result != tt.expected {
				t.Errorf("IsHelpRequested() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHelpHandler_GenerateFlagHelp(t *testing.T) {
	handler := NewHelpHandler()

	tests := []struct {
		name     string
		def      *Definition
		expected string
	}{
		{
			name: "basic definition",
			def: &Definition{
				key:          "PORT",
				valueType:    TypeInt64,
				description:  "Server port",
				required:     false,
				secret:       false,
				defaultValue: 8080,
			},
			expected: "Server port (default: 8080)",
		},
		{
			name: "required with env var",
			def: &Definition{
				key:         "DATABASE_URL",
				valueType:   TypeString,
				description: "Database connection URL",
				required:    true,
				envVar:      "DATABASE_URL",
			},
			expected: "Database connection URL (env: DATABASE_URL, required)",
		},
		{
			name: "secret with default",
			def: &Definition{
				key:          "API_KEY",
				valueType:    TypeString,
				description:  "API authentication key",
				secret:       true,
				defaultValue: "secret123",
			},
			expected: "API authentication key (default: '[hidden]', secret)",
		},
		{
			name: "string default with quotes",
			def: &Definition{
				key:          "HOST",
				valueType:    TypeString,
				description:  "Server host",
				defaultValue: "localhost",
			},
			expected: "Server host (default: 'localhost')",
		},
		{
			name: "no indicators",
			def: &Definition{
				key:         "VERBOSE",
				valueType:   TypeBool,
				description: "Enable verbose logging",
			},
			expected: "Enable verbose logging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.GenerateFlagHelp(tt.def)
			if result != tt.expected {
				t.Errorf("GenerateFlagHelp() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestHelpHandler_GetCommandHelp(t *testing.T) {
	handler := NewHelpHandler()

	cmd := &Command{
		Name:      "start",
		ShortHelp: "Start the service",
		LongHelp:  "This command starts the service with the specified configuration.\nIt initializes all components and begins serving requests.",
		Definitions: map[string]*Definition{
			"PORT": {
				flag:         "port",
				description:  "Server port",
				required:     false,
				defaultValue: 8080,
			},
			"DAEMON": {
				flag:        "daemon",
				description: "Run in background",
				required:    true,
			},
		},
		SubCommands: map[string]*Command{
			"web": {
				Name:      "web",
				ShortHelp: "Start web server",
				Aliases:   []string{"www"},
			},
		},
	}

	help := handler.GetCommandHelp(cmd)

	// Check that help contains expected content
	expectedParts := []string{
		"This command starts the service",
		"Options:",
		"--port",
		"Server port",
		"default: 8080",
		"--daemon",
		"Run in background",
		"required",
		"Subcommands:",
		"web",
		"Start web server",
		"aliases: www",
	}

	for _, part := range expectedParts {
		if !contains(help, part) {
			t.Errorf("Expected help to contain %q, but it didn't. Full help:\n%s", part, help)
		}
	}
}

func TestHelpHandler_ShowSubcommandHelp(t *testing.T) {
	handler := NewHelpHandler()

	subcommands := map[string]*Command{
		"create": {
			Name:      "create",
			ShortHelp: "Create a new user",
		},
		"delete": {
			Name:      "delete",
			ShortHelp: "Delete a user",
		},
		"list": {
			Name:      "list",
			ShortHelp: "List all users",
			Aliases:   []string{"ls"},
		},
	}

	ctx := NewCommandContext([]string{}, New(), "user", "")
	help := handler.ShowSubcommandHelp("user", subcommands, ctx)

	// Check that help contains expected content
	expectedParts := []string{
		"Subcommands for user:",
		"create       Create a new user",
		"delete       Delete a user",
		"list         List all users",
		"user <command> --help",
	}

	for _, part := range expectedParts {
		if !contains(help, part) {
			t.Errorf("Expected subcommand help to contain %q, but it didn't. Full help:\n%s", part, help)
		}
	}

	// Check that subcommands are sorted (create should come before delete)
	createPos := strings.Index(help, "create")
	deletePos := strings.Index(help, "delete")
	if createPos > deletePos || createPos == -1 || deletePos == -1 {
		t.Error("Subcommands should be sorted alphabetically")
	}
}

func TestHelpHandler_ShowCommandHelp(t *testing.T) {
	handler := NewHelpHandler()

	cmd := &Command{
		Name: "deploy",
		Definitions: map[string]*Definition{
			"ENV": {
				flag:        "env",
				envVar:      "ENVIRONMENT",
				description: "Deployment environment",
				required:    true,
			},
			"SECRET": {
				flag:        "",
				envVar:      "API_SECRET",
				description: "API secret key",
				secret:      true,
			},
		},
	}

	ctx := NewCommandContext([]string{}, New(), "deploy", "")

	// This test captures stdout to verify help is displayed
	// In a real scenario, this would print to console
	// For testing purposes, we just verify it doesn't panic
	handler.ShowCommandHelp(cmd, ctx)
}

func TestHelpHandler_Integration(t *testing.T) {
	// Test that HelpHandler works correctly with the service factory
	services := NewCommandServices()
	handler := services.HelpHandler

	// Test all methods through the service factory
	if !handler.IsHelpRequested([]string{"--help"}) {
		t.Error("HelpHandler from service factory should detect help request")
	}

	def := &Definition{
		key:         "TEST",
		description: "Test flag",
		required:    true,
	}

	help := handler.GenerateFlagHelp(def)
	expected := "Test flag (required)"
	if help != expected {
		t.Errorf("Expected %q, got %q", expected, help)
	}
}
