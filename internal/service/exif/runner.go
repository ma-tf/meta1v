//go:generate mockgen -destination=./mocks/runner_mock.go -package=exif_test github.com/ma-tf/meta1v/internal/service/exif ToolRunner
package exif

import (
	"bytes"
	"context"
	_ "embed"
	"errors"

	"github.com/ma-tf/meta1v/internal/service/osfs"
)

//go:embed exiftool.config
var exiftoolConfig string

var (
	ErrCreatePipe          = errors.New("failed to create pipe")
	ErrStartExifTool       = errors.New("failed to start exiftool")
	ErrExifToolFailed      = errors.New("exiftool failed")
	ErrContextDone         = errors.New("context done before writing config")
	ErrWriteExifToolConfig = errors.New("failed to write exiftool config")
)

// ToolRunner executes exiftool with metadata and configuration.
type ToolRunner interface {
	// Run executes exiftool on the target file with the provided metadata tags.
	Run(ctx context.Context, targetFile string, metadata string) error
}

type exifToolRunner struct {
	fs      osfs.FileSystem
	factory ExiftoolCommandFactory
}

func NewExifToolRunner(
	fs osfs.FileSystem,
	factory ExiftoolCommandFactory,
) ToolRunner {
	return &exifToolRunner{
		fs:      fs,
		factory: factory,
	}
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

	var out bytes.Buffer

	cmd := r.factory.CreateCommand(ctx, targetFile, &out, metadata, rPipe)

	if err = cmd.Start(); err != nil {
		return errors.Join(ErrStartExifTool, err)
	}

	// Write config in a goroutine so we don't risk blocking if the child
	// doesn't read immediately. Close writer when done.
	writeErr := make(chan error, 1)

	go func() {
		defer wPipe.Close()
		defer close(writeErr)

		select {
		case <-ctx.Done():
			writeErr <- errors.Join(ErrContextDone, ctx.Err())
		default:
			_, cfgWriteError := wPipe.WriteString(exiftoolConfig)
			if cfgWriteError != nil {
				writeErr <- errors.Join(ErrWriteExifToolConfig, cfgWriteError)
			}
		}
	}()

	if err = cmd.Wait(); err != nil {
		return errors.Join(ErrExifToolFailed, err)
	}

	if err = <-writeErr; err != nil {
		return err
	}

	return nil
}
