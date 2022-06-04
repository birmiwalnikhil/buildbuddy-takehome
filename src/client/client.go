package client

import (
  "bytes"
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
  _, err := 
    http.Post(
      "http://localhost:8080/set", "application/json", bytes.NewBuffer(value)) 
  if err != nil {
    return err
  }

  // TODO: Case on the HttpErrorCode
  return nil
}

func MakeClient() *Client {
  c := &Client {}

  c.httpClient = &http.Client {}

  return c
}
