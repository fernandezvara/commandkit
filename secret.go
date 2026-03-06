// commandkit/secret.go
package commandkit

import (
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/awnumar/memguard"
)

// Secret wraps a memguard LockedBuffer for secure secret storage
type Secret struct {
	buffer  *memguard.LockedBuffer
	cleaned int32 // atomic flag for cleanup tracking
}

// newSecret creates a new Secret from a string value
func newSecret(value string) *Secret {
	if value == "" {
		return &Secret{}
	}

	buf := memguard.NewBufferFromBytes([]byte(value))
	s := &Secret{buffer: buf}

	// Set finalizer for automatic cleanup
	runtime.SetFinalizer(s, func(s *Secret) {
		s.finalize()
	})

	return s
}

// Bytes returns the secret value as bytes
// The returned slice is only valid until Destroy() is called
func (s *Secret) Bytes() []byte {
	if s.isDestroyed() {
		return nil
	}
	if s.buffer == nil {
		return nil
	}
	return s.buffer.Bytes()
}

// String returns the secret value as a string
// The returned string is only valid until Destroy() is called
func (s *Secret) String() string {
	if s.isDestroyed() {
		return ""
	}
	if s.buffer == nil {
		return ""
	}
	return string(s.buffer.Bytes())
}

// Destroy securely wipes the secret from memory
// Multiple calls are safe and will not panic
func (s *Secret) Destroy() {
	s.finalize()
}

// finalize performs the actual cleanup with atomic protection
func (s *Secret) finalize() {
	if atomic.CompareAndSwapInt32(&s.cleaned, 0, 1) {
		if s.buffer != nil {
			s.buffer.Destroy()
			s.buffer = nil
		}
		// Prevent finalizer from running again
		runtime.SetFinalizer(s, nil)
	}
}

// isDestroyed returns true if the secret has been destroyed
func (s *Secret) isDestroyed() bool {
	return atomic.LoadInt32(&s.cleaned) == 1
}

// IsSet returns true if the secret has a value and hasn't been destroyed
func (s *Secret) IsSet() bool {
	return !s.isDestroyed() && s.buffer != nil && s.buffer.Size() > 0
}

// Size returns the length of the secret
func (s *Secret) Size() int {
	if s.isDestroyed() {
		return 0
	}
	if s.buffer == nil {
		return 0
	}
	return s.buffer.Size()
}

// IsDestroyed returns true if the secret has been securely destroyed
func (s *Secret) IsDestroyed() bool {
	return s.isDestroyed()
}

// VerifyDestroyed returns true if the secret memory has been securely wiped
// This provides verification that cleanup was successful
func (s *Secret) VerifyDestroyed() bool {
	if !s.isDestroyed() {
		return false
	}
	// Additional verification: check if buffer is nil and size is 0
	return s.buffer == nil
}

// SecretStore holds all secrets for cleanup with thread-safe operations
type SecretStore struct {
	secrets   map[string]*Secret
	mu        sync.RWMutex
	destroyed int32 // atomic flag for store-wide cleanup tracking
}

func newSecretStore() *SecretStore {
	ss := &SecretStore{
		secrets: make(map[string]*Secret),
	}

	// Set finalizer for automatic cleanup of the entire store
	runtime.SetFinalizer(ss, func(ss *SecretStore) {
		ss.DestroyAll()
	})

	return ss
}

// Store securely stores a secret with thread safety
func (ss *SecretStore) Store(key, value string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Destroy existing secret if present
	if existing, exists := ss.secrets[key]; exists {
		existing.Destroy()
	}

	ss.secrets[key] = newSecret(value)
}

// Get retrieves a secret with thread safety
func (ss *SecretStore) Get(key string) *Secret {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if s, ok := ss.secrets[key]; ok {
		return s
	}
	return &Secret{}
}

// DestroyAll securely destroys all secrets with verification
func (ss *SecretStore) DestroyAll() {
	// Use atomic operation to prevent double cleanup
	if !atomic.CompareAndSwapInt32(&ss.destroyed, 0, 1) {
		return // Already destroyed
	}

	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Destroy all secrets
	for _, s := range ss.secrets {
		s.Destroy()
	}

	// Clear the map
	ss.secrets = make(map[string]*Secret)

	// Prevent finalizer from running again
	runtime.SetFinalizer(ss, nil)
}

// Has checks if a secret exists and is set
func (ss *SecretStore) Has(key string) bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if s, ok := ss.secrets[key]; ok {
		return s.IsSet()
	}
	return false
}

// Size returns the number of stored secrets
func (ss *SecretStore) Size() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	return len(ss.secrets)
}

// Keys returns all secret keys
func (ss *SecretStore) Keys() []string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	keys := make([]string, 0, len(ss.secrets))
	for key := range ss.secrets {
		keys = append(keys, key)
	}
	return keys
}

// VerifyAllDestroyed returns true if all secrets have been securely destroyed
func (ss *SecretStore) VerifyAllDestroyed() bool {
	if atomic.LoadInt32(&ss.destroyed) != 1 {
		return false
	}

	ss.mu.RLock()
	defer ss.mu.RUnlock()

	// Verify all secrets are destroyed and map is empty
	if len(ss.secrets) != 0 {
		return false
	}

	return true
}

// IsDestroyed returns true if the store has been destroyed
func (ss *SecretStore) IsDestroyed() bool {
	return atomic.LoadInt32(&ss.destroyed) == 1
}
