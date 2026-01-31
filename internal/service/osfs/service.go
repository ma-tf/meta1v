//go:generate mockgen -destination=./mocks/service_mock.go -package=osfs_test github.com/ma-tf/meta1v/internal/service/osfs FileSystem,File
package osfs

import (
	"io"
	"os"
)

type FileSystem interface {
	Create(name string) (File, error)
	Open(name string) (File, error)
	Pipe() (r *os.File, w *os.File, err error)
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
func (osFS) Create(name string) (File, error) { return os.Create(name) }

//nolint:wrapcheck // os package errors are sufficient
func (osFS) Open(name string) (File, error) { return os.Open(name) }

//nolint:wrapcheck // os package errors are sufficient
func (osFS) Pipe() (*os.File, *os.File, error) { return os.Pipe() }

func NewFileSystem() FileSystem {
	return osFS{}
}
