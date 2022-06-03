package server 

import(
  "fmt"
  "log"
  "net/http"
)

// A collection of configuration parameters
// for the server.
type Server struct {}

// Restore all in-memory state of the server.
func (s *Server) restoreBackup() {
  // Do nothing; we currently have no in-memory state.
}

// Handler for a /get call.
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintln(w, "Server handling GET")
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
