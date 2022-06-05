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

func TestCacheSetsUpstream(t *testing.T) {
  upstream := &FakeKeyValueStore{}
  cache := MakeCache(10, upstream)
  cache.Set(KEY, VALUE)
  
  if len(upstream.SetCalls) != 1 || upstream.SetCalls[0].Key != KEY ||
upstream.SetCalls[0].Value != VALUE {
    t.Errorf("Missing upstream set; expected %v->%v", KEY, VALUE)
  }
}

func TestCacheUpdatesAccessedTime(t *testing.T) {
  upstream := &FakeKeyValueStore{}
  cache := MakeCache(50, upstream)
  cache.Set(KEY, VALUE)

  if cache.cache[KEY].lastAccessedTimestamp != 0 {
    t.Errorf("Invalid timestamp for %v, expected %v", KEY, 0)
  }
  
  cache.Set(KEY, VALUE_THAT_FITS)
  if cache.cache[KEY].lastAccessedTimestamp != 1 {
    t.Errorf("Invalid timestamp for %v, expected %v", KEY, 1)
  }
}

func TestCacheDeletesOldKey(t *testing.T) {
  upstream := &FakeKeyValueStore{}
  cache := MakeCache(50, upstream)
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
  upstream := &FakeKeyValueStore{}
  cache := MakeCache(10, upstream)
  
  if err := cache.Set(KEY, VALUE_LARGE); err != VALUE_TOO_LARGE {
    t.Errorf("Expected %v when setting %v->%v", VALUE_TOO_LARGE, KEY,
VALUE_LARGE)
  } 

}

func TestCacheEvictsIfNeeded(t *testing.T) {
  upstream := &FakeKeyValueStore{}
  cache := MakeCache(len(VALUE) + len(VALUE_THAT_FITS) - 1, upstream)

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
  upstream := &FakeKeyValueStore{}
  cache := MakeCache(50, upstream)

  cache.Set(KEY, VALUE)
  cache.Get(KEY)
  
  if entry, _ := cache.cache[KEY]; entry.lastAccessedTimestamp != 1 {
    t.Errorf("Expected GET to increment the accessed timestamp of %v", KEY)
  }
}

func TestCacheUsesUpstreamOnCacheMiss(t *testing.T) {
  upstream := &FakeKeyValueStore{}
  cache := MakeCache(50, upstream)

  cache.Get(KEY)
  if len(upstream.GetCalls) != 1 || upstream.GetCalls[0] != KEY {
    t.Errorf("Expected upstream GET call on cache miss of %v", KEY)
  }
}

func TestCacheUpdatesCacheFromUpstream(t *testing.T) {
  upstream := &FakeKeyValueStore{}
  cache := MakeCache(50, upstream)

  upstream.SetNextGet(VALUE, nil)
  cache.Get("asd")
  
  if entry, ok := cache.cache["asd"]; !ok || entry.value != VALUE {
    t.Errorf("Expected upstream to populate the in memory cache for %v", KEY)
  } 
}
