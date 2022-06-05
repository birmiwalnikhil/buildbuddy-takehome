package store 

import (
  "testing"
)

const (
  KEY = "a key"
  KEY2 = "another key"
  KEY3 = "the third key"
  VALUE = "some value 123"
  VALUE_THAT_FITS = "another "
  VALUE_LARGE = "..............1234567890abcdef..............."
)

func TestCacheUpdatesLruOrder(t *testing.T) {
  cache, _ := MakeCache(50)
  cache.Set(KEY, VALUE)
  cache.Set(KEY2, VALUE)
  
  // Current LRU order is KEY->KEY2
  if cache.evictionList.Front().Value.(Key) != KEY {
    t.Errorf("Invalid LRU ordering; expected %v at front", KEY)
  }

  if cache.evictionList.Back().Value.(Key) != KEY2 {
    t.Errorf("Invalid LRU ordering; expected %v at back.", KEY)
  }
}

func TestCacheSetsEntry(t *testing.T) {
  cache, _ := MakeCache(50)
  if err := cache.Set(KEY, VALUE); err != nil {
    t.Errorf("Error when setting %v->%v in cache", KEY, VALUE)
  }

  if val, err := cache.Get(KEY); err != nil || val != VALUE {
    t.Errorf("Error retrieving %v from cache", KEY)
  }
}

func TestCacheDeletesOldKey(t *testing.T) {
  cache, _ := MakeCache(50)
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
  cache, _ := MakeCache(10)
  
  if err := cache.Set(KEY, VALUE_LARGE); err == nil {
    t.Errorf("Expected error when setting %v->%v",  KEY, VALUE_LARGE)
  } 

}

func TestCacheEvictsIfNeeded(t *testing.T) {
  cache, _ := MakeCache(len(VALUE) + len(VALUE_THAT_FITS) - 1)

  cache.Set(KEY, VALUE)
  cache.Set(KEY2, VALUE_THAT_FITS) // Evicts KEY->VALUE

  if _, ok := cache.cache[KEY]; ok {
    t.Errorf("Expected eviction of %v", KEY)
  }

  if val, _ := cache.Get(KEY2); val != VALUE_THAT_FITS {
    t.Errorf("Expected %v-%v", KEY2, VALUE_THAT_FITS)
  }
}

func TestCacheEvictsLRU(t *testing.T) {
  // The cache can fit three key/value pairs.
  cache, _ := MakeCache(25)
  value1 := Value("aaaaa") // 5 bytes
  
  cache.Set("key1", value1)
  cache.Set("key2", value1)
  cache.Set("key3", value1)
  cache.Set("key4", value1)
  cache.Set("key5", value1)

  // LRU ordering is 1->2->3->4->5.
  cache.Get("key3")
  cache.Get("key2")
  
  // LRU ordering is 1->4->5->3->2.
  value2 := Value("aaaaabbbbbcccccddddd") // 20 bytes.
  cache.Set("key6", value2)

  // Values 1, 4, 5, 3 should all be evicted. 2 and 6 should be present.
  errorIfCacheContains(cache, "key1", t)
  errorIfCacheContains(cache, "key4", t)
  errorIfCacheContains(cache, "key5", t)
  errorIfCacheContains(cache, "key3", t)

  if val, err := cache.Get("key2"); err != nil || val != value1 {
    t.Errorf("Expected %v->%v in cache.", "key2", value1)
  }   

  if val, err := cache.Get("key6"); err != nil || val != value2 {
    t.Errorf("Expected %v->%v in cache.", "key6", value2)
  } 
}

func TestCacheGetUpdatesTimestamp(t *testing.T) {
  cache, _ := MakeCache(50)

  cache.Set(KEY, VALUE)
  cache.Set(KEY2, VALUE_THAT_FITS) 

  // LRU Ordering is KEY->KEY2.
  if cache.evictionList.Front().Value.(Key) != KEY {
    t.Errorf("Invalid LRU ordering; expected %v at front", KEY)
  }

  if cache.evictionList.Back().Value.(Key) != KEY2 {
    t.Errorf("Invalid LRU ordering; expected %v at back.", KEY2)
  }

  cache.Get(KEY)

  // LRU Ordering is KEY2->KEY.  
  if cache.evictionList.Front().Value.(Key) != KEY2 {
    t.Errorf("Invalid LRU ordering; expected %v at front", KEY2)
  }

  if cache.evictionList.Back().Value.(Key) != KEY {
    t.Errorf("Invalid LRU ordering; expected %v at back.", KEY)
  }
}

func errorIfCacheContains(c *Cache, key Key, t *testing.T) {
  if _, err := c.Get(key); err == nil {
    t.Errorf("Expected %v to be missing from cache.", key)
  }
}
