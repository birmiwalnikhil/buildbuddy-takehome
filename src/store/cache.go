package store

import (
  "errors"
  "fmt"
)

var (
  // An error to report when a Value cannot be stored in the cache because
  // the value is too large.
  VALUE_TOO_LARGE = 
    errors.New("Value too large; will not be stored in cache.")
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
  // An in-memory key/value store.
  cache map[Key]*cacheEntry 
}

/** 
 * Set the key/value pair in memory, possibly performing eviction. 
 */
func (c *Cache) Set(key Key, value Value) error {
  // Maintain the upstream and the cache in sync.
  // On failure, do not set the key in the in-memory store either.
  if err := c.upstream.Set(key, value); err != nil {
    return err
  }

  if cachedValue, ok := c.cache[key]; ok {
    if cachedValue.value == value {    
      c.onKeyTouched(key)
      return nil
    } else {
      // Delete the value from the cache to maintain consistency with upstream.
      c.sizeBytes = c.sizeBytes - cachedValue.sizeBytes
      delete(c.cache, key) 
    }
  }

  return c.addEntry(key, value) 
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

  // Attempt to store the value into the cache.
  // If adding to the cache fails, log an error to Telemetry but 
  // return the value to the caller.
  c.addEntry(key, value) // Return value ignored.
  
  return value, nil
}

// Add a key/value pair to the cache.
func (c *Cache) addEntry(key Key, value Value) error {
  entry := &cacheEntry{}
  entry.value = value
  entry.sizeBytes = value.SizeOfBytes()
  entry.lastAccessedTimestamp = c.getAndIncreaseTimestamp()
  
  if entry.sizeBytes >= c.capacityBytes {
    // We cannot store this entry in memory. Do nothing.
    return VALUE_TOO_LARGE 
  }

  // Continually evict LRU elements until we have sufficient space.
  for c.sizeBytes + entry.sizeBytes >= c.capacityBytes {
      err := c.evictLru()
      if err != nil { 
        return err
      }
  }

  c.sizeBytes = c.sizeBytes + entry.sizeBytes
  c.cache[key] = entry
  return nil 
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
    return errors.New(fmt.Sprintf("Key missing in cache %v", key))
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
func MakeCache(capacityBytes int, upstream KeyValueStore) *Cache {
  c := &Cache{}

  c.upstream = upstream
  c.capacityBytes = capacityBytes
  c.sizeBytes = 0
  c.timestamp = 0
  c.cache = make(map[Key]*cacheEntry) 

  return c
}
