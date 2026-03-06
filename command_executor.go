// commandkit/command_executor.go
package commandkit

import "fmt"

// CommandExecutor orchestrates command execution flow
type CommandExecutor interface {
	// Execute orchestrates the complete command execution using all services
	Execute(cmd *Command, ctx *CommandContext, services *CommandServices) *CommandResult
}

// commandExecutor implements CommandExecutor interface
type commandExecutor struct{}

// NewCommandExecutor creates a new CommandExecutor instance
func NewCommandExecutor() CommandExecutor {
	return &commandExecutor{}
}

// Execute orchestrates the complete command execution using all services
func (ce *commandExecutor) Execute(cmd *Command, ctx *CommandContext, services *CommandServices) *CommandResult {
	if cmd == nil {
		return Error(fmt.Errorf("command cannot be nil"))
	}
	
	if ctx == nil {
		return Error(fmt.Errorf("context cannot be nil"))
	}
	
	if services == nil {
		return Error(fmt.Errorf("services cannot be nil"))
	}
	
	// Ensure execution context exists
	if ctx.execution == nil {
		ctx.execution = NewExecutionContext(ctx.Command)
	}

	// 1. Handle help requests
	if err := ce.handleHelp(cmd, ctx, services); err != nil {
		return err
	}

	// 2. Validate command state
	if err := ce.validateCommand(cmd, ctx); err != nil {
		return err
	}

	// 3. Process configuration if needed
	if err := ce.processConfiguration(cmd, ctx, services); err != nil {
		return err
	}

	// 4. Apply and execute middleware
	return ce.executeWithMiddleware(cmd, ctx, services)
}

// handleHelp checks for help requests and displays appropriate help
func (ce *commandExecutor) handleHelp(cmd *Command, ctx *CommandContext, services *CommandServices) *CommandResult {
	helpHandler := services.HelpHandler

	// Check for help request before any processing
	if helpHandler.IsHelpRequested(ctx.Args) {
		helpHandler.ShowCommandHelp(cmd, ctx)
		return Success() // Help was shown successfully
	}

	return nil // Continue with execution
}

// validateCommand validates the command state and returns appropriate help for subcommands
func (ce *commandExecutor) validateCommand(cmd *Command, ctx *CommandContext) *CommandResult {
	// Check if command has no function but has subcommands
	if cmd.Func == nil && len(cmd.SubCommands) > 0 {
		services := NewCommandServices()
		helpText := services.HelpHandler.ShowSubcommandHelp(ctx.Command, cmd.SubCommands, ctx)
		return Error(fmt.Errorf("%s", helpText))
	}

	if cmd.Func == nil {
		return Error(fmt.Errorf("command '%s' has no implementation", ctx.Command))
	}

	return nil // Continue with execution
}

// processConfiguration handles command-specific configuration processing
func (ce *commandExecutor) processConfiguration(cmd *Command, ctx *CommandContext, services *CommandServices) *CommandResult {
	// Process command-specific configuration if any
	if len(cmd.Definitions) > 0 {
		configProcessor := services.ConfigProcessor
		
		// Process configuration
		if result := configProcessor.ProcessCommandConfig(cmd, ctx); result.Error != nil {
			return result // Errors already collected in ctx.execution
		}
		
		// Validate required flags and log warnings for designers
		if result := configProcessor.ValidateRequiredFlags(cmd, ctx); result.Error != nil {
			return result // Should not error for validation, but check anyway
		}
	}

	return nil // Continue with execution
}

// executeWithMiddleware applies middleware and executes the command
func (ce *commandExecutor) executeWithMiddleware(cmd *Command, ctx *CommandContext, services *CommandServices) *CommandResult {
	middlewareChain := services.MiddlewareChain
	
	// Apply middleware using MiddlewareChain service
	finalFunc := middlewareChain.ApplyCommandOnly(cmd, cmd.Func)

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
