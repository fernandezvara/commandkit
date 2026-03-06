// commandkit/help_factory.go
package commandkit

// HelpFactory creates help objects for different contexts
type HelpFactory interface {
	// Core help creation methods
	CreateGlobalHelp(commands map[string]*Command, executable string) *GlobalHelp
	CreateCommandHelp(cmd *Command, executable string) *CommandHelp
	CreateSubcommandHelp(parent string, subcommands map[string]*Command) *SubcommandHelp
	CreateFlagHelp(command string, defs map[string]*Definition) *FlagHelp

	// Help detection and parsing
	DetectHelpRequest(args []string) *HelpRequest
	IsHelpRequested(args []string) bool
	GetHelpType(args []string) HelpType

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
	TemplateSubcommand
	TemplateFlag
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
func (hf *helpFactory) IsHelpRequested(args []string) bool {
	return hf.detector.IsHelpRequested(args)
}

// GetHelpType gets the type of help request
func (hf *helpFactory) GetHelpType(args []string) HelpType {
	return hf.detector.GetHelpType(args)
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

	return &CommandHelp{
		Command:     cmd,
		Usage:       commandInfo.Usage,
		Description: commandInfo.Description,
		Flags:       commandInfo.Flags,
		Subcommands: commandInfo.Subcommands,
		Template:    hf.templates[TemplateCommand],
	}
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
	hf.templates[TemplateSubcommand] = DefaultSubcommandTemplate
	hf.templates[TemplateFlag] = DefaultFlagTemplate
}

// Default templates (can be overridden)
const (
	DefaultGlobalTemplate = `Usage: {{.Executable}} <command> [options]

Available commands:

{{range .Commands}}{{if .Aliases}}  {{printf "%-12s" .Name}} (aliases: {{join .Aliases ", "}}) {{.Description}}
{{else}}  {{printf "%-12s" .Name}} {{.Description}}
{{end}}{{end}}

Use '{{.Executable}} <command> --help' for command-specific help`

	DefaultCommandTemplate = `Usage: {{.Command.Name}} [options]

{{.Description}}

{{if .Flags}}Options:
{{range .Flags}}{{if .NoFlag}}  (no flag) {{.Type}} (env: {{.EnvVar}}{{if .Required}}, required{{end}}{{if .Default}}, default: {{.Default}}{{end}}{{if .Secret}}, secret{{end}})
        {{.Description}}
{{else}}  --{{.Name}} {{.Type}}{{if .Required}} (required){{end}}{{if .Default}} (default: {{.Default}}){{end}}{{if .EnvVar}} (env: {{.EnvVar}}){{end}}{{if .Validations}} ({{join .Validations ", "}}){{end}}{{if .Secret}} (secret){{end}}
        {{.Description}}
{{end}}{{end}}{{end}}

{{if .Subcommands}}Subcommands:
{{range .Subcommands}}  {{printf "%-12s" .Name}} {{.Description}}{{if .Aliases}} (aliases: {{join .Aliases ", "}}){{end}}
{{end}}{{end}}`

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
)
