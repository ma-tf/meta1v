//go:generate mockgen -destination=./mocks/service_mock.go -package=osfs_test github.com/ma-tf/meta1v/internal/service/osfs FileSystem,File
package osfs

import (
	"io"
	"os"
)

var _ FileSystem = osFS{}

type FileSystem interface {
	Create(name string) (File, error)
	Open(name string) (File, error)
	Stat(name string) (os.FileInfo, error)
}

type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Writer
	Stat() (os.FileInfo, error)
}

type osFS struct{}

//nolint:wrapcheck // os package errors are sufficient
func (osFS) Create(name string) (File, error) { return os.Create(name) }

//nolint:wrapcheck // os package errors are sufficient
func (osFS) Open(name string) (File, error) { return os.Open(name) }

//nolint:wrapcheck // os package errors are sufficient
func (osFS) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }

func NewFileSystem() FileSystem {
	return osFS{}
}
