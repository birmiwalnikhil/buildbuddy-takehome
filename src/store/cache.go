package store

import (
  "errors"
  "fmt"
)

// A value and cache-relevant metadata, e.g. last accessed timestamp.
type cacheEntry struct {
  value Value
  // The most recently accessed time of the CacheEntry. 
  lastAccessedTimestamp int
  sizeBytes int 
}

// An LRU cache that supports a Key/Value store.
// This class is not thread safe; callers are expected to serialize
// reads and writes appropriately.
type Cache struct {
  // The upstream KeyValueStore of the cache.
  upstream KeyValueStore
  // The maximum size of the cache, in bytes.
  capacityBytes int
  // The current size of the cache, in bytes.
  sizeBytes int
  // A monotonically increasing timestamp.
  timestamp int
  // A min-heap of the least recently used CacheKey.
  // An in-memory key/value store.
  cache map[Key]*cacheEntry 
}

/** 
 * Set the key/value store in memory, possibly performing eviction. 
 */
func (c *Cache) Set(key Key, value Value) error {
  // Maintain the upstream and the cache in sync.
  if err := c.upstream.Set(key, value); err != nil {
    return err
  }  

  // If we need to evict, evict the least recently used key.

  // Increment the size of the cache.
  entry := &cacheEntry{}
  entry.value = value
  entry.sizeBytes = value.SizeOfBytes()
  entry.lastAccessedTimestamp = c.getAndIncreaseTimestamp()
  c.cache[key] = entry
  return nil
}

/**
 * Retrieve the key/value from memory. On cache miss, consult the upstream
 * store.
 */
func (c *Cache) Get(key Key) (Value, error) {
  if entry, ok := c.cache[key]; ok {
    c.onKeyTouched(key); 
    return entry.value, nil
  }

  value, err := c.upstream.Get(key)
  if err != nil {
    return "", err 
  }

  return value, nil
}

// Evict the key from the cache.
func (c *Cache) evict(key Key) error {
  // Identify the size of the value.
  return nil
}

func (c *Cache) onKeyTouched(key Key) error {
  entry, ok := c.cache[key]
  if !ok {
    return errors.New(fmt.Sprintf("Key missing in cache", key))
  }

  entry.lastAccessedTimestamp = c.getAndIncreaseTimestamp()
  return nil
}

func (c *Cache) getAndIncreaseTimestamp() int {
  timestamp := c.timestamp
  c.timestamp++
  return timestamp
}

// Construct a new Cache instance.
func MakeCache(capacity int, upstream KeyValueStore) *Cache {
  c := &Cache{}

  c.upstream = upstream
  c.capacityBytes = capacity
  c.sizeBytes = 0
  c.timestamp = 0
  c.cache = make(map[Key]*cacheEntry) 

  return c
}
