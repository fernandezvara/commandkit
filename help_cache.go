package commandkit

import (
	"crypto/sha256"
	"fmt"
	"sync"
)

// HelpCache provides caching for generated help text
type HelpCache struct {
	cache sync.Map
}

// NewHelpCache creates a new help cache
func NewHelpCache() *HelpCache {
	return &HelpCache{}
}

// GetHelp retrieves cached help or generates it using the provided function
func (hc *HelpCache) GetHelp(cacheKey string, generator func() (string, error)) (string, error) {
	// Try cache first
	if cached, ok := hc.cache.Load(cacheKey); ok {
		return cached.(string), nil
	}

	// Generate help
	result, err := generator()
	if err != nil {
		return "", err
	}

	// Cache the result
	hc.cache.Store(cacheKey, result)
	return result, nil
}

// Clear clears the cache
func (hc *HelpCache) Clear() {
	hc.cache = sync.Map{}
}

// generateCacheKey creates a cache key from help data
func (hc *HelpCache) generateCacheKey(data interface{}) string {
	// Simple hash-based cache key
	// In a real implementation, you might want to be more sophisticated
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%v", data))))
}

// CachedHelpFormatter wraps a HelpFormatter with caching
type CachedHelpFormatter struct {
	HelpFormatter
	cache *HelpCache
}

// NewCachedHelpFormatter creates a new cached help formatter
func NewCachedHelpFormatter(formatter HelpFormatter) HelpFormatter {
	return &CachedHelpFormatter{
		HelpFormatter: formatter,
		cache:         NewHelpCache(),
	}
}

// FormatGlobalHelp formats global help with caching
func (chf *CachedHelpFormatter) FormatGlobalHelp(help *GlobalHelp) (string, error) {
	cacheKey := chf.cache.generateCacheKey(help)

	return chf.cache.GetHelp(cacheKey, func() (string, error) {
		return chf.HelpFormatter.FormatGlobalHelp(help)
	})
}

// FormatCommandHelp formats command help with caching
func (chf *CachedHelpFormatter) FormatCommandHelp(help *CommandHelp) (string, error) {
	cacheKey := chf.cache.generateCacheKey(help)

	return chf.cache.GetHelp(cacheKey, func() (string, error) {
		return chf.HelpFormatter.FormatCommandHelp(help)
	})
}

// FormatSubcommandHelp formats subcommand help with caching
func (chf *CachedHelpFormatter) FormatSubcommandHelp(help *SubcommandHelp) (string, error) {
	cacheKey := chf.cache.generateCacheKey(help)

	return chf.cache.GetHelp(cacheKey, func() (string, error) {
		return chf.HelpFormatter.FormatSubcommandHelp(help)
	})
}

// FormatFlagHelp formats flag help with caching
func (chf *CachedHelpFormatter) FormatFlagHelp(help *FlagHelp) (string, error) {
	cacheKey := chf.cache.generateCacheKey(help)

	return chf.cache.GetHelp(cacheKey, func() (string, error) {
		return chf.HelpFormatter.FormatFlagHelp(help)
	})
}

// SetTemplate sets a template and clears cache
func (chf *CachedHelpFormatter) SetTemplate(templateType TemplateType, template string) {
	chf.HelpFormatter.SetTemplate(templateType, template)
	// Clear cache when templates change
	chf.cache.Clear()
}

// GetTemplate gets a template
func (chf *CachedHelpFormatter) GetTemplate(templateType TemplateType) string {
	return chf.HelpFormatter.GetTemplate(templateType)
}

// SetRenderer sets a renderer and clears cache
func (chf *CachedHelpFormatter) SetRenderer(renderer TemplateRenderer) {
	chf.HelpFormatter.SetRenderer(renderer)
	// Clear cache when renderer changes
	chf.cache.Clear()
}

// GetRenderer gets the renderer
func (chf *CachedHelpFormatter) GetRenderer() TemplateRenderer {
	return chf.HelpFormatter.GetRenderer()
}

// ClearCache clears the help cache
func (chf *CachedHelpFormatter) ClearCache() {
	chf.cache.Clear()
}
