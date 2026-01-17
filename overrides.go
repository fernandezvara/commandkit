// commandkit/overrides.go
package commandkit

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// OverrideWarning represents a configuration override warning
type OverrideWarning struct {
	Key        string // Configuration key
	Command    string // Command name (if command-specific)
	Source     string // Source being overridden (global, flag, env, etc.)
	OverrideBy string // Source doing the overriding
	OldValue   string // Previous value (masked if secret)
	NewValue   string // New value (masked if secret)
	Message    string // Warning message
}

// OverrideWarnings holds all override warnings
type OverrideWarnings struct {
	warnings []OverrideWarning
}

// NewOverrideWarnings creates a new warnings collector
func NewOverrideWarnings() *OverrideWarnings {
	return &OverrideWarnings{
		warnings: make([]OverrideWarning, 0),
	}
}

// Add adds a new override warning
func (ow *OverrideWarnings) Add(warning OverrideWarning) {
	ow.warnings = append(ow.warnings, warning)
}

// HasWarnings returns true if there are warnings
func (ow *OverrideWarnings) HasWarnings() bool {
	return len(ow.warnings) > 0
}

// GetWarnings returns all warnings
func (ow *OverrideWarnings) GetWarnings() []OverrideWarning {
	return ow.warnings
}

// FormatWarnings formats all warnings for display
func (ow *OverrideWarnings) FormatWarnings() string {
	if len(ow.warnings) == 0 {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("Warning: Configuration overrides detected\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n")

	for i, warning := range ow.warnings {
		// Key and command
		if warning.Command != "" {
			sb.WriteString(fmt.Sprintf("%s (command: %s)\n", warning.Key, warning.Command))
		} else {
			sb.WriteString(fmt.Sprintf("%s\n", warning.Key))
		}

		// Override information
		overrideInfo := fmt.Sprintf("  %s -> %s", warning.Source, warning.OverrideBy)
		if warning.OldValue != "" || warning.NewValue != "" {
			overrideInfo += fmt.Sprintf(" (%s -> %s)", warning.OldValue, warning.NewValue)
		}
		sb.WriteString(fmt.Sprintf("%s\n", overrideInfo))

		// Message
		if warning.Message != "" {
			sb.WriteString(fmt.Sprintf("  Note: %s\n", warning.Message))
		}

		// Separator between warnings
		if i < len(ow.warnings)-1 {
			sb.WriteString("\n")
		}
	}

	sb.WriteString(strings.Repeat("=", 50) + "\n")
	sb.WriteString(fmt.Sprintf("Total: %d override(s)\n", len(ow.warnings)))

	return sb.String()
}

// LogWarnings logs all warnings
func (ow *OverrideWarnings) LogWarnings() {
	if ow.HasWarnings() {
		log.Printf("Configuration overrides detected:\n%s", ow.FormatWarnings())
	}
}

// checkCommandOverrides checks for command-specific config overriding global config
func (c *Config) checkCommandOverrides(commandName string, commandDefs map[string]*Definition) *OverrideWarnings {
	warnings := NewOverrideWarnings()

	for key, cmdDef := range commandDefs {
		// Check if this key exists in global config
		if globalDef, exists := c.definitions[key]; exists {
			// Check if command definition overrides global definition
			if c.shouldWarnAboutOverride(globalDef, cmdDef) {
				warning := OverrideWarning{
					Key:        key,
					Command:    commandName,
					Source:     "global config",
					OverrideBy: "command config",
					Message:    "Command-specific configuration overrides global configuration",
				}

				// Add value information if available
				if c.Has(key) {
					if c.IsSecret(key) {
						secret := c.GetSecret(key)
						if secret.IsSet() {
							warning.OldValue = fmt.Sprintf("[SECRET:%d bytes]", secret.Size())
						} else {
							warning.OldValue = "[SECRET:not set]"
						}
					} else {
						warning.OldValue = fmt.Sprintf("%v", Get[any](c, key))
					}
				}

				warnings.Add(warning)
			}
		}
	}

	return warnings
}

// shouldWarnAboutOverride determines if an override should generate a warning
func (c *Config) shouldWarnAboutOverride(globalDef, cmdDef *Definition) bool {
	// Don't warn if the definitions are identical
	if c.definitionsEqual(globalDef, cmdDef) {
		return false
	}

	// Warn if command has different flag, env var, or default value
	if globalDef.flag != cmdDef.flag && cmdDef.flag != "" {
		return true
	}

	if globalDef.envVar != cmdDef.envVar && cmdDef.envVar != "" {
		return true
	}

	// Check if defaults are different
	globalHasDefault := globalDef.defaultValue != nil
	cmdHasDefault := cmdDef.defaultValue != nil

	if globalHasDefault && cmdHasDefault {
		// Both have defaults, check if they're different
		return fmt.Sprintf("%v", globalDef.defaultValue) != fmt.Sprintf("%v", cmdDef.defaultValue)
	} else if globalHasDefault && !cmdHasDefault {
		// Global has default but command doesn't - this is an override
		return true
	} else if !globalHasDefault && cmdHasDefault {
		// Global doesn't have default but command does - this is an override
		return true
	}

	// Check if command adds validation that global doesn't have
	if len(cmdDef.validations) > len(globalDef.validations) {
		return true
	}

	return false
}

// definitionsEqual checks if two definitions are effectively the same
func (c *Config) definitionsEqual(def1, def2 *Definition) bool {
	if def1.valueType != def2.valueType {
		return false
	}

	if def1.flag != def2.flag {
		return false
	}

	if def1.envVar != def2.envVar {
		return false
	}

	if fmt.Sprintf("%v", def1.defaultValue) != fmt.Sprintf("%v", def2.defaultValue) {
		return false
	}

	if def1.required != def2.required {
		return false
	}

	if def1.secret != def2.secret {
		return false
	}

	return true
}

// checkSourceOverrides checks for higher priority sources overriding lower priority sources
func (c *Config) checkSourceOverrides() *OverrideWarnings {
	warnings := NewOverrideWarnings()

	for key, def := range c.definitions {
		// Check each source in priority order and detect overrides
		c.checkSourceOverridesForKey(key, def, warnings)
	}

	return warnings
}

// checkSourceOverridesForKey checks overrides for a specific configuration key
func (c *Config) checkSourceOverridesForKey(key string, def *Definition, warnings *OverrideWarnings) {
	var flagValue, envValue, defaultValue string
	var hasFlag, hasEnv, hasDefault bool

	// Check each source
	// 1. Command flags (highest priority)
	if def.flag != "" {
		if flagVal, ok := c.flagValues[key]; ok && flagVal != nil && *flagVal != "" {
			flagValue = *flagVal
			hasFlag = true
		}
	}

	// 2. Environment variables
	if def.envVar != "" {
		if envVal := c.getValueFromEnv(def.envVar); envVal != "" {
			envValue = envVal
			hasEnv = true
		}
	}

	// 3. Default values
	if def.defaultValue != nil {
		defaultValue = fmt.Sprintf("%v", def.defaultValue)
		hasDefault = true
	}

	// Check for overrides
	// Flag overrides env
	if hasFlag && hasEnv {
		warnings.Add(OverrideWarning{
			Key:        key,
			Source:     "environment",
			OverrideBy: "flag",
			OldValue:   c.maskValueIfNeeded(key, envValue),
			NewValue:   c.maskValueIfNeeded(key, flagValue),
			Message:    "Command-line flag overrides environment variable",
		})
	}

	// Flag overrides default
	if hasFlag && hasDefault {
		warnings.Add(OverrideWarning{
			Key:        key,
			Source:     "default",
			OverrideBy: "flag",
			OldValue:   c.maskValueIfNeeded(key, defaultValue),
			NewValue:   c.maskValueIfNeeded(key, flagValue),
			Message:    "Command-line flag overrides default value",
		})
	}

	// Env overrides default (only if no flag)
	if hasEnv && hasDefault && !hasFlag {
		warnings.Add(OverrideWarning{
			Key:        key,
			Source:     "default",
			OverrideBy: "environment",
			OldValue:   c.maskValueIfNeeded(key, defaultValue),
			NewValue:   c.maskValueIfNeeded(key, envValue),
			Message:    "Environment variable overrides default value",
		})
	}
}

// getValueFromEnv gets value from environment variable
func (c *Config) getValueFromEnv(envVar string) string {
	return os.Getenv(envVar)
}

// getDefaultValue gets the default value for a key
func (c *Config) getDefaultValue(key string) string {
	if def, exists := c.definitions[key]; exists && def.defaultValue != nil {
		return fmt.Sprintf("%v", def.defaultValue)
	}
	return ""
}

// maskValueIfNeeded masks a value if it's a secret
func (c *Config) maskValueIfNeeded(key, value string) string {
	if def, exists := c.definitions[key]; exists && def.secret {
		return maskSecret(value)
	}
	return value
}
