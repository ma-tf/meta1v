//go:generate mockgen -destination=./mocks/factory_mock.go -package=exif_test github.com/ma-tf/meta1v/internal/service/exif ExiftoolCommandFactory
package exif

import (
	"bytes"
	"context"
	"os"
	"os/exec"

	"github.com/ma-tf/meta1v/internal/service/osexec"
)

type exiftoolCommandFactory struct {
	lookPath osexec.LookPath
}

type ExiftoolCommandFactory interface {
	CreateCommand(
		ctx context.Context,
		targetFile string,
		out *bytes.Buffer,
		metadata string,
		rPipe *os.File,
	) osexec.Command
}

func NewExiftoolCommandFactory(
	lookPath osexec.LookPath,
) ExiftoolCommandFactory {
	if _, err := lookPath.LookPath("exiftool"); err != nil {
		panic("exiftool binary not found in PATH")
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
