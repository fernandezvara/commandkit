// commandkit/config.go
package commandkit

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Config holds configuration definitions and values
type Config struct {
	definitions      map[string]*Definition
	values           map[string]any
	secrets          *SecretStore
	flagSet          *flag.FlagSet
	flagValues       map[string]*string
	fileConfig       *FileConfig
	commands         map[string]*Command
	globalMiddleware []CommandMiddleware
	overrideWarnings *OverrideWarnings
	processed        bool
}

// New creates a new Config instance
func New() *Config {
	return &Config{
		definitions:      make(map[string]*Definition),
		values:           make(map[string]any),
		secrets:          newSecretStore(),
		flagSet:          flag.NewFlagSet(os.Args[0], flag.ContinueOnError),
		flagValues:       make(map[string]*string),
		fileConfig:       nil,
		commands:         make(map[string]*Command),
		globalMiddleware: make([]CommandMiddleware, 0),
		overrideWarnings: NewOverrideWarnings(),
		processed:        false,
	}
}

// Define starts a new configuration definition
func (c *Config) Define(key string) *DefinitionBuilder {
	builder := newDefinitionBuilder(c, key)
	c.definitions[key] = builder.def
	return builder
}

// Command starts a new command definition
func (c *Config) Command(name string) *CommandBuilder {
	builder := newCommandBuilder(c, name)
	c.commands[name] = builder.cmd
	return builder
}

// UseMiddleware adds global middleware that applies to all commands
func (c *Config) UseMiddleware(middleware CommandMiddleware) {
	c.globalMiddleware = append(c.globalMiddleware, middleware)
}

// UseMiddlewareForCommands adds middleware only for specific commands
func (c *Config) UseMiddlewareForCommands(commandNames []string, middleware CommandMiddleware) {
	// Create a wrapper middleware that only executes for specified commands
	wrapper := func(next CommandFunc) CommandFunc {
		return func(ctx *CommandContext) error {
			for _, name := range commandNames {
				if ctx.Command == name {
					return middleware(next)(ctx)
				}
			}
			return next(ctx)
		}
	}
	c.globalMiddleware = append(c.globalMiddleware, wrapper)
}

// UseMiddlewareForSubcommands adds middleware only for specific subcommands of a command
func (c *Config) UseMiddlewareForSubcommands(commandName string, subcommandNames []string, middleware CommandMiddleware) {
	// Create a wrapper middleware that only executes for specified subcommands
	wrapper := func(next CommandFunc) CommandFunc {
		return func(ctx *CommandContext) error {
			if ctx.Command == commandName {
				for _, name := range subcommandNames {
					if ctx.SubCommand == name {
						return middleware(next)(ctx)
					}
				}
			}
			return next(ctx)
		}
	}
	c.globalMiddleware = append(c.globalMiddleware, wrapper)
}

// Process parses flags and environment variables, validates all definitions,
// and populates the values map. Returns a CommandResult for unified error handling.
func (c *Config) Process() *CommandResult {
	// Clear previous values if re-processing
	if c.processed {
		c.values = make(map[string]any)
		c.secrets.DestroyAll()
		c.secrets = newSecretStore()
	}
	c.processed = true

	var errs []ConfigError

	// Use centralized FlagParser for consistent flag parsing
	services := NewCommandServices()
	flagParser := services.FlagParser

	// Parse global flags using the centralized service
	parsedFlags, err := flagParser.ParseGlobal(os.Args[1:], c.definitions)
	if err != nil {
		// Collect any parsing errors
		errs = append(errs, ConfigError{
			Key:     "flag_parsing",
			Source:  "flag",
			Message: fmt.Sprintf("Flag parsing error: %v", err),
		})
	}

	// Update Config's flag values with parsed results
	c.flagValues = parsedFlags.Values

	// Process each definition
	for key, def := range c.definitions {
		value, source, err := c.resolveValueWithFiles(key, def)
		if err != nil {
			displayValue := ""
			if value != nil && !def.secret {
				displayValue = fmt.Sprintf("%v", value)
			} else if value != nil && def.secret {
				displayValue = maskSecret(fmt.Sprintf("%v", value))
			}
			errs = append(errs, ConfigError{
				Key:     key,
				Source:  source,
				Value:   displayValue,
				Message: err.Error(),
			})
			continue
		}

		// Store the value
		if def.secret && value != nil {
			// Store secrets in memguard
			strValue := fmt.Sprintf("%v", value)
			c.secrets.Store(key, strValue)
			// Also store a placeholder in values for Has() checks
			c.values[key] = "[SECRET]"
		} else {
			c.values[key] = value
		}
	}

	// Check for source overrides and store warnings
	overrideWarnings := c.checkSourceOverrides()
	if overrideWarnings.HasWarnings() {
		c.overrideWarnings = overrideWarnings
		c.overrideWarnings.LogWarnings()
	}

	// Convert ConfigError slice to CommandResult
	if len(errs) > 0 {
		var errorMessages []string
		for _, configErr := range errs {
			errorMessages = append(errorMessages, configErr.Error())
		}
		return ConfigErrorResult(formatErrors(errs))
	}

	return Success()
}

// PrintErrors prints formatted error messages to stderr
func (c *Config) PrintErrors(errs []ConfigError) {
	fmt.Fprint(os.Stderr, formatErrors(errs))
}

// Destroy cleans up all secrets from memory
func (c *Config) Destroy() {
	c.secrets.DestroyAll()
}

// IsSecret checks if a configuration key is defined as a secret
func (c *Config) IsSecret(key string) bool {
	if def, exists := c.definitions[key]; exists {
		return def.secret
	}
	return false
}

// GetOverrideWarnings returns all override warnings
func (c *Config) GetOverrideWarnings() *OverrideWarnings {
	return c.overrideWarnings
}

// HasOverrideWarnings returns true if there are override warnings
func (c *Config) HasOverrideWarnings() bool {
	return c.overrideWarnings.HasWarnings()
}

// PrintOverrideWarnings prints override warnings to stderr
func (c *Config) PrintOverrideWarnings() {
	if c.overrideWarnings.HasWarnings() {
		fmt.Fprint(os.Stderr, c.overrideWarnings.FormatWarnings())
	}
}

// Dump returns a map of all configuration values (secrets masked)
func (c *Config) Dump() map[string]string {
	result := make(map[string]string)
	for key, def := range c.definitions {
		if def.secret {
			if c.secrets.Get(key).IsSet() {
				result[key] = "[SECRET:" + fmt.Sprintf("%d", c.secrets.Get(key).Size()) + " bytes]"
			} else {
				result[key] = "[SECRET:not set]"
			}
		} else if val, ok := c.values[key]; ok && val != nil {
			result[key] = fmt.Sprintf("%v", val)
		} else {
			result[key] = "[not set]"
		}
	}
	return result
}

// GenerateHelp creates a help message with all configuration options
func (c *Config) GenerateHelp() string {
	var sb strings.Builder

	sb.WriteString("Configuration Options:\n\n")

	for key, def := range c.definitions {
		sb.WriteString(fmt.Sprintf("  %s\n", key))
		sb.WriteString(fmt.Sprintf("    Type: %s\n", def.valueType))

		if def.envVar != "" {
			sb.WriteString(fmt.Sprintf("    Env:  %s\n", def.envVar))
		}
		if def.flag != "" {
			sb.WriteString(fmt.Sprintf("    Flag: --%s\n", def.flag))
		}
		if def.required {
			sb.WriteString("    Required: yes\n")
		}
		if def.secret {
			sb.WriteString("    Secret: yes (protected in memory)\n")
		}
		if def.defaultValue != nil {
			if def.secret {
				sb.WriteString("    Default: [hidden]\n")
			} else {
				sb.WriteString(fmt.Sprintf("    Default: %v\n", def.defaultValue))
			}
		}
		if def.description != "" {
			sb.WriteString(fmt.Sprintf("    Description: %s\n", def.description))
		}

		// List validations
		if len(def.validations) > 0 {
			var valNames []string
			for _, v := range def.validations {
				valNames = append(valNames, v.Name)
			}
			sb.WriteString(fmt.Sprintf("    Validations: %s\n", strings.Join(valNames, ", ")))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// Execute parses command line arguments and executes the appropriate command
func (c *Config) Execute(args []string) error {
	// Create services for routing
	services := NewCommandServices()
	router := services.CommandRouter

	// Route command with error handling
	cmd, ctx, err := router.RouteWithErrorHandling(args, c)
	if err != nil {
		return err
	}

	// If no command to execute (help was shown), return success
	if cmd == nil {
		return nil
	}

	// Handle subcommands
	finalCmd, finalCtx, err := router.HandleSubcommands(cmd, ctx)
	if err != nil {
		return err
	}

	// Execute command with global middleware
	return c.executeWithGlobalMiddleware(finalCmd, finalCtx)
}

// executeWithGlobalMiddleware wraps command execution with global middleware
func (c *Config) executeWithGlobalMiddleware(cmd *Command, ctx *CommandContext) error {
	// Create services for middleware handling
	services := NewCommandServices()
	middlewareChain := services.MiddlewareChain

	// Create the final execution function that runs the command
	execFunc := func(ctx *CommandContext) error {
		result := cmd.Execute(ctx)
		if result.Error != nil {
			// Always display the message if it exists
			if result.Message != "" {
				fmt.Fprintln(os.Stderr, result.Message)
			}
			if result.ShouldExit {
				result.Handle()
			}
		}
		return result.Error
	}

	// Apply global middleware using MiddlewareChain service
	finalFunc := middlewareChain.ApplyGlobalOnly(c.globalMiddleware, execFunc)

	return finalFunc(ctx)
}

// ShowGlobalHelp displays help for all commands
func (c *Config) ShowGlobalHelp() error {
	fmt.Printf("Usage: %s <command> [options]\n\n", os.Args[0])
	fmt.Printf("Available commands:\n\n")

	for name, cmd := range c.commands {
		aliases := ""
		if len(cmd.Aliases) > 0 {
			aliases = fmt.Sprintf(" (aliases: %s)", strings.Join(cmd.Aliases, ", "))
		}
		fmt.Printf("  %-12s %s%s\n", name, cmd.ShortHelp, aliases)
	}

	fmt.Printf("\nUse '%s <command> --help' for command-specific help\n", os.Args[0])
	return nil
}

// ShowCommandHelp displays help for a specific command
func (c *Config) ShowCommandHelp(commandName string) error {
	cmd, exists := c.commands[commandName]
	if !exists {
		return fmt.Errorf("unknown command: %s", commandName)
	}

	fmt.Printf("Usage: %s %s [options]\n\n", os.Args[0], commandName)

	// Use HelpHandler service to get command help
	services := NewCommandServices()
	helpText := services.HelpHandler.GetCommandHelp(cmd)
	fmt.Printf("%s\n", helpText)
	return nil
}

// findSuggestions finds similar command names for suggestions
func (c *Config) findSuggestions(input string) string {
	var suggestions []string
	minDistance := 3

	for name := range c.commands {
		distance := levenshteinDistance(input, name)
		if distance <= minDistance {
			suggestions = append(suggestions, name)
		}
	}

	if len(suggestions) == 0 {
		return "no similar commands found"
	}

	return strings.Join(suggestions, ", ")
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
	}

	for i := 0; i <= len(a); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
