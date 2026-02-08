//go:generate mockgen -destination=./mocks/service_mock.go -package=osfs_test github.com/ma-tf/meta1v/internal/service/osfs FileSystem,File

// Package osfs provides an abstraction layer over the os package's filesystem operations.
//
// This package enables dependency injection and mocking of filesystem operations for testing,
// while maintaining compatibility with standard library io interfaces.
package osfs

import (
	"io"
	"os"
)

// FileSystem provides filesystem operations that can be mocked for testing.
type FileSystem interface {
	// OpenFile opens a file with the given flags and permissions.
	OpenFile(name string, flag int, perm os.FileMode) (File, error)

	// Open opens a file for reading.
	Open(name string) (File, error)

	// Pipe creates a pair of connected files for inter-process communication.
	Pipe() (r *os.File, w *os.File, err error)

	// Stat returns file information.
	Stat(name string) (os.FileInfo, error)
}

// File is a mockable file interface combining standard io operations.
type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Writer
}

type osFS struct{}

//nolint:wrapcheck // os package errors are sufficient
func (osFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return os.OpenFile(name, flag, perm)
}

//nolint:wrapcheck // os package errors are sufficient
func (osFS) Open(name string) (File, error) { return os.Open(name) }

//nolint:wrapcheck // os package errors are sufficient
func (osFS) Pipe() (*os.File, *os.File, error) { return os.Pipe() }

//nolint:wrapcheck // os package errors are sufficient
func (osFS) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }

// NewFileSystem creates a FileSystem that delegates to the standard os package.
func NewFileSystem() FileSystem {
	return osFS{}
}
