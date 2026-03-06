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
func NewCommandRouter() CommandRouter {
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

	helpIntegration := config.getHelpIntegration()
	if helpIntegration.GetHelpService().IsHelpRequested(helpArgs) {
		err := helpIntegration.ShowHelp(helpArgs, config.commands)
		return nil, nil, err // Help shown, no command to execute
	}

	// Handle no command case - show global help
	if len(args) < 2 {
		err := helpIntegration.ShowHelp([]string{"--help"}, config.commands)
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
		helpIntegration := config.getHelpIntegration()
		err := helpIntegration.ShowHelp([]string{"--help"}, config.commands)
		return nil, nil, err // Help shown, no command to execute
	}

	// Extract command name and remaining args
	commandName := args[1]
	remainingArgs := args[2:]

	// Check if the command name is actually a help flag
	if commandName == "--help" || commandName == "-h" || commandName == "help" {
		helpIntegration := config.getHelpIntegration()
		err := helpIntegration.ShowHelp([]string{"--help"}, config.commands)
		return nil, nil, err // Help shown, no command to execute
	}

	// Find command
	cmd, exists := config.commands[commandName]
	if !exists {
		suggestions := config.findSuggestions(commandName)
		return nil, nil, fmt.Errorf("unknown command: %q\nDid you mean: %s?", commandName, suggestions)
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

	// Check for help requests using context-aware detection
	helpIntegration := config.getHelpIntegration()
	helpService := helpIntegration.GetHelpService()
	helpFactory := helpService.GetFactory()

	// Use the remaining args after subcommand routing for help detection
	helpRequest := helpFactory.DetectHelpRequestWithContext(finalCtx.Args, commandPath)

	if helpRequest.Type != HelpTypeNone {
		// Show appropriate help based on the detected type
		switch helpRequest.Type {
		case HelpTypeGlobal:
			err := helpIntegration.ShowHelp([]string{"--help"}, config.commands)
			return nil, nil, err
		case HelpTypeCommand:
			err := helpIntegration.ShowHelp([]string{helpRequest.Command, "--help"}, config.commands)
			return nil, nil, err
		case HelpTypeSubcommand:
			err := cr.showSubcommandHelp(helpRequest.Command, helpRequest.Subcommand, config)
			return nil, nil, err
		}
	}

	return finalCmd, finalCtx, nil
}

// showSubcommandHelp displays help for a specific subcommand
func (cr *commandRouter) showSubcommandHelp(parentCommand, subcommandName string, config *Config) error {
	parentCmd, exists := config.commands[parentCommand]
	if !exists {
		return fmt.Errorf("parent command %s not found", parentCommand)
	}

	subCmd := parentCmd.FindSubCommand(subcommandName)
	if subCmd == nil {
		return fmt.Errorf("subcommand %s not found in command %s", subcommandName, parentCommand)
	}

	// Use the help service directly to generate command help for the subcommand
	helpIntegration := config.getHelpIntegration()
	helpService := helpIntegration.GetHelpService()

	// Generate command help for the subcommand
	executable := "example_commands" // This should be dynamic but for now it's ok
	commandHelp := helpService.GetFactory().CreateCommandHelp(subCmd, executable)

	// Format the help using the command help formatter
	formatter := helpService.GetFormatter()
	helpText, err := formatter.FormatCommandHelp(commandHelp)
	if err != nil {
		return err
	}

	// Print the help
	output := helpService.GetOutput()
	return output.Print(helpText)
}
