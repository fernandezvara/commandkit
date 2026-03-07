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
		definitions: cmd.Definitions,
		values:      make(map[string]any),
		secrets:     newSecretStore(),
		flagSet:     parsedFlags.FlagSet,
		flagValues:  parsedFlags.Values,
		fileConfig:  ctx.GlobalConfig.fileConfig,
		commands:    ctx.GlobalConfig.commands,
		processed:   false,
	}

	// Process the command-specific configuration using parsed flags
	// We need to manually process the definitions since we can't call tempConfig.Process()
	// as it would re-parse global flags instead of command-specific flags
	var configErrs []ConfigError

	for key, def := range cmd.Definitions {
		var value any
		var err error
		source := "none"
		rawValue := ""

		// Check flag value from parsed flags
		if flagVal, exists := parsedFlags.Values[key]; exists && flagVal != nil && *flagVal != "" {
			source = "flag"
			rawValue = *flagVal
			value, err = parseValue(*flagVal, def.valueType, ",")
			if err != nil {
				configErrs = append(configErrs, newParseConfigError(key, def, source, rawValue, err))
				continue
			}
		} else {
			// Check environment variable
			if def.envVar != "" {
				if envVal := os.Getenv(def.envVar); envVal != "" {
					source = "env"
					rawValue = envVal
					value, err = parseValue(envVal, def.valueType, ",")
					if err != nil {
						configErrs = append(configErrs, newParseConfigError(key, def, source, rawValue, err))
						continue
					}
				}
			}

			// Check default value if no flag or env value
			if value == nil && def.defaultValue != nil {
				source = "default"
				rawValue = fmt.Sprintf("%v", def.defaultValue)
				value = def.defaultValue
			}
		}

		// Check if required field is missing
		if def.required && value == nil {
			configErrs = append(configErrs, newRequiredConfigError(key, def))
			continue
		}

		if value != nil {
			validationFailed := false
			for _, validation := range def.validations {
				if source == "default" && validation.Name == "required" {
					continue
				}
				if err := validation.Check(value); err != nil {
					configErrs = append(configErrs, newValidationConfigError(key, def, source, rawValue, value, validation.Name, err))
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

	// Return errors if any occurred
	if len(configErrs) > 0 {
		// Collect errors in execution context for display
		for _, configErr := range configErrs {
			ctx.execution.CollectConfigError(tempConfig, configErr)
		}

		return ErrorWithMessage(fmt.Errorf("configuration errors detected"), ctx.execution.GetFormattedErrors())
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
