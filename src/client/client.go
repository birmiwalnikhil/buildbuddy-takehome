package client

import (
  "log"
  "io/ioutil"
  "net/http"
)

var (
  emptyBuffer []byte
)

// A collection of configuration parameters for 
// an HTTP Client that interacts with the server.
type Client struct {
    httpClient *http.Client
}

/** 
 * A utility method for the HttpClient to invoke 
 * /get from the API server. Returns the stored
 * value from a previous /set call, or an error 
 * on failure.
 */
func (c *Client) Get (key string) ([]byte, error) {
  req, err := http.NewRequest("GET", "http://localhost:8080/get", nil)
  if err != nil {
    log.Fatal(err)
    return emptyBuffer, err
  }

  // Add the key as a query parameter to the request.
  query := req.URL.Query()
  query.Add("key", key)
  req.URL.RawQuery = query.Encode()

  // Execute the request.
  resp, err := c.httpClient.Do(req) 
  if err != nil {
    log.Fatal(err)
    return emptyBuffer, err
  }

  defer resp.Body.Close()
  return ioutil.ReadAll(resp.Body)
}
