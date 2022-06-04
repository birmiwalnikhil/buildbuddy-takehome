package store

type Key string
type Value string

// Utilty methods around the Value type.
func (v Value) SizeOfBytes() int { 
  return len(v)
}


// An interface for a KeyValue store.
type KeyValueStore interface {
  /**
   * Associate the {@code key} with the {@code value}.
   * Return any error encountered (e.g. an IO failure 
   * if writing to disk), or nil if the operation was succesful.
   */
  Set(key Key, value Value) error
  
  /** 
   * Retrieve the value associated with this key, or
   * an error if no value is stored for this key.
   */
  Get(key Key) (Value, error)
}

