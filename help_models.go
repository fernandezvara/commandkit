// commandkit/help_models.go
package commandkit

import "slices"

// helpType represents the type of help request
type helpType int

const (
	helpTypeNone helpType = iota
	helpTypeGlobal
	helpTypeCommand
	helpTypeSubcommand
)

// helpMode represents the help mode (currently essential vs full)
type helpMode int

const (
	helpModeEssential helpMode = iota
	helpModeFull
)

// commandSummary represents a command summary for global help
type commandSummary struct {
	Name        string
	Description string
	Aliases     []string
}

// flagInfo represents flag information for help display
type flagInfo struct {
	Key           string
	Name          string
	Description   string
	Type          string
	Required      bool
	Default       any
	EnvVar        string
	Validations   []string
	Secret        bool
	NoFlag        bool
	DisplayLine   string
	EnvVarDisplay string
}

// subcommandInfo represents subcommand information for help display
type subcommandInfo struct {
	Name        string
	Description string
	Aliases     []string
}

// --- Centralized help flag detection functions ---

// isHelpFlag returns true if the given argument is any help flag
func isHelpFlag(arg string) bool {
	return arg == "--help" || arg == "-h" || arg == "help" || arg == "--full-help"
}

// isFullHelpFlag returns true if the given argument requests full/extended help
func isFullHelpFlag(arg string) bool {
	return arg == "--full-help"
}

// argsContainHelpFlag returns true if any argument in the slice is a help flag
func argsContainHelpFlag(args []string) bool {
	return slices.ContainsFunc(args, isHelpFlag)
}

// argsContainFullHelp returns true if any argument requests full help
func argsContainFullHelp(args []string) bool {
	return slices.ContainsFunc(args, isFullHelpFlag)
}

// lastArgIsHelpFlag returns true if the last argument is a help flag
func lastArgIsHelpFlag(args []string) bool {
	if len(args) == 0 {
		return false
	}
	return isHelpFlag(args[len(args)-1])
}

// helpModeFromArgs returns helpModeFull if --full-help is present, otherwise helpModeEssential
func helpModeFromArgs(args []string) helpMode {
	if argsContainFullHelp(args) {
		return helpModeFull
	}
	return helpModeEssential
}

// usageData represents data for usage layer rendering
type usageData struct {
	command    string
	subcommand string
	executable string
}

// commandsData represents data for commands layer rendering
type commandsData struct {
	commands   []commandSummary
	executable string
}

// flagsData represents data for flags layer rendering
type flagsData struct {
	flags []flagInfo
}

// envVarsData represents data for environment variables layer rendering
type envVarsData struct {
	envVars []flagInfo
	mode    helpMode // essential vs full
}

// subcommandsData represents data for subcommands layer rendering
type subcommandsData struct {
	subcommands []subcommandInfo
}

// errorsData represents data for errors layer rendering
type errorsData struct {
	errors []GetError
}
