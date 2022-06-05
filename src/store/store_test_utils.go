package store

type KeyValuePair struct {
  Key Key
  Value Value
}

// A collection of testing utilities for the KeyValueStore.
type FakeKeyValueStore struct {
  // An ordered list of Get calls.
  GetCalls []Key
  // An ordered list of Set calls.
  SetCalls []*KeyValuePair
}

func (f *FakeKeyValueStore) Get(key Key) (Value, error) {
  f.GetCalls = append(f.GetCalls, key)
  return "", nil
}

func (f *FakeKeyValueStore) Set(key Key, value Value) error {
  f.SetCalls = append(f.SetCalls, &KeyValuePair{ Key: key, Value: value })
  return nil
}
