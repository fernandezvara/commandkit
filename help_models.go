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
	Command     *Command
	Usage       string
	Description string
	Flags       []FlagInfo
	Subcommands []SubcommandInfo
	Template    string
}

// FlagInfo represents information about a configuration flag
type FlagInfo struct {
	Name         string
	Description  string
	Type         string
	Required     bool
	Default      interface{}
	EnvVar       string
	Validations  []string
	Secret       bool
	NoFlag       bool // Environment-only configuration
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
	Command    string
	Flags      []FlagInfo
	Template   string
}

// HelpRequest represents a parsed help request
type HelpRequest struct {
	Type        HelpType
	Command     string
	Subcommand  string
	Args        []string
	Original    []string
}
