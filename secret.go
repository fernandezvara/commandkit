// commandkit/secret.go
package commandkit

import (
	"github.com/awnumar/memguard"
)

// Secret wraps a memguard LockedBuffer for secure secret storage
type Secret struct {
	buffer *memguard.LockedBuffer
}

// newSecret creates a new Secret from a string value
func newSecret(value string) *Secret {
	if value == "" {
		return &Secret{}
	}

	buf := memguard.NewBufferFromBytes([]byte(value))
	return &Secret{buffer: buf}
}

// Bytes returns the secret value as bytes
// The returned slice is only valid until Destroy() is called
func (s *Secret) Bytes() []byte {
	if s.buffer == nil {
		return nil
	}
	return s.buffer.Bytes()
}

// String returns the secret value as a string
// The returned string is only valid until Destroy() is called
func (s *Secret) String() string {
	if s.buffer == nil {
		return ""
	}
	return string(s.buffer.Bytes())
}

// Destroy securely wipes the secret from memory
func (s *Secret) Destroy() {
	if s.buffer != nil {
		s.buffer.Destroy()
		s.buffer = nil
	}
}

// IsSet returns true if the secret has a value
func (s *Secret) IsSet() bool {
	return s.buffer != nil && s.buffer.Size() > 0
}

// Size returns the length of the secret
func (s *Secret) Size() int {
	if s.buffer == nil {
		return 0
	}
	return s.buffer.Size()
}

// SecretStore holds all secrets for cleanup
type SecretStore struct {
	secrets map[string]*Secret
}

func newSecretStore() *SecretStore {
	return &SecretStore{
		secrets: make(map[string]*Secret),
	}
}

func (ss *SecretStore) Store(key, value string) {
	ss.secrets[key] = newSecret(value)
}

func (ss *SecretStore) Get(key string) *Secret {
	if s, ok := ss.secrets[key]; ok {
		return s
	}
	return &Secret{}
}

func (ss *SecretStore) DestroyAll() {
	for _, s := range ss.secrets {
		s.Destroy()
	}
	ss.secrets = make(map[string]*Secret)
}
