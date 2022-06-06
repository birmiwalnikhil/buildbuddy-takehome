package server 

import(
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "buildbuddy.takehome.com/src/store"
)

// An HTTP Server that supports GET and SET operations. There may be multiple
// GET operations running at the same time, but only one SET may be executed.
type Server struct {
  // The file store.
  store *store.FileStore 
  // An optionally enabled cache.
  cache *store.Cache
}

// Handler for a /get call. Reads a key/value pair from the underlying
// store, and returns the value.
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
  // Extract the query parameter `key`.
  query := r.URL.Query()
  keyQuery, ok := query["key"]
  if !ok || len(keyQuery) != 1 {
    // Return an StatusBadRequest; the query parameter `key` is malformed.
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  key := keyQuery[0]

  // First check the cache to see if the value is present.
  if s.cache != nil {
    if value, err := s.cache.Get(store.Key(key)); err == nil {
      fmt.Fprintln(w, value)
      return
    }
  }

  // Retrieve the value from the filestore.
  value, err := s.store.Get(store.Key(key))
  if err != nil {
    // Return a StatusNotFoundError; failure retrieving the value.
    w.WriteHeader(http.StatusNotFound)
    return 
  }

  // Write the value into the cache. Any errors here are non-fatal; they should
  // be logged to Telemetry. TODO: Migrate this logic off the critical path of
  // GET.
  if s.cache != nil {
    if cacheSetErr := s.cache.Set(store.Key(key), store.Value(value)); 
        cacheSetErr != nil {
        // Log this error to telemetry.
        fmt.Println("\tCache error:", cacheSetErr) 
    }
  }
  // Output the value back to the caller.
  fmt.Fprintln(w, value)
}

// Handler for a /set call. The HTTP Body is a JSON containing a 
// Key/Value Pair (e.g. { "key" : "a key", "value": "an arbitrary value" })
func (s *Server) handleSet(w http.ResponseWriter, r *http.Request) {
  // Unmarshal the POST Body into the key/value pair.
  defer r.Body.Close()
  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
    // Return a StatusInternalServerError; error reading the POST body.
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  var kv map[string][]byte
  if err := json.Unmarshal(body, &kv); err != nil {
    // Return a StatusInternalServerError; error unmarshaling the POST body.
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  key, ok1 := kv["key"]
  if !ok1 {
    // Return a StatusBadRequest; POST body missing the key.
    w.WriteHeader(http.StatusBadRequest)
    return 
  }

  value, ok2 := kv["value"]
  if !ok2 {
    // Return a StatusBadRequest; POST body missing the value.
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  // Attempt to write the value to the filestore.
  if err := s.store.Set(store.Key(key), store.Value(value)); err != nil {
    // Failure writing to fliestore; return a 500.
    w.WriteHeader(http.StatusInternalServerError)
    return
  } 
  
  // Maintain consistency between the cache and the filestore.
  // Any errors thrown here are non-fatal; they should be logged to telemetry.
  if s.cache != nil {  
    if err := s.cache.Set(store.Key(key), store.Value(value)); err != nil {
      // Log a caching failure to telemetry.
      fmt.Println("\tCache Set error:", err)
    }
  }
}

// Start the server. Initializes any in-memory state, then begins
// accepting API calls.
func (s *Server) Start() {
  http.HandleFunc("/get", s.handleGet)
  http.HandleFunc("/set", s.handleSet)

  if err := http.ListenAndServe(":8080", nil); err != nil {
    log.Fatal(err)
  }
}

// Make a server, providing some configuration parameters.
func MakeServer(fs *store.FileStore, cache *store.Cache) *Server {
  server := &Server {}  
  server.store = fs
  server.cache = cache 
  return server
}
