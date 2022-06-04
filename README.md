# buildbuddy-takehome

A key/value store implemented on a Go server which accepts arbitrary length
inputs. The server runs on `localhost:8080` and can be started via `go run
./src/main/` from the root directory. There are two APIs exposed to clients:

1) `/set`. A HTTP Post method which stores a key/value pair from the POST Body.
2) `/get/<key>`. Returns the value of a previously `/set/` key/value pair. 

The key/value store is recovery resistant: server resets will continue to operate.

The following optimizations can be enabled via command line flags:
- `--enable_caching`: Enables an in-memory cache
