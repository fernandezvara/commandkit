// commandkit/help_handler.go
package commandkit

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

// HelpHandler provides help generation and display functionality
type HelpHandler interface {
	// ShowCommandHelp displays comprehensive help for a command including flags and environment-only configurations
	ShowCommandHelp(cmd *Command, ctx *CommandContext)
	
	// ShowSubcommandHelp returns help text for subcommands of a command
	ShowSubcommandHelp(parent string, subcommands map[string]*Command, ctx *CommandContext) string
	
	// IsHelpRequested checks if help is requested in the arguments
	IsHelpRequested(args []string) bool
	
	// GenerateFlagHelp generates enhanced help text with required/default indicators and validations
	GenerateFlagHelp(def *Definition) string
	
	// GetCommandHelp returns help text for a command
	GetCommandHelp(cmd *Command) string
}

// helpHandler implements HelpHandler interface
type helpHandler struct{}

// NewHelpHandler creates a new HelpHandler instance
func NewHelpHandler() HelpHandler {
	return &helpHandler{}
}

// ShowCommandHelp displays comprehensive help for a command including flags and environment-only configurations
func (h *helpHandler) ShowCommandHelp(cmd *Command, ctx *CommandContext) {
	// Create a temporary config to properly set up flags for help display
	tempConfig := &Config{
		flagSet:    flag.NewFlagSet("", flag.ContinueOnError),
		flagValues: make(map[string]*string),
	}

	// Register flags to show proper help with enhanced descriptions
	for key, def := range cmd.Definitions {
		if def.flag != "" {
			enhancedDescription := h.GenerateFlagHelp(def)
			tempConfig.flagValues[key] = tempConfig.flagSet.String(def.flag, "", enhancedDescription)
		}
	}

	// Show flag help using the flag package's built-in help
	tempConfig.flagSet.PrintDefaults()

	// Show environment-only configurations (no flag)
	var envOnlyConfigs []*Definition
	for _, def := range cmd.Definitions {
		if def.flag == "" && def.envVar != "" {
			envOnlyConfigs = append(envOnlyConfigs, def)
		}
	}

	// Print environment-only configurations if any exist
	if len(envOnlyConfigs) > 0 {
		fmt.Println()
		for _, def := range envOnlyConfigs {
			enhancedDescription := h.GenerateFlagHelp(def)
			fmt.Printf("  (no flag) string %s\n", enhancedDescription)
			fmt.Printf("        %s\n", def.description)
		}
	}
}

// ShowSubcommandHelp returns help text for subcommands of a command
func (h *helpHandler) ShowSubcommandHelp(parent string, subcommands map[string]*Command, ctx *CommandContext) string {
	var sb strings.Builder

	// Get executable name from os.Args[0] or use a default
	executable := "command"
	if len(os.Args) > 0 {
		executable = os.Args[0]
	}

	fmt.Fprintf(&sb, "Subcommands for %s:\n\n", parent)

	// Sort subcommands for consistent display
	names := make([]string, 0, len(subcommands))
	for name := range subcommands {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		subCmd := subcommands[name]
		fmt.Fprintf(&sb, "  %-12s %s\n", name, subCmd.ShortHelp)
	}

	fmt.Fprintf(&sb, "\nUse '%s %s <command> --help' for more information on a specific command.\n", executable, parent)

	return sb.String()
}

// IsHelpRequested checks if help is requested in the arguments
func (h *helpHandler) IsHelpRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			return true
		}
	}
	return false
}

// GenerateFlagHelp generates enhanced help text with required/default indicators and validations
func (h *helpHandler) GenerateFlagHelp(def *Definition) string {
	var indicators []string

	// 1. Environment variable context
	if def.envVar != "" {
		indicators = append(indicators, fmt.Sprintf("env: %s", def.envVar))
	}

	// 2. Required indicator
	if def.required {
		indicators = append(indicators, "required")
	}

	// 3. Default value (masked for secrets)
	if def.defaultValue != nil {
		if def.secret {
			indicators = append(indicators, "default: '[hidden]'")
		} else if def.valueType == TypeString {
			indicators = append(indicators, fmt.Sprintf("default: '%v'", def.defaultValue))
		} else {
			indicators = append(indicators, fmt.Sprintf("default: %v", def.defaultValue))
		}
	}

	// 4. Validations
	validations := formatValidation(def.validations)
	indicators = append(indicators, validations...)

	// 5. Secret indicator
	if def.secret {
		indicators = append(indicators, "secret")
	}

	// Combine description with indicators
	if len(indicators) > 0 {
		return fmt.Sprintf("%s (%s)", def.description, strings.Join(indicators, ", "))
	}

	return def.description
}

// GetCommandHelp returns help text for a command
func (h *helpHandler) GetCommandHelp(cmd *Command) string {
	var sb strings.Builder

	if cmd.LongHelp != "" {
		sb.WriteString(cmd.LongHelp)
		sb.WriteString("\n\n")
	} else if cmd.ShortHelp != "" {
		sb.WriteString(cmd.ShortHelp)
		sb.WriteString("\n\n")
	}

	// Show options if any
	if len(cmd.Definitions) > 0 {
		sb.WriteString("Options:\n")
		for key, def := range cmd.Definitions {
			flag := "--" + def.flag
			if def.flag == "" {
				flag = "--" + strings.ToLower(strings.ReplaceAll(key, "_", "-"))
			}

			required := ""
			if def.required {
				required = " (required)"
			}

			defaultValue := ""
			if def.defaultValue != nil && !def.secret {
				defaultValue = fmt.Sprintf(" (default: %v)", def.defaultValue)
			} else if def.defaultValue != nil && def.secret {
				defaultValue = " (default: [hidden])"
			}

			fmt.Fprintf(&sb, "  %-20s %s%s%s\n", flag, def.description, required, defaultValue)
		}
		sb.WriteString("\n")
	}

	// Show subcommands if any
	if len(cmd.SubCommands) > 0 {
		sb.WriteString("Subcommands:\n")
		for name, subCmd := range cmd.SubCommands {
			aliases := ""
			if len(subCmd.Aliases) > 0 {
				aliases = fmt.Sprintf(" (aliases: %s)", strings.Join(subCmd.Aliases, ", "))
			}
			fmt.Fprintf(&sb, "  %-12s %s%s\n", name, subCmd.ShortHelp, aliases)
		}
	}

	return sb.String()
}
