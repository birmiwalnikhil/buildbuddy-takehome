package store

type Key string
type Value string

var (
  EMPTY_VALUE = Value("")
)

// Utilty methods around the Value type.
func (v Value) SizeOfBytes() int { 
  return len(v)
}

type KeyValuePair struct {
  Key Key
  Value Value
}

// An interface for a KeyValue store.
type KeyValueStore interface {
  /**
   * Associate the {@code key} with the {@code value}.
   * If the key/value pair could not be set, return the error that occurred
   * (e.g. an IO failure if writing to disk).
   */
  Set(key Key, value Value) error
  
  /** 
   * Retrieve the value associated with this key, or
   * an error if no value is stored for this key.
   */
  Get(key Key) (Value, error)
}

