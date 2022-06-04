package store

import (
  "crypto/md5"
  "hash"
  "encoding/hex"
  "fmt"

)

/**
 * A KeyValueStore that stores key/value pairs on disk.
 * Every key/value pair is allocated its own file.  
 *
 * <p> We enforce a deterministic strategy for identifying a 
 * filename given a Key. Currently, we filename for key == hash(key).
 */
type FileStore struct {
  // The absolute path where we should allocat files.
  directory string 
  // The hashing strategy for this file store.
  hash hash.Hash
}

func (f *FileStore) Set(key Key, value Value) error {
    fmt.Println("SET", string(key), " will be stored in", f.getFilePath(key))
    return nil
}

func (f *FileStore) Get(key Key) (Value, error) {
  fmt.Println("GET", key, " will be looked up in", f.getFilePath(key))
  return "", nil
}

func (f *FileStore) getFilePath(key Key) string {
  hash := f.hash.Sum([]byte(key))
  return fmt.Sprintf(f.directory + "/%s", hex.EncodeToString(hash))
}

func MakeFileStore(directory string) *FileStore {
  fs := &FileStore{}
  fs.directory = directory
  fs.hash = md5.New()
  return fs
} 
