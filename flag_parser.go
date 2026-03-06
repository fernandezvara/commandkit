// commandkit/flag_parser.go
package commandkit

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// FlagParser provides centralized flag parsing functionality
type FlagParser interface {
	// ParseCommand parses flags for command-specific configuration
	ParseCommand(args []string, defs map[string]*Definition) (*ParsedFlags, error)

	// ParseGlobal parses flags for global configuration
	ParseGlobal(args []string, defs map[string]*Definition) (*ParsedFlags, error)

	// GenerateHelp generates consistent help text for flags
	GenerateHelp(defs map[string]*Definition) string
}

// ParsedFlags contains the results of flag parsing
type ParsedFlags struct {
	Values  map[string]*string // Parsed flag values
	FlagSet *flag.FlagSet      // The actual FlagSet used
	Errors  []error            // Any parsing errors encountered
	Args    []string           // Remaining arguments after flag parsing
}

// flagParser implements FlagParser interface
type flagParser struct{}

// NewFlagParser creates a new FlagParser instance
func NewFlagParser() FlagParser {
	return &flagParser{}
}

// ParseCommand parses flags for command-specific configuration
func (fp *flagParser) ParseCommand(args []string, defs map[string]*Definition) (*ParsedFlags, error) {
	return fp.parseFlags(args, defs, "")
}

// ParseGlobal parses flags for global configuration
func (fp *flagParser) ParseGlobal(args []string, defs map[string]*Definition) (*ParsedFlags, error) {
	// For global parsing, use the executable name as the FlagSet name
	executable := os.Args[0]
	if executable == "" {
		executable = "command"
	}

	// Filter out test flags that might interfere (from original config.go logic)
	filteredArgs := make([]string, 0)
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-test.") {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	return fp.parseFlags(filteredArgs, defs, executable)
}

// parseFlags is the core flag parsing implementation
func (fp *flagParser) parseFlags(args []string, defs map[string]*Definition, flagSetName string) (*ParsedFlags, error) {
	// Create FlagSet with ContinueOnError to collect errors instead of exiting
	flagSet := flag.NewFlagSet(flagSetName, flag.ContinueOnError)

	// Create values map and register flags with correct types
	values := make(map[string]*string)
	for key, def := range defs {
		if def.flag != "" {
			// Use string values for all flags to maintain consistency
			// The type conversion will happen during config processing
			values[key] = flagSet.String(def.flag, "", def.description)
		}
	}

	// Parse flags and collect any errors
	err := flagSet.Parse(args)

	// Create ParsedFlags result
	result := &ParsedFlags{
		Values:  values,
		FlagSet: flagSet,
		Args:    flagSet.Args(),
	}

	// Collect parsing errors
	if err != nil {
		result.Errors = []error{err}
	}

	return result, nil
}

// GenerateHelp generates consistent help text for flags
func (fp *flagParser) GenerateHelp(defs map[string]*Definition) string {
	var sb strings.Builder

	// Create a temporary FlagSet for help generation
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)

	// Register flags with enhanced descriptions
	for _, def := range defs {
		if def.flag != "" {
			enhancedDescription := fp.generateEnhancedDescription(def)
			flagSet.String(def.flag, "", enhancedDescription)
		}
	}

	// Track environment-only configurations (no flag)
	var envOnlyConfigs []*Definition
	for _, def := range defs {
		if def.flag == "" && def.envVar != "" {
			envOnlyConfigs = append(envOnlyConfigs, def)
		}
	}

	// Print flag help to the string builder
	flagSet.SetOutput(&sb)
	flagSet.PrintDefaults()

	// Print environment-only configurations if any exist
	if len(envOnlyConfigs) > 0 {
		sb.WriteString("\n")
		for _, def := range envOnlyConfigs {
			enhancedDescription := fp.generateEnhancedDescription(def)
			sb.WriteString(fmt.Sprintf("  (no flag) string %s\n", enhancedDescription))
			sb.WriteString(fmt.Sprintf("        %s\n", def.description))
		}
	}

	return sb.String()
}

// generateEnhancedDescription creates the enhanced description with indicators
func (fp *flagParser) generateEnhancedDescription(def *Definition) string {
	var indicators []string

	// 1. Environment variable context
	if def.envVar != "" {
		indicators = append(indicators, fmt.Sprintf("env: %s", def.envVar))
	}

	// 2. Required indicator
	if def.required {
		indicators = append(indicators, "required")
	}

	// 3. Default value (masked for secrets)
	if def.defaultValue != nil {
		if def.secret {
			indicators = append(indicators, "default: '[hidden]'")
		} else if def.valueType == TypeString {
			indicators = append(indicators, fmt.Sprintf("default: '%v'", def.defaultValue))
		} else {
			indicators = append(indicators, fmt.Sprintf("default: %v", def.defaultValue))
		}
	}

	// 4. Validations
	validations := formatValidation(def.validations)
	indicators = append(indicators, validations...)

	// 5. Secret indicator
	if def.secret {
		indicators = append(indicators, "secret")
	}

	// Combine description with indicators
	if len(indicators) > 0 {
		return fmt.Sprintf("%s (%s)", def.description, strings.Join(indicators, ", "))
	}

	return def.description
}
