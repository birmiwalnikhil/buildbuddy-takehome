package server 

import(
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "buildbuddy.takehome.com/src/store"
)

// A collection of configuration parameters
// for the server.
type Server struct {
  // The key/value store.
  store store.KeyValueStore 
}

// Restore all in-memory state of the server.
func (s *Server) restoreBackup() {
  // Do nothing; we currently have no in-memory state.
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
  resp, err := s.store.Get(store.Key(key))
  if err != nil {
    // Return a StatusInternalServerError; failure retrieving the value.
    w.WriteHeader(http.StatusInternalServerError)
    return 
  }

  // Output the value back to the caller.
  fmt.Fprintln(w, resp)
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
  
  fmt.Println("SET", string(key), "->", string(value))

  storeErr := s.store.Set(store.Key(key), store.Value(value))
  if storeErr != nil {
    // Return a StatusInternalServerError; error when storing 
    // the key/value pair.
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprintln(w, storeErr)
    return
  }
}

// Start the server. Initializes any in-memory state, then begins
// accepting API calls.
func (s *Server) Start() {
  s.restoreBackup()

  http.HandleFunc("/get", s.handleGet)
  http.HandleFunc("/set", s.handleSet)

  if err := http.ListenAndServe(":8080", nil); err != nil {
    log.Fatal(err)
  }
}

// Make a server, providing some configuration parameters.
func MakeServer(store store.KeyValueStore) *Server {
  server := &Server {}  
  server.store = store
  return server
}
