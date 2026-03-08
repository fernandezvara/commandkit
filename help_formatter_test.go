// commandkit/help_formatter_test.go
package commandkit

import (
	"strings"
	"testing"
)

func TestNewTemplateHelpFormatter(t *testing.T) {
	formatter := NewTemplateHelpFormatter()
	if formatter == nil {
		t.Error("Expected non-nil formatter")
	}

	// Check default templates are set
	if formatter.GetTemplate(TemplateGlobal) == "" {
		t.Error("Expected default global template")
	}

	if formatter.GetTemplate(TemplateCommand) == "" {
		t.Error("Expected default command template")
	}

	// Check renderer is set
	if formatter.GetRenderer() == nil {
		t.Error("Expected non-nil renderer")
	}
}

func TestTemplateHelpFormatter_FormatGlobalHelp(t *testing.T) {
	formatter := NewTemplateHelpFormatter()

	help := &GlobalHelp{
		Executable: "testapp",
		Commands: []CommandSummary{
			{Name: "start", Description: "Start the service", Aliases: []string{"run"}},
			{Name: "stop", Description: "Stop the service"},
		},
		Template: "Usage: {{.Executable}} <command>\n\nCommands:\n{{range .Commands}}  {{.Name}} - {{.Description}}\n{{end}}",
	}

	result, err := formatter.FormatGlobalHelp(help)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "Usage: testapp <command>\n\nCommands:\n  start - Start the service\n  stop - Stop the service\n"
	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

func TestTemplateHelpFormatter_FormatCommandHelp(t *testing.T) {
	formatter := NewTemplateHelpFormatter()

	cmd := &Command{
		Name:      "deploy",
		ShortHelp: "Deploy the application",
		LongHelp:  "Deploy the application to production",
	}

	help := &CommandHelp{
		Command:     cmd,
		Usage:       "testapp deploy [options]",
		Description: "Deploy the application to production",
		Flags: []FlagInfo{
			{Name: "port", Description: "HTTP server port", Type: "int64", Required: true, Default: 8080},
			{Name: "", Description: "API key", Type: "string", Required: true, Secret: true, EnvVar: "API_KEY", NoFlag: true},
		},
		Template: "{{.Description}}\n\nUsage: {{.Usage}}\n\n{{range .Flags}}--{{.Name}}: {{.Description}}\n{{end}}",
	}

	result, err := formatter.FormatCommandHelp(help)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that basic template rendering works
	if !strings.Contains(result, "Deploy the application to production") {
		t.Error("Expected description in result")
	}

	if !strings.Contains(result, "testapp deploy [options]") {
		t.Error("Expected usage in result")
	}

	if !strings.Contains(result, "port: HTTP server port") {
		t.Error("Expected flag description in result")
	}
}

func TestTemplateHelpFormatter_FormatSubcommandHelp(t *testing.T) {
	formatter := NewTemplateHelpFormatter()

	help := &SubcommandHelp{
		Parent: "deploy",
		Subcommands: []SubcommandInfo{
			{Name: "server", Description: "Deploy server component", Aliases: []string{"srv"}},
			{Name: "database", Description: "Deploy database component"},
		},
		Template: "Subcommands for {{.Parent}}:\n{{range .Subcommands}}  {{.Name}} - {{.Description}}\n{{end}}",
	}

	result, err := formatter.FormatSubcommandHelp(help)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "Subcommands for deploy:\n  server - Deploy server component\n  database - Deploy database component\n"
	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

func TestTemplateHelpFormatter_FormatFlagHelp(t *testing.T) {
	formatter := NewTemplateHelpFormatter()

	help := &FlagHelp{
		Command: "deploy",
		Flags: []FlagInfo{
			{Name: "port", Description: "HTTP server port", Type: "int64", Required: true, Default: 8080},
			{Name: "workers", Description: "Number of workers", Type: "int64", Default: 4},
		},
		Template: "Flags for {{.Command}}:\n{{range .Flags}}--{{.Name}}: {{.Description}} ({{.Type}})\n{{end}}",
	}

	result, err := formatter.FormatFlagHelp(help)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "Flags for deploy:\n--port: HTTP server port (int64)\n--workers: Number of workers (int64)\n"
	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

func TestTemplateHelpFormatter_SetTemplate(t *testing.T) {
	formatter := NewTemplateHelpFormatter()

	customTemplate := "Custom: {{.Executable}}"
	formatter.SetTemplate(TemplateGlobal, customTemplate)

	if formatter.GetTemplate(TemplateGlobal) != customTemplate {
		t.Error("Expected custom template to be set")
	}
}

func TestTemplateHelpFormatter_SetRenderer(t *testing.T) {
	formatter := NewTemplateHelpFormatter()

	newRenderer := NewGoTemplateRenderer()
	formatter.SetRenderer(newRenderer)

	if formatter.GetRenderer() != newRenderer {
		t.Error("Expected renderer to be set")
	}
}

func TestTemplateHelpFormatter_DefaultTemplates(t *testing.T) {
	formatter := NewTemplateHelpFormatter()

	// Test that default templates are set
	globalTemplate := formatter.GetTemplate(TemplateGlobal)
	if globalTemplate == "" {
		t.Error("Expected non-empty global template")
	}

	commandTemplate := formatter.GetTemplate(TemplateCommand)
	if commandTemplate == "" {
		t.Error("Expected non-empty command template")
	}

	subcommandTemplate := formatter.GetTemplate(TemplateSubcommand)
	if subcommandTemplate == "" {
		t.Error("Expected non-empty subcommand template")
	}

	flagTemplate := formatter.GetTemplate(TemplateFlag)
	if flagTemplate == "" {
		t.Error("Expected non-empty flag template")
	}
}

func TestTemplateHelpFormatter_EmptyTemplate(t *testing.T) {
	formatter := NewTemplateHelpFormatter()

	help := &GlobalHelp{
		Executable: "testapp",
		Commands:   []CommandSummary{},
		Template:   "", // Empty template should use default
	}

	result, err := formatter.FormatGlobalHelp(help)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should use default template
	if result == "" {
		t.Error("Expected non-empty result using default template")
	}
}

func TestTemplateHelpFormatter_TemplateerrorResult(t *testing.T) {
	formatter := NewTemplateHelpFormatter()

	help := &GlobalHelp{
		Executable: "testapp",
		Commands:   []CommandSummary{},
		Template:   "{{.InvalidField", // Invalid template
	}

	_, err := formatter.FormatGlobalHelp(help)
	if err == nil {
		t.Error("Expected template parsing error")
	}
}

func TestTemplateHelpFormatter_ComplexTemplate(t *testing.T) {
	formatter := NewTemplateHelpFormatter()

	help := &CommandHelp{
		Command: &Command{
			Name: "deploy",
		},
		Usage:       "testapp deploy [options]",
		Description: "Deploy the application",
		Flags: []FlagInfo{
			{Name: "port", Description: "HTTP server port", Type: "int64", Required: true},
			{Name: "env", Description: "Deployment environment", Type: "string", Default: "dev"},
		},
		Template: `{{.Description | upper}}

{{.Usage}}

{{if .Flags}}OPTIONS:
{{range .Flags}}  --{{.Name}}: {{.Description}} ({{.Type}})
{{end}}{{end}}`,
	}

	result, err := formatter.FormatCommandHelp(help)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check template functions work
	if !strings.Contains(result, "DEPLOY THE APPLICATION") {
		t.Error("Expected uppercase description")
	}

	if !strings.Contains(result, "testapp deploy [options]") {
		t.Error("Expected usage")
	}

	if !strings.Contains(result, "OPTIONS:") {
		t.Error("Expected options section")
	}
}

func TestTemplateHelpFormatter_CustomFunctions(t *testing.T) {
	formatter := NewTemplateHelpFormatter()

	// Add custom function to renderer
	renderer := formatter.GetRenderer()
	renderer.AddFunction("reverse", func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	})

	help := &GlobalHelp{
		Executable: "testapp",
		Commands:   []CommandSummary{},
		Template:   "{{.Executable | reverse}}",
	}

	result, err := formatter.FormatGlobalHelp(help)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "ppatset"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
