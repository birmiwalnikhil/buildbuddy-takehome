package server 

import(
  "fmt"
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

// Handler for a /set call.
func (s *Server) handleSet(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintln(w, "Server handling SET")
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
