package main

import (
  "bufio"
  "fmt"
  "strings"
  "os"
  "buildbuddy.takehome.com/src/server"
  "buildbuddy.takehome.com/src/client"
  "buildbuddy.takehome.com/src/store"
)

const (
  flagEnableCaching = "--enable_caching"
)

// Entry point for the server runner.
// HttpClients may be instantiated to interact with the Server.
func main() {
  var fs store.KeyValueStore
  var err error
  // Configure the server via command line arguments. 
  fs, err = store.MakeFileStore("/tmp/buildbuddy")
  if err != nil {
    fmt.Println("Error making filestore; aborting.")
    return 
  }

  // Optionally configure a cache.
  if (flagEnabled(flagEnableCaching, os.Args)) {
    cache := store.MakeCache(/* capacityBytes= */ 50, fs)
    fs = cache
  }
   
  s := server.MakeServer(fs)
  c := client.MakeClient()
  reader := bufio.NewReader(os.Stdin)
  go s.Start() // This is a blocking call.
  
  // Accept user input, and convert it into either a Get or Set.
  for {
    userInput, err := reader.ReadString('\n')
    if err != nil {
      fmt.Println("Error when reading input", err)
      return
    }

    // Split on empty space, and execute either a GET or SET call.
    tokens := strings.Fields(userInput) 
    operation := tokens[0]
    // Require exactly two tokens, e.g. GET <key>.
    if strings.EqualFold(operation, "GET") && len(tokens) == 2 {
      
      key := tokens[1]
      resp := c.Get(key)
      fmt.Println("GET", key, "->", string(resp)) 
   } else if strings.EqualFold(operation, "SET") && len(tokens) >= 3 {
      // Require 3+ tokens, e.g. GET <key> <value with spaces>
      key := tokens[1]
      valueIdxStart := strings.Index(userInput, tokens[2])
      value := userInput[valueIdxStart:]
      
      fmt.Println("SET", key, "->", value)
      if err := c.Set(key, []byte(value)); err != nil {
        fmt.Println("Error setting", key, "->", value, "error:", err)
      }     
    } else if operation == "exit" {
      return 
    } else {
      fmt.Println("Invalid input:", userInput)

    } 
  }
}

func flagEnabled(flag string, args []string) bool {
  for _, value := range args {
    if value == flag  {
      return true
    }
  }
  return false
}
