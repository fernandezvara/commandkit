// commandkit/command.go
package commandkit

import (
	"flag"
	"fmt"
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

// Execute executes the command with the given context
func (cmd *Command) Execute(ctx *CommandContext) error {
	// Process command-specific configuration if any
	if len(cmd.Definitions) > 0 {
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
		if errs := tempConfig.Process(); len(errs) > 0 {
			return fmt.Errorf("command configuration errors: %v", errs)
		}

		// Update the context with the processed config
		ctx.Config = tempConfig
	}

	// Apply middleware in reverse order (last added wraps first)
	var finalFunc CommandFunc = cmd.Func

	for i := len(cmd.Middleware) - 1; i >= 0; i-- {
		finalFunc = cmd.Middleware[i](finalFunc)
	}

	return finalFunc(ctx)
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

			sb.WriteString(fmt.Sprintf("  %-20s %s%s%s\n", flag, def.description, required, defaultValue))
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
			sb.WriteString(fmt.Sprintf("  %-12s %s%s\n", name, subCmd.ShortHelp, aliases))
		}
	}

	return sb.String()
}
