package cache

import (
	"encoding/json"
	"fmt"
	"os"
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
	Dir string
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

// Get retrieves an item from the cache if it exists and is not expired.
func (c *Cache[T]) Get(key string, ttl time.Duration) (*T, bool, error) {
	path := filepath.Join(c.Dir, key+".json")
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	defer f.Close()

	var entry Entry[T]
	if err := json.NewDecoder(f).Decode(&entry); err != nil {
		// If we can't decode it, treat it as a miss (and maybe it will be overwritten later)
		return nil, false, nil
	}

	if time.Since(entry.Timestamp) > ttl {
		return nil, false, nil
	}

	return &entry.Data, true, nil
}

// Set writes an item to the cache.
func (c *Cache[T]) Set(key string, data T) error {
	path := filepath.Join(c.Dir, key+".json")
	entry := Entry[T]{
		Timestamp: time.Now(),
		Data:      data,
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(entry)
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
