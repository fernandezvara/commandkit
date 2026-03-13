// commandkit/help_data.go
package commandkit

import (
	"fmt"
	"sort"
	"strings"
)

// UnifiedHelpData represents the complete help data for a command
type UnifiedHelpData struct {
	Command     *Command
	Usage       string
	Description string
	Flags       []flagInfo
	EnvVars     []flagInfo
	Subcommands []subcommandInfo
	Errors      []GetError
	Mode        helpMode
	HasErrors   bool
}

// unifiedExtractor extracts and processes help data
type unifiedExtractor struct{}

// NewUnifiedExtractor creates a new unified extractor
func newUnifiedExtractor() *unifiedExtractor {
	return &unifiedExtractor{}
}

// ExtractHelpData extracts unified help data for a command
func (ue *unifiedExtractor) ExtractHelpData(cmd *Command, mode helpMode, errors []GetError) *UnifiedHelpData {
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

	// Use LongHelp if available, fall back to ShortHelp
	description := cmd.LongHelp
	if description == "" {
		description = cmd.ShortHelp
	}

	return &UnifiedHelpData{
		Command:     cmd,
		Usage:       usage,
		Description: description,
		Flags:       flags,
		EnvVars:     filteredEnvVars,
		Subcommands: subcommands,
		Errors:      errors,
		Mode:        mode,
		HasErrors:   len(errors) > 0,
	}
}

// ExtractFlags extracts flag information from definitions
func (ue *unifiedExtractor) ExtractFlags(defs map[string]*Definition) []flagInfo {
	var flags []flagInfo

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

		flag := flagInfo{
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

		flags = append(flags, flag)
	}

	return flags
}

// ExtractEnvVars extracts environment variable information from definitions
func (ue *unifiedExtractor) ExtractEnvVars(defs map[string]*Definition, mode helpMode) []flagInfo {
	var envVars []flagInfo

	// Sort definitions for consistent display
	keys := make([]string, 0, len(defs))
	for key := range defs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		def := defs[key]

		// Include environment variables (those with env var name)
		if def.envVar == "" {
			continue // Skip non-environment variables
		}

		envVar := flagInfo{
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
		envVar.EnvVarDisplay = envVar.DisplayLine

		envVars = append(envVars, envVar)
	}

	return envVars
}

// FilterEnvVars filters environment variables based on help mode
func (ue *unifiedExtractor) FilterEnvVars(envVars []flagInfo, mode helpMode) []flagInfo {
	var result []flagInfo

	for _, envVar := range envVars {
		// Only include items that actually have environment variables
		if envVar.EnvVar == "" {
			continue
		}

		if mode == helpModeFull {
			// Include all environment variables in full mode
			result = append(result, envVar)
		} else {
			// Only include required environment variables in essential mode
			if envVar.Required {
				result = append(result, envVar)
			}
		}
	}
	return result
}

// ExtractSubcommands extracts subcommand information
func (ue *unifiedExtractor) ExtractSubcommands(cmd *Command) []subcommandInfo {
	if cmd == nil || len(cmd.SubCommands) == 0 {
		return []subcommandInfo{}
	}

	var subcommands []subcommandInfo

	// Sort subcommands for consistent display
	names := make([]string, 0, len(cmd.SubCommands))
	for name := range cmd.SubCommands {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		subCmd := cmd.SubCommands[name]
		// Use LongHelp if available, fall back to ShortHelp
		desc := subCmd.LongHelp
		if desc == "" {
			desc = subCmd.ShortHelp
		}
		subcommands = append(subcommands, subcommandInfo{
			Name:        name,
			Description: desc,
			Aliases:     subCmd.Aliases,
		})
	}

	return subcommands
}

// buildUsage builds the usage string for a command
func (ue *unifiedExtractor) buildUsage(cmd *Command) string {
	if cmd == nil {
		return ""
	}

	if cmd.Name == "" {
		return "Usage: [options]"
	}

	return fmt.Sprintf("Usage: %s [options]", cmd.Name)
}

// extractValidations extracts validation descriptions
func (ue *unifiedExtractor) extractValidations(validations []Validation) []string {
	var descriptions []string
	for _, validation := range validations {
		descriptions = append(descriptions, validation.Name)
	}
	return descriptions
}

// --- Data extraction methods for layer coordinator ---

// extractUsageData extracts usage layer data
func (ue *unifiedExtractor) extractUsageData(command, subcommand, executable string) *usageData {
	return &usageData{
		command:    command,
		subcommand: subcommand,
		executable: executable,
	}
}

// extractCommandsData extracts commands layer data
func (ue *unifiedExtractor) extractCommandsData(commands map[string]*Command, executable string) *commandsData {
	var commandSummaries []commandSummary
	for name, cmd := range commands {
		if name != "" { // Skip empty string command
			// Use LongHelp if available, fall back to ShortHelp
			description := cmd.LongHelp
			if description == "" {
				description = cmd.ShortHelp
			}
			// Get first line for summary
			lines := strings.Split(description, "\n")
			if len(lines) > 0 {
				description = lines[0]
			}

			commandSummaries = append(commandSummaries, commandSummary{
				Name:        name,
				Description: description,
				Aliases:     cmd.Aliases,
			})
		}
	}

	return &commandsData{
		commands:   commandSummaries,
		executable: executable,
	}
}

// extractFlagsData extracts flags layer data
func (ue *unifiedExtractor) extractFlagsData(cmd *Command) *flagsData {
	if cmd == nil {
		return &flagsData{}
	}

	flags := ue.ExtractFlags(cmd.Definitions)
	return &flagsData{
		flags: flags,
	}
}

// extractEnvVarsData extracts environment variables layer data
func (ue *unifiedExtractor) extractEnvVarsData(cmd *Command, mode helpMode) *envVarsData {
	if cmd == nil {
		return &envVarsData{}
	}

	envVars := ue.ExtractEnvVars(cmd.Definitions, mode)
	filteredEnvVars := ue.FilterEnvVars(envVars, mode)

	return &envVarsData{
		envVars: filteredEnvVars,
		mode:    mode,
	}
}

// extractSubcommandsData extracts subcommands layer data
func (ue *unifiedExtractor) extractSubcommandsData(cmd *Command) *subcommandsData {
	if cmd == nil {
		return &subcommandsData{}
	}

	subcommands := ue.ExtractSubcommands(cmd)
	return &subcommandsData{
		subcommands: subcommands,
	}
}

// extractErrorsData extracts errors layer data
func (ue *unifiedExtractor) extractErrorsData(errors []GetError) *errorsData {
	return &errorsData{
		errors: errors,
	}
}
