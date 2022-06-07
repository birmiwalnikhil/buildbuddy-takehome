package server

import (
  "bytes"
  "errors"
  "encoding/json"
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
  s := &Server {
    filestore: nil,
    cache: nil,
    mutex: &sync.Mutex{},
  }

  w := httptest.NewRecorder()
  req := httptest.NewRequest("POST", "http://localhost:8080/set", nil)
  req.Header.Set("Content-Type", "application/json")

  s.handleSet(w, req)

  // Failure parsing the POST body.
  if w.Result().StatusCode != http.StatusInternalServerError {
    t.Errorf("Expected http %v, received %v", http.StatusInternalServerError, w.Result().StatusCode)
  } 
}

func TestSetFilestoreFailureReturns500(t *testing.T) { 
  fs := &store.FakeKeyValueStore{}
  s := &Server {
    filestore: fs,
    cache: nil,
    mutex: &sync.Mutex{},
  }
  
  fs.SetNextSet(errors.New("File store SET error."))

  w := httptest.NewRecorder()
  kv := make(map[string][]byte)
  kv["key"] = []byte("a key")
  kv["value"] = []byte("a value")

  jsonKv, _ := json.Marshal(&kv)
  req2, _ := http.NewRequest("POST", "http://localhost:8080/set", bytes.NewBuffer(jsonKv))

  s.handleSet(w, req2)
  if len(fs.SetCalls) != 1 || fs.SetCalls[0].Key != store.Key("a key") ||
fs.SetCalls[0].Value != store.Value("a value") {
    t.Errorf("Expected file store set call.")
  }

  if w.Result().StatusCode != http.StatusInternalServerError {
    t.Errorf("Expected http %v, received %v", http.StatusInternalServerError,
w.Result().StatusCode)
  }
}

func TestSetFilestoreSuccessSynchronizesCache(t *testing.T) {
  fs := &store.FakeKeyValueStore{}
  cache := &store.FakeKeyValueStore{}
  s := &Server {
    filestore: fs,
    cache: cache,
    mutex: &sync.Mutex{},
  }
  
  w := httptest.NewRecorder()
  kv := make(map[string][]byte)
  kv["key"] = []byte("a key")
  kv["value"] = []byte("a value")

  jsonKv, _ := json.Marshal(&kv)
  req2, _ := http.NewRequest("POST", "http://localhost:8080/set", bytes.NewBuffer(jsonKv))

  s.handleSet(w, req2)
  if len(fs.SetCalls) != 1 || fs.SetCalls[0].Key != store.Key("a key") ||
fs.SetCalls[0].Value != store.Value("a value") {
    t.Errorf("Expected file store set call.")
  }

  if len(cache.SetCalls) != 1 || cache.SetCalls[0].Key != store.Key("a key") ||
cache.SetCalls[0].Value != store.Value("a value") {
    t.Errorf("Expected cache set call.")
  }

  if w.Result().StatusCode != http.StatusOK {
    t.Errorf("Expected http %v, received %v", http.StatusOK,
w.Result().StatusCode)
  }
}
