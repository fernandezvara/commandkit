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

// ProcessCommandConfig handles command-specific configuration processing
func (cp *configProcessor) ProcessCommandConfig(cmd *Command, ctx *CommandContext) *CommandResult {
	if cmd == nil {
		return Error(fmt.Errorf("command cannot be nil"))
	}

	if ctx == nil {
		return Error(fmt.Errorf("context cannot be nil"))
	}

	// Use centralized FlagParser for command-specific flag parsing
	services := NewCommandServices()
	flagParser := services.FlagParser

	// Parse command-specific flags using the centralized service
	parsedFlags, err := flagParser.ParseCommand(ctx.Args, cmd.Definitions)
	if err != nil {
		// Collect any parsing errors
		return Error(fmt.Errorf("command flag parsing error: %v", err))
	}

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

			// Check if this is a "Not provided" error and create proper ConfigError
			if err.Error() == "Not provided" {
				configErr := newRequiredConfigError(key, def)
				configErrs = append(configErrs, configErr)
			} else {
				// This is a validation or parsing error - create proper ConfigError with display
				configErr := ConfigError{
					Key:              key,
					Source:           source.String(),
					Value:            displayValue,
					Message:          err.Error(),
					Display:          buildErrorDisplay(def),
					ErrorDescription: err.Error(),
				}
				configErrs = append(configErrs, configErr)
			}
			continue
		}

		// Check if required field is missing
		if def.required && value == nil {
			configErr := newRequiredConfigError(key, def)
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
					configErr := newValidationConfigError(key, def, source.String(), rawValue, value, validation.Name, err)
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

		return Error(fmt.Errorf("configuration errors detected"))
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
