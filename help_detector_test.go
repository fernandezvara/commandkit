// commandkit/help_detector_test.go
package commandkit

import (
	"testing"
)

func TestNewHelpDetector(t *testing.T) {
	detector := NewHelpDetector()
	if detector == nil {
		t.Error("Expected non-nil detector")
	}
}

func TestHelpDetector_IsHelpRequested(t *testing.T) {
	detector := NewHelpDetector()

	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{"No help", []string{"start", "server"}, false},
		{"Help flag", []string{"--help"}, true},
		{"Short help flag", []string{"-h"}, true},
		{"Help command", []string{"help"}, true},
		{"Help with command", []string{"start", "--help"}, true},
		{"Help in middle", []string{"start", "server", "--help"}, true},
		{"Multiple flags", []string{"-h", "start"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.IsHelpRequested(tt.args)
			if result != tt.expected {
				t.Errorf("IsHelpRequested(%v) = %v, expected %v", tt.args, result, tt.expected)
			}
		})
	}
}

func TestHelpDetector_GetHelpType(t *testing.T) {
	detector := NewHelpDetector()

	tests := []struct {
		name     string
		args     []string
		expected HelpType
	}{
		{"No help", []string{"start", "server"}, HelpTypeNone},
		{"Global help flag first", []string{"--help"}, HelpTypeGlobal},
		{"Global help command first", []string{"help"}, HelpTypeGlobal},
		{"Global help no args", []string{}, HelpTypeNone},
		{"Command help", []string{"start", "--help"}, HelpTypeCommand},
		{"Command help short", []string{"start", "-h"}, HelpTypeCommand},
		{"Command help command", []string{"start", "help"}, HelpTypeCommand},
		{"Subcommand help", []string{"start", "server", "--help"}, HelpTypeSubcommand},
		{"Subcommand help short", []string{"start", "server", "-h"}, HelpTypeSubcommand},
		{"Command without help flag", []string{"start", "server"}, HelpTypeNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.GetHelpType(tt.args)
			if result != tt.expected {
				t.Errorf("GetHelpType(%v) = %v, expected %v", tt.args, result, tt.expected)
			}
		})
	}
}

func TestHelpDetector_ExtractCommandFromArgs(t *testing.T) {
	detector := NewHelpDetector()

	tests := []struct {
		name           string
		args           []string
		expectedCmd    string
		expectedRemain []string
	}{
		{"Empty args", []string{}, "", []string{}},
		{"Help flag only", []string{"--help"}, "", []string{"--help"}},
		{"Single command", []string{"start"}, "start", []string{}},
		{"Command with help", []string{"start", "--help"}, "start", []string{"--help"}},
		{"Command with options", []string{"start", "server", "--port", "8080"}, "start", []string{"server", "--port", "8080"}},
		{"Help flag first", []string{"--help", "start"}, "", []string{"--help", "start"}},
		{"Command with subcommand", []string{"start", "server"}, "start", []string{"server"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, remain := detector.ExtractCommandFromArgs(tt.args)
			if cmd != tt.expectedCmd {
				t.Errorf("ExtractCommandFromArgs(%v) command = %s, expected %s", tt.args, cmd, tt.expectedCmd)
			}

			if len(remain) != len(tt.expectedRemain) {
				t.Errorf("ExtractCommandFromArgs(%v) remain length = %d, expected %d", tt.args, len(remain), len(tt.expectedRemain))
			}

			for i, arg := range remain {
				if i < len(tt.expectedRemain) && arg != tt.expectedRemain[i] {
					t.Errorf("ExtractCommandFromArgs(%v) remain[%d] = %s, expected %s", tt.args, i, arg, tt.expectedRemain[i])
				}
			}
		})
	}
}

func TestHelpDetector_ParseHelpRequest(t *testing.T) {
	detector := NewHelpDetector()

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
		{"Command help command", []string{"start", "help"}, HelpTypeCommand, "start", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := detector.ParseHelpRequest(tt.args)

			if request.Type != tt.expectedType {
				t.Errorf("ParseHelpRequest(%v) type = %v, expected %v", tt.args, request.Type, tt.expectedType)
			}

			if request.Command != tt.expectedCmd {
				t.Errorf("ParseHelpRequest(%v) command = %s, expected %s", tt.args, request.Command, tt.expectedCmd)
			}

			if request.Subcommand != tt.expectedSubcmd {
				t.Errorf("ParseHelpRequest(%v) subcommand = %s, expected %s", tt.args, request.Subcommand, tt.expectedSubcmd)
			}
		})
	}
}

func TestHelpDetector_GetExecutableName(t *testing.T) {
	detector := NewHelpDetector()

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{"Empty args", []string{}, "command"},
		{"Simple path", []string{"/usr/bin/testapp"}, "testapp"},
		{"Relative path", []string{"./testapp"}, "testapp"},
		{"Complex path", []string{"/opt/myapp/bin/testapp"}, "testapp"},
		{"No path separators", []string{"testapp"}, "testapp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.GetExecutableName(tt.args)
			if result != tt.expected {
				t.Errorf("GetExecutableName(%v) = %s, expected %s", tt.args, result, tt.expected)
			}
		})
	}
}

func TestHelpDetector_FormatHelpUsage(t *testing.T) {
	detector := NewHelpDetector()

	usage := detector.FormatHelpUsage("testapp")
	expected := "Usage: testapp <command> [options]\n\nUse 'testapp <command> --help' for command-specific help\nUse 'testapp --help' for global help"

	if usage != expected {
		t.Errorf("FormatHelpUsage() = %q, expected %q", usage, expected)
	}
}

func TestHelpDetector_isHelpFlag(t *testing.T) {
	detector := NewHelpDetector()

	tests := []struct {
		name     string
		arg      string
		expected bool
	}{
		{"Long help", "--help", true},
		{"Short help", "-h", true},
		{"Help command", "help", true},
		{"Not help", "start", false},
		{"Partial help", "--hel", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't test the private method directly, but we can test IsHelpRequested
			result := detector.IsHelpRequested([]string{tt.arg})
			if result != tt.expected {
				t.Errorf("isHelpFlag(%s) via IsHelpRequested = %v, expected %v", tt.arg, result, tt.expected)
			}
		})
	}
}
