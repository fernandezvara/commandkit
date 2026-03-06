package commandkit

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestNewSecret(t *testing.T) {
	// Test with value
	s := newSecret("my-secret-value")
	if s == nil {
		t.Fatal("newSecret should not return nil")
	}
	if !s.IsSet() {
		t.Error("Secret should be set")
	}
	defer s.Destroy()

	// Test with empty value
	s2 := newSecret("")
	if s2 == nil {
		t.Fatal("newSecret with empty string should not return nil")
	}
	if s2.IsSet() {
		t.Error("Secret with empty value should not be set")
	}
}

func TestSecretBytes(t *testing.T) {
	s := newSecret("test-secret")
	defer s.Destroy()

	bytes := s.Bytes()
	if string(bytes) != "test-secret" {
		t.Errorf("Secret.Bytes() = %s, expected 'test-secret'", string(bytes))
	}

	// Test nil buffer
	s2 := &Secret{}
	if s2.Bytes() != nil {
		t.Error("Secret.Bytes() should return nil for nil buffer")
	}
}

func TestSecretString(t *testing.T) {
	s := newSecret("test-secret")
	defer s.Destroy()

	str := s.String()
	if str != "test-secret" {
		t.Errorf("Secret.String() = %s, expected 'test-secret'", str)
	}

	// Test nil buffer
	s2 := &Secret{}
	if s2.String() != "" {
		t.Error("Secret.String() should return empty string for nil buffer")
	}
}

func TestSecretDestroy(t *testing.T) {
	s := newSecret("test-secret")

	// Verify it's set before destroy
	if !s.IsSet() {
		t.Error("Secret should be set before destroy")
	}

	s.Destroy()

	// Verify it's not set after destroy
	if s.IsSet() {
		t.Error("Secret should not be set after destroy")
	}
	if s.Size() != 0 {
		t.Error("Secret size should be 0 after destroy")
	}

	// Double destroy should not panic
	s.Destroy()
}

func TestSecretIsSet(t *testing.T) {
	// Set secret
	s := newSecret("value")
	if !s.IsSet() {
		t.Error("Secret with value should be set")
	}
	s.Destroy()

	// Empty secret
	s2 := newSecret("")
	if s2.IsSet() {
		t.Error("Secret with empty value should not be set")
	}

	// Nil buffer
	s3 := &Secret{}
	if s3.IsSet() {
		t.Error("Secret with nil buffer should not be set")
	}
}

func TestSecretSize(t *testing.T) {
	s := newSecret("12345")
	defer s.Destroy()

	if s.Size() != 5 {
		t.Errorf("Secret.Size() = %d, expected 5", s.Size())
	}

	// Nil buffer
	s2 := &Secret{}
	if s2.Size() != 0 {
		t.Error("Secret.Size() should return 0 for nil buffer")
	}
}

func TestSecretStore(t *testing.T) {
	ss := newSecretStore()

	// Store secrets
	ss.Store("key1", "value1")
	ss.Store("key2", "value2")

	// Get existing secrets
	s1 := ss.Get("key1")
	if !s1.IsSet() || s1.String() != "value1" {
		t.Error("SecretStore.Get('key1') failed")
	}

	s2 := ss.Get("key2")
	if !s2.IsSet() || s2.String() != "value2" {
		t.Error("SecretStore.Get('key2') failed")
	}

	// Get non-existent secret
	s3 := ss.Get("nonexistent")
	if s3.IsSet() {
		t.Error("SecretStore.Get('nonexistent') should return unset secret")
	}

	// Destroy all
	ss.DestroyAll()

	// After destroy, getting should return unset secrets
	s4 := ss.Get("key1")
	if s4.IsSet() {
		t.Error("SecretStore.Get after DestroyAll should return unset secret")
	}
}

func TestSecretStoreOverwrite(t *testing.T) {
	ss := newSecretStore()

	ss.Store("key", "value1")
	s1 := ss.Get("key")
	if s1.String() != "value1" {
		t.Error("Initial store failed")
	}

	// Overwrite
	ss.Store("key", "value2")
	s2 := ss.Get("key")
	if s2.String() != "value2" {
		t.Error("Overwrite failed")
	}

	ss.DestroyAll()
}

func TestSecretStoreDestroyAllMultipleTimes(t *testing.T) {
	ss := newSecretStore()

	ss.Store("key1", "value1")
	ss.Store("key2", "value2")

	// Multiple destroys should not panic
	ss.DestroyAll()
	ss.DestroyAll()
	ss.DestroyAll()

	// Should be able to store again after destroy
	ss.Store("key3", "value3")
	s := ss.Get("key3")
	if !s.IsSet() || s.String() != "value3" {
		t.Error("Store after DestroyAll failed")
	}

	ss.DestroyAll()
}

// TestSecretRAII tests automatic cleanup with finalizers
func TestSecretRAII(t *testing.T) {
	// Create a secret
	s := newSecret("test-secret-raii")
	if !s.IsSet() {
		t.Error("Secret should be set initially")
	}

	// Destroy it
	s.Destroy()

	// Verify it's destroyed
	if !s.IsDestroyed() {
		t.Error("Secret should be marked as destroyed")
	}

	if !s.VerifyDestroyed() {
		t.Error("Secret should verify as destroyed")
	}

	if s.IsSet() {
		t.Error("Destroyed secret should not be set")
	}

	if s.Size() != 0 {
		t.Error("Destroyed secret should have size 0")
	}

	if s.String() != "" {
		t.Error("Destroyed secret should return empty string")
	}

	if s.Bytes() != nil {
		t.Error("Destroyed secret should return nil bytes")
	}
}

// TestSecretThreadSafety tests concurrent access to secrets
func TestSecretThreadSafety(t *testing.T) {
	ss := newSecretStore()

	// Store multiple secrets
	for i := 0; i < 10; i++ {
		ss.Store(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}

	var wg sync.WaitGroup
	errs := make(chan error, 20)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			s := ss.Get(fmt.Sprintf("key%d", index))
			if !s.IsSet() || s.String() != fmt.Sprintf("value%d", index) {
				errs <- fmt.Errorf("concurrent read failed for key%d", index)
			}
		}(i)
	}

	// Concurrent writes
	for i := 10; i < 20; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			ss.Store(fmt.Sprintf("key%d", index), fmt.Sprintf("value%d", index))
			s := ss.Get(fmt.Sprintf("key%d", index))
			if !s.IsSet() || s.String() != fmt.Sprintf("value%d", index) {
				errs <- fmt.Errorf("concurrent write failed for key%d", index)
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	// Check for errors
	for err := range errs {
		t.Error(err)
	}

	ss.DestroyAll()
}

// TestSecretMemoryLeak tests for memory leaks under pressure
func TestSecretMemoryLeak(t *testing.T) {
	ss := newSecretStore()

	// Create many secrets to test memory pressure
	for i := 0; i < 1000; i++ {
		ss.Store(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d-with-longer-content-to-test-memory-management", i))
	}

	// Verify all secrets are set
	if ss.Size() != 1000 {
		t.Errorf("Expected 1000 secrets, got %d", ss.Size())
	}

	// Destroy all
	ss.DestroyAll()

	// Verify cleanup
	if !ss.VerifyAllDestroyed() {
		t.Error("SecretStore should verify all secrets destroyed")
	}

	if ss.Size() != 0 {
		t.Error("SecretStore should have size 0 after destroy")
	}

	// Force garbage collection to trigger finalizers
	runtime.GC()
	time.Sleep(100 * time.Millisecond) // Allow finalizers to run
	runtime.GC()

	// Verify store is still destroyed
	if !ss.IsDestroyed() {
		t.Error("SecretStore should still be marked as destroyed")
	}
}

// TestSecretSecurityViolation tests that Get[T] properly blocks secret access
func TestSecretSecurityViolation(t *testing.T) {
	cfg := New()
	cfg.Define("SECRET_KEY").String().Secret().Default("secret-value")
	cfg.Define("NORMAL_KEY").String().Default("normal-value")

	// Set environment variable for secret (optional, will use default)
	t.Setenv("SECRET_KEY", "secret-value")
	t.Setenv("NORMAL_KEY", "env-value")

	result := cfg.Process()
	if result.Error != nil {
		t.Fatalf("Config processing failed: %v", result.Error)
	}

	// Create command context
	ctx := &CommandContext{
		GlobalConfig: cfg,
		execution:    NewExecutionContext("test"),
	}

	// Try to access secret with Get[T] - should fail
	secretResult := Get[string](ctx, "SECRET_KEY")
	if secretResult.Error == nil {
		t.Error("Get[string] should fail for secret keys")
	}

	if secretResult.Error.Error() != "validation error: configuration 'SECRET_KEY' is secret, use GetSecret() instead" {
		t.Errorf("Unexpected error message: %s", secretResult.Error.Error())
	}

	// Access secret properly with GetSecret
	secret := cfg.GetSecret("SECRET_KEY")
	if !secret.IsSet() || secret.String() != "secret-value" {
		t.Error("GetSecret should work for secret keys")
	}

	// Normal key should work with Get[T]
	normalResult := Get[string](ctx, "NORMAL_KEY")
	if normalResult.Error != nil {
		t.Errorf("Get[string] should work for normal keys: %v", normalResult.Error)
	}

	if GetValue[string](normalResult) != "normal-value" {
		t.Errorf("Expected 'normal-value', got '%s'", GetValue[string](normalResult))
	}

	// Test Has method behavior
	if cfg.Has("SECRET_KEY") {
		t.Error("Has should return false for secret keys")
	}

	if !cfg.HasSecret("SECRET_KEY") {
		t.Error("HasSecret should return true for secret keys")
	}

	if !cfg.Has("NORMAL_KEY") {
		t.Error("Has should return true for normal keys")
	}

	cfg.Destroy()
}

// TestSecretStoreVerification tests cleanup verification methods
func TestSecretStoreVerification(t *testing.T) {
	ss := newSecretStore()

	// Initially not destroyed
	if ss.IsDestroyed() {
		t.Error("New SecretStore should not be destroyed")
	}

	if ss.VerifyAllDestroyed() {
		t.Error("New SecretStore should not verify as destroyed")
	}

	// Add some secrets
	ss.Store("key1", "value1")
	ss.Store("key2", "value2")

	// Test Has method
	if !ss.Has("key1") {
		t.Error("SecretStore should have key1")
	}

	if ss.Has("nonexistent") {
		t.Error("SecretStore should not have nonexistent key")
	}

	// Test Keys method
	keys := ss.Keys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}

	// Destroy and verify
	ss.DestroyAll()

	if !ss.IsDestroyed() {
		t.Error("SecretStore should be destroyed")
	}

	if !ss.VerifyAllDestroyed() {
		t.Error("SecretStore should verify as destroyed")
	}

	if ss.Has("key1") {
		t.Error("Destroyed SecretStore should not have any keys")
	}

	if len(ss.Keys()) != 0 {
		t.Error("Destroyed SecretStore should have no keys")
	}
}
