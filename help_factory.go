// commandkit/help_factory.go
package commandkit

import "fmt"

// HelpFactory creates help objects for different contexts
type HelpFactory interface {
	// Core help creation methods
	CreateGlobalHelp(commands map[string]*Command, executable string) *GlobalHelp
	CreateCommandHelp(cmd *Command, executable string) *CommandHelp
	CreateCommandHelpWithErrors(cmd *Command, executable string, errors []GetError) *CommandHelp
	CreateCommandHelpWithMode(cmd *Command, executable string, mode HelpMode) (*CommandHelp, error)
	CreateSubcommandHelp(parent string, subcommands map[string]*Command) *SubcommandHelp
	CreateFlagHelp(command string, defs map[string]*Definition) *FlagHelp

	// Help detection and parsing
	DetectHelpRequest(args []string) *HelpRequest
	IsHelpRequested(args []string) bool
	GetHelpType(args []string) HelpType
	GetHelpMode(args []string) HelpMode

	// Context-aware help detection methods
	DetectHelpRequestWithContext(args []string, commandPath []string) *HelpRequest

	// Template management
	SetTemplate(templateType TemplateType, template string)
	GetTemplate(templateType TemplateType) string
}

// TemplateType represents different help template types
type TemplateType int

const (
	TemplateGlobal TemplateType = iota
	TemplateCommand
	TemplateCommandError
	TemplateCustomHelp      // For custom help (LongHelp only)
	TemplateCustomHelpError // For custom help with errors (LongHelp + errors)
	TemplateSubcommand
	TemplateFlag
	TemplateEssential      // For essential help (all flags + required env vars)
	TemplateEssentialError // For essential help with errors
	TemplateFull           // For full help (all flags + all env vars)
)

// helpFactory implements HelpFactory
type helpFactory struct {
	detector  HelpDetector
	extractor HelpExtractor
	templates map[TemplateType]string
}

// NewHelpFactory creates a new help factory
func NewHelpFactory() HelpFactory {
	factory := &helpFactory{
		detector:  NewHelpDetector(),
		extractor: NewHelpExtractor(),
		templates: make(map[TemplateType]string),
	}

	// Set default templates
	factory.setDefaultTemplates()

	return factory
}

// DetectHelpRequest detects and parses a help request
func (hf *helpFactory) DetectHelpRequest(args []string) *HelpRequest {
	return hf.detector.ParseHelpRequest(args)
}

// IsHelpRequested checks if help is requested
// Delegates to HelpDetector for centralized help flag detection
func (hf *helpFactory) IsHelpRequested(args []string) bool {
	return hf.detector.IsHelpRequested(args)
}

// GetHelpType gets the type of help request
func (hf *helpFactory) GetHelpType(args []string) HelpType {
	return hf.detector.GetHelpType(args)
}

// GetHelpMode gets the help mode (essential vs full)
func (hf *helpFactory) GetHelpMode(args []string) HelpMode {
	return hf.detector.GetHelpMode(args)
}

// DetectHelpRequestWithContext detects and parses a help request with command path context
func (hf *helpFactory) DetectHelpRequestWithContext(args []string, commandPath []string) *HelpRequest {
	return hf.detector.DetectHelpWithContext(args, commandPath)
}

// CreateGlobalHelp creates global help for all commands
func (hf *helpFactory) CreateGlobalHelp(commands map[string]*Command, executable string) *GlobalHelp {
	summaries := hf.extractor.ExtractGlobalSummary(commands)

	return &GlobalHelp{
		Executable: executable,
		Commands:   summaries,
		Template:   hf.templates[TemplateGlobal],
	}
}

// CreateCommandHelp creates detailed help for a specific command
func (hf *helpFactory) CreateCommandHelp(cmd *Command, executable string) *CommandHelp {
	commandInfo := hf.extractor.ExtractCommandInfo(cmd, executable)

	// Choose template based on customHelp flag
	templateType := TemplateCommand
	if cmd.customHelp {
		templateType = TemplateCustomHelp
	}

	return &CommandHelp{
		Command:     cmd,
		Usage:       commandInfo.Usage,
		Description: commandInfo.Description,
		Flags:       commandInfo.Flags,
		Subcommands: commandInfo.Subcommands,
		Template:    hf.templates[templateType],
	}
}

// CreateCommandHelpWithErrors creates detailed help for a specific command with errors
func (hf *helpFactory) CreateCommandHelpWithErrors(cmd *Command, executable string, errors []GetError) *CommandHelp {
	commandInfo := hf.extractor.ExtractCommandInfo(cmd, executable)

	// Match errors to flags
	flagsWithErrors := hf.matchErrorsToFlags(commandInfo.Flags, errors)
	orderedErrors := hf.orderErrors(flagsWithErrors, errors)

	// Choose template based on customHelp flag
	templateType := TemplateCommandError
	if cmd.customHelp {
		templateType = TemplateCustomHelpError
	}

	return &CommandHelp{
		Command:     cmd,
		Usage:       commandInfo.Usage,
		Description: commandInfo.Description,
		Flags:       flagsWithErrors,
		Subcommands: commandInfo.Subcommands,
		Template:    hf.templates[templateType],
		Errors:      orderedErrors,
		HasErrors:   len(errors) > 0,
	}
}

// CreateCommandHelpWithMode creates command help with filtering mode (essential vs full)
func (hf *helpFactory) CreateCommandHelpWithMode(cmd *Command, executable string, mode HelpMode) (*CommandHelp, error) {
	// Build usage string
	usage := fmt.Sprintf("%s %s [options]", executable, cmd.Name)

	// Build description
	description := cmd.LongHelp
	if description == "" {
		description = cmd.ShortHelp
	}

	// Extract filtered flag information
	filteredHelp, err := hf.extractor.ExtractFlagInfoFiltered(cmd.Definitions, mode)
	if err != nil {
		return nil, err
	}

	// Extract subcommand information
	subcommands := hf.extractor.ExtractSubcommandInfo(cmd.SubCommands)

	// Choose template based on mode and customHelp flag
	var templateType TemplateType
	if cmd.customHelp {
		templateType = TemplateCustomHelp
	} else if mode == HelpModeEssential {
		templateType = TemplateEssential
	} else {
		templateType = TemplateFull
	}

	return &CommandHelp{
		Command:         cmd,
		Usage:           usage,
		Description:     description,
		Flags:           filteredHelp.Flags,
		RequiredEnvVars: filteredHelp.RequiredEnvVars,
		AllEnvVars:      filteredHelp.AllEnvVars,
		Subcommands:     subcommands,
		Template:        hf.templates[templateType],
	}, nil
}

func (hf *helpFactory) orderErrors(flags []FlagInfo, errors []GetError) []GetError {
	ordered := make([]GetError, 0, len(errors))
	used := make(map[string]bool)
	errorMap := make(map[string]GetError)
	for _, err := range errors {
		errorMap[err.Key] = err
	}

	for _, flag := range flags {
		if err, ok := errorMap[flag.Key]; ok {
			ordered = append(ordered, err)
			used[flag.Key] = true
		}
	}

	for _, err := range errors {
		if !used[err.Key] {
			ordered = append(ordered, err)
		}
	}

	return ordered
}

// matchErrorsToFlags matches errors to their corresponding flags
func (hf *helpFactory) matchErrorsToFlags(flags []FlagInfo, errors []GetError) []FlagInfo {
	result := make([]FlagInfo, len(flags))
	copy(result, flags)

	// Create error map for quick lookup
	errorMap := make(map[string]GetError)
	for _, err := range errors {
		errorMap[err.Key] = err
	}

	// Update flags with error information
	for i, flag := range result {
		if err, hasError := errorMap[flag.Key]; hasError {
			result[i].HasError = true
			result[i].ErrorMsg = err.ErrorDescription
		}
	}

	return result
}

// CreateSubcommandHelp creates help for subcommands
func (hf *helpFactory) CreateSubcommandHelp(parent string, subcommands map[string]*Command) *SubcommandHelp {
	subcommandInfo := hf.extractor.ExtractSubcommandInfo(subcommands)

	return &SubcommandHelp{
		Parent:      parent,
		Subcommands: subcommandInfo,
		Template:    hf.templates[TemplateSubcommand],
	}
}

// CreateFlagHelp creates help for flags
func (hf *helpFactory) CreateFlagHelp(command string, defs map[string]*Definition) *FlagHelp {
	flagInfo := hf.extractor.ExtractFlagInfo(defs)

	return &FlagHelp{
		Command:  command,
		Flags:    flagInfo,
		Template: hf.templates[TemplateFlag],
	}
}

// SetTemplate sets a custom template
func (hf *helpFactory) SetTemplate(templateType TemplateType, template string) {
	hf.templates[templateType] = template
}

// GetTemplate gets the current template
func (hf *helpFactory) GetTemplate(templateType TemplateType) string {
	return hf.templates[templateType]
}

// setDefaultTemplates sets the default templates
func (hf *helpFactory) setDefaultTemplates() {
	hf.templates[TemplateGlobal] = DefaultGlobalTemplate
	hf.templates[TemplateCommand] = DefaultCommandTemplate
	hf.templates[TemplateCommandError] = DefaultCommandErrorTemplate
	hf.templates[TemplateCustomHelp] = DefaultCustomHelpTemplate
	hf.templates[TemplateCustomHelpError] = DefaultCustomHelpErrorTemplate
	hf.templates[TemplateSubcommand] = DefaultSubcommandTemplate
	hf.templates[TemplateFlag] = DefaultFlagTemplate
	hf.templates[TemplateEssential] = DefaultEssentialTemplate
	hf.templates[TemplateEssentialError] = DefaultEssentialErrorTemplate
	hf.templates[TemplateFull] = DefaultFullTemplate
}

// Default templates (can be overridden)
const (
	DefaultGlobalTemplate = `Usage: {{.Executable}} <command> [options]

{{if .Commands}}Available commands:

{{range .Commands}}{{if .Aliases}}  {{printf "%-12s" .Name}} (aliases: {{join .Aliases ", "}}) {{.Description}}
{{else}}  {{printf "%-12s" .Name}} {{.Description}}
{{end}}{{end}}

Use '{{.Executable}} <command> --help' for command-specific help{{else}}{{if .Description}}{{.Description}}

{{end}}Use '{{.Executable}} --help' for configuration options{{end}}`

	DefaultCommandTemplate = `Usage: {{.Command.Name}} [options]

{{.Description}}

{{if .Flags}}Flags:
{{range .Flags}}  {{.DisplayLine}}
{{if .Description}}        {{.Description}}
{{end}}{{end}}{{end}}

{{if .Subcommands}}Subcommands:
{{range .Subcommands}}  {{printf "%-12s" .Name}} {{.Description}}{{if .Aliases}} (aliases: {{join .Aliases ", "}}){{end}}
{{end}}{{end}}`

	DefaultCommandErrorTemplate = `Usage: {{.Command.Name}} [options]

{{.Description}}

{{if .HasErrors}}Configuration errors:
{{range .Errors}}  {{.Display}} -> {{.ErrorDescription}}
{{end}}

{{end}}{{if .Flags}}Flags:
{{range .Flags}}  {{.DisplayLine}}
{{if .Description}}        {{.Description}}
{{end}}{{end}}{{end}}

{{if .Subcommands}}Subcommands:
{{range .Subcommands}}  {{printf "%-12s" .Name}} {{.Description}}{{if .Aliases}} (aliases: {{join .Aliases ", "}}){{end}}
{{end}}{{end}}`

	DefaultCustomHelpTemplate = `{{.Command.LongHelp}}

{{if .Command.ShortHelp}}Use '{{.Command.Name}} <options>' to {{.Command.ShortHelp}}{{else}}Use '{{.Command.Name}} <options>' to execute the command{{end}}`

	DefaultCustomHelpErrorTemplate = `{{.Command.LongHelp}}

{{if .HasErrors}}Configuration errors:
{{range .Errors}}  {{.Display}} -> {{.ErrorDescription}}
{{end}}

{{end}}{{if .Command.ShortHelp}}Use '{{.Command.Name}} <options>' to {{.Command.ShortHelp}}{{else}}Use '{{.Command.Name}} <options>' to execute the command{{end}}`

	DefaultSubcommandTemplate = `Subcommands for {{.Parent}}:

{{range .Subcommands}}  {{printf "%-12s" .Name}} {{.Description}}{{if .Aliases}} (aliases: {{join .Aliases ", "}}){{end}}
{{end}}

Use '{{.Parent}} <subcommand> --help' for more information on a specific subcommand.`

	DefaultFlagTemplate = `Usage of {{.Command}}:
{{range .Flags}}{{if .NoFlag}}  (no flag) {{.Type}} (env: {{.EnvVar}}{{if .Required}}, required{{end}}{{if .Default}}, default: {{.Default}}{{end}}{{if .Secret}}, secret{{end}})
        {{.Description}}
{{else}}  --{{.Name}} {{.Type}}{{if .Required}} (required){{end}}{{if .Default}} (default: {{.Default}}){{end}}{{if .EnvVar}} (env: {{.EnvVar}}){{end}}{{if .Validations}} ({{join .Validations ", "}}){{end}}{{if .Secret}} (secret){{end}}
        {{.Description}}
{{end}}{{end}}`

	DefaultEssentialTemplate = `Usage: {{.Command.Name}} [options]

{{.Description}}

{{if .Flags}}Flags:
{{range .Flags}}  {{.DisplayLine}}
{{if .Description}}        {{.Description}}
{{end}}{{end}}{{end}}

{{if .RequiredEnvVars}}Environment Variables:
{{range .RequiredEnvVars}}  {{.EnvVarDisplay}}
        {{.Description}}
{{end}}{{end}}

{{if .Subcommands}}Subcommands:
{{range .Subcommands}}  {{printf "%-12s" .Name}} {{.Description}}{{if .Aliases}} (aliases: {{join .Aliases ", "}}){{end}}
{{end}}{{end}}`

	DefaultFullTemplate = `Usage: {{.Command.Name}} [options]

{{.Description}}

{{if .Flags}}Flags:
{{range .Flags}}  {{.DisplayLine}}
{{if .Description}}        {{.Description}}
{{end}}{{end}}{{end}}

{{if .AllEnvVars}}Environment Variables:
{{range .AllEnvVars}}  {{.EnvVarDisplay}}
        {{.Description}}
{{end}}{{end}}

{{if .Subcommands}}Subcommands:
{{range .Subcommands}}  {{printf "%-12s" .Name}} {{.Description}}{{if .Aliases}} (aliases: {{join .Aliases ", "}}){{end}}
{{end}}{{end}}`

	DefaultEssentialErrorTemplate = `Usage: {{.Command.Name}} [options]

{{.Description}}

{{if .HasErrors}}Configuration errors:
{{range .Errors}}  {{.Display}} -> {{.ErrorDescription}}
{{end}}

{{end}}{{if .Flags}}Flags:
{{range .Flags}}  {{.DisplayLine}}
{{if .Description}}        {{.Description}}
{{end}}{{end}}{{end}}

{{if .RequiredEnvVars}}Environment Variables:
{{range .RequiredEnvVars}}  {{.EnvVarDisplay}}
        {{.Description}}
{{end}}{{end}}

{{if .Subcommands}}Subcommands:
{{range .Subcommands}}  {{printf "%-12s" .Name}} {{.Description}}{{if .Aliases}} (aliases: {{join .Aliases ", "}}){{end}}
{{end}}{{end}}`
)
