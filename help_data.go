// commandkit/help_data.go
package commandkit

import (
	"fmt"
	"sort"
)

// UnifiedHelpData represents the unified help data structure
type UnifiedHelpData struct {
	Command     *Command
	Usage       string
	Description string
	Flags       []FlagInfo // All flags (unified display format)
	EnvVars     []FlagInfo // Filtered based on full/basic mode
	Subcommands []SubcommandInfo
	Errors      []GetError // Optional error list
	Mode        HelpMode   // Essential or Full
	HasErrors   bool
}

// GetDisplayLine returns the appropriate display line based on flag type
func (fi *FlagInfo) GetDisplayLine() string {
	if fi.NoFlag {
		return fi.EnvVarDisplay // Use new layout for env-only vars
	}
	return fi.DisplayLine // Use traditional format for flags
}

// UnifiedExtractor extracts and processes help data
type UnifiedExtractor struct{}

// NewUnifiedExtractor creates a new unified extractor
func NewUnifiedExtractor() *UnifiedExtractor {
	return &UnifiedExtractor{}
}

// ExtractHelpData extracts unified help data for a command
func (ue *UnifiedExtractor) ExtractHelpData(cmd *Command, mode HelpMode, errors []GetError) *UnifiedHelpData {
	if cmd == nil {
		return &UnifiedHelpData{}
	}

	// Extract flags and environment variables separately
	flags := ue.ExtractFlags(cmd.Definitions)
	envVars := ue.ExtractEnvVars(cmd.Definitions, mode)

	// Filter environment variables based on mode
	filteredEnvVars := ue.FilterEnvVars(envVars, mode)

	// Extract subcommands
	subcommands := ue.ExtractSubcommands(cmd)

	// Build usage string
	usage := ue.buildUsage(cmd)

	return &UnifiedHelpData{
		Command:     cmd,
		Usage:       usage,
		Description: cmd.LongHelp,
		Flags:       flags,
		EnvVars:     filteredEnvVars,
		Subcommands: subcommands,
		Errors:      errors,
		Mode:        mode,
		HasErrors:   len(errors) > 0,
	}
}

// ExtractFlags extracts flag information from definitions
func (ue *UnifiedExtractor) ExtractFlags(defs map[string]*Definition) []FlagInfo {
	var flags []FlagInfo

	// Sort definitions for consistent display
	keys := make([]string, 0, len(defs))
	for key := range defs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		def := defs[key]

		// Only include actual flags (those with a flag name), not environment-only variables
		if def.flag == "" {
			continue // Skip environment-only variables
		}

		flag := FlagInfo{
			Key:         key,
			Name:        def.flag,
			Description: def.description,
			Type:        def.valueType.String(),
			Required:    def.required,
			Default:     def.defaultValue,
			EnvVar:      def.envVar,
			Secret:      def.secret,
			NoFlag:      def.flag == "",
			Validations: ue.extractValidations(def.validations),
		}

		// Set DisplayLine for flags
		flag.DisplayLine = buildFlagDisplay(def)

		// Also set EnvVarDisplay if it has an environment variable
		if def.envVar != "" {
			flag.EnvVarDisplay = buildEnvVarDisplay(def)
		}

		// Mask secret defaults
		if def.secret && flag.Default != "" {
			flag.Default = "[secret]"
		}

		flags = append(flags, flag)
	}

	return flags
}

// ExtractEnvVars extracts environment variables from definitions
func (ue *UnifiedExtractor) ExtractEnvVars(defs map[string]*Definition, mode HelpMode) []FlagInfo {
	var envVars []FlagInfo

	// Sort definitions for consistent display
	keys := make([]string, 0, len(defs))
	for key := range defs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		def := defs[key]

		// Only include environment variables (those with an env var)
		if def.envVar == "" {
			continue // Skip non-environment variables
		}

		envVar := FlagInfo{
			Key:         key,
			Name:        def.flag,
			Description: def.description,
			Type:        def.valueType.String(),
			Required:    def.required,
			Default:     def.defaultValue,
			EnvVar:      def.envVar,
			Secret:      def.secret,
			NoFlag:      def.flag == "",
			Validations: ue.extractValidations(def.validations),
		}

		// Set DisplayLine for environment variables
		envVar.DisplayLine = buildEnvVarDisplay(def)
		envVar.EnvVarDisplay = buildEnvVarDisplay(def)

		// Mask secret defaults
		if def.secret && envVar.Default != "" {
			envVar.Default = "[secret]"
		}

		envVars = append(envVars, envVar)
	}

	return envVars
}

// FilterEnvVars filters environment variables based on help mode
func (ue *UnifiedExtractor) FilterEnvVars(flags []FlagInfo, mode HelpMode) []FlagInfo {
	var envVars []FlagInfo

	for _, flag := range flags {
		// Only include environment-only variables
		if flag.NoFlag && flag.EnvVar != "" {
			// In essential mode, only include required env vars
			if mode == HelpModeEssential && flag.Required {
				envVars = append(envVars, flag)
			} else if mode == HelpModeFull {
				// In full mode, include all env vars
				envVars = append(envVars, flag)
			}
		}
	}

	return envVars
}

// ExtractSubcommands extracts subcommand information
func (ue *UnifiedExtractor) ExtractSubcommands(cmd *Command) []SubcommandInfo {
	if cmd == nil || len(cmd.SubCommands) == 0 {
		return []SubcommandInfo{}
	}

	var subcommands []SubcommandInfo

	// Sort subcommands for consistent display
	names := make([]string, 0, len(cmd.SubCommands))
	for name := range cmd.SubCommands {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		subCmd := cmd.SubCommands[name]
		subcommands = append(subcommands, SubcommandInfo{
			Name:        name,
			Description: subCmd.LongHelp,
			Aliases:     subCmd.Aliases,
		})
	}

	return subcommands
}

// buildUsage builds the usage string for a command
func (ue *UnifiedExtractor) buildUsage(cmd *Command) string {
	if cmd == nil {
		return ""
	}

	if cmd.Name == "" {
		return "Usage: [options]"
	}

	return fmt.Sprintf("Usage: %s [options]", cmd.Name)
}

// extractValidations extracts validation information
func (ue *UnifiedExtractor) extractValidations(validations []Validation) []string {
	var result []string

	// For now, skip complex validation extraction to avoid type issues
	// This can be enhanced later once the validation types are properly exposed
	if len(validations) > 0 {
		// Add generic validation info
		result = append(result, "validation")
	}

	return result
}

// ExtractGlobalCommands extracts command summaries for global help
func (ue *UnifiedExtractor) ExtractGlobalCommands(commands map[string]*Command) []CommandSummary {
	var summaries []CommandSummary

	// Sort commands for consistent display
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		cmd := commands[name]
		summaries = append(summaries, CommandSummary{
			Name:        name,
			Description: cmd.LongHelp,
			Aliases:     cmd.Aliases,
		})
	}

	return summaries
}
