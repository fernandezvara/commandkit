// commandkit/help_coordinator.go
package commandkit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// HelpCoordinator provides a unified entry point for all help functionality
type HelpCoordinator struct {
	templates *TemplateComposer
	extractor *UnifiedExtractor
	cache     map[string]string
	output    HelpOutputInterface
}

// HelpOutputInterface defines the interface for help output (avoiding conflicts)
type HelpOutputInterface interface {
	Print(text string) error
}

// NewHelpCoordinator creates a new help coordinator
func NewHelpCoordinator() *HelpCoordinator {
	return &HelpCoordinator{
		templates: NewTemplateComposer(),
		extractor: NewUnifiedExtractor(),
		cache:     make(map[string]string),
		output:    &DefaultHelpOutput{},
	}
}

// SetOutput sets the help output interface
func (hc *HelpCoordinator) SetOutput(output HelpOutputInterface) {
	hc.output = output
}

// TriggerHelp is the main entry point for all help scenarios
func (hc *HelpCoordinator) TriggerHelp(ctx *CommandContext, errors []GetError) error {
	if ctx == nil {
		return fmt.Errorf("command context cannot be nil")
	}

	// Detect help mode from arguments
	full := hc.detectHelpMode(ctx.Args)

	// Extract command and subcommand from context
	command := ctx.Command
	subcommand := ctx.SubCommand

	return hc.ShowHelp(command, subcommand, full, errors)
}

// ShowHelp displays help for a specific command/subcommand with optional errors
func (hc *HelpCoordinator) ShowHelp(command, subcommand string, full bool, errors []GetError) error {
	return hc.ShowHelpWithCommands(command, subcommand, full, errors, nil)
}

// ShowHelpWithCommands displays help with access to commands map
func (hc *HelpCoordinator) ShowHelpWithCommands(command, subcommand string, full bool, errors []GetError, commands map[string]*Command) error {
	// Get the command
	var cmd *Command

	if command == "" {
		// Global help - pass commands to showGlobalHelp
		return hc.showGlobalHelp(commands)
	}

	if commands != nil {
		// Use provided commands map
		if subcommand != "" {
			// Subcommand help
			parentCmd, exists := commands[command]
			if !exists {
				return fmt.Errorf("parent command %s not found", command)
			}

			subCmd, exists := parentCmd.SubCommands[subcommand]
			if !exists {
				return fmt.Errorf("subcommand %s not found in command %s", subcommand, command)
			}
			cmd = subCmd
		} else {
			// Command help
			var exists bool
			cmd, exists = commands[command]
			if !exists {
				// Fall back to simple command help for unknown commands
				return hc.showSimpleCommandHelp(command, subcommand, errors)
			}
		}
	} else {
		// Fallback to simple command help without full details
		return hc.showSimpleCommandHelp(command, subcommand, errors)
	}

	// Determine help mode
	mode := HelpModeEssential
	if full {
		mode = HelpModeFull
	}

	// Extract help data
	helpData := hc.extractor.ExtractHelpData(cmd, mode, errors)

	// Compose template
	template := hc.templates.ComposeTemplate(len(errors) > 0, full)

	// Format and render help using template
	text, err := hc.formatHelp(helpData, template)
	if err != nil {
		return fmt.Errorf("failed to format help: %w", err)
	}

	return hc.output.Print(text)
}

// showSimpleCommandHelp shows basic help when commands map is not available
func (hc *HelpCoordinator) showSimpleCommandHelp(command, subcommand string, errors []GetError) error {
	// For now, fall back to a simple format
	var builder strings.Builder

	if subcommand != "" {
		fmt.Fprintf(&builder, "Usage: %s %s [options]\n\n", command, subcommand)
	} else {
		fmt.Fprintf(&builder, "Usage: %s [options]\n\n", command)
	}

	if len(errors) > 0 {
		fmt.Fprintf(&builder, "Configuration errors:\n")
		for _, err := range errors {
			fmt.Fprintf(&builder, "  %s -> %s\n", err.Display, err.ErrorDescription)
		}
		fmt.Fprintf(&builder, "\n")
	}

	fmt.Fprintf(&builder, "Use %s %s", hc.getExecutableName(), command)
	if subcommand != "" {
		fmt.Fprintf(&builder, " %s", subcommand)
	}
	fmt.Fprintf(&builder, " --help' or '--full-help' for more information\n")

	return hc.output.Print(builder.String())
}

// formatHelp formats help data using the template
func (hc *HelpCoordinator) formatHelp(helpData *UnifiedHelpData, templateString string) (string, error) {
	// Create template data structure that matches the template expectations
	templateData := struct {
		Command         *Command
		Name            string // For command name
		Usage           string
		Description     string
		Flags           []FlagInfo
		RequiredEnvVars []FlagInfo // For basic mode
		AllEnvVars      []FlagInfo // For full mode
		Subcommands     []SubcommandInfo
		Errors          []GetError
		HasErrors       bool
		Executable      string
	}{
		Command:         helpData.Command,
		Name:            helpData.Usage, // Extract command name from usage
		Usage:           helpData.Usage,
		Description:     helpData.Description,
		Flags:           helpData.Flags,
		RequiredEnvVars: helpData.EnvVars, // In basic mode, this is filtered to required only
		AllEnvVars:      helpData.EnvVars, // In full mode, this includes all
		Subcommands:     helpData.Subcommands,
		Errors:          helpData.Errors,
		HasErrors:       helpData.HasErrors,
		Executable:      hc.getExecutableName(),
	}

	// Create template with custom functions
	funcMap := template.FuncMap{
		"join": strings.Join,
	}

	// Parse and execute template
	tmpl, err := template.New("help").Funcs(funcMap).Parse(templateString)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var builder strings.Builder
	err = tmpl.Execute(&builder, templateData)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return builder.String(), nil
}

// detectHelpMode determines if full help is requested
func (hc *HelpCoordinator) detectHelpMode(args []string) bool {
	for _, arg := range args {
		if arg == "--full-help" {
			return true
		}
	}
	return false
}

// showGlobalHelp displays help for all commands
func (hc *HelpCoordinator) showGlobalHelp(commands map[string]*Command) error {
	executable := hc.getExecutableName()

	// Create a simple global help
	var builder strings.Builder
	fmt.Fprintf(&builder, "Usage: %s <command> [options]\n\n", executable)

	if len(commands) > 0 {
		fmt.Fprintf(&builder, "Available commands:\n\n")

		// List commands with descriptions
		for name, cmd := range commands {
			fmt.Fprintf(&builder, "  %-12s", name)
			if cmd.LongHelp != "" {
				// Get first line of description
				lines := strings.Split(cmd.LongHelp, "\n")
				if len(lines) > 0 {
					fmt.Fprintf(&builder, "%s", lines[0])
				}
			}
			// Add aliases if present
			if len(cmd.Aliases) > 0 {
				fmt.Fprintf(&builder, " (aliases: ")
				for i, alias := range cmd.Aliases {
					if i > 0 {
						fmt.Fprintf(&builder, ", ")
					}
					fmt.Fprintf(&builder, "%s", alias)
				}
				fmt.Fprintf(&builder, ")")
			}
			fmt.Fprintf(&builder, "\n")
		}
		fmt.Fprintf(&builder, "\n")
	}

	fmt.Fprintf(&builder, "Use '%s <command> --help' for command-specific help\n", executable)

	return hc.output.Print(builder.String())
}

// getExecutableName returns the executable name
func (hc *HelpCoordinator) getExecutableName() string {
	if len(os.Args) > 0 {
		return filepath.Base(os.Args[0])
	}
	return "commandkit"
}

// DefaultHelpOutput provides default help output to stdout
type DefaultHelpOutput struct{}

// Print prints text to stdout
func (dho *DefaultHelpOutput) Print(text string) error {
	fmt.Print(text)
	return nil
}

// StringHelpOutput captures help output as a string (for testing)
type StringHelpOutputInterface struct {
	Output strings.Builder
}

// Print appends text to the string builder
func (sho *StringHelpOutputInterface) Print(text string) error {
	sho.Output.WriteString(text)
	return nil
}

// GetString returns the accumulated output
func (sho *StringHelpOutputInterface) GetString() string {
	return sho.Output.String()
}
