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
 * Re
 */
func (c *Cache) Set(key Key, value Value) error {
  if cachedEntry, ok := c.cache[key]; ok {
    // We aren't overwriting the value; update the LACT.
    if cachedEntry.value == value {    
      c.onKeyTouched(key)
      return nil
    } else {
      // Overwrite the value by deleting then adding.
      c.sizeBytes = c.sizeBytes - cachedEntry.sizeBytes
      delete(c.cache, key) 
    }
  }

  return c.addEntry(key, value) 
}

/**
 * Retrieve the key/value from memory, or return an error if the value is
 * missing. 
 */
func (c *Cache) Get(key Key) (Value, error) {
  if entry, ok := c.cache[key]; ok {
    fmt.Println("\tCache hit!")
    c.onKeyTouched(key); 
    return entry.value, nil
  } else {
  }
 
  fmt.Println("\tCache miss!")
  return "", errors.New(fmt.Sprintf("Cache miss for %v", key))
}

// Add a key/value pair to the cache, performing LRU if possible.
// Return an error if the key/value pair was not added to memory.
func (c *Cache) addEntry(key Key, value Value) error {
  entry := &cacheEntry{}
  entry.value = value
  entry.sizeBytes = value.SizeOfBytes()
  entry.lastAccessedTimestamp = c.getAndIncreaseTimestamp()
  if entry.sizeBytes >= c.capacityBytes {
    // We cannot store this entry in memory. Do nothing.
    return errors.New(fmt.Sprintf("Value too large; cannot store %v->%v in cache of size %v", key, value, c.capacityBytes))
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
func MakeCache(capacityBytes int) (*Cache, error) {
  if capacityBytes <= 0 {
    return nil, errors.New(fmt.Sprintf("Cannot create a cache of capacity %v",
capacityBytes))
  }
  
  c := &Cache{}

  c.capacityBytes = capacityBytes
  c.sizeBytes = 0
  c.timestamp = 0
  c.cache = make(map[Key]*cacheEntry) 

  return c, nil
}

