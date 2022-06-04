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

  entry := &cacheEntry{}
  entry.value = value
  entry.sizeBytes = value.SizeOfBytes()
  entry.lastAccessedTimestamp = c.getAndIncreaseTimestamp()

  // Continually evict LRU elements until we have sufficient space.
  for c.sizeBytes + entry.sizeBytes >= c.capacityBytes {
    err := c.evictLru()
    if err != nil { 
      return err
    }
  }

  c.sizeBytes += entry.sizeBytes
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

// Evict the least recently used key from the cache, returning 
// an error in case of failure.
// TODO: Improve to a non O(|cache|) operation, e.g. via min heaps.
func (c *Cache) evictLru() error {
  if len(c.cache) == 0 {
    return errors.New("Cannot evict an empty cache.")
  }

  // Arbitrarily assign the lruKey, lruValue to be any entry in the cache.
  var lruKey Key
  var lruEntry *cacheEntry

  for lruKey, lruEntry = range c.cache {
    break
  }

  for key, entry := range c.cache {
    if entry.lastAccessedTimestamp < lruEntry.lastAccessedTimestamp {
      lruKey = key
      lruEntry = entry
    } 
  }

  c.sizeBytes = c.sizeBytes - lruEntry.sizeBytes
  delete(c.cache, lruKey)
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
