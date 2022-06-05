package store 

import (
  "testing"
)

func TestCacheSetsUpstream(t *testing.T) {
  key := "a key"
  value := "some value 123"
  upstream := &FakeKeyValueStore{}
  cache := MakeCache(10, upstream)
  cache.Set("a key", "some value 123")
  
  if len(upstream.SetCalls) != 1 || upstream.SetCalls[0].Key != "a key" ||
upstream.SetCalls[0].Value != "some value 123" {
    t.Errorf("Missing upstream set; expected %v->%v", key, value)
  }
}

func TestCacheUpdatesAccessedTime(t *testing.T) {}

func TestCacheDeletesOldKey(t *testing.T) {}

func TestCacheAddsEntryIfFits(t *testing.T) {}

func TestCacheAddsEntryThrowsValueTooLarge(t *testing.T) {}

func TestCacheEvictsIfNeeded(t *testing.T) {}

func TestCacheGetFromMemory(t *testing.T) {}

func TestCacheGetUpdatesTimestamp(t *testing.T) {}

func TestCacheUsesUpstreamOnCacheMiss(t *testing.T) {}

func TestCacheUpdatesCacheFromUpstream(t *testing.T) {}
