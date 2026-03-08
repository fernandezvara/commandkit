// commandkit/config.go
package commandkit

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	helpService      HelpService
	defaultPriority  SourcePriority // Fallback priority for definitions without explicit priority
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
		defaultPriority:  PriorityFlagEnvDefault, // Flag > Env > Default to match test expectations
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

// SetDefaultPriority sets the default priority order for all definitions
// that don't have an explicit priority set
func (c *Config) SetDefaultPriority(priority SourcePriority) *Config {
	c.defaultPriority = append(SourcePriority(nil), priority...)
	return c
}

// GetDefaultPriority returns the current default priority order
func (c *Config) GetDefaultPriority() SourcePriority {
	return append(SourcePriority(nil), c.defaultPriority...)
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
	services := c.createServices()
	flagParser := services.FlagParser

	// Parse global flags using the centralized service
	parsedFlags, err := flagParser.ParseGlobal(os.Args[1:], c.definitions)
	if err != nil {
		// Collect any parsing errors
		errs = append(errs, ConfigError{
			Key:              "flag_parsing",
			Source:           "flag",
			Display:          "",
			ErrorDescription: fmt.Sprintf("Flag parsing error: %v", err),
		})
	}

	// Update Config's flag values with parsed results
	c.flagValues = parsedFlags.Values

	// Process each definition
	for key, def := range c.definitions {
		value, source, err := c.resolveValueWithPriority(key, def)
		if err != nil {
			displayValue := ""
			if value != nil && !def.secret {
				displayValue = fmt.Sprintf("%v", value)
			} else if value != nil && def.secret {
				displayValue = maskSecret(fmt.Sprintf("%v", value))
			}
			errs = append(errs, ConfigError{
				Key:              key,
				Source:           source.String(),
				Value:            displayValue,
				Display:          buildErrorDisplay(def),
				ErrorDescription: err.Error(),
			})
			continue
		}

		// Store the value
		if def.secret && value != nil {
			// Store secrets in memguard only - no placeholders in values map
			strValue := fmt.Sprintf("%v", value)
			c.secrets.Store(key, strValue)
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

	// Convert ConfigError slice to CommandResult using templated help
	if len(errs) > 0 {
		// Create execution context for error display
		executable := filepath.Base(os.Args[0])
		ctx := NewExecutionContext(executable)
		for _, configErr := range errs {
			ctx.CollectConfigError(c, configErr)
		}

		// Use templated help system for consistent display
		helpText, err := ctx.renderErrorsWithCommand(nil, c.getHelpService())
		if err != nil {
			return errorResult(err)
		}

		return configErrorResult(helpText)
	}

	return success()
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

// GenerateHelp creates a help message using the new template-based help system
func (c *Config) GenerateHelp() string {
	text, _ := c.getHelpService().GenerateHelp([]string{"--help"}, c.commands)
	return text
}

// getHelpService returns the help service instance, creating it if needed
func (c *Config) getHelpService() HelpService {
	if c.helpService == nil {
		c.helpService = NewHelpService()
	}
	return c.helpService
}

// createServices creates a new CommandServices instance for internal use
func (c *Config) createServices() *CommandServices {
	return newCommandServices()
}

func (c *Config) Execute(args []string) error {
	// Check if this is a no-command application
	if len(c.commands) == 0 {
		// Handle no-command application directly
		result := c.Process()
		if result.Error != nil {
			// Return the error - let the caller handle display
			return result.Error
		}
		return nil
	}

	// Create services for routing
	services := c.createServices()
	router := services.CommandRouter

	// Route command with integrated help handling
	cmd, ctx, err := router.RouteWithHelpHandling(args, c)
	if err != nil {
		return err
	}

	// If no command to execute (help was shown), return success
	if cmd == nil {
		return nil
	}

	// Execute command with global middleware
	return c.executeWithGlobalMiddleware(cmd, ctx)
}

// executeWithGlobalMiddleware wraps command execution with global middleware
func (c *Config) executeWithGlobalMiddleware(cmd *Command, ctx *CommandContext) error {
	// Create services for middleware handling
	services := c.createServices()
	middlewareChain := services.MiddlewareChain

	// Create the final execution function that runs the command
	execFunc := func(ctx *CommandContext) error {
		result := cmd.Execute(ctx)
		if result.Error != nil {
			// Check if execution context has errors and display them
			if ctx.execution != nil && ctx.execution.HasErrors() {
				helpText, err := ctx.execution.renderErrorsWithCommand(cmd, c.getHelpService())
				if err != nil {
					return err
				}
				fmt.Fprintln(os.Stderr, helpText)
				os.Exit(1)
			}

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

// ShowGlobalHelp displays help for all commands using the new template-based help system
func (c *Config) ShowGlobalHelp() error {
	return c.getHelpService().ShowHelp([]string{"--help"}, c.commands)
}

// ShowCommandHelp displays help for a specific command using the new template-based help system
func (c *Config) ShowCommandHelp(commandName string) error {
	return c.getHelpService().ShowHelp([]string{commandName, "--help"}, c.commands)
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

// createSyntheticDefaultCommand creates a default command for no-command applications
func (c *Config) createSyntheticDefaultCommand() {
	// Create a default command that includes all global definitions
	c.Command("default").
		Func(func(ctx *CommandContext) error {
			return nil // No-op for synthetic command
		}).
		ShortHelp("Run the application").
		LongHelp("Runs the application with the specified configuration.").
		Config(func(cc *CommandConfig) {
			// All global definitions are automatically copied by the Config() method
			// No additional configuration needed for synthetic command
		})
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
