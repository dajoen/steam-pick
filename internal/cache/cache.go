package cache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Entry represents a cached item with a timestamp.
type Entry[T any] struct {
	Timestamp time.Time `json:"timestamp"`
	Data      T         `json:"data"`
}

// Cache manages storage of items in the filesystem.
type Cache[T any] struct {
	Dir       string
	Encrypted bool
	GPGKey    string // GPG Key ID (email or hex ID)
}

// New creates a new Cache instance.
func New[T any](appName string) (*Cache[T], error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user cache dir: %w", err)
	}
	dir := filepath.Join(userCacheDir, appName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache dir: %w", err)
	}
	return &Cache[T]{Dir: dir}, nil
}

// WithEncryption enables GPG encryption for the cache.
func (c *Cache[T]) WithEncryption(gpgKey string) *Cache[T] {
	c.Encrypted = true
	c.GPGKey = gpgKey
	return c
}

// Get retrieves an item from the cache if it exists and is not expired.
func (c *Cache[T]) Get(key string, ttl time.Duration) (*T, bool, error) {
	ext := ".json"
	if c.Encrypted {
		ext = ".json.gpg"
	}
	path := filepath.Join(c.Dir, key+ext)

	var r io.Reader

	if c.Encrypted {
		// Decrypt using gpg
		cmd := exec.Command("gpg", "--decrypt", "--quiet", path)
		out, err := cmd.Output()
		if err != nil {
			if os.IsNotExist(err) {
				return nil, false, nil
			}
			// If file exists but decryption fails (e.g. cancelled), treat as miss or error?
			// Check if file exists first
			if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
				return nil, false, nil
			}
			return nil, false, fmt.Errorf("gpg decryption failed: %w", err)
		}
		r = bytes.NewReader(out)
	} else {
		f, err := os.Open(path)
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		if err != nil {
			return nil, false, err
		}
		defer f.Close()
		r = f
	}

	var entry Entry[T]
	if err := json.NewDecoder(r).Decode(&entry); err != nil {
		return nil, false, nil
	}

	if time.Since(entry.Timestamp) > ttl {
		return nil, false, nil
	}

	return &entry.Data, true, nil
}

// Set writes an item to the cache.
func (c *Cache[T]) Set(key string, data T) error {
	ext := ".json"
	if c.Encrypted {
		ext = ".json.gpg"
	}
	path := filepath.Join(c.Dir, key+ext)

	entry := Entry[T]{
		Timestamp: time.Now(),
		Data:      data,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	if c.Encrypted {
		// Encrypt using gpg
		args := []string{"--encrypt", "--recipient", c.GPGKey, "--output", path, "--yes"}
		cmd := exec.Command("gpg", args...)
		cmd.Stdin = bytes.NewReader(jsonData)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("gpg encryption failed: %s: %w", string(out), err)
		}
		return nil
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(jsonData)
	return err
}

// Clear removes the cache directory.
func (c *Cache[T]) Clear() error {
	return os.RemoveAll(c.Dir)
}

// Stats returns the number of files and total size in bytes.
func (c *Cache[T]) Stats() (int, int64, error) {
	var count int
	var size int64

	entries, err := os.ReadDir(c.Dir)
	if os.IsNotExist(err) {
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		count++
		size += info.Size()
	}

	return count, size, nil
}

// DirPath returns the cache directory path.
func (c *Cache[T]) DirPath() string {
	return c.Dir
}
