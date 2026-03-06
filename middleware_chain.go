// commandkit/middleware_chain.go
package commandkit

// MiddlewareChain manages middleware application and composition
type MiddlewareChain interface {
	// Apply combines command-specific and global middleware and applies them in correct order
	Apply(cmd *Command, globalMiddleware []CommandMiddleware, baseFunc CommandFunc) CommandFunc

	// ApplyCommandOnly applies only command-specific middleware (for use within Command.Execute)
	ApplyCommandOnly(cmd *Command, baseFunc CommandFunc) CommandFunc

	// ApplyGlobalOnly applies only global middleware (for use within Config.Execute)
	ApplyGlobalOnly(globalMiddleware []CommandMiddleware, baseFunc CommandFunc) CommandFunc
}

// middlewareChain implements MiddlewareChain interface
type middlewareChain struct{}

// NewMiddlewareChain creates a new MiddlewareChain instance
func NewMiddlewareChain() MiddlewareChain {
	return &middlewareChain{}
}

// Apply combines command-specific and global middleware and applies them in correct order
// Middleware is applied in reverse order (last added wraps first) for proper chaining
func (mc *middlewareChain) Apply(cmd *Command, globalMiddleware []CommandMiddleware, baseFunc CommandFunc) CommandFunc {
	if cmd == nil {
		return baseFunc
	}

	// Start with the base function (usually the command function)
	finalFunc := baseFunc

	// Combine all middleware: global middleware first, then command-specific middleware
	// This ensures command middleware wraps global middleware
	var allMiddleware []CommandMiddleware

	// Add global middleware first
	allMiddleware = append(allMiddleware, globalMiddleware...)

	// Add command-specific middleware
	allMiddleware = append(allMiddleware, cmd.Middleware...)

	// Apply middleware in reverse order (last added wraps first)
	for i := len(allMiddleware) - 1; i >= 0; i-- {
		finalFunc = allMiddleware[i](finalFunc)
	}

	return finalFunc
}

// ApplyCommandOnly applies only command-specific middleware (for use within Command.Execute)
func (mc *middlewareChain) ApplyCommandOnly(cmd *Command, baseFunc CommandFunc) CommandFunc {
	if cmd == nil {
		return baseFunc
	}

	// Start with the base function
	finalFunc := baseFunc

	// Apply command-specific middleware in reverse order
	for i := len(cmd.Middleware) - 1; i >= 0; i-- {
		finalFunc = cmd.Middleware[i](finalFunc)
	}

	return finalFunc
}

// ApplyGlobalOnly applies only global middleware (for use within Config.Execute)
func (mc *middlewareChain) ApplyGlobalOnly(globalMiddleware []CommandMiddleware, baseFunc CommandFunc) CommandFunc {
	// Start with the base function
	finalFunc := baseFunc

	// Apply global middleware in reverse order
	for i := len(globalMiddleware) - 1; i >= 0; i-- {
		finalFunc = globalMiddleware[i](finalFunc)
	}

	return finalFunc
}
