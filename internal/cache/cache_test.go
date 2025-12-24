package cache

import (
	"os"
	"testing"
	"time"
)

type TestData struct {
	Value string `json:"value"`
}

func TestCache(t *testing.T) {
	appName := "steam-pick-test"
	c, err := New[TestData](appName)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer os.RemoveAll(c.Dir)

	key := "test-key"
	data := TestData{Value: "hello"}

	// Test Set
	if err := c.Set(key, data); err != nil {
		t.Errorf("Set() error = %v", err)
	}

	// Test Get Hit
	got, found, err := c.Get(key, time.Minute)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if !found {
		t.Errorf("Get() found = false, want true")
	}
	if got.Value != data.Value {
		t.Errorf("Get() value = %v, want %v", got.Value, data.Value)
	}

	// Test Get Expired
	// Let's sleep for 10ms and pass 1ns TTL.
	time.Sleep(10 * time.Millisecond)
	_, found, _ = c.Get(key, 1*time.Nanosecond)
	if found {
		t.Errorf("Get() with small TTL found = true, want false (expired)")
	}

	// Test Get Miss
	_, found, _ = c.Get("missing", time.Minute)
	if found {
		t.Errorf("Get() missing found = true, want false")
	}

	// Test Stats
	count, size, err := c.Stats()
	if err != nil {
		t.Errorf("Stats() error = %v", err)
	}
	if count != 1 {
		t.Errorf("Stats() count = %d, want 1", count)
	}
	if size == 0 {
		t.Errorf("Stats() size = 0, want > 0")
	}

	// Test Clear
	if err := c.Clear(); err != nil {
		t.Errorf("Clear() error = %v", err)
	}
	// Verify dir is gone
	if _, err := os.Stat(c.Dir); !os.IsNotExist(err) {
		t.Errorf("Clear() failed to remove dir")
	}
}

type MockRunner struct {
}

func (m *MockRunner) Run(name string, args ...string) ([]byte, error) {
	if name == "gpg" && args[0] == "--decrypt" {
		// Find input path (last arg)
		path := args[len(args)-1]
		return os.ReadFile(path)
	}
	return nil, nil
}

func (m *MockRunner) RunWithInput(name string, input []byte, args ...string) ([]byte, error) {
	if name == "gpg" && args[0] == "--encrypt" {
		// Find output path
		for i, arg := range args {
			if arg == "--output" && i+1 < len(args) {
				path := args[i+1]
				// Write "encrypted" data (just the input for mock)
				return nil, os.WriteFile(path, input, 0644)
			}
		}
	}
	return nil, nil
}

func TestCacheEncryption(t *testing.T) {
	appName := "steam-pick-test-enc"
	c, err := New[TestData](appName)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer os.RemoveAll(c.Dir)

	c.WithEncryption("test-key")
	c.Runner = &MockRunner{}

	key := "test-key-enc"
	data := TestData{Value: "secret"}

	// Test Set
	if err := c.Set(key, data); err != nil {
		t.Errorf("Set() error = %v", err)
	}

	// Test Get Hit
	got, found, err := c.Get(key, time.Minute)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if !found {
		t.Errorf("Get() found = false, want true")
	}
	if got.Value != data.Value {
		t.Errorf("Get() value = %v, want %v", got.Value, data.Value)
	}
}
