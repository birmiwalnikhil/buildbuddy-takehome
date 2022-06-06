package store

import (
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
 * filename given a Key. Currently, the key is the filename. We may use
 * hashing to give some security, and compression for the value type.
 *
 * <p> Files are created in a temporary subdirectory of `FileStore.directory`.
 * On write completion, they are moved into `FileStore.directory`. This enables
 * protection against partial writes due to server failure.  
 */
type FileStore struct {
  // The absolute path where which holds permanent files.
  directory string 
  // The temporary directory which holds temporary files. This directory 
  // will be cleared on FileStore instantiation. 
  tempDirectory string
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

/** 
 * Construct a file path given a key and a parent directory.
 */
func (f *FileStore) getFilePath(key Key, dir string) string {
  return fmt.Sprintf(dir + "/%s", string(key))
}

/**
 * Clean up a temporary file, e.g. by closing the file handle and moving it to 
 * the parent directory. Return nil if this operation was successful, or the 
 * error that occurred.
 *
 * <p> This method assumes the mutex is held.
 */
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
  return fs, nil
} 
