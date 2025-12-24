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
	// Manually modify timestamp to be old
	// path := filepath.Join(c.Dir, key+".json")
	// We can't easily modify the file content without re-writing it,
	// but we can just wait if the TTL is small, or use a negative TTL for testing?
	// No, Get checks time.Since(timestamp) > ttl.
	// If we pass a negative TTL (or zero), it might not work as expected depending on logic.
	// Let's just use a very short TTL and sleep.

	// Actually, let's just pass a 0 duration TTL, which means any existing file (created "now")
	// will have time.Since > 0.
	// Wait, time.Since(now) is approx 0.
	// If I pass 0 TTL, time.Since > 0 is likely true.

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
