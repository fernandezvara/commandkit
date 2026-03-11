// commandkit/help_service_new.go
package commandkit

import (
	"fmt"
	"strings"
)

// HelpService provides help display and generation functionality
type HelpService interface {
	ShowHelp(args []string, commands map[string]*Command) error
	GenerateHelp(args []string, commands map[string]*Command) (string, error)
	ShowGlobalHelp(commands map[string]*Command) error
	ShowCommandHelp(commandName string, commands map[string]*Command) error
	SetOutput(output HelpOutput)
	GetOutput() HelpOutput
	IsHelpRequested(args []string) bool
	GetHelpType(args []string) HelpType
	GetHelpMode(args []string) HelpMode
	// Primary unified methods
	ShowHelpUnified(command, subcommand string, full bool, errors []GetError, commands map[string]*Command) error
	TriggerHelpUnified(ctx *CommandContext, errors []GetError) error
}

// HelpOutput handles help output destinations
type HelpOutput interface {
	Print(text string) error
	Get() string
	Reset()
}

// ConsoleHelpOutput implements HelpOutput for console output
type ConsoleHelpOutput struct{}

// Print prints text to console
func (cho *ConsoleHelpOutput) Print(text string) error {
	fmt.Print(text)
	return nil
}

// Get returns the accumulated output (not applicable for console)
func (cho *ConsoleHelpOutput) Get() string {
	return ""
}

// Reset resets the output (not applicable for console)
func (cho *ConsoleHelpOutput) Reset() {
	// No-op for console output
}

// StringHelpOutput captures help output as a string
type StringHelpOutput struct {
	buffer strings.Builder
}

// Print appends text to the string builder
func (sho *StringHelpOutput) Print(text string) error {
	sho.buffer.WriteString(text)
	return nil
}

// Get returns the accumulated output
func (sho *StringHelpOutput) Get() string {
	return sho.buffer.String()
}

// Reset clears the accumulated output
func (sho *StringHelpOutput) Reset() {
	sho.buffer.Reset()
}

// helpService implements HelpService using the unified system
type helpService struct {
	coordinator *HelpCoordinator
	output      HelpOutput
}

// NewHelpService creates a new help service
func NewHelpService() HelpService {
	coordinator := NewHelpCoordinator()
	return &helpService{
		coordinator: coordinator,
		output:      &ConsoleHelpOutput{},
	}
}

// ShowHelpUnified provides a unified help interface using the new HelpCoordinator
func (hs *helpService) ShowHelpUnified(command, subcommand string, full bool, errors []GetError, commands map[string]*Command) error {
	hs.coordinator.SetOutput(hs.output)
	return hs.coordinator.ShowHelpWithCommands(command, subcommand, full, errors, commands)
}

// TriggerHelpUnified triggers help using the new unified system
func (hs *helpService) TriggerHelpUnified(ctx *CommandContext, errors []GetError) error {
	hs.coordinator.SetOutput(hs.output)
	return hs.coordinator.TriggerHelp(ctx, errors)
}

// Legacy methods - implemented using unified system

// ShowHelp detects help type and displays appropriate help
func (hs *helpService) ShowHelp(args []string, commands map[string]*Command) error {
	// Simple help detection
	if len(args) == 0 {
		return hs.ShowGlobalHelp(commands)
	}

	lastArg := args[len(args)-1]
	if lastArg == "--help" || lastArg == "-h" || lastArg == "help" {
		if len(args) >= 2 {
			return hs.ShowHelpUnified(args[1], "", false, []GetError{}, commands)
		}
		return hs.ShowGlobalHelp(commands)
	}

	return fmt.Errorf("no help requested")
}

// GenerateHelp generates help text without displaying it
func (hs *helpService) GenerateHelp(args []string, commands map[string]*Command) (string, error) {
	// Capture output to string
	stringOutput := &StringHelpOutput{}
	originalOutput := hs.output
	hs.SetOutput(stringOutput)
	defer hs.SetOutput(originalOutput)

	err := hs.ShowHelp(args, commands)
	if err != nil {
		return "", err
	}

	return stringOutput.Get(), nil
}

// ShowGlobalHelp displays help for all commands
func (hs *helpService) ShowGlobalHelp(commands map[string]*Command) error {
	return hs.ShowHelpUnified("", "", false, []GetError{}, commands)
}

// ShowCommandHelp displays help for a specific command
func (hs *helpService) ShowCommandHelp(commandName string, commands map[string]*Command) error {
	return hs.ShowHelpUnified(commandName, "", false, []GetError{}, commands)
}

// SetOutput sets the help output destination
func (hs *helpService) SetOutput(output HelpOutput) {
	hs.output = output
	hs.coordinator.SetOutput(output)
}

// GetOutput returns the current help output destination
func (hs *helpService) GetOutput() HelpOutput {
	return hs.output
}

// IsHelpRequested checks if help is requested in arguments
func (hs *helpService) IsHelpRequested(args []string) bool {
	if len(args) == 0 {
		return false
	}

	lastArg := args[len(args)-1]
	return lastArg == "--help" || lastArg == "-h" || lastArg == "help" || lastArg == "--full-help"
}

// GetHelpType gets the type of help request
func (hs *helpService) GetHelpType(args []string) HelpType {
	if len(args) == 0 {
		return HelpTypeNone
	}

	lastArg := args[len(args)-1]
	if lastArg == "--help" || lastArg == "-h" || lastArg == "help" || lastArg == "--full-help" {
		if len(args) >= 2 {
			return HelpTypeCommand
		}
		return HelpTypeGlobal
	}

	return HelpTypeNone
}

// GetHelpMode gets the help mode (essential vs full)
func (hs *helpService) GetHelpMode(args []string) HelpMode {
	for _, arg := range args {
		if arg == "--full-help" {
			return HelpModeFull
		}
	}
	return HelpModeEssential
}
