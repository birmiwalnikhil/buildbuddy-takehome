package store 

import (
  "testing"
)

const (
  KEY = "a key"
  KEY2 = "another key"
  VALUE = "some value 123"
  VALUE_THAT_FITS = "another "
  VALUE_LARGE = "..............1234567890abcdef..............."
)

func TestCacheUpdatesAccessedTime(t *testing.T) {
  cache := MakeCache(50)
  cache.Set(KEY, VALUE)

  if cache.cache[KEY].lastAccessedTimestamp != 0 {
    t.Errorf("Invalid timestamp for %v, expected %v", KEY, 0)
  }
  
  cache.Set(KEY, VALUE_THAT_FITS)
  if cache.cache[KEY].lastAccessedTimestamp != 1 {
    t.Errorf("Invalid timestamp for %v, expected %v", KEY, 1)
  }
}

func TestCacheSetsEntry(t *testing.T) {
  cache := MakeCache(50)
  if err := cache.Set(KEY, VALUE); err != nil {
    t.Errorf("Error when setting %v->%v in cache", KEY, VALUE)
  }

  if val, err := cache.Get(KEY); err != nil || val != VALUE {
    t.Errorf("Error retrieving %v from cache", KEY)
  }
}

func TestCacheDeletesOldKey(t *testing.T) {
  cache := MakeCache(50)
  cache.Set(KEY, VALUE)

  if val, _ := cache.Get(KEY); val != VALUE {
    t.Errorf("Expected a %v->%v store", KEY, VALUE)
  }

  cache.Set(KEY, VALUE_THAT_FITS)
  if val, _ := cache.Get(KEY); val != VALUE_THAT_FITS {
    t.Errorf("Expected a %v->%v store", KEY, VALUE_THAT_FITS)
  }  
}

func TestCacheAddsEntryThrowsValueTooLarge(t *testing.T) {
  cache := MakeCache(10)
  
  if err := cache.Set(KEY, VALUE_LARGE); err == nil {
    t.Errorf("Expected error when setting %v->%v",  KEY, VALUE_LARGE)
  } 

}

func TestCacheEvictsIfNeeded(t *testing.T) {
  cache := MakeCache(len(VALUE) + len(VALUE_THAT_FITS) - 1)

  cache.Set(KEY, VALUE)
  cache.Set(KEY2, VALUE_THAT_FITS) // Evicts KEY->VALUE

  if _, ok := cache.cache[KEY]; ok {
    t.Errorf("Expected eviction of %v", KEY)
  }

  if val, _ := cache.Get(KEY2); val != VALUE_THAT_FITS {
    t.Errorf("Expected %v-%v", KEY2, VALUE_THAT_FITS)
  }
}

func TestCacheGetUpdatesTimestamp(t *testing.T) {
  cache := MakeCache(50)

  cache.Set(KEY, VALUE)
  cache.Get(KEY)
  
  if entry, _ := cache.cache[KEY]; entry.lastAccessedTimestamp != 1 {
    t.Errorf("Expected GET to increment the accessed timestamp of %v", KEY)
  }
}

