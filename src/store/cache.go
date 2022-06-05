package store

import (
  "container/list"
  "errors"
  "fmt"
  "sync"
)

// A value and cache-relevant metadata, e.g. its eviction order priority.
type cacheEntry struct {
  value Value
  evictionListElement *list.Element
  sizeBytes int 
}

// An LRU cache that supports a Key/Value store.
type Cache struct {
 // The maximum size of the cache, in bytes.
  capacityBytes int
  // The current size of the cache, in bytes.
  sizeBytes int
  // An in-memory key/value store.
  cache map[Key]*cacheEntry 
  // A doubly-linked ordered list of Keys. Ordered by eviction priority;
  // The first element should be evicted first, and the last element should be
  // evicted last. 
  evictionList *list.List
  // A mutex to allow multiple GoRoutines to utilize the cache.
  // Note that we cannot use a RW lock; there may be contention if multiple
  // GET threads are modifying the eviction list.
  mutex *sync.Mutex
}

/** 
 * Set the key/value pair in memory, possibly performing eviction if need be. 
 */
func (c *Cache) Set(key Key, value Value) error {
  defer c.mutex.Unlock()
  c.mutex.Lock()
  if cachedEntry, ok := c.cache[key]; ok {
    // We aren't overwriting the value; update the eviction preference
    // of this key.
    if cachedEntry.value.Equals(value) {    
      c.onKeyTouched(key)
      return nil
    } else {
      // Overwrite the value by deleting then adding.
      c.sizeBytes = c.sizeBytes - cachedEntry.sizeBytes
      delete(c.cache, key) 
    }
  }

  // Add the key/value pair to the cache, performing LRU if needed.
  if value.SizeOfBytes() >= c.capacityBytes {
    // We cannot store this entry in memory. Do nothing.
    return errors.New(fmt.Sprintf("Value too large; cannot store %v->%v in cache of size %v", key, value, c.capacityBytes))
  }

  // Continually evict LRU elements until we have sufficient space.
  for c.sizeBytes + value.SizeOfBytes() > c.capacityBytes {
      err := c.evictLru()
      if err != nil { 
        return err
      }
  }

  entry := &cacheEntry{}
  entry.value = value
  entry.sizeBytes = value.SizeOfBytes()
  entry.evictionListElement = c.evictionList.PushBack(key)

  c.sizeBytes = c.sizeBytes + entry.sizeBytes
  c.cache[key] = entry
  return nil 
}

/**
 * Retrieve the key/value from memory, or return an error if the value is
 * missing. 
 */
func (c *Cache) Get(key Key) (Value, error) {
  defer c.mutex.Unlock()
  c.mutex.Lock()
  if entry, ok := c.cache[key]; ok {
    fmt.Println("\tCache hit!")
    c.onKeyTouched(key); 
    return entry.value, nil
  }
 
  fmt.Println("\tCache miss!")
  return "", errors.New(fmt.Sprintf("Cache miss for %v", key))
}

// Evict the least recently used key from the cache, returning 
// an error in case of failure.
func (c *Cache) evictLru() error {
  frontElement := c.evictionList.Front()
  if frontElement == nil {
    return errors.New("Cannot evict an empty cache.")
  }

  key := c.evictionList.Remove(frontElement).(Key)
  cacheEntry, ok := c.cache[key]
  if !ok {
    return errors.New(fmt.Sprintf("Key %v missing from cache during eviction.", key))
  }

  c.sizeBytes = c.sizeBytes - cacheEntry.sizeBytes
  delete(c.cache, key)
  return nil
}

func (c *Cache) onKeyTouched(key Key) error {
  entry, ok := c.cache[key]
  if !ok {
    return errors.New(fmt.Sprintf("Key missing in cache %v", key))
  }

  if entry.evictionListElement == nil {
    return errors.New(fmt.Sprintf("Nil pointer in eviction list for %v", key))
  }

  c.evictionList.MoveToBack(entry.evictionListElement)
  return nil
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
  c.cache = make(map[Key]*cacheEntry) 
  c.evictionList = list.New()
  c.mutex = &sync.Mutex{}

  return c, nil
}

