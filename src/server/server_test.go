package server

import (
  "errors"
  "fmt"
  "sync"
  "testing"
  "net/http"
  "net/http/httptest"
  
  "buildbuddy.takehome.com/src/store"
)

func TestGetMissingKeyReturns400(t *testing.T) { 
  fs := &store.FakeKeyValueStore{} 
  req := httptest.NewRequest("GET", "http://localhost:8080/get", nil)
  w := httptest.NewRecorder()

  s := &Server {
    filestore: fs,
    cache: nil,
    mutex: &sync.Mutex{},
  }  
  s.handleGet(w, req)
 
  if w.Result().StatusCode != http.StatusBadRequest {
    t.Errorf(fmt.Sprintf("Expected error code %v for this req %v",
http.StatusBadRequest, req))
  } 
}

func TestGetReturnsCacheHit(t *testing.T) {
  cache := &store.FakeKeyValueStore{}
  fs := &store.FakeKeyValueStore{}
  req := httptest.NewRequest("GET", "http://localhost:8080/get", nil)
  w := httptest.NewRecorder()
  s := &Server {
    filestore: fs,
    cache: cache,
    mutex: &sync.Mutex{},
  }  

  // Specify a cache hit.
  value := "value"
  cache.SetNextGet(store.Value(value), nil)

  // Add a key to the query.
  key := "this is a key"
  query := req.URL.Query()
  query.Add("key", key) 
  req.URL.RawQuery = query.Encode()

  s.handleGet(w, req)
  
  if len(cache.GetCalls) != 1 || cache.GetCalls[0] != store.Key(key) {
    t.Errorf("Expected a cache GET call of %v", key) 
  }

  if w.Result().StatusCode != http.StatusOK {
    t.Errorf("Expected 200 on cache hit.")
  }

  if string(w.Body.Bytes()) != value {
    t.Errorf(
      "Expected cache result= received %v, expected %v", string(w.Body.Bytes()), value)
  }
}

func TestGetTriesFilestoreOnCacheMiss(t *testing.T) {
  cache := &store.FakeKeyValueStore{}
  fs := &store.FakeKeyValueStore{}
  req := httptest.NewRequest("GET", "http://localhost:8080/get", nil)
  w := httptest.NewRecorder()
  s := &Server {
    filestore: fs,
    cache: cache,
    mutex: &sync.Mutex{},
  }  

  key := "this is a key"

  // Specify a cache miss.
  err := errors.New("Cache miss!")
  cache.SetNextGet(store.Value(""), err)

  // Add a key to the query.
  query := req.URL.Query()
  query.Add("key", key)
  req.URL.RawQuery = query.Encode()

  s.handleGet(w, req)

  if len(cache.GetCalls) != 1 || cache.GetCalls[0] != store.Key(key) {
    t.Errorf("Expected a cache GET call of %v", key)
  }

  if len(fs.GetCalls) != 1 || fs.GetCalls[0] != store.Key(key) {
    t.Errorf("Expected a filestore GET call of %v", key)
  }
}

func TestGetReportsFilestoreFailure(t *testing.T) {
  fs := &store.FakeKeyValueStore{}
  req := httptest.NewRequest("GET", "http://localhost:8080/get", nil)
  w := httptest.NewRecorder()
  s := &Server {
    filestore: fs,
    cache: nil,
    mutex: &sync.Mutex{},
  }

  // Add a key to the query.
  key := "key"
  query := req.URL.Query()
  query.Add("key", key)
  req.URL.RawQuery = query.Encode()
  
  // Specify a filestore miss.
  fs.SetNextGet("", errors.New("Filestore error!"))

  s.handleGet(w, req)

  if w.Result().StatusCode != http.StatusNotFound {
    t.Errorf(
      "Expected http status code %v, got %v", 
      http.StatusNotFound,
      w.Result().StatusCode)
  }
}

func TestGetSynchronizesFilestoreAndCache(t *testing.T) {
  fs := &store.FakeKeyValueStore{}
  c := &store.FakeKeyValueStore{}
  req := httptest.NewRequest("GET", "http://localhost:8080/get", nil)
  w := httptest.NewRecorder()
  s := &Server {
    filestore: fs,
    cache: c,
    mutex: &sync.Mutex{},
  }


  // Add a key to the query.
  key := "key"
  value := "value"
  query := req.URL.Query()
  query.Add("key", key)
  req.URL.RawQuery = query.Encode()
 
  c.SetNextGet(store.Value(""), errors.New("cache miss")) 
  // Specify a filestore hit.
  fs.SetNextGet(store.Value(value), nil)

  s.handleGet(w, req)

  if len(c.SetCalls) != 1 || c.SetCalls[0].Key != store.Key(key) ||
c.SetCalls[0].Value != store.Value(value) {
    t.Errorf("Expected 1 SET cache call with %v->%v", key, value)
  }
}

func TestSetRequiresKeyValuePair (t *testing.T) {
}

func TestSetFilestoreFailureReturns500(t *testing.T) {}

func TestSetFilestoreSuccessSynchronziesCache(t *testing.T) {}
