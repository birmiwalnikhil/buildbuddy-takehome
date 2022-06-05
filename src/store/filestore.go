package store

import (
  "crypto/md5"
  "hash"
  "encoding/hex"
  "fmt"
  "os"
)

/**
 * A KeyValueStore that stores key/value pairs on disk.
 * Every key/value pair is allocated its own file.  
 *
 * <p> We enforce a deterministic strategy for identifying a 
 * filename given a Key. Currently, we filename for key == hash(key).
 *
 * <p> This class is not thread-safe; callers are expected to serialize
 * reads and writes appropriately.
 */
type FileStore struct {
  // The absolute path where we should allocate files.
  directory string 
  // The hashing strategy for this file store.
  hash hash.Hash
}

func (f *FileStore) Set(key Key, value Value) error {
    filePath := f.getFilePath(key)
    file, err := 
      os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
    if err != nil {
      // IO Error when opening the file; return the error.
      return err
    }
    
    // Write the value into the opened file.
    defer file.Close()
    _, err2 := file.Write([]byte(value))
    if err2 != nil {
      return err2
    }

    return nil
}

func (f *FileStore) Get(key Key) (Value, error) {
  filePath := f.getFilePath(key)
  value, err := os.ReadFile(filePath)
  if err != nil {
    // Error when reading the file (e.g. corrupted file, file missing).
    return "", err
  }
 
  return Value(value), nil
}

func (f *FileStore) getFilePath(key Key) string {
  hash := f.hash.Sum([]byte(key))
  return fmt.Sprintf(f.directory + "/%s", hex.EncodeToString(hash))
}

func MakeFileStore(directory string) (*FileStore, error) {
  fs := &FileStore{}
  fs.directory = directory
  // Make the directory if it does not already exist.
  if err := os.Mkdir(directory, 0644); err != nil && !os.IsExist(err) {
    return nil, err
  } 
  fs.hash = md5.New()
  return fs, nil
} 
