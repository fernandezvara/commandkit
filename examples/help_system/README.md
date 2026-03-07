# Help System Example

This example demonstrates the template-based help system in CommandKit.

## Features Demonstrated

### 1. Template Customization
- Custom global help template with banner formatting
- Custom command help template with enhanced formatting
- Custom template functions (`reverse`, `banner`)

### 2. Help Service
- Direct `HelpService` usage for help system access
- Template customization through formatter interface
- Output management for testing and display

### 3. Output Management
- String output for testing and automation
- Console output for direct display
- Multiple output destinations

### 4. Help Types
- Global help (`--help`)
- Command help (`<command> --help`)
- Subcommand help (`<command> <subcommand> --help`)

## Running the Example

```bash
cd examples/help_system
go run example_help_system.go
```

## Key Features

### Template Functions
The example demonstrates custom template functions:

```go
helpService := commandkit.NewHelpService()
formatter := helpService.GetFormatter()

if templateFormatter, ok := formatter.(*commandkit.TemplateHelpFormatter); ok {
    renderer := templateFormatter.GetRenderer()
    renderer.AddFunction("reverse", func(s string) string {
        // Reverse string functionality
    })

    renderer.AddFunction("banner", func(s string) string {
        // Banner formatting functionality
    })
}
```

### Custom Templates
Shows how to override default templates:

```go
customGlobalTemplate := `{{banner "MyApp CLI Tool"}}

{{.Executable | upper}} - Command Line Interface

USAGE:
  {{.Executable}} <command> [options]

{{if .Commands}}AVAILABLE COMMANDS:
{{range .Commands}}  {{.Name | printf "%-12s"}} {{.Description}}
{{end}}{{end}}`

if templateFormatter, ok := formatter.(*commandkit.TemplateHelpFormatter); ok {
    templateFormatter.SetTemplate(commandkit.TemplateGlobal, customGlobalTemplate)
}
```

### Help Generation
Demonstrates both display and text generation:

```go
// Display help directly
err := helpService.ShowHelp([]string{"--help"}, commands)

// Generate help text
text, err := helpService.GenerateHelp([]string{"server", "--help"}, commands)
```

## Output Examples

The example produces help output like:

```
===================================
= MyApp CLI Tool                =
===================================

EXAMPLE -- COMMAND LINE INTERFACE

USAGE:
  Example <command> [options]

AVAILABLE COMMANDS:
  server        Start the HTTP server
  deploy        Deploy the application
  admin         Administrative operations

EXAMPLES:
  Example server --host 0.0.0.0 --port 3000
  Example deploy --target production
  Example admin users --action list
```

## Integration Points

### HelpService
Main interface for the help system:

```go
helpService := commandkit.NewHelpService()

// Set custom templates
formatter := helpService.GetFormatter()
if templateFormatter, ok := formatter.(*commandkit.TemplateHelpFormatter); ok {
    templateFormatter.SetTemplate(commandkit.TemplateGlobal, customTemplate)
}

// Add custom functions
renderer := templateFormatter.GetRenderer()
renderer.AddFunction("customFunc", customFunction)
```

### Help Output Control
Flexible output management:

```go
// String output for testing
stringOutput := commandkit.NewStringHelpOutput()
helpService.SetOutput(stringOutput)

// Generate help without displaying
text, err := helpService.GenerateHelp([]string{"--help"}, commands)
```

## Benefits

1. **Template-Based**: Full control over help output formatting
2. **Customizable**: Add custom functions and templates
3. **Flexible**: Multiple output destinations
4. **Direct**: Simple, direct API without unnecessary layers
5. **Extensible**: Easy to extend with new features

This example showcases the power and flexibility of the help system while maintaining a clean, simple architecture.
