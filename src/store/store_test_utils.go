package store

type KeyValuePair struct {
  Key Key
  Value Value
}

// A collection of testing utilities for the KeyValueStore.
type FakeKeyValueStore struct {
  // An ordered list of Get calls.
  GetCalls []Key
  // The result of the next Get call.
  NextGet struct{Value; error}
  // An ordered list of Set calls.
  SetCalls []*KeyValuePair
}

func (f *FakeKeyValueStore) Get(key Key) (Value, error) {
  f.GetCalls = append(f.GetCalls, key)
  return f.NextGet.Value, f.NextGet.error
}

func (f *FakeKeyValueStore) SetNextGet(v Value, err error) {
  f.NextGet = struct {Value; error} {v, err}
}

func (f *FakeKeyValueStore) Set(key Key, value Value) error {
  f.SetCalls = append(f.SetCalls, &KeyValuePair{ Key: key, Value: value })
  return nil
}

// A collection of testing utilities for the Cache.
type FakeCache struct {
  // An ordered list of Get calls.
  GetCalls []Key
  // An ordered list of Set calls.
  SetCalls []*KeyValuePair
}

func (f *FakeCache) Get(key Key) (Value, error) {
  f.GetCalls = append(f.GetCalls, key)
  return "", nil
}
