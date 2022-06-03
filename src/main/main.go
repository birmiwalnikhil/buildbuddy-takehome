package main

import (
  "buildbuddy.takehome.com/src/server"
)


// Entry point for the server runner.
// HttpClients may be instantiated to interact with the Server.
func main() {
  s := &server.Server {}
  s.Start() // This is a blocking call.
}
