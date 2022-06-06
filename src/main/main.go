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

func main() {
  var fs *store.FileStore
  var cache *store.Cache

  var err error
  fs, err = store.MakeFileStore("/tmp/buildbuddy")
  if err != nil {
    fmt.Println("Error making filestore; aborting.")
    return 
  }

  
  // Optionally configure a cache.
  if (flagEnabled(flagEnableCaching, os.Args)) {
    cache, err = store.MakeCache(/* capacityBytes= */ 10)
    if err != nil {
      fmt.Println("Error making cache; aborting.")
      return
    }
  }
   
  s := server.MakeServer(fs, cache)
  c := client.MakeClient("http://localhost:8080")
  reader := bufio.NewReader(os.Stdin)
  go s.Start() // Spin the server on a background thread. 
  
  // Accept user input, and convert it into either a Get or Set.
  for {
    userInput, err := reader.ReadString('\n')
    if err != nil {
      fmt.Println("Error when reading input", err)
      return
    }

    // Split on empty space, and execute either a GET or SET call.
    tokens := strings.Fields(userInput) 
    if strings.EqualFold(userInput, "exit") {
      return
    }

    if len(tokens) < 2 {
      fmt.Println("Invalid input.")
      continue
    }

    operation := tokens[0]
    // Require exactly two tokens, e.g. GET <key>.
    if strings.EqualFold(operation, "GET") && len(tokens) == 2 {
      key := tokens[1]
      resp, err := c.Get(key)
      if err != nil {
        fmt.Println("Error getting", key, ":", err)
      } else {
        fmt.Println("GET", key, "->", string(resp)) 
      }
   } else if strings.EqualFold(operation, "SET") && len(tokens) >= 3 {
      // Require 3+ tokens, e.g. SET <key> <one space> <value with spaces>
      key := tokens[1]
      keyIdxStart := strings.Index(userInput, tokens[1])

      // The value is one character (i.e. a space) after the end of <key>. 
      value := userInput[keyIdxStart + len(tokens[1]) + 1:]
      
      if err := c.Set(key, []byte(value)); err != nil {
        fmt.Println("Error setting", key, "->", value, "error:", err)
      } else {
        fmt.Println("SET", key, "->", value)
      }
    } else {
      fmt.Println("Invalid input:", userInput)
    } 
  }
}

// Return whether the flag is enabled from the command line invocation,
// e.g. `./execute_target --enable-caching`.
func flagEnabled(flag string, args []string) bool {
  for _, value := range args {
    if value == flag  {
      return true
    }
  }
  return false
}
