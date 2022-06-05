package store

import (
  "crypto/md5"
  "hash"
  "encoding/hex"
  "fmt"
  "os"
  "sync"
)

const (
  TEMP_DIRECTORY_NAME = "tmp"
)

/**
 * A KeyValueStore that stores key/value pairs on disk.
 * Every key/value pair is allocated its own file.  
 *
 * <p> We enforce a deterministic strategy for identifying a 
 * filename given a Key. Currently, we filename for key == hash(key).
 *
 * <p> Files are created in a temporary subdirectory of `FileStore.directory`.
 * On write completion, they are moved into `FileStore.directory`. This enables
 * protection against partial writes due to server failure.  
 */
type FileStore struct {
  // The absolute path where which holds permanent files.
  directory string 
  // The temporary directory which holds temporary files. Reset on
  // instantiation of the file store. 
  tempDirectory string
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

    // 1) Create a temp file in the temporary directory.
    // 2) Write to the temporary file.
    // 3) On completion:
    //    - Close the file.
    //    - Promote the temporary file to a permanent file?
    //    - Move the temporary file from the temp directory.
    filePath := f.getFilePath(key, f.tempDirectory)
    tmpFile, err := 
      os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
    if err != nil {
      // IO Error when opening the file; return the error.
      return err
    }
   
    // Write the value into the opened file.
    _, err2 := tmpFile.Write([]byte(value))
    if err2 != nil {
      // On failure, close the opened file handle.
      tmpFile.Close()
      return err2
    }
  
    return f.onTmpFileComplete(key, tmpFile)
}

/** 
 * Read the key/value pair from disk. Return the stored value, or any errors
 * that may have occurred when reading the file.
 */
func (f *FileStore) Get(key Key) (Value, error) {
  defer f.mutex.Unlock()
  f.mutex.Lock()
  // Only search the directory of fully written files.
  filePath := f.getFilePath(key, f.directory)
  value, err := os.ReadFile(filePath)
  if err != nil {
    // Error when reading the file (e.g. corrupted file, file missing).
    return "", err
  }
 
  return Value(value), nil
}

// Identify a file path given a Key. If isTmpFile is true, the path
// will include the temporary directory. 
func (f *FileStore) getFilePath(key Key, dir string) string {
  hash := f.hash.Sum([]byte(key))
  return fmt.Sprintf(dir + "/%s", hex.EncodeToString(hash))
}

// To be called on completion of writing to a temporary file, e.g.
// to move the temporary file out of the temporary directory. Return
// any errors, e.g. an OS failure.
func (f *FileStore) onTmpFileComplete(key Key, tmpFile *os.File) error {
  if err := tmpFile.Close(); err != nil {
    return err
  }

  oldPath := f.getFilePath(key, f.tempDirectory)
  newPath := f.getFilePath(key, f.directory)
  if err := os.Rename(oldPath, newPath); err != nil {
    return err
  }
  
  return nil
}

func MakeFileStore(directory string) (*FileStore, error) {
  fs := &FileStore{}
  fs.directory = directory
  fs.tempDirectory = fmt.Sprintf(directory + "/%s", TEMP_DIRECTORY_NAME) 
  // Make the directory if it does not already exist.
  if err := os.Mkdir(directory, 0644); err != nil && !os.IsExist(err) {
    return nil, err
  } 
  
  if err := os.RemoveAll(fs.tempDirectory); err != nil {
    // Remove all the temporary files.
    return nil, err
  }

  // Create the temp directory if it does not already exist.
  if err := os.Mkdir(fs.tempDirectory, 0644); err != nil && !os.IsExist(err) {
    return nil, err
  }

  fs.mutex = &sync.Mutex{} 
  fs.hash = md5.New()
  return fs, nil
} 
