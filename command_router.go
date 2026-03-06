// commandkit/command_router.go
package commandkit

import (
	"fmt"
	"os"
	"strings"
)

// CommandRouter routes commands and handles subcommands
type CommandRouter interface {
	// RouteCommand parses arguments and routes to the appropriate command
	RouteCommand(args []string, config *Config) (*Command, *CommandContext, error)

	// HandleSubcommands checks for and handles subcommand routing
	HandleSubcommands(cmd *Command, ctx *CommandContext) (*Command, *CommandContext, error)

	// RouteWithErrorHandling is a convenience method that handles special routing cases
	RouteWithErrorHandling(args []string, config *Config) (*Command, *CommandContext, error)
}

// commandRouter implements CommandRouter interface
type commandRouter struct{}

// NewCommandRouter creates a new CommandRouter instance
func NewCommandRouter() CommandRouter {
	return &commandRouter{}
}

// RouteCommand parses arguments and routes to the appropriate command
func (cr *commandRouter) RouteCommand(args []string, config *Config) (*Command, *CommandContext, error) {
	if config == nil {
		return nil, nil, fmt.Errorf("config cannot be nil")
	}

	// Create execution context at entry point
	execCtx := NewExecutionContext("")

	// Handle no command case
	if len(args) < 2 {
		execCtx.SetCommand("help")
		if result := config.Process(); result.Error != nil {
			if result.Message != "" {
				fmt.Fprintln(os.Stderr, result.Message)
			}
			return nil, nil, fmt.Errorf("global configuration errors")
		}
		return nil, nil, fmt.Errorf("show global help") // Special case for help
	}

	// Handle help commands
	if isHelpCommand(args[1]) {
		execCtx.SetCommand("help")
		if len(args) > 2 {
			return nil, nil, fmt.Errorf("show command help: %s", args[2])
		}
		return nil, nil, fmt.Errorf("show global help") // Special case for help
	}

	commandName := args[1]
	remainingArgs := args[2:]
	execCtx.SetCommand(commandName)

	// Find command
	cmd, exists := config.commands[commandName]
	if !exists {
		suggestions := config.findSuggestions(commandName)
		return nil, nil, fmt.Errorf("unknown command: %q\nDid you mean: %s?", commandName, suggestions)
	}

	// Create command context with execution context
	ctx := NewCommandContext(remainingArgs, config, commandName, "")
	ctx.execution = execCtx

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

// isHelpCommand checks if the argument is a help command
func isHelpCommand(arg string) bool {
	return arg == "help" || arg == "--help" || arg == "-h"
}

// RouteWithErrorHandling is a convenience method that handles special routing cases
func (cr *commandRouter) RouteWithErrorHandling(args []string, config *Config) (*Command, *CommandContext, error) {
	cmd, ctx, err := cr.RouteCommand(args, config)

	// Handle special routing cases
	if err != nil {
		if err.Error() == "show global help" {
			if execErr := config.ShowGlobalHelp(); execErr != nil {
				return nil, nil, execErr
			}
			return nil, nil, nil // Success, but no command to execute
		}

		if strings.HasPrefix(err.Error(), "show command help:") {
			commandName := strings.TrimPrefix(err.Error(), "show command help: ")
			if execErr := config.ShowCommandHelp(commandName); execErr != nil {
				return nil, nil, execErr
			}
			return nil, nil, nil // Success, but no command to execute
		}

		// Return actual error
		return nil, nil, err
	}

	return cmd, ctx, nil
}
