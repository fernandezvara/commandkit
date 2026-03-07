// commandkit/help_service.go
package commandkit

import (
	"fmt"
	"os"
	"strings"
)

// HelpService provides help display and generation functionality
type HelpService interface {
	ShowHelp(args []string, commands map[string]*Command) error
	GenerateHelp(args []string, commands map[string]*Command) (string, error)
	ShowGlobalHelp(commands map[string]*Command) error
	ShowCommandHelp(commandName string, commands map[string]*Command) error
	ShowSubcommandHelp(parent string, subcommands map[string]*Command) error
	SetOutput(output HelpOutput)
	GetOutput() HelpOutput
	IsHelpRequested(args []string) bool
	GetHelpType(args []string) HelpType
	GetFormatter() HelpFormatter
	GetFactory() HelpFactory
}

// HelpOutput handles help output destinations
type HelpOutput interface {
	Print(text string) error
	Get() string
	Reset()
}

// helpService implements HelpService
type helpService struct {
	factory   HelpFactory
	formatter HelpFormatter
	output    HelpOutput
}

// NewHelpService creates a new help service
func NewHelpService() HelpService {
	factory := NewHelpFactory()
	formatter := NewTemplateHelpFormatter()

	return &helpService{
		factory:   factory,
		formatter: formatter,
		output:    NewConsoleHelpOutput(),
	}
}

// ShowHelp detects help type and displays appropriate help
func (hs *helpService) ShowHelp(args []string, commands map[string]*Command) error {
	request := hs.factory.DetectHelpRequest(args)

	switch request.Type {
	case HelpTypeNone:
		return fmt.Errorf("no help requested")
	case HelpTypeGlobal:
		return hs.ShowGlobalHelp(commands)
	case HelpTypeCommand:
		return hs.ShowCommandHelp(request.Command, commands)
	case HelpTypeSubcommand:
		return hs.ShowSubcommandHelp(request.Command, commands)
	default:
		return fmt.Errorf("unsupported help type: %v", request.Type)
	}
}

// GenerateHelp generates help text without displaying it
func (hs *helpService) GenerateHelp(args []string, commands map[string]*Command) (string, error) {
	request := hs.factory.DetectHelpRequest(args)

	switch request.Type {
	case HelpTypeNone:
		return "", fmt.Errorf("no help requested")
	case HelpTypeGlobal:
		return hs.generateGlobalHelp(commands)
	case HelpTypeCommand:
		return hs.generateCommandHelp(request.Command, commands)
	case HelpTypeSubcommand:
		return hs.generateSubcommandHelp(request.Command, commands)
	default:
		return "", fmt.Errorf("unsupported help type: %v", request.Type)
	}
}

// ShowGlobalHelp displays help for all commands
func (hs *helpService) ShowGlobalHelp(commands map[string]*Command) error {
	text, err := hs.generateGlobalHelp(commands)
	if err != nil {
		return err
	}

	return hs.output.Print(text)
}

// ShowCommandHelp displays help for a specific command
func (hs *helpService) ShowCommandHelp(commandName string, commands map[string]*Command) error {
	text, err := hs.generateCommandHelp(commandName, commands)
	if err != nil {
		return err
	}

	return hs.output.Print(text)
}

// ShowSubcommandHelp displays help for subcommands
func (hs *helpService) ShowSubcommandHelp(parent string, commands map[string]*Command) error {
	text, err := hs.generateSubcommandHelp(parent, commands)
	if err != nil {
		return err
	}

	return hs.output.Print(text)
}

// generateGlobalHelp generates global help text
func (hs *helpService) generateGlobalHelp(commands map[string]*Command) (string, error) {
	// Get executable name
	executable := hs.getExecutableName()

	// Create global help data
	globalHelp := hs.factory.CreateGlobalHelp(commands, executable)

	// Format using template
	return hs.formatter.FormatGlobalHelp(globalHelp)
}

// generateCommandHelp generates command help text
func (hs *helpService) generateCommandHelp(commandName string, commands map[string]*Command) (string, error) {
	cmd, exists := commands[commandName]
	if !exists {
		return "", fmt.Errorf("unknown command: %s", commandName)
	}

	// Get executable name
	executable := hs.getExecutableName()

	// Create command help data
	commandHelp := hs.factory.CreateCommandHelp(cmd, executable)

	// Format using template
	return hs.formatter.FormatCommandHelp(commandHelp)
}

// generateSubcommandHelp generates subcommand help text
func (hs *helpService) generateSubcommandHelp(parent string, commands map[string]*Command) (string, error) {
	parentCmd, exists := commands[parent]
	if !exists {
		return "", fmt.Errorf("unknown command: %s", parent)
	}

	// Create subcommand help data
	subcommandHelp := hs.factory.CreateSubcommandHelp(parent, parentCmd.SubCommands)

	// Format using template
	return hs.formatter.FormatSubcommandHelp(subcommandHelp)
}

// getExecutableName gets the executable name from args or default
func (hs *helpService) getExecutableName() string {
	if len(os.Args) > 0 {
		parts := strings.Split(os.Args[0], "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	return "command"
}

// SetOutput sets the help output destination
func (hs *helpService) SetOutput(output HelpOutput) {
	hs.output = output
}

// GetOutput returns the current help output
func (hs *helpService) GetOutput() HelpOutput {
	return hs.output
}

// IsHelpRequested checks if help is requested in arguments
func (hs *helpService) IsHelpRequested(args []string) bool {
	return hs.factory.IsHelpRequested(args)
}

// GetHelpType gets the type of help request
func (hs *helpService) GetHelpType(args []string) HelpType {
	return hs.factory.GetHelpType(args)
}

// GetFormatter returns the help formatter
func (hs *helpService) GetFormatter() HelpFormatter {
	return hs.formatter
}

// GetFactory returns the help factory
func (hs *helpService) GetFactory() HelpFactory {
	return hs.factory
}

// ConsoleHelpOutput implements HelpOutput for console output
type ConsoleHelpOutput struct{}

// NewConsoleHelpOutput creates a new console help output
func NewConsoleHelpOutput() HelpOutput {
	return &ConsoleHelpOutput{}
}

// Print prints text to console
func (cho *ConsoleHelpOutput) Print(text string) error {
	fmt.Print(text)
	return nil
}

// Get returns accumulated output (always empty for console)
func (cho *ConsoleHelpOutput) Get() string {
	return ""
}

// Reset resets the output (no-op for console)
func (cho *ConsoleHelpOutput) Reset() {
	// No-op for console output
}

// StringHelpOutput implements HelpOutput for string accumulation
type StringHelpOutput struct {
	buffer strings.Builder
}

// NewStringHelpOutput creates a new string help output
func NewStringHelpOutput() HelpOutput {
	return &StringHelpOutput{
		buffer: strings.Builder{},
	}
}

// Print appends text to buffer
func (sho *StringHelpOutput) Print(text string) error {
	sho.buffer.WriteString(text)
	return nil
}

// Get returns accumulated output
func (sho *StringHelpOutput) Get() string {
	return sho.buffer.String()
}

// Reset clears the buffer
func (sho *StringHelpOutput) Reset() {
	sho.buffer.Reset()
}

// MultiHelpOutput implements HelpOutput for multiple destinations
type MultiHelpOutput struct {
	outputs []HelpOutput
}

// NewMultiHelpOutput creates a new multi help output
func NewMultiHelpOutput(outputs ...HelpOutput) HelpOutput {
	return &MultiHelpOutput{
		outputs: outputs,
	}
}

// Print sends text to all outputs
func (mho *MultiHelpOutput) Print(text string) error {
	var lastErr error
	for _, output := range mho.outputs {
		if err := output.Print(text); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// Get returns output from first string output or empty string
func (mho *MultiHelpOutput) Get() string {
	for _, output := range mho.outputs {
		if strOutput, ok := output.(*StringHelpOutput); ok {
			return strOutput.Get()
		}
	}
	return ""
}

// Reset resets all outputs
func (mho *MultiHelpOutput) Reset() {
	for _, output := range mho.outputs {
		output.Reset()
	}
}

// ShowCommandHelpWithErrors displays command help with configuration errors
func (hs *helpService) ShowCommandHelpWithErrors(commandName string, commands map[string]*Command, errors []GetError) error {
	helpText, err := hs.GenerateCommandHelpWithErrors(commandName, commands, errors)
	if err != nil {
		return err
	}
	return hs.output.Print(helpText)
}

// GenerateCommandHelpWithErrors generates help string with errors
func (hs *helpService) GenerateCommandHelpWithErrors(commandName string, commands map[string]*Command, errors []GetError) (string, error) {
	command, exists := commands[commandName]
	if !exists {
		return "", fmt.Errorf("command '%s' not found", commandName)
	}

	// Create command help with errors
	executable := hs.getExecutableName()
	help := hs.factory.CreateCommandHelpWithErrors(command, executable, errors)

	return hs.formatter.FormatCommandHelp(help)
}
