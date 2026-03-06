// commandkit/command.go
package commandkit

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Command represents a CLI command with its configuration
type Command struct {
	Name        string
	Func        CommandFunc
	ShortHelp   string
	LongHelp    string
	Aliases     []string
	Definitions map[string]*Definition
	SubCommands map[string]*Command
	Middleware  []CommandMiddleware
}

// CommandFunc represents the function that executes a command
type CommandFunc func(*CommandContext) error

// CommandMiddleware represents middleware that can wrap command execution
type CommandMiddleware func(next CommandFunc) CommandFunc

// NewCommand creates a new command instance
func NewCommand(name string) *Command {
	return &Command{
		Name:        name,
		Definitions: make(map[string]*Definition),
		SubCommands: make(map[string]*Command),
		Middleware:  make([]CommandMiddleware, 0),
	}
}

// AddSubCommand adds a subcommand to this command
func (cmd *Command) AddSubCommand(name string, subCmd *Command) {
	cmd.SubCommands[name] = subCmd
}

// validateRequiredFlags checks if all required flags have values and logs warnings for missing ones
func validateRequiredFlags(cmd *Command, ctx *CommandContext) {
	for key, def := range cmd.Definitions {
		if def.required {
			// Check if value is provided in any source (flag, env, or default)
			hasValue := false

			// Check flag value
			if def.flag != "" {
				if flagVal, ok := ctx.Config.flagValues[key]; ok && flagVal != nil && *flagVal != "" {
					hasValue = true
				}
			}

			// Check environment variable
			if !hasValue && def.envVar != "" {
				if envVal := os.Getenv(def.envVar); envVal != "" {
					hasValue = true
				}
			}

			// Check default value
			if !hasValue && def.defaultValue != nil {
				hasValue = true
			}

			// Log warning if required flag is missing
			if !hasValue {
				var displayName string
				if def.flag != "" && def.envVar != "" {
					displayName = fmt.Sprintf("--%s (env: %s)", def.flag, def.envVar)
				} else if def.flag != "" {
					displayName = fmt.Sprintf("--%s", def.flag)
				} else if def.envVar != "" {
					displayName = fmt.Sprintf("env: %s", def.envVar)
				} else {
					displayName = key
				}

				logWarningForDesigner(fmt.Sprintf("Required configuration '%s' is not provided", displayName))
			}
		}
	}
}

// isHelpRequested checks if help is requested in the arguments
func isHelpRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			return true
		}
	}
	return false
}

// Execute executes the command with the given context and returns CommandResult for unified error handling
func (cmd *Command) Execute(ctx *CommandContext) *CommandResult {
	// Ensure execution context exists
	if ctx.execution == nil {
		ctx.execution = NewExecutionContext(ctx.Command)
	}

	// Check for help request before any processing
	if isHelpRequested(ctx.Args) {
		cmd.showEnhancedHelp(ctx)
		return Success() // Help was shown successfully
	}

	// Check if command has no function but has subcommands
	if cmd.Func == nil && len(cmd.SubCommands) > 0 {
		return Error(fmt.Errorf("%s", cmd.GetSubcommandHelp(ctx.Command)))
	}

	if cmd.Func == nil {
		return Error(fmt.Errorf("command '%s' has no implementation", ctx.Command))
	}

	// Process command-specific configuration if any
	if len(cmd.Definitions) > 0 {
		if result := cmd.processCommandConfig(ctx); result.Error != nil {
			return result // Errors already collected in ctx.execution
		}
	}

	// Validate required flags and log warnings for designers
	cmd.validateRequiredFlags(ctx)

	// Apply middleware in reverse order (last added wraps first)
	var finalFunc CommandFunc = cmd.Func

	for i := len(cmd.Middleware) - 1; i >= 0; i-- {
		finalFunc = cmd.Middleware[i](finalFunc)
	}

	// Execute the command function
	err := finalFunc(ctx)

	// Check for collected errors and create appropriate result
	if ctx.execution.HasErrors() {
		// Instead of exiting, return a CommandResult that should exit
		errorMessages := ctx.execution.GetFormattedErrors()
		return ErrorWithExit(err, errorMessages)
	}

	// Return success or error result
	if err != nil {
		return Error(err)
	}

	return Success()
}

// showEnhancedHelp displays comprehensive help including flag help and environment-only configurations
func (cmd *Command) showEnhancedHelp(ctx *CommandContext) {
	// Create a temporary config to properly set up flags for help display
	tempConfig := &Config{
		flagSet:    flag.NewFlagSet("", flag.ContinueOnError),
		flagValues: make(map[string]*string),
	}

	// Register flags to show proper help with enhanced descriptions
	for key, def := range cmd.Definitions {
		if def.flag != "" {
			enhancedDescription := formatFlagHelp(def)
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
			enhancedDescription := formatFlagHelp(def)
			fmt.Printf("  (no flag) string %s\n", enhancedDescription)
			fmt.Printf("        %s\n", def.description)
		}
	}
}

// processCommandConfig handles command-specific configuration processing
func (cmd *Command) processCommandConfig(ctx *CommandContext) *CommandResult {
	// Create a temporary config with command-specific definitions
	tempConfig := &Config{
		definitions: cmd.Definitions,
		values:      make(map[string]any),
		secrets:     newSecretStore(),
		flagSet:     flag.NewFlagSet("", flag.ContinueOnError),
		flagValues:  make(map[string]*string),
		fileConfig:  ctx.Config.fileConfig,
		commands:    ctx.Config.commands,
		processed:   false,
	}

	// Register command-specific flags
	for key, def := range cmd.Definitions {
		if def.flag != "" {
			tempConfig.flagValues[key] = tempConfig.flagSet.String(def.flag, "", def.description)
		}
	}

	// Parse command-specific flags from context.Args
	tempConfig.flagSet.Parse(ctx.Args)

	// Process the command-specific configuration
	result := tempConfig.Process()
	if result.Error != nil {
		// Collect errors in context
		for _, configErr := range result.Context {
			if errMsg, ok := configErr.(string); ok {
				errorType := "not found"
				if strings.Contains(errMsg, "validation") ||
					strings.Contains(errMsg, "greater than") ||
					strings.Contains(errMsg, "less than") ||
					strings.Contains(errMsg, "oneOf") ||
					strings.Contains(errMsg, "required") {
					errorType = "validation"
				}
				// Extract key from context if available
				key := "unknown"
				if k, exists := result.Context["key"]; exists {
					key = fmt.Sprintf("%v", k)
				}
				ctx.execution.CollectErrorWithConfig(tempConfig, key, errorType, "", errMsg, false)
			}
		}
		// Return the actual detailed error message, not a generic one
		return ConfigErrorResult(result.Message)
	}

	// Update the context config
	ctx.Config = tempConfig
	return Success()
}

// validateRequiredFlags checks if all required flags have values and logs warnings for missing ones
func (cmd *Command) validateRequiredFlags(ctx *CommandContext) {
	for key, def := range cmd.Definitions {
		if def.required {
			// Check if value is provided in any source (flag, env, or default)
			hasValue := false

			// Check flag value
			if def.flag != "" {
				if flagVal, ok := ctx.Config.flagValues[key]; ok && flagVal != nil && *flagVal != "" {
					hasValue = true
				}
			}

			// Check environment variable
			if !hasValue && def.envVar != "" {
				if envVal := os.Getenv(def.envVar); envVal != "" {
					hasValue = true
				}
			}

			// Check default value
			if !hasValue && def.defaultValue != nil {
				hasValue = true
			}

			// Log warning if required flag is missing
			if !hasValue {
				var displayName string
				if def.flag != "" && def.envVar != "" {
					displayName = fmt.Sprintf("--%s (env: %s)", def.flag, def.envVar)
				} else if def.flag != "" {
					displayName = fmt.Sprintf("--%s", def.flag)
				} else if def.envVar != "" {
					displayName = fmt.Sprintf("env: %s", def.envVar)
				} else {
					displayName = key
				}

				logWarningForDesigner(fmt.Sprintf("Required configuration '%s' is not provided", displayName))
			}
		}
	}
}

// FindSubCommand finds a subcommand by name or alias
func (cmd *Command) FindSubCommand(name string) *Command {
	// Check exact name first
	if subCmd, exists := cmd.SubCommands[name]; exists {
		return subCmd
	}

	// Check aliases
	for _, subCmd := range cmd.SubCommands {
		for _, alias := range subCmd.Aliases {
			if alias == name {
				return subCmd
			}
		}
	}

	return nil
}

// GetSubcommandHelp returns help text for subcommands of this command
func (cmd *Command) GetSubcommandHelp(commandPath string) string {
	var sb strings.Builder

	// Get executable name from os.Args[0] or use a default
	executable := "command"
	if len(os.Args) > 0 {
		executable = os.Args[0]
	}

	fmt.Fprintf(&sb, "Subcommands for %s:\n\n", commandPath)

	// Sort subcommands for consistent display
	names := make([]string, 0, len(cmd.SubCommands))
	for name := range cmd.SubCommands {
		names = append(names, name)
	}

	for _, name := range names {
		subCmd := cmd.SubCommands[name]
		fmt.Fprintf(&sb, "  %-12s %s\n", name, subCmd.ShortHelp)
	}

	fmt.Fprintf(&sb, "\nUse '%s %s <command> --help' for more information on a specific command.\n", executable, commandPath)

	return sb.String()
}

// GetHelp returns help text for this command
func (cmd *Command) GetHelp() string {
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
