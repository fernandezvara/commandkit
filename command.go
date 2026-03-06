// commandkit/command.go
package commandkit

import (
	"fmt"
	"os"
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

// Execute executes the command with the given context and returns CommandResult for unified error handling
func (cmd *Command) Execute(ctx *CommandContext) *CommandResult {
	// Create services and delegate to CommandExecutor
	services := NewCommandServices()
	executor := services.Executor
	return executor.Execute(cmd, ctx, services)
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
