// commandkit/template_renderer.go
package commandkit

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
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

// CachedTemplateRenderer implements TemplateRenderer with template caching and string builder pool
type CachedTemplateRenderer struct {
	*GoTemplateRenderer
	templateCache sync.Map // Use sync.Map for better performance in concurrent scenarios
	builderPool   sync.Pool
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

// NewCachedTemplateRenderer creates a new cached template renderer
func NewCachedTemplateRenderer() TemplateRenderer {
	base := &GoTemplateRenderer{
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

	return &CachedTemplateRenderer{
		GoTemplateRenderer: base,
		builderPool: sync.Pool{
			New: func() interface{} {
				return &strings.Builder{}
			},
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

// Render renders a template with caching and string builder pool optimization
func (r *CachedTemplateRenderer) Render(templateStr string, data interface{}) (string, error) {
	// Create cache key from template string
	cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte(templateStr)))

	// Try cache first
	if cached, ok := r.templateCache.Load(cacheKey); ok {
		tmpl := cached.(*template.Template)
		return r.executeTemplate(tmpl, data)
	}

	// Parse and cache
	tmpl, err := template.New("cached").Funcs(r.funcMap).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	r.templateCache.Store(cacheKey, tmpl)
	return r.executeTemplate(tmpl, data)
}

// executeTemplate executes a template using the string builder pool
func (r *CachedTemplateRenderer) executeTemplate(tmpl *template.Template, data interface{}) (string, error) {
	// Get builder from pool
	builder := r.builderPool.Get().(*strings.Builder)
	defer func() {
		builder.Reset()
		r.builderPool.Put(builder)
	}()

	// Execute template
	if err := tmpl.Execute(builder, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return builder.String(), nil
}

// SetFuncMap sets the custom function map for templates
func (r *GoTemplateRenderer) SetFuncMap(funcMap template.FuncMap) {
	r.funcMap = funcMap
}

// GetFuncMap returns the current function map
func (r *GoTemplateRenderer) GetFuncMap() template.FuncMap {
	return r.funcMap
}

// SetFuncMap sets the custom function map for templates and clears cache
func (r *CachedTemplateRenderer) SetFuncMap(funcMap template.FuncMap) {
	r.GoTemplateRenderer.SetFuncMap(funcMap)
	// Clear cache when function map changes
	r.templateCache = sync.Map{}
}

// GetFuncMap returns the current function map
func (r *CachedTemplateRenderer) GetFuncMap() template.FuncMap {
	return r.GoTemplateRenderer.GetFuncMap()
}

// AddFunction adds a custom function to the function map
func (r *GoTemplateRenderer) AddFunction(name string, fn interface{}) {
	if r.funcMap == nil {
		r.funcMap = make(template.FuncMap)
	}
	r.funcMap[name] = fn
}

// AddFunction adds a custom function to the function map and clears cache
func (r *CachedTemplateRenderer) AddFunction(name string, fn interface{}) {
	r.GoTemplateRenderer.AddFunction(name, fn)
	// Clear cache when function map changes
	r.templateCache = sync.Map{}
}
