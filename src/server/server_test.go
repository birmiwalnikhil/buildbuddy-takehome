package server

import (
  "fmt"
  "testing"
  "net/http"
  "net/http/httptest"
  
  "buildbuddy.takehome.com/src/store"
)

func TestGetMissingKeyReturns400(t *testing.T) { 
  fs := &store.FakeKeyValueStore{} 
  req := httptest.NewRequest("GET", "http://localhost:8080/get", nil)
  w := httptest.NewRecorder()

  s := MakeServer(fs, nil)
  s.handleGet(w, req)
 
  if w.Result().StatusCode != http.StatusBadRequest {
    t.Errorf(fmt.Sprintf("Expected error code %v for this req %v",
http.StatusBadRequest, req))
  } 
}

func TestCacheGetTriedCache(t *testing.T) {
  cache := &store.FakeCache{}
  fs := &store.FakeKeyValueStore{}
  req := httptest.NewRequest("GET", "http://localhost:8080/get", nil)
  w := httptest.NewRecorder()
  s := MakeServer(fs, cache)
  
  // Add a key to the query.
  key := "this is a key"
  query := req.URL.Query()
  query.Add("key", key) 
  req.URL.RawQuery = query.Encode()

  s.handleGet(w, req)
  
  if len(cache.GetCalls) != 1 || cache.GetCalls[0] != store.Key(key) {
    t.Errorf("Expected a cache GET call of %v", key) 
  }
}

