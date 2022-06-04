package client

import (
  "bytes"
  "encoding/json"
  "errors"
  "fmt"
  "io/ioutil"
  "net/http"
)

var (
  EMPTY_BUFFER []byte
)

// A collection of configuration parameters for 
// an HTTP Client that interacts with the server.
type Client struct {
    httpClient *http.Client
}

/** 
 * A utility method for the HttpClient to invoke 
 * /get on the API server. Returns the stored
 * value from a previous /set call, or an error 
 * on failure.
 */
func (c *Client) Get(key string) []byte {
  req, err := http.NewRequest("GET", "http://localhost:8080/get", nil)
  if err != nil {
    return EMPTY_BUFFER
  }

  // Add the key as a query parameter to the request.
  query := req.URL.Query()
  query.Add("key", key)
  req.URL.RawQuery = query.Encode()

  // Execute the request.
  resp, err := c.httpClient.Do(req) 
  if err != nil {
    return EMPTY_BUFFER
  }
  
  if resp.StatusCode != http.StatusOK {
    // The server was not able to service this request.
    return EMPTY_BUFFER
  }

  defer resp.Body.Close()
  buffer, err := ioutil.ReadAll(resp.Body)

  if err != nil {
    // Error reading the response body.
    return EMPTY_BUFFER
  }
  
  return buffer
}

/** 
 * A utility method for the HttpClient to invoke /set 
 * on the API server. Returns an error when invoking the request,
 * if any.
 */
func (c *Client) Set(key string, value []byte) error {
  // Marshal the key/value pair into a JSON.
  kv := map[string][]byte {
    "key": []byte(key),
    "value": value,
  }
  
  jsonKv, err := json.Marshal(kv)
  if err != nil {
    return err
  }
   
  resp, postErr := 
    http.Post(
      "http://localhost:8080/set", "application/json", bytes.NewBuffer(jsonKv)) 
  if postErr != nil {
    return postErr
  }

  if resp.StatusCode != http.StatusOK {
    return errors.New(
      fmt.Sprintf("HttpError %v when setting %v -> %v", resp.StatusCode, key, value))
  }

  return nil
}

func MakeClient() *Client {
  c := &Client {}

  c.httpClient = &http.Client {}

  return c
}
