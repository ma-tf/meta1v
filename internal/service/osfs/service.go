//go:generate mockgen -destination=./mocks/service_mock.go -package=osfs_test github.com/ma-tf/meta1v/internal/service/osfs FileSystem,File
package osfs

import (
	"io"
	"os"
)

type FileSystem interface {
	OpenFile(name string, flag int, perm os.FileMode) (File, error)
	Open(name string) (File, error)
	Pipe() (r *os.File, w *os.File, err error)
	Stat(name string) (os.FileInfo, error)
}

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

func NewFileSystem() FileSystem {
	return osFS{}
}
