// commandkit/template_renderer.go
package commandkit

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// TemplateRenderer renders templates with data
type TemplateRenderer interface {
	Render(templateStr string, data interface{}) (string, error)
	SetFuncMap(funcMap template.FuncMap)
	GetFuncMap() template.FuncMap
	AddFunction(name string, fn interface{})
}

// GoTemplateRenderer implements TemplateRenderer using Go's text/template
type GoTemplateRenderer struct {
	funcMap template.FuncMap
}

// NewGoTemplateRenderer creates a new Go template renderer
func NewGoTemplateRenderer() TemplateRenderer {
	return &GoTemplateRenderer{
		funcMap: template.FuncMap{
			"join":      strings.Join,
			"upper":     strings.ToUpper,
			"lower":     strings.ToLower,
			"title":     strings.Title,
			"trim":      strings.TrimSpace,
			"split":     strings.Split,
			"replace":   strings.ReplaceAll,
			"contains":  strings.Contains,
			"hasPrefix": strings.HasPrefix,
			"hasSuffix": strings.HasSuffix,
			"format":    fmt.Sprintf,
		},
	}
}

// Render renders a template with the given data
func (r *GoTemplateRenderer) Render(templateStr string, data interface{}) (string, error) {
	tmpl, err := template.New("help").Funcs(r.funcMap).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// SetFuncMap sets the custom function map for templates
func (r *GoTemplateRenderer) SetFuncMap(funcMap template.FuncMap) {
	r.funcMap = funcMap
}

// GetFuncMap returns the current function map
func (r *GoTemplateRenderer) GetFuncMap() template.FuncMap {
	return r.funcMap
}

// AddFunction adds a custom function to the function map
func (r *GoTemplateRenderer) AddFunction(name string, fn interface{}) {
	if r.funcMap == nil {
		r.funcMap = make(template.FuncMap)
	}
	r.funcMap[name] = fn
}
