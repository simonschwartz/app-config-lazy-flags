package filecache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/simonschwartz/app-config-lazy-flags/internal/appconfig"
)

const (
	filename = ".cache"
	dir = "LazyFlags"
	ttl = time.Minute
)

type cache struct {
	Value   []appconfig.Result `json:"value"`
	Expires time.Time          `json:"expires"`
}

type Cache struct {
	// cache key will be appId:configId
	entries map[string]cache
}

func New() (*Cache, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	appCacheDir := filepath.Join(cacheDir, dir)

	if err := os.MkdirAll(appCacheDir, 0755); err != nil {
		return nil, err
	}

	fc := &Cache{
		entries: make(map[string]cache),
	}

	// Try to load existing cache file
	cachePath := filepath.Join(appCacheDir, filename)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		// File doesn't exist or can't be read - start with empty cache
		if !os.IsNotExist(err) {
			// Log non-existence errors but continue with empty cache
			return fc, nil
		}
		return fc, nil
	}

	// Deserialize cache data
	if err := json.Unmarshal(data, &fc.entries); err != nil {
		// Corrupted cache - start fresh
		return fc, nil
	}

	return fc, nil
}

func (fc *Cache) Get(key string) ([]appconfig.Result, bool) {
	entry, exists := fc.entries[key]
	if !exists {
		return nil, false
	}

	// Check if cache entry has expired
	if time.Now().After(entry.Expires) {
		delete(fc.entries, key)
		return nil, false
	}

	return entry.Value, true
}

func (fc *Cache) Add(key string, value []appconfig.Result) error {
	fc.entries[key] = cache{
		Value:   value,
		Expires: time.Now().Add(ttl),
	}

	return fc.persist()
}

func (fc *Cache) persist() error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}

	cachePath := filepath.Join(cacheDir, dir, filename)
	data, err := json.Marshal(fc.entries)
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0644)
}
