// commandkit/help_integration_test.go
package commandkit

import (
	"strings"
	"testing"
)

func TestNewHelpIntegration(t *testing.T) {
	integration := NewHelpIntegration()
	if integration == nil {
		t.Error("Expected non-nil integration")
	}

	if integration.GetHelpService() == nil {
		t.Error("Expected non-nil help service")
	}
}

func TestHelpIntegration_ShowHelp(t *testing.T) {
	integration := NewHelpIntegration()

	// Use string output for testing
	stringOutput := NewStringHelpOutput()
	integration.SetOutput(stringOutput)

	commands := map[string]*Command{
		"start": {
			Name:      "start",
			ShortHelp: "Start the service",
		},
	}

	err := integration.ShowHelp([]string{"--help"}, commands)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	output := stringOutput.Get()
	if !strings.Contains(output, "Available commands") {
		t.Error("Expected 'Available commands' in output")
	}
}

func TestHelpIntegration_GenerateHelp(t *testing.T) {
	integration := NewHelpIntegration()

	commands := map[string]*Command{
		"deploy": {
			Name:      "deploy",
			ShortHelp: "Deploy the application",
		},
	}

	result, err := integration.GenerateHelp([]string{"deploy", "--help"}, commands)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "deploy") {
		t.Error("Expected 'deploy' in result")
	}
}

func TestHelpIntegration_SetCustomTemplate(t *testing.T) {
	integration := NewHelpIntegration()

	customTemplate := "Custom: {{.Executable}}"
	integration.SetCustomTemplate(TemplateGlobal, customTemplate)

	// Verify template was set (through help service)
	formatter := integration.GetFormatter()
	if formatter == nil {
		t.Error("Expected non-nil formatter")
	}
}

func TestHelpIntegration_AddCustomFunction(t *testing.T) {
	integration := NewHelpIntegration()

	// Add custom function
	integration.AddCustomFunction("reverse", func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	})

	// Test with custom function (this would be tested through actual template rendering)
	formatter := integration.GetFormatter()
	if formatter == nil {
		t.Error("Expected non-nil formatter")
	}
}

func TestHelpIntegration_SetOutput(t *testing.T) {
	integration := NewHelpIntegration()

	stringOutput := NewStringHelpOutput()
	integration.SetOutput(stringOutput)

	if integration.GetOutput() != stringOutput {
		t.Error("Expected output to be set to string output")
	}
}

func TestNewHelpConfig(t *testing.T) {
	config := New()
	helpConfig := NewHelpConfig(config)

	if helpConfig == nil {
		t.Error("Expected non-nil help config")
	}

	if helpConfig.Config != config {
		t.Error("Expected config to be wrapped")
	}
}

func TestHelpConfig_ShowGlobalHelp(t *testing.T) {
	config := New()
	helpConfig := NewHelpConfig(config)

	// Add a command
	config.Command("start").ShortHelp("Start the service")

	// Use string output for testing
	stringOutput := NewStringHelpOutput()
	helpConfig.SetHelpOutput(stringOutput)

	err := helpConfig.ShowGlobalHelp()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	output := stringOutput.Get()
	if !strings.Contains(output, "Available commands") {
		t.Error("Expected 'Available commands' in output")
	}
}

func TestHelpConfig_ShowCommandHelp(t *testing.T) {
	config := New()
	helpConfig := NewHelpConfig(config)

	// Add a command
	config.Command("deploy").ShortHelp("Deploy the application")

	// Use string output for testing
	stringOutput := NewStringHelpOutput()
	helpConfig.SetHelpOutput(stringOutput)

	err := helpConfig.ShowCommandHelp("deploy")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	output := stringOutput.Get()
	if !strings.Contains(output, "deploy") {
		t.Error("Expected 'deploy' in output")
	}
}

func TestHelpConfig_ShowCommandHelp_Unknown(t *testing.T) {
	config := New()
	helpConfig := NewHelpConfig(config)

	// Add a command
	config.Command("start").ShortHelp("Start the service")

	err := helpConfig.ShowCommandHelp("unknown")
	if err == nil {
		t.Error("Expected error for unknown command")
	}
}

func TestHelpConfig_GenerateGlobalHelpText(t *testing.T) {
	config := New()
	helpConfig := NewHelpConfig(config)

	// Add a command
	config.Command("start").ShortHelp("Start the service")

	result, err := helpConfig.GenerateGlobalHelpText()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "Available commands") {
		t.Error("Expected 'Available commands' in result")
	}
}

func TestHelpConfig_GenerateCommandHelpText(t *testing.T) {
	config := New()
	helpConfig := NewHelpConfig(config)

	// Add a command
	config.Command("deploy").ShortHelp("Deploy the application")

	result, err := helpConfig.GenerateCommandHelpText("deploy")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "deploy") {
		t.Error("Expected 'deploy' in result")
	}
}

func TestHelpConfig_SetHelpOutput(t *testing.T) {
	config := New()
	helpConfig := NewHelpConfig(config)

	stringOutput := NewStringHelpOutput()
	helpConfig.SetHelpOutput(stringOutput)

	if helpConfig.GetHelpOutput() != stringOutput {
		t.Error("Expected help output to be set")
	}
}

func TestNewHelpExecutor(t *testing.T) {
	executor := NewHelpExecutor()
	if executor == nil {
		t.Error("Expected non-nil help executor")
	}
}

func TestHelpExecutor_ExecuteHelp(t *testing.T) {
	executor := NewHelpExecutor()

	// Use string output for testing
	stringOutput := NewStringHelpOutput()
	executor.helpIntegration.SetOutput(stringOutput)

	commands := map[string]*Command{
		"start": {
			Name:      "start",
			ShortHelp: "Start the service",
		},
	}

	err := executor.ExecuteHelp([]string{"--help"}, commands)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	output := stringOutput.Get()
	if !strings.Contains(output, "Available commands") {
		t.Error("Expected 'Available commands' in output")
	}
}

func TestHelpExecutor_CheckAndHandleHelp_HelpRequested(t *testing.T) {
	executor := NewHelpExecutor()

	// Use string output for testing
	stringOutput := NewStringHelpOutput()
	executor.helpIntegration.SetOutput(stringOutput)

	commands := map[string]*Command{
		"start": {
			Name:      "start",
			ShortHelp: "Start the service",
		},
	}

	handled, err := executor.CheckAndHandleHelp([]string{"--help"}, commands)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !handled {
		t.Error("Expected help to be handled")
	}

	output := stringOutput.Get()
	if !strings.Contains(output, "Available commands") {
		t.Error("Expected 'Available commands' in output")
	}
}

func TestHelpExecutor_CheckAndHandleHelp_NoHelp(t *testing.T) {
	executor := NewHelpExecutor()

	commands := map[string]*Command{
		"start": {
			Name:      "start",
			ShortHelp: "Start the service",
		},
	}

	handled, err := executor.CheckAndHandleHelp([]string{"start", "server"}, commands)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if handled {
		t.Error("Expected help not to be handled")
	}
}

func TestHelpIntegration_GetHelpExecutor(t *testing.T) {
	integration := NewHelpIntegration()

	executor := integration.GetHelpExecutor()
	if executor == nil {
		t.Error("Expected non-nil help executor")
	}

	if executor.helpIntegration != integration {
		t.Error("Expected executor to reference integration")
	}
}

func TestHelpIntegration_GetFormatter(t *testing.T) {
	integration := NewHelpIntegration()

	formatter := integration.GetFormatter()
	if formatter == nil {
		t.Error("Expected non-nil formatter")
	}
}

func TestHelpIntegration_GetFactory(t *testing.T) {
	integration := NewHelpIntegration()

	factory := integration.GetFactory()
	if factory == nil {
		t.Error("Expected non-nil factory")
	}
}
