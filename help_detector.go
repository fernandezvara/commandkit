// commandkit/help_detector.go
package commandkit

import (
	"fmt"
	"strings"
)

// HelpDetector detects help requests in command arguments
type HelpDetector interface {
	IsHelpRequested(args []string) bool
	GetHelpType(args []string) HelpType
	ExtractCommandFromArgs(args []string) (string, []string)
	ParseHelpRequest(args []string) *HelpRequest
	GetExecutableName(args []string) string
	FormatHelpUsage(executable string) string

	// Context-aware help detection methods
	DetectHelpWithContext(args []string, commandPath []string) *HelpRequest
	ParseHelpRequestWithContext(args []string, commandPath []string) *HelpRequest
}

// helpDetector implements HelpDetector
type helpDetector struct {
	helpFlags []string
}

// NewHelpDetector creates a new help detector
func NewHelpDetector() HelpDetector {
	return &helpDetector{
		helpFlags: []string{"--help", "-h", "help"},
	}
}

// IsHelpRequested checks if help is requested in the arguments
func (hd *helpDetector) IsHelpRequested(args []string) bool {
	for _, arg := range args {
		for _, helpFlag := range hd.helpFlags {
			if arg == helpFlag {
				return true
			}
		}
	}
	return false
}

// GetHelpType determines the type of help request
func (hd *helpDetector) GetHelpType(args []string) HelpType {
	if !hd.IsHelpRequested(args) {
		return HelpTypeNone
	}

	// Check for global help (no command specified)
	if len(args) <= 1 {
		return HelpTypeGlobal
	}

	// Check if first arg is a help flag
	if hd.isHelpFlag(args[0]) {
		return HelpTypeGlobal
	}

	// Check for command-specific help
	if len(args) >= 2 {
		if hd.isHelpFlag(args[1]) {
			return HelpTypeCommand
		}

		// Check for subcommand help
		if len(args) >= 3 && hd.isHelpFlag(args[2]) {
			return HelpTypeSubcommand
		}
	}

	return HelpTypeCommand
}

// ExtractCommandFromArgs extracts the command name from arguments
func (hd *helpDetector) ExtractCommandFromArgs(args []string) (string, []string) {
	if len(args) == 0 {
		return "", args
	}

	// If first arg is help flag, no command specified
	if hd.isHelpFlag(args[0]) {
		return "", args
	}

	// First non-help arg is the command
	for i, arg := range args {
		if !hd.isHelpFlag(arg) {
			return arg, args[i+1:]
		}
	}

	return "", args
}

// ParseHelpRequest parses a complete help request
func (hd *helpDetector) ParseHelpRequest(args []string) *HelpRequest {
	helpType := hd.GetHelpType(args)

	request := &HelpRequest{
		Type:     helpType,
		Args:     args,
		Original: args,
	}

	switch helpType {
	case HelpTypeGlobal:
		// No command specified
		request.Command = ""

	case HelpTypeCommand:
		// Extract command name
		command, remaining := hd.ExtractCommandFromArgs(args)
		request.Command = command
		request.Args = remaining

	case HelpTypeSubcommand:
		// Extract command and subcommand
		if len(args) >= 2 {
			request.Command = args[0]
			if len(args) >= 3 {
				request.Subcommand = args[1]
			}
		}
	}

	return request
}

// isHelpFlag checks if an argument is a help flag
func (hd *helpDetector) isHelpFlag(arg string) bool {
	for _, helpFlag := range hd.helpFlags {
		if arg == helpFlag {
			return true
		}
	}
	return false
}

// GetExecutableName returns the executable name from arguments or a default
func (hd *helpDetector) GetExecutableName(args []string) string {
	if len(args) > 0 {
		// Extract executable from first argument (typically the program path)
		parts := strings.Split(args[0], "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	return "command"
}

// FormatHelpUsage formats help usage information
func (hd *helpDetector) FormatHelpUsage(executable string) string {
	return fmt.Sprintf("Usage: %s <command> [options]\n\nUse '%s <command> --help' for command-specific help\nUse '%s --help' for global help", executable, executable, executable)
}

// DetectHelpWithContext detects help requests considering the command path context
func (hd *helpDetector) DetectHelpWithContext(args []string, commandPath []string) *HelpRequest {
	return hd.ParseHelpRequestWithContext(args, commandPath)
}

// ParseHelpRequestWithContext parses a help request with command path context
func (hd *helpDetector) ParseHelpRequestWithContext(args []string, commandPath []string) *HelpRequest {
	if !hd.IsHelpRequested(args) {
		return &HelpRequest{
			Type:     HelpTypeNone,
			Args:     args,
			Original: args,
		}
	}

	// Find the position of the help flag
	helpIndex := -1
	for i, arg := range args {
		if hd.isHelpFlag(arg) {
			helpIndex = i
			break
		}
	}

	if helpIndex == -1 {
		return &HelpRequest{
			Type:     HelpTypeNone,
			Args:     args,
			Original: args,
		}
	}

	request := &HelpRequest{
		Args:     args,
		Original: args,
	}

	// Determine help type based on command path and help position
	if len(commandPath) == 0 {
		// No command path - global help
		request.Type = HelpTypeGlobal
		request.Command = ""
	} else if helpIndex == len(args)-1 {
		// Help flag is at the end - help for the last command in path
		if len(commandPath) == 1 {
			// Single command - command help
			request.Type = HelpTypeCommand
			request.Command = commandPath[0]
		} else {
			// Multiple commands - subcommand help
			request.Type = HelpTypeSubcommand
			request.Command = commandPath[0]
			request.Subcommand = commandPath[len(commandPath)-1]
		}
	} else {
		// Help flag in middle - determine based on position
		if helpIndex < len(commandPath) {
			// Help flag within command path - help for command at that position
			if helpIndex == 0 {
				request.Type = HelpTypeGlobal
				request.Command = ""
			} else if helpIndex == 1 {
				request.Type = HelpTypeCommand
				request.Command = commandPath[0]
			} else {
				request.Type = HelpTypeSubcommand
				request.Command = commandPath[0]
				if helpIndex < len(commandPath) {
					request.Subcommand = commandPath[helpIndex-1]
				}
			}
		} else {
			// Help flag after command path - help for last command
			if len(commandPath) == 1 {
				request.Type = HelpTypeCommand
				request.Command = commandPath[0]
			} else {
				request.Type = HelpTypeSubcommand
				request.Command = commandPath[0]
				request.Subcommand = commandPath[len(commandPath)-1]
			}
		}
	}

	return request
}
