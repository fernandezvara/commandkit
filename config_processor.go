// commandkit/config_processor.go
package commandkit

import (
	"fmt"
	"os"
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

// newConfigProcessor creates a new ConfigProcessor instance for internal use
func newConfigProcessor() ConfigProcessor {
	return &configProcessor{}
}

// ProcessCommandConfig handles command-specific configuration processing
func (cp *configProcessor) ProcessCommandConfig(cmd *Command, ctx *CommandContext) *CommandResult {
	if cmd == nil {
		return errorResult(fmt.Errorf("command cannot be nil"))
	}

	if ctx == nil {
		return errorResult(fmt.Errorf("context cannot be nil"))
	}

	// Use centralized FlagParser for command-specific flag parsing
	services := newCommandServices()
	flagParser := services.FlagParser

	// Parse command-specific flags using the centralized service
	parsedFlags, err := flagParser.ParseCommand(ctx.Args, cmd.Definitions)

	// Create a temporary config with command-specific definitions and parsed flags
	tempConfig := &Config{
		definitions:      cmd.Definitions,
		values:           make(map[string]any),
		secrets:          newSecretStore(),
		flagSet:          parsedFlags.FlagSet,
		flagValues:       parsedFlags.Values,
		fileConfig:       ctx.GlobalConfig.fileConfig,
		commands:         ctx.GlobalConfig.commands,
		defaultPriority:  ctx.GlobalConfig.defaultPriority, // Inherit default priority
		overrideWarnings: NewOverrideWarnings(),            // Initialize override warnings
		processed:        false,
	}

	// Check for flag parsing errors (either from err return or from ParsedFlags.Errors)
	if err != nil || len(parsedFlags.Errors) > 0 {
		var allErrors []error
		if err != nil {
			allErrors = append(allErrors, err)
		}
		allErrors = append(allErrors, parsedFlags.Errors...)

		// Convert flag parsing errors to ConfigError instances and collect them
		flagConfigErrs := flagParser.ConvertFlagErrorsToConfigErrors(allErrors, cmd.Definitions)

		// Clear any previous errors BEFORE collecting flag errors
		if ctx.execution != nil {
			ctx.execution.Clear()
		}

		// Collect flag errors in execution context using the same pattern as other config errors
		for _, configErr := range flagConfigErrs {
			ctx.execution.CollectConfigError(tempConfig, configErr)
		}

		return configErrorResult("configuration errors detected")
	}

	// Process the command-specific configuration using the same priority system as global config
	// Clear any previous errors to prevent accumulation
	if ctx.execution != nil {
		ctx.execution.Clear()
	}

	var configErrs []ConfigError

	for key, def := range cmd.Definitions {
		// Use the same priority-based resolution as global config
		value, source, err := tempConfig.resolveValueWithPriority(key, def)
		if err != nil {
			displayValue := ""
			if value != nil && !def.secret {
				displayValue = fmt.Sprintf("%v", value)
			} else if value != nil && def.secret {
				displayValue = maskSecret(fmt.Sprintf("%v", value))
			}

			// Create ConfigError using unified function
			configErr := newConfigError(key, def, source.String(), displayValue, err)
			configErrs = append(configErrs, configErr)
			continue
		}

		// Check if required field is missing
		if def.required && value == nil {
			configErr := newConfigError(key, def, "validation", "", fmt.Errorf("Not provided"))
			configErrs = append(configErrs, configErr)
			continue
		}

		if value != nil {
			validationFailed := false
			for _, validation := range def.validations {
				if source == SourceDefault && validation.Name == "required" {
					continue
				}
				if err := validation.Check(value); err != nil {
					rawValue := ""
					if !def.secret {
						rawValue = fmt.Sprintf("%v", value)
					} else {
						rawValue = "[secret]"
					}
					configErr := newConfigError(key, def, source.String(), rawValue, err)
					configErrs = append(configErrs, configErr)
					validationFailed = true
					break
				}
			}
			if validationFailed {
				continue
			}
		}

		// Store the value
		if def.secret && value != nil {
			// Store secrets in memguard only - no placeholders in values map
			strValue := fmt.Sprintf("%v", value)
			tempConfig.secrets.Store(key, strValue)
		} else if value != nil {
			tempConfig.values[key] = value
		}
	}

	// Check for source overrides and store warnings for command-specific config
	overrideWarnings := tempConfig.checkSourceOverrides()
	if overrideWarnings.HasWarnings() {
		tempConfig.overrideWarnings = overrideWarnings
		tempConfig.overrideWarnings.LogWarnings()
	}

	// Return errors if any occurred
	if len(configErrs) > 0 {
		// Collect errors in execution context for display
		for _, configErr := range configErrs {
			ctx.execution.CollectConfigError(tempConfig, configErr)
		}

		return errorResult(fmt.Errorf("configuration errors detected"))
	}

	// Set the command config instead of mutating the context
	ctx.CommandConfig = tempConfig
	return success()
}

// ValidateRequiredFlags checks if all required flags have values and logs warnings for missing ones
func (cp *configProcessor) ValidateRequiredFlags(cmd *Command, ctx *CommandContext) *CommandResult {
	if cmd == nil {
		return success() // Nothing to validate
	}

	if ctx == nil {
		return errorResult(fmt.Errorf("context cannot be nil"))
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

	return success()
}
