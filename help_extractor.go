// commandkit/help_extractor.go
package commandkit

import (
	"fmt"
	"sort"
	"strings"
)

// HelpExtractor extracts help information from commands and definitions
type HelpExtractor interface {
	ExtractGlobalSummary(commands map[string]*Command) []CommandSummary
	ExtractCommandInfo(cmd *Command, executable string) *CommandHelp
	ExtractFlagInfo(defs map[string]*Definition) []FlagInfo
	ExtractFlagInfoFiltered(defs map[string]*Definition, mode HelpMode) (*CommandHelp, error)
	ExtractSubcommandInfo(subcommands map[string]*Command) []SubcommandInfo
}

// helpExtractor implements HelpExtractor
type helpExtractor struct{}

// NewHelpExtractor creates a new help extractor
func NewHelpExtractor() HelpExtractor {
	return &helpExtractor{}
}

// ExtractGlobalSummary extracts command summaries for global help
func (he *helpExtractor) ExtractGlobalSummary(commands map[string]*Command) []CommandSummary {
	var summaries []CommandSummary

	// Sort commands for consistent display
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		cmd := commands[name]
		summary := CommandSummary{
			Name:        name,
			Description: cmd.ShortHelp,
			Aliases:     cmd.Aliases,
		}
		summaries = append(summaries, summary)
	}

	return summaries
}

// ExtractCommandInfo extracts detailed information for command help
func (he *helpExtractor) ExtractCommandInfo(cmd *Command, executable string) *CommandHelp {
	// Build usage string
	usage := fmt.Sprintf("%s %s [options]", executable, cmd.Name)

	// Build description
	description := cmd.LongHelp
	if description == "" {
		description = cmd.ShortHelp
	}

	// Extract flag information
	flags := he.ExtractFlagInfo(cmd.Definitions)

	// Extract subcommand information
	subcommands := he.ExtractSubcommandInfo(cmd.SubCommands)

	return &CommandHelp{
		Command:     cmd,
		Usage:       usage,
		Description: description,
		Flags:       flags,
		Subcommands: subcommands,
	}
}

// ExtractFlagInfo extracts information about flags/definitions
func (he *helpExtractor) ExtractFlagInfo(defs map[string]*Definition) []FlagInfo {
	var flags []FlagInfo

	// Sort definitions for consistent display
	keys := make([]string, 0, len(defs))
	for key := range defs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		def := defs[key]

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
		}

		// Set DisplayLine based on whether it's a flag or environment-only variable
		if def.flag != "" {
			// This is a flag - use traditional flag display
			flag.DisplayLine = buildFlagDisplay(def)
		} else if def.envVar != "" {
			// This is environment-only - use new layout format
			flag.DisplayLine = buildEnvVarDisplay(def)
			flag.EnvVarDisplay = buildEnvVarDisplay(def)
		} else {
			// Fallback - shouldn't normally happen
			flag.DisplayLine = buildFlagDisplay(def)
		}

		// Extract validations
		flag.Validations = he.extractValidations(def.validations)

		// Set EnvVarDisplay for environment-only variables (redundant but safe)
		if def.flag == "" && def.envVar != "" {
			flag.EnvVarDisplay = buildEnvVarDisplay(def)
		}

		// Mask secret defaults
		if flag.Secret && flag.Default != nil {
			flag.Default = "[hidden]"
		}

		// Handle flag name for environment-only configurations
		if flag.NoFlag && flag.Name == "" {
			flag.Name = key
		}

		flags = append(flags, flag)
	}

	return flags
}

// ExtractSubcommandInfo extracts information about subcommands
func (he *helpExtractor) ExtractSubcommandInfo(subcommands map[string]*Command) []SubcommandInfo {
	var subcommandInfo []SubcommandInfo

	// Sort subcommands for consistent display
	names := make([]string, 0, len(subcommands))
	for name := range subcommands {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		subCmd := subcommands[name]
		info := SubcommandInfo{
			Name:        name,
			Description: subCmd.ShortHelp,
			Aliases:     subCmd.Aliases,
		}
		subcommandInfo = append(subcommandInfo, info)
	}

	return subcommandInfo
}

// extractValidations extracts validation information from definitions
func (he *helpExtractor) extractValidations(validations []Validation) []string {
	var result []string
	var minVal, maxVal string

	for _, validation := range validations {
		switch {
		case validation.Name == "required":
			// Skip required as it's handled separately
			continue
		case strings.HasPrefix(validation.Name, "min("):
			minVal = he.extractValue(validation.Name, "min(")
		case strings.HasPrefix(validation.Name, "max("):
			maxVal = he.extractValue(validation.Name, "max(")
		case strings.HasPrefix(validation.Name, "oneOf("):
			// Extract values from oneOf(format)
			values := he.extractOneOfValues(validation.Name)
			result = append(result, fmt.Sprintf("oneOf: %s", values))
		case strings.HasPrefix(validation.Name, "minLength("):
			min := he.extractValue(validation.Name, "minLength(")
			result = append(result, fmt.Sprintf("minLength: %s", min))
		case strings.HasPrefix(validation.Name, "maxLength("):
			max := he.extractValue(validation.Name, "maxLength(")
			result = append(result, fmt.Sprintf("maxLength: %s", max))
		case strings.HasPrefix(validation.Name, "regexp("):
			pattern := he.extractValue(validation.Name, "regexp(")
			result = append(result, fmt.Sprintf("pattern: %s", pattern))
		default:
			// For other validations, use the name as-is
			result = append(result, validation.Name)
		}
	}

	// Handle min/max range
	if minVal != "" && maxVal != "" {
		result = append([]string{fmt.Sprintf("valid: %s-%s", minVal, maxVal)}, result...)
	} else if minVal != "" {
		result = append([]string{fmt.Sprintf("min: %s", minVal)}, result...)
	} else if maxVal != "" {
		result = append([]string{fmt.Sprintf("max: %s", maxVal)}, result...)
	}

	return result
}

// extractValue extracts numeric value from validation name like "min(8080)"
func (he *helpExtractor) extractValue(name, prefix string) string {
	start := strings.Index(name, prefix)
	if start == -1 {
		return ""
	}
	start += len(prefix)
	end := strings.Index(name[start:], ")")
	if end == -1 {
		return name[start:]
	}
	return name[start : start+end]
}

// extractOneOfValues extracts values from oneOf(['a', 'b', 'c']) format
func (he *helpExtractor) extractOneOfValues(name string) string {
	start := strings.Index(name, "oneOf(")
	if start == -1 {
		return ""
	}
	start += 6 // len("oneOf(")
	end := strings.Index(name[start:], ")")
	if end == -1 {
		return name[start:]
	}

	values := name[start : start+end]
	// Remove brackets and quotes, clean up spacing
	values = strings.ReplaceAll(values, "[", "")
	values = strings.ReplaceAll(values, "]", "")
	values = strings.ReplaceAll(values, "'", "")
	values = strings.ReplaceAll(values, `"`, "")

	return strings.TrimSpace(values)
}

// ExtractFlagInfoFiltered extracts flag information with filtering for essential vs full help
func (he *helpExtractor) ExtractFlagInfoFiltered(defs map[string]*Definition, mode HelpMode) (*CommandHelp, error) {
	var allFlags []FlagInfo
	var requiredEnvVars []FlagInfo
	var allEnvVars []FlagInfo

	// Sort definitions for consistent display
	keys := make([]string, 0, len(defs))
	for key := range defs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		def := defs[key]

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
		}

		// Set DisplayLine based on whether it's a flag or environment-only variable
		if def.flag != "" {
			// This is a flag - use traditional flag display
			flag.DisplayLine = buildFlagDisplay(def)
		} else if def.envVar != "" {
			// This is environment-only - use new layout format
			flag.DisplayLine = buildEnvVarDisplay(def)
			flag.EnvVarDisplay = buildEnvVarDisplay(def)
		} else {
			// Fallback - shouldn't normally happen
			flag.DisplayLine = buildFlagDisplay(def)
		}

		// Extract validations
		flag.Validations = he.extractValidations(def.validations)

		// Set EnvVarDisplay for environment-only variables
		if def.flag == "" && def.envVar != "" {
			flag.EnvVarDisplay = buildEnvVarDisplay(def)
		}

		// Mask secret defaults
		if flag.Secret && flag.Default != nil {
			flag.Default = "[hidden]"
		}

		// Handle flag name for environment-only configurations
		if flag.NoFlag && flag.Name == "" {
			flag.Name = key
		}

		// Separate flags from environment-only variables
		if def.flag != "" {
			// This is a flag (may have associated env var)
			allFlags = append(allFlags, flag)
		} else if def.envVar != "" {
			// This is an environment-only variable
			allEnvVars = append(allEnvVars, flag)
			if def.required {
				requiredEnvVars = append(requiredEnvVars, flag)
			}
		}
	}

	return &CommandHelp{
		Flags:           allFlags,
		RequiredEnvVars: requiredEnvVars,
		AllEnvVars:      allEnvVars,
	}, nil
}
