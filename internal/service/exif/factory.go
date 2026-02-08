//go:generate mockgen -destination=./mocks/factory_mock.go -package=exif_test github.com/ma-tf/meta1v/internal/service/exif ExiftoolCommandFactory
package exif

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"

	"github.com/ma-tf/meta1v/internal/service/osexec"
)

var ErrExifToolBinaryNotFound = errors.New("exiftool binary not found in PATH")

// ExiftoolCommandFactory creates configured exiftool command instances.
type ExiftoolCommandFactory interface {
	// CreateCommand builds an exiftool command with all necessary arguments,
	// pipes, and file descriptors configured for metadata writing.
	CreateCommand(
		ctx context.Context,
		targetFile string,
		out *bytes.Buffer,
		metadata string,
		rPipe *os.File,
	) osexec.Command
}

type exiftoolCommandFactory struct {
	lookPath osexec.LookPath
}

// NewExiftoolCommandFactory creates an ExiftoolCommandFactory.
// Panics if the exiftool binary is not found in PATH.
func NewExiftoolCommandFactory(
	lookPath osexec.LookPath,
) ExiftoolCommandFactory {
	if _, err := lookPath.LookPath("exiftool"); err != nil {
		panic(ErrExifToolBinaryNotFound)
	}

	return &exiftoolCommandFactory{
		lookPath: lookPath,
	}
}

func (f *exiftoolCommandFactory) CreateCommand(
	ctx context.Context,
	targetFile string,
	out *bytes.Buffer,
	metadata string,
	rPipe *os.File,
) osexec.Command {
	cmd := exec.CommandContext(ctx, "exiftool",
		"-config", "/proc/self/fd/3",
		"-m",
		"-@", "-",
		targetFile,
	)

	cmd.Stderr = out
	cmd.Stdout = out
	cmd.Stdin = bytes.NewBufferString(metadata)
	cmd.ExtraFiles = []*os.File{rPipe}

	return osexec.NewCommand(cmd)
}
