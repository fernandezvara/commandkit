// commandkit/config_processor.go
package commandkit

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// ConfigProcessor processes command-specific configuration
type ConfigProcessor interface {
	// ProcessCommandConfig handles command-specific configuration processing
	ProcessCommandConfig(cmd *Command, ctx *CommandContext) *CommandResult

	// ValidateRequiredFlags checks if all required flags have values and logs warnings for missing ones
	ValidateRequiredFlags(cmd *Command, ctx *CommandContext) *CommandResult
}

// configProcessor implements ConfigProcessor interface
type configProcessor struct{}

// NewConfigProcessor creates a new ConfigProcessor instance
func NewConfigProcessor() ConfigProcessor {
	return &configProcessor{}
}

// ProcessCommandConfig handles command-specific configuration processing
func (cp *configProcessor) ProcessCommandConfig(cmd *Command, ctx *CommandContext) *CommandResult {
	if cmd == nil {
		return Error(fmt.Errorf("command cannot be nil"))
	}

	if ctx == nil {
		return Error(fmt.Errorf("context cannot be nil"))
	}

	// Create a temporary config with command-specific definitions
	tempConfig := &Config{
		definitions: cmd.Definitions,
		values:      make(map[string]any),
		secrets:     newSecretStore(),
		flagSet:     flag.NewFlagSet("", flag.ContinueOnError),
		flagValues:  make(map[string]*string),
		fileConfig:  ctx.GlobalConfig.fileConfig,
		commands:    ctx.GlobalConfig.commands,
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

	// Set the command config instead of mutating the context
	ctx.CommandConfig = tempConfig
	return Success()
}

// ValidateRequiredFlags checks if all required flags have values and logs warnings for missing ones
func (cp *configProcessor) ValidateRequiredFlags(cmd *Command, ctx *CommandContext) *CommandResult {
	if cmd == nil {
		return Success() // Nothing to validate
	}

	if ctx == nil {
		return Error(fmt.Errorf("context cannot be nil"))
	}

	for key, def := range cmd.Definitions {
		if def.required {
			// Check if value is provided in any source (flag, env, or default)
			hasValue := false

			// Check flag value
			if def.flag != "" {
				var flagVal *string
				if ctx.CommandConfig != nil {
					flagVal, _ = ctx.CommandConfig.flagValues[key]
				} else {
					// Fall back to global config when no command config
					flagVal, _ = ctx.GlobalConfig.flagValues[key]
				}
				if flagVal != nil && *flagVal != "" {
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

	return Success()
}
