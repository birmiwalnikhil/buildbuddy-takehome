package store

import (
  "crypto/md5"
  "hash"
  "encoding/hex"
  "fmt"
  "os"
  "sync"
)

/**
 * A KeyValueStore that stores key/value pairs on disk.
 * Every key/value pair is allocated its own file.  
 *
 * <p> We enforce a deterministic strategy for identifying a 
 * filename given a Key. Currently, we filename for key == hash(key).
*/
type FileStore struct {
  // The absolute path where we should allocate files.
  directory string 
  // The hashing strategy for this file store.
  hash hash.Hash
  // A mutex used to synchronize access to the underlying file directory.
  // A RW lock _may_ improve performance; I'm not sure what the concurrency
  // requirements are of a UNIX based file system.
  mutex *sync.Mutex
}

/** 
 * Store the key/value pair on disk. Overwrite any prexisting files on key
 * conflict. If the operation was successful, return nil; otherwise, report the
 * error that occurred (e.g. an IO failure during file creation.)
 */
func (f *FileStore) Set(key Key, value Value) error {
    defer f.mutex.Unlock()
    f.mutex.Lock()
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

/** 
 * Read the key/value pair from disk. Return the stored value, or any errors
 * that may have occurred when reading the file.
 */
func (f *FileStore) Get(key Key) (Value, error) {
  defer f.mutex.Unlock()
  f.mutex.Lock()
  filePath := f.getFilePath(key)
  value, err := os.ReadFile(filePath)
  if err != nil {
    // Error when reading the file (e.g. corrupted file, file missing).
    return "", err
  }
 
  return Value(value), nil
}

// Identify a file path given a Key. 
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
  fs.mutex = &sync.Mutex{} 
  fs.hash = md5.New()
  return fs, nil
} 
