# Help System Example

This example demonstrates the new template-based help system in CommandKit.

## Features Demonstrated

### 1. Template Customization
- Custom global help template with banner formatting
- Custom command help template with enhanced formatting
- Custom template functions (`reverse`, `banner`)

### 2. Help Integration
- `HelpIntegration` for seamless help system access
- `HelpConfig` wrapper for Config instances
- `HelpExecutor` for command execution integration

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
integration.AddCustomFunction("reverse", func(s string) string {
    // Reverse string functionality
})

integration.AddCustomFunction("banner", func(s string) string {
    // Banner formatting functionality
})
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

integration.SetCustomTemplate(TemplateGlobal, customGlobalTemplate)
```

### Help Generation
Demonstrates both display and text generation:

```go
// Display help directly
err := integration.ShowHelp([]string{"--help"}, commands)

// Generate help text
text, err := integration.GenerateHelp([]string{"server", "--help"}, commands)
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

### HelpIntegration
Main integration point for the new help system:

```go
integration := commandkit.NewHelpIntegration()
integration.SetCustomTemplate(commandkit.TemplateGlobal, customTemplate)
integration.AddCustomFunction("customFunc", customFunction)
```

### HelpConfig
Wrapper for Config instances:

```go
helpConfig := commandkit.NewHelpConfig(cfg)
err := helpConfig.ShowGlobalHelp()
```

### HelpExecutor
Integration for command execution:

```go
executor := commandkit.NewHelpExecutor()
handled, err := executor.CheckAndHandleHelp(args, commands)
```

## Benefits

1. **Template-Based**: Full control over help output formatting
2. **Customizable**: Add custom functions and templates
3. **Flexible**: Multiple output destinations
4. **Integrated**: Seamless integration with existing CommandKit
5. **Extensible**: Easy to extend with new features

This example showcases the power and flexibility of the new help system while maintaining backward compatibility with existing CommandKit code.
