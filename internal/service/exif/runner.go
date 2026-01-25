//go:generate mockgen -destination=./mocks/runner_mock.go -package=exif_test github.com/ma-tf/meta1v/internal/service/exif ToolRunner
package exif

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"os"
	"os/exec"

	"github.com/ma-tf/meta1v/internal/service/osfs"
)

//go:embed exiftool.config
var exiftoolConfig string

var (
	ErrCreatePipe          = errors.New("failed to create pipe")
	ErrStartExifTool       = errors.New("failed to start exiftool")
	ErrExifToolFailed      = errors.New("exiftool failed")
	ErrWriteExifToolConfig = errors.New("failed to write exiftool config")
)

type ToolRunner interface {
	Run(ctx context.Context, targetFile string, metadata string) error
}

type exifToolRunner struct {
	fs osfs.FileSystem
}

func NewExifToolRunner(fs osfs.FileSystem) ToolRunner {
	return &exifToolRunner{fs: fs}
}

// Run executes exiftool with a config passed via fd 3 and metadata on stdin.
func (r *exifToolRunner) Run(
	ctx context.Context,
	targetFile string,
	metadata string,
) error {
	rPipe, wPipe, err := r.fs.Pipe()
	if err != nil {
		return errors.Join(ErrCreatePipe, err)
	}

	defer rPipe.Close()

	cmd := exec.CommandContext(ctx, "exiftool",
		"-config", "/proc/self/fd/3",
		"-m",
		"-@", "-",
		targetFile,
	)

	var out bytes.Buffer

	cmd.Stderr = &out
	cmd.Stdout = &out
	cmd.Stdin = bytes.NewBufferString(metadata)
	cmd.ExtraFiles = []*os.File{rPipe}

	if err = cmd.Start(); err != nil {
		return errors.Join(ErrStartExifTool, err)
	}

	// Write config in a goroutine so we don't risk blocking if the child
	// doesn't read immediately. Close writer when done.
	writeErr := make(chan error, 1)

	go func() {
		defer wPipe.Close()

		select {
		case <-ctx.Done():
			writeErr <- ctx.Err()
		default:
			_, cfgWriteError := wPipe.WriteString(exiftoolConfig)
			writeErr <- cfgWriteError
		}
	}()

	if err = cmd.Wait(); err != nil {
		return errors.Join(ErrExifToolFailed, err)
	}

	if err = <-writeErr; err != nil {
		return errors.Join(ErrWriteExifToolConfig, err)
	}

	return nil
}
