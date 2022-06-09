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

// A thin wrapper around a HTTP Client. Used to 
// marshal Get and Set calls to the API server.
type Client struct {
    // The URL of the Get Endpoint, e.g. `http://localhost:8080/get`.
    getUrl string
    // The URL of the Set Endpoint, e.g. `http://localhost:8080/set`.
    setUrl string
    httpClient *http.Client
}

type KeyValuePair struct {
  Key string
  Value string
}

/** 
 * Invoke a /get request for a specified `key` on the API server. Returns the 
 * stored value, if any, or any errors (e.g. a network connection failure, an 
 * HTTP error code, etc.) 
*/
func (c *Client) Get(key string) ([]byte, error) {
  if len(key) == 0 {
    return EMPTY_BUFFER, errors.New("GET cannot be called on an empty key.")
  }

  req, err := http.NewRequest("GET", c.getUrl, nil)
  if err != nil {
    return EMPTY_BUFFER, err
  }

  // Add the key as a query parameter to the request.
  query := req.URL.Query()
  query.Add("key", key)
  req.URL.RawQuery = query.Encode()

  // Execute the request.
  resp, err := c.httpClient.Do(req) 
  if err != nil {
    return EMPTY_BUFFER, err
  }
  
  if resp.StatusCode != http.StatusOK {
    // The server was not able to service this request.
    return EMPTY_BUFFER, errors.New(fmt.Sprintf("HttpError %v from server",
resp.StatusCode))
  }

  defer resp.Body.Close()
  buffer, err := ioutil.ReadAll(resp.Body)

  if err != nil {
    // Error reading the response body.
    return EMPTY_BUFFER, err
  }
  
  return buffer, nil
}

/**
 * Invoke the /set API with the provided `key`->`value` pair. Return any
 * failures (e.g. a malformed request, a connection failure, etc.) or nil
 * otherwise.
 */
func (c *Client) Set(key string, value []byte) error {
  if len(key) == 0 {
    return errors.New("Cannot SET an empty key.")
  }

  if len(value) == 0 {
    return errors.New("Cannot SET an empty value.")
  }

  // Marshal the key/value pair into a JSON.
  kv := &KeyValuePair {
    Key: key,
    Value: string(value),
  }
 
  jsonKv, err := json.Marshal(kv)
  if err != nil {
    return err
  }

  // Execute the /set request.   
  resp, postErr := 
    http.Post(c.setUrl, "application/json", bytes.NewBuffer(jsonKv)) 
  if postErr != nil {
    return postErr
  }

  if resp.StatusCode != http.StatusOK {
    return errors.New(
      fmt.Sprintf("HttpError %v when setting %v -> %v", resp.StatusCode, key, value))
  }

  return nil
}

// Construct Client instances.
func MakeClient(serverUrl string) *Client {
  c := &Client {}

  c.httpClient = &http.Client {}
  c.getUrl = fmt.Sprintf("%s/get", serverUrl)
  c.setUrl = fmt.Sprintf("%s/set", serverUrl)

  return c
}
