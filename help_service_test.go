// commandkit/help_service_test.go
package commandkit

import (
	"strings"
	"testing"
)

func TestNewHelpService(t *testing.T) {
	service := NewHelpService()
	if service == nil {
		t.Error("Expected non-nil service")
	}
	
	// Check default output is console output
	if service.GetOutput() == nil {
		t.Error("Expected non-nil output")
	}
}

func TestHelpService_ShowHelp_Global(t *testing.T) {
	service := NewHelpService()
	
	// Use string output for testing
	stringOutput := NewStringHelpOutput()
	service.SetOutput(stringOutput)
	
	commands := map[string]*Command{
		"start": {
			Name:      "start",
			ShortHelp: "Start the service",
		},
		"stop": {
			Name:      "stop",
			ShortHelp: "Stop the service",
		},
	}
	
	err := service.ShowHelp([]string{"--help"}, commands)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	output := stringOutput.Get()
	if !strings.Contains(output, "Available commands") {
		t.Error("Expected 'Available commands' in output")
	}
	
	if !strings.Contains(output, "start") {
		t.Error("Expected 'start' command in output")
	}
}

func TestHelpService_ShowHelp_Command(t *testing.T) {
	service := NewHelpService()
	
	// Use string output for testing
	stringOutput := NewStringHelpOutput()
	service.SetOutput(stringOutput)
	
	commands := map[string]*Command{
		"deploy": {
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
		},
	}
	
	err := service.ShowHelp([]string{"deploy", "--help"}, commands)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	output := stringOutput.Get()
	if !strings.Contains(output, "deploy") {
		t.Error("Expected 'deploy' in output")
	}
	
	if !strings.Contains(output, "HTTP server port") {
		t.Error("Expected flag description in output")
	}
}

func TestHelpService_ShowHelp_NoHelp(t *testing.T) {
	service := NewHelpService()
	
	commands := map[string]*Command{
		"start": {
			Name:      "start",
			ShortHelp: "Start the service",
		},
	}
	
	err := service.ShowHelp([]string{"start", "server"}, commands)
	if err == nil {
		t.Error("Expected error for no help request")
	}
	
	expected := "no help requested"
	if err.Error() != expected {
		t.Errorf("Expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestHelpService_ShowHelp_UnknownCommand(t *testing.T) {
	service := NewHelpService()
	
	commands := map[string]*Command{
		"start": {
			Name:      "start",
			ShortHelp: "Start the service",
		},
	}
	
	err := service.ShowHelp([]string{"unknown", "--help"}, commands)
	if err == nil {
		t.Error("Expected error for unknown command")
	}
	
	expected := "unknown command: unknown"
	if err.Error() != expected {
		t.Errorf("Expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestHelpService_GenerateHelp_Global(t *testing.T) {
	service := NewHelpService()
	
	commands := map[string]*Command{
		"start": {
			Name:      "start",
			ShortHelp: "Start the service",
		},
	}
	
	result, err := service.GenerateHelp([]string{"--help"}, commands)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if !strings.Contains(result, "Available commands") {
		t.Error("Expected 'Available commands' in result")
	}
}

func TestHelpService_GenerateHelp_Command(t *testing.T) {
	service := NewHelpService()
	
	commands := map[string]*Command{
		"deploy": {
			Name:      "deploy",
			ShortHelp: "Deploy the application",
		},
	}
	
	result, err := service.GenerateHelp([]string{"deploy", "--help"}, commands)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if !strings.Contains(result, "deploy") {
		t.Error("Expected 'deploy' in result")
	}
}

func TestHelpService_GenerateHelp_NoHelp(t *testing.T) {
	service := NewHelpService()
	
	commands := map[string]*Command{
		"start": {
			Name:      "start",
			ShortHelp: "Start the service",
		},
	}
	
	_, err := service.GenerateHelp([]string{"start", "server"}, commands)
	if err == nil {
		t.Error("Expected error for no help request")
	}
}

func TestHelpService_ShowGlobalHelp(t *testing.T) {
	service := NewHelpService()
	
	// Use string output for testing
	stringOutput := NewStringHelpOutput()
	service.SetOutput(stringOutput)
	
	commands := map[string]*Command{
		"start": {
			Name:      "start",
			ShortHelp: "Start the service",
		},
	}
	
	err := service.ShowGlobalHelp(commands)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	output := stringOutput.Get()
	if !strings.Contains(output, "Available commands") {
		t.Error("Expected 'Available commands' in output")
	}
}

func TestHelpService_ShowCommandHelp(t *testing.T) {
	service := NewHelpService()
	
	// Use string output for testing
	stringOutput := NewStringHelpOutput()
	service.SetOutput(stringOutput)
	
	commands := map[string]*Command{
		"deploy": {
			Name:      "deploy",
			ShortHelp: "Deploy the application",
		},
	}
	
	err := service.ShowCommandHelp("deploy", commands)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	output := stringOutput.Get()
	if !strings.Contains(output, "deploy") {
		t.Error("Expected 'deploy' in output")
	}
}

func TestHelpService_ShowCommandHelp_Unknown(t *testing.T) {
	service := NewHelpService()
	
	commands := map[string]*Command{
		"start": {
			Name:      "start",
			ShortHelp: "Start the service",
		},
	}
	
	err := service.ShowCommandHelp("unknown", commands)
	if err == nil {
		t.Error("Expected error for unknown command")
	}
	
	expected := "unknown command: unknown"
	if err.Error() != expected {
		t.Errorf("Expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestHelpService_ShowSubcommandHelp(t *testing.T) {
	service := NewHelpService()
	
	// Use string output for testing
	stringOutput := NewStringHelpOutput()
	service.SetOutput(stringOutput)
	
	commands := map[string]*Command{
		"deploy": {
			Name: "deploy",
			SubCommands: map[string]*Command{
				"server": {
					Name:      "server",
					ShortHelp: "Deploy server component",
				},
				"database": {
					Name:      "database",
					ShortHelp: "Deploy database component",
				},
			},
		},
	}
	
	err := service.ShowSubcommandHelp("deploy", commands)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	output := stringOutput.Get()
	if !strings.Contains(output, "Subcommands for deploy") {
		t.Error("Expected 'Subcommands for deploy' in output")
	}
	
	if !strings.Contains(output, "server") {
		t.Error("Expected 'server' subcommand in output")
	}
}

func TestHelpService_SetOutput(t *testing.T) {
	service := NewHelpService()
	
	// Test setting string output
	stringOutput := NewStringHelpOutput()
	service.SetOutput(stringOutput)
	
	if service.GetOutput() != stringOutput {
		t.Error("Expected output to be set to string output")
	}
}

func TestNewConsoleHelpOutput(t *testing.T) {
	output := NewConsoleHelpOutput()
	if output == nil {
		t.Error("Expected non-nil console output")
	}
	
	// Test print (should not error)
	err := output.Print("test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	// Test get (should return empty)
	if output.Get() != "" {
		t.Error("Expected empty string from console output")
	}
	
	// Test reset (should not error)
	output.Reset()
}

func TestNewStringHelpOutput(t *testing.T) {
	output := NewStringHelpOutput()
	if output == nil {
		t.Error("Expected non-nil string output")
	}
	
	// Test print
	err := output.Print("test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	// Test get
	if output.Get() != "test" {
		t.Errorf("Expected 'test', got '%s'", output.Get())
	}
	
	// Test reset
	output.Reset()
	if output.Get() != "" {
		t.Error("Expected empty string after reset")
	}
}

func TestNewMultiHelpOutput(t *testing.T) {
	stringOutput1 := NewStringHelpOutput()
	stringOutput2 := NewStringHelpOutput()
	
	multiOutput := NewMultiHelpOutput(stringOutput1, stringOutput2)
	if multiOutput == nil {
		t.Error("Expected non-nil multi output")
	}
	
	// Test print
	err := multiOutput.Print("test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	// Test get (should return first string output)
	if multiOutput.Get() != "test" {
		t.Errorf("Expected 'test', got '%s'", multiOutput.Get())
	}
	
	// Test reset
	multiOutput.Reset()
	if stringOutput1.Get() != "" || stringOutput2.Get() != "" {
		t.Error("Expected empty strings after reset")
	}
}

func TestHelpService_ComplexCommand(t *testing.T) {
	service := NewHelpService()
	
	// Use string output for testing
	stringOutput := NewStringHelpOutput()
	service.SetOutput(stringOutput)
	
	commands := map[string]*Command{
		"deploy": {
			Name:      "deploy",
			ShortHelp: "Deploy the application",
			LongHelp:  "Deploy the application to production environment with all components",
			Definitions: map[string]*Definition{
				"PORT": {
					key:         "PORT",
					flag:        "port",
					description: "HTTP server port",
					valueType:   TypeInt64,
					required:    true,
					defaultValue: int64(8080),
					envVar:      "PORT",
				},
				"API_KEY": {
					key:          "API_KEY",
					envVar:       "API_KEY",
					description:  "API authentication key",
					valueType:    TypeString,
					required:     true,
					secret:       true,
				},
			},
			SubCommands: map[string]*Command{
				"server": {
					Name:      "server",
					ShortHelp: "Deploy server component",
					Aliases:   []string{"srv"},
				},
			},
		},
	}
	
	err := service.ShowHelp([]string{"deploy", "--help"}, commands)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	output := stringOutput.Get()
	
	// Check for various components
	if !strings.Contains(output, "deploy") {
		t.Error("Expected command name in output")
	}
	
	if !strings.Contains(output, "HTTP server port") {
		t.Error("Expected flag description in output")
	}
	
	if !strings.Contains(output, "API authentication key") {
		t.Error("Expected secret flag description in output")
	}
	
	if !strings.Contains(output, "server") {
		t.Error("Expected subcommand in output")
	}
}
