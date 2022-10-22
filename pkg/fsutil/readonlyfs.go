package fsutil

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type ReadOnlyFS struct {
	f filesys.FileSystem
}

func NewReadOnlyFS(f filesys.FileSystem) filesys.FileSystem {
	return &ReadOnlyFS{
		f: f,
	}
}

// Create a file.
func (fs *ReadOnlyFS) Create(path string) (filesys.File, error) {
	return nil, fmt.Errorf("readonly fs: %s", path)
}

// MkDir makes a directory.
func (fs *ReadOnlyFS) Mkdir(path string) error {
	return fmt.Errorf("readonly fs: %s", path)
}

// MkDirAll makes a directory path, creating intervening directories.
func (fs *ReadOnlyFS) MkdirAll(path string) error {
	return fmt.Errorf("readonly fs: %s", path)
}

// RemoveAll removes path and any children it contains.
func (fs *ReadOnlyFS) RemoveAll(path string) error {
	return fmt.Errorf("readonly fs: %s", path)
}

// Open opens the named file for reading.
func (fs *ReadOnlyFS) Open(path string) (filesys.File, error) {
	file, err := fs.f.Open(path)

	if err != nil {
		return nil, err
	}

	return &ReadOnlyFile{
		fileName: path,
		f:        file,
	}, nil
}

// IsDir returns true if the path is a directory.
func (fs *ReadOnlyFS) IsDir(path string) bool {
	return fs.f.IsDir(path)
}

// ReadDir returns a list of files and directories within a directory.
func (fs *ReadOnlyFS) ReadDir(path string) ([]string, error) {
	return fs.f.ReadDir(path)
}

// CleanedAbs converts the given path into a
// directory and a file name, where the directory
// is represented as a ConfirmedDir and all that implies.
// If the entire path is a directory, the file component
// is an empty string.
func (fs *ReadOnlyFS) CleanedAbs(path string) (filesys.ConfirmedDir, string, error) {
	return fs.f.CleanedAbs(path)
}

// Exists is true if the path exists in the file system.
func (fs *ReadOnlyFS) Exists(path string) bool {
	return fs.f.Exists(path)
}

// Glob returns the list of matching files,
// emulating https://golang.org/pkg/path/filepath/#Glob
func (fs *ReadOnlyFS) Glob(pattern string) ([]string, error) {
	return fs.f.Glob(pattern)
}

// ReadFile returns the contents of the file at the given path.
func (fs *ReadOnlyFS) ReadFile(path string) ([]byte, error) {
	return fs.f.ReadFile(path)
}

// WriteFile writes the data to a file at the given path,
// overwriting anything that's already there.
func (fs *ReadOnlyFS) WriteFile(path string, data []byte) error {
	return fmt.Errorf("readonly fs: %s", path)
}

// Walk walks the file system with the given WalkFunc.
func (fs *ReadOnlyFS) Walk(path string, walkFn filepath.WalkFunc) error {
	return fs.f.Walk(path, walkFn)
}
