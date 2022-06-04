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


// Entry point for the server runner.
// HttpClients may be instantiated to interact with the Server.
func main() {
  // Configure the server via command line arguments.
  s := server.MakeServer(store.MakeFileStore("/tmp/buildbuddy"))
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
    if operation == "GET" && len(tokens) == 2 {
      key := tokens[1]
      resp := c.Get(key)
      fmt.Println("GET", key, "->", string(resp)) 
   } else if operation == "SET" && len(tokens) == 3 {
      key := tokens[1]
      value := tokens[2]
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
