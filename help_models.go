// commandkit/help_models.go
package commandkit

// HelpType represents the type of help request
type HelpType int

const (
	HelpTypeNone HelpType = iota
	HelpTypeGlobal
	HelpTypeCommand
	HelpTypeSubcommand
)

// HelpMode represents the help mode (essential vs full)
type HelpMode int

const (
	HelpModeEssential HelpMode = iota
	HelpModeFull
)

// GlobalHelp represents help for all commands
type GlobalHelp struct {
	Executable string
	Commands   []CommandSummary
	Template   string
}

// CommandSummary represents a brief command summary for global help
type CommandSummary struct {
	Name        string
	Description string
	Aliases     []string
}

// CommandHelp represents detailed help for a specific command
type CommandHelp struct {
	Command         *Command
	Usage           string
	Description     string
	Flags           []FlagInfo // All flags (including those with env vars)
	RequiredEnvVars []FlagInfo // Only required env-only vars (essential mode)
	AllEnvVars      []FlagInfo // All env-only vars (full mode)
	Subcommands     []SubcommandInfo
	Template        string
	// NEW: Error information
	Errors    []GetError
	HasErrors bool
}

// FlagInfo represents information about a configuration flag
type FlagInfo struct {
	Key           string
	Name          string
	DisplayLine   string // For flags: "--flag type (default: value)"
	EnvVarDisplay string // For env-only vars: "ENV_VAR type (attributes)"
	Description   string
	Type          string
	Required      bool
	Default       interface{}
	EnvVar        string
	Validations   []string
	Secret        bool
	NoFlag        bool // Environment-only configuration
	// NEW: Error information
	HasError bool
	ErrorMsg string
}

// SubcommandInfo represents information about a subcommand
type SubcommandInfo struct {
	Name        string
	Description string
	Aliases     []string
}

// SubcommandHelp represents help for subcommands of a command
type SubcommandHelp struct {
	Parent      string
	Subcommands []SubcommandInfo
	Template    string
}

// FlagHelp represents help for flags/definitions
type FlagHelp struct {
	Command  string
	Flags    []FlagInfo
	Template string
}

// HelpRequest represents a parsed help request
type HelpRequest struct {
	Type       HelpType
	Mode       HelpMode
	Command    string
	Subcommand string
	Args       []string
	Original   []string
}
