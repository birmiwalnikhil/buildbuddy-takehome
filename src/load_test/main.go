package main

import (
  "fmt"
  "sync"
  "sync/atomic"
  "buildbuddy.takehome.com/src/server"
  "buildbuddy.takehome.com/src/client"
  "buildbuddy.takehome.com/src/store"
)

// A load test for the Key/Value server.
func main() {
  // 10 GoRoutines are constantly hitting the /get API at a random cadence.
  // 10 GoRoutines are constantly hitting the /set API at a random cadence.
  // Assert that things are as expected.
  //
  var value1 atomic.Value
  
  values := []string { "value 1", "another value", "a third value", "a really long value" }
  key1 := "this is the first key"
  
  // Initialize the clients and the servers.
  fs, _ := store.MakeFileStore("/tmp/buildbuddy")
  cache := store.MakeCache(/* capacityBytes= */ 50, fs)
  
  s := server.MakeServer(cache)
  c1 := client.MakeClient()
  mu := &sync.Mutex{}

  go s.Start()
  var wg sync.WaitGroup // Will block until all GoRoutines are completed.

  // One reader GoRoutine, one writer GoRoutine.
  wg.Add(2)
  go func() {
    for i := 0; i < 100; i++ {
      val := values[i % len(values)]
      c1.Set(key1, []byte(val))
      mu.Lock()
      fmt.Println("SET ", key1, "->", val)
      mu.Unlock()
      value1.Store(values[i % len(values)])
    }
    wg.Done()
  }()

  go func() {
    for i := 0; i < 100; i++ {
      res := c1.Get(key1)
      mu.Lock()
      fmt.Println("GET ", key1, "->", string(res))
      mu.Unlock()
    }
    wg.Done()
  }()


  wg.Wait()
}
