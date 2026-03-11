// commandkit/command_router.go
package commandkit

import (
	"fmt"
)

// CommandRouter routes commands and handles subcommands
type CommandRouter interface {
	// RouteCommand parses arguments and routes to the appropriate command
	RouteCommand(args []string, config *Config) (*Command, *CommandContext, error)

	// HandleSubcommands checks for and handles subcommand routing
	HandleSubcommands(cmd *Command, ctx *CommandContext) (*Command, *CommandContext, error)

	// RouteWithErrorHandling is a convenience method that handles special routing cases
	RouteWithErrorHandling(args []string, config *Config) (*Command, *CommandContext, error)

	// RouteWithHelpHandling integrates help detection with command routing
	RouteWithHelpHandling(args []string, config *Config) (*Command, *CommandContext, error)
}

// commandRouter implements CommandRouter interface
type commandRouter struct{}

// NewCommandRouter creates a new CommandRouter instance
func newCommandRouter() CommandRouter {
	return &commandRouter{}
}

// RouteCommand parses arguments and routes to the appropriate command
func (cr *commandRouter) RouteCommand(args []string, config *Config) (*Command, *CommandContext, error) {
	if config == nil {
		return nil, nil, fmt.Errorf("config cannot be nil")
	}

	// Check for help requests first using the new help system
	// Pass args without program name to help system
	var helpArgs []string
	if len(args) > 0 {
		helpArgs = args[1:] // Skip program name
	}

	helpService := config.getHelpService()
	if helpService.IsHelpRequested(helpArgs) {
		err := helpService.ShowHelp(helpArgs, config.commands)
		return nil, nil, err // Help shown, no command to execute
	}

	// Handle no command case - show global help
	if len(args) < 2 {
		err := config.getHelpService().ShowHelp([]string{"--help"}, config.commands)
		return nil, nil, err // Help shown, no command to execute
	}

	commandName := args[1]
	remainingArgs := args[2:]

	// Find command
	cmd, exists := config.commands[commandName]
	if !exists {
		suggestions := config.findSuggestions(commandName)
		return nil, nil, fmt.Errorf("unknown command: %q\nDid you mean: %s?", commandName, suggestions)
	}

	// Create command context
	ctx := NewCommandContext(remainingArgs, config, commandName, "")

	return cmd, ctx, nil
}

// HandleSubcommands checks for and handles subcommand routing
func (cr *commandRouter) HandleSubcommands(cmd *Command, ctx *CommandContext) (*Command, *CommandContext, error) {
	if cmd == nil {
		return nil, ctx, fmt.Errorf("command cannot be nil")
	}

	if ctx == nil {
		return nil, nil, fmt.Errorf("context cannot be nil")
	}

	// Check for subcommands
	if len(ctx.Args) > 0 {
		subCmdName := ctx.Args[0]
		if subCmd := cmd.FindSubCommand(subCmdName); subCmd != nil {
			// Update context for subcommand
			ctx.SubCommand = subCmdName
			ctx.Args = ctx.Args[1:]

			// Update execution context to include subcommand
			if ctx.execution != nil {
				ctx.execution.SetCommand(ctx.Command + " " + subCmdName)
			}

			return subCmd, ctx, nil
		}
	}

	return cmd, ctx, nil
}

// RouteWithErrorHandling is a convenience method that routes commands with error handling
func (cr *commandRouter) RouteWithErrorHandling(args []string, config *Config) (*Command, *CommandContext, error) {
	cmd, ctx, err := cr.RouteCommand(args, config)
	return cmd, ctx, err
}

// RouteWithHelpHandling integrates help detection with command routing
func (cr *commandRouter) RouteWithHelpHandling(args []string, config *Config) (*Command, *CommandContext, error) {
	if config == nil {
		return nil, nil, fmt.Errorf("config cannot be nil")
	}

	// Handle no command case - show global help
	if len(args) < 2 {
		err := config.getHelpService().ShowHelpUnified("", "", false, []GetError{}, config.commands)
		return nil, nil, err // Help shown, no command to execute
	}

	// Extract command name and remaining args
	commandName := args[1]
	remainingArgs := args[2:]

	// Check if the command name is actually a help flag
	if commandName == "--help" || commandName == "-h" || commandName == "help" {
		err := config.getHelpService().ShowHelpUnified("", "", false, []GetError{}, config.commands)
		return nil, nil, err // Help shown, no command to execute
	}

	// Find command
	cmd, exists := config.commands[commandName]
	if !exists {
		// For no-command apps with synthetic default, route unknown commands to default
		if defaultCmd, hasDefault := config.commands["default"]; hasDefault && len(config.commands) == 1 {
			// Only default command exists, treat args[1] as a flag, not a command
			cmd = defaultCmd
			remainingArgs = args[1:] // Include the "unknown" command as a flag
		} else {
			suggestions := config.findSuggestions(commandName)
			return nil, nil, fmt.Errorf("unknown command: %q\nDid you mean: %s?", commandName, suggestions)
		}
	}

	// Create initial command context
	ctx := NewCommandContext(remainingArgs, config, commandName, "")

	// Handle subcommands first to get the full command path
	finalCmd, finalCtx, err := cr.HandleSubcommands(cmd, ctx)
	if err != nil {
		return nil, nil, err
	}

	// Build command path for help detection
	var commandPath []string
	commandPath = append(commandPath, commandName)
	if finalCtx.SubCommand != "" {
		commandPath = append(commandPath, finalCtx.SubCommand)
	}

	// Check for help requests using unified system
	helpService := config.getHelpService()

	// Detect help mode from args
	full := false
	for _, arg := range finalCtx.Args {
		if arg == "--full-help" {
			full = true
			break
		}
	}

	// Check if help is requested
	if len(finalCtx.Args) > 0 {
		lastArg := finalCtx.Args[len(finalCtx.Args)-1]
		if lastArg == "--help" || lastArg == "-h" || lastArg == "help" || lastArg == "--full-help" {
			// Show subcommand help using unified system
			err := helpService.ShowHelpUnified(finalCtx.Command, finalCtx.SubCommand, full, []GetError{}, config.commands)
			if err != nil {
				return nil, nil, err
			}
			// Help was shown successfully, return nil to prevent command execution
			return nil, nil, nil
		}
	}

	return finalCmd, finalCtx, nil
}

// showSubcommandHelp displays help for a specific subcommand
func (cr *commandRouter) showSubcommandHelp(parentCommand, subcommandName string, config *Config) error {
	// Use the unified help system
	helpService := config.getHelpService()
	return helpService.ShowHelpUnified(parentCommand, subcommandName, false, []GetError{}, config.commands)
}
