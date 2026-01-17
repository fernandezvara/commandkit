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
// and populates the values map. Returns any configuration errors.
func (c *Config) Process() []ConfigError {
	// Clear previous values if re-processing
	if c.processed {
		c.values = make(map[string]any)
		c.secrets.DestroyAll()
		c.secrets = newSecretStore()
	}
	c.processed = true

	var errs []ConfigError

	// Register all flags first (only if not already registered)
	for key, def := range c.definitions {
		if def.flag != "" {
			if _, exists := c.flagValues[key]; !exists {
				c.flagValues[key] = c.flagSet.String(def.flag, "", def.description)
			}
		}
	}

	// Parse command line flags
	// Filter out test flags that might interfere
	filteredArgs := make([]string, 0)
	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "-test.") {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	// Ignore errors from unknown flags to allow partial parsing
	c.flagSet.Parse(filteredArgs)

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

	return errs
}

// resolveValue determines the value from flags, env, or default
func (c *Config) resolveValue(key string, def *Definition) (any, string, error) {
	var rawValue string
	var source string

	// Priority 1: Command line flags
	if def.flag != "" {
		if flagVal, ok := c.flagValues[key]; ok && flagVal != nil && *flagVal != "" {
			rawValue = *flagVal
			source = "flag"
		}
	}

	// Priority 2: Environment variables
	if rawValue == "" && def.envVar != "" {
		if envVal := os.Getenv(def.envVar); envVal != "" {
			rawValue = envVal
			source = "env"
		}
	}

	// Priority 3: Default value
	if rawValue == "" && def.defaultValue != nil {
		source = "default"
		// Default is already the correct type, validate and return
		for _, v := range def.validations {
			if v.Name == "required" {
				continue // Skip required check for defaults
			}
			if err := v.Check(def.defaultValue); err != nil {
				return def.defaultValue, source, err
			}
		}
		return def.defaultValue, source, nil
	}

	// No value found
	if rawValue == "" {
		source = "none"
		if def.required {
			return nil, source, fmt.Errorf("required value not provided (set %s or --%s)", def.envVar, def.flag)
		}
		return nil, source, nil
	}

	// Parse the raw string value into the expected type
	parsedValue, err := parseValue(rawValue, def.valueType, def.delimiter)
	if err != nil {
		return rawValue, source, err
	}

	// Run validations
	for _, validation := range def.validations {
		if err := validation.Check(parsedValue); err != nil {
			return parsedValue, source, err
		}
	}

	return parsedValue, source, nil
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
	if len(args) < 2 {
		// No command provided, process global config and show help
		if errs := c.Process(); len(errs) > 0 {
			c.PrintErrors(errs)
			return fmt.Errorf("global configuration errors")
		}
		return c.ShowGlobalHelp()
	}

	// Handle help commands
	if args[1] == "help" || args[1] == "--help" || args[1] == "-h" {
		if len(args) > 2 {
			return c.ShowCommandHelp(args[2])
		}
		return c.ShowGlobalHelp()
	}

	commandName := args[1]
	remainingArgs := args[2:]

	// Find command
	cmd, exists := c.commands[commandName]
	if !exists {
		suggestions := c.findSuggestions(commandName)
		return fmt.Errorf("unknown command: %q\nDid you mean: %s?", commandName, suggestions)
	}

	// Create command context
	ctx := NewCommandContext(remainingArgs, c, commandName, "")

	// Check for subcommands
	if len(remainingArgs) > 0 {
		subCmdName := remainingArgs[0]
		if subCmd := cmd.FindSubCommand(subCmdName); subCmd != nil {
			ctx.SubCommand = subCmdName
			ctx.Args = remainingArgs[1:]
			return c.executeWithGlobalMiddleware(subCmd, ctx)
		}
	}

	// Execute command with global middleware
	return c.executeWithGlobalMiddleware(cmd, ctx)
}

// executeWithGlobalMiddleware wraps command execution with global middleware
func (c *Config) executeWithGlobalMiddleware(cmd *Command, ctx *CommandContext) error {
	// Create the final execution function that runs the command
	execFunc := func(ctx *CommandContext) error {
		return cmd.Execute(ctx)
	}

	// Apply global middleware in reverse order (last added wraps first)
	for i := len(c.globalMiddleware) - 1; i >= 0; i-- {
		execFunc = c.globalMiddleware[i](execFunc)
	}

	return execFunc(ctx)
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
	fmt.Printf("%s\n", cmd.GetHelp())
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
