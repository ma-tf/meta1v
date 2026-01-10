package exif

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/ma-tf/meta1v/pkg/records"
)

//go:embed exiftool.config
var exiftoolConfig string

type Service interface {
	WriteEXIF(
		ctx context.Context,
		r records.Root,
		frameNumber int,
		targetFile string,
	) error
}

type service struct {
	log *slog.Logger
}

// WriteEXIF runs exiftool with a user-defined config. It avoids shells and
// temporary files by streaming the config over an anonymous pipe and passing
// the read end as fd 3 to the child process (accessible as /proc/self/fd/3).
func (s service) WriteEXIF(
	ctx context.Context,
	r records.Root,
	frameNumber int,
	targetFile string,
) error {
	emf, err := newExifBuilder(r, frameNumber).
		WithAvs().
		WithTv().
		WithFocalLength().
		WithIso().
		WithRemarks().
		Build()
	if err != nil {
		return fmt.Errorf("failed to build exportable data: %w", err)
	}

	return s.runExifTool(
		ctx,
		exiftoolConfig,
		emf.GetMetadataToWrite(),
		targetFile,
	)
}

// runExifTool creates an anonymous pipe to pass the exiftool config via fd 3
// to the child process and streams the metadata on stdin.
func (s service) runExifTool(
	ctx context.Context,
	cfg string,
	metadataToWrite string,
	targetFile string,
) error {
	r, w, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("create pipe: %w", err)
	}

	defer r.Close()

	cmd := exec.CommandContext(ctx, "exiftool",
		"-config", "/proc/self/fd/3",
		"-m",
		"-@", "-",
		targetFile,
	)
	cmd.ExtraFiles = []*os.File{r}
	cmd.Stdin = bytes.NewBufferString(metadataToWrite)

	var out bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &out

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("start exiftool: %w", err)
	}

	// Write config in a goroutine so we don't risk blocking if the child
	// doesn't read immediately. Close writer when done.
	writeErr := make(chan error, 1)

	go func() {
		defer w.Close()

		select {
		case <-ctx.Done():
			writeErr <- ctx.Err()
		default:
			_, writeError := w.WriteString(cfg)
			writeErr <- writeError
		}
	}()

	waitErr := cmd.Wait()
	wErr := <-writeErr

	if waitErr != nil {
		return fmt.Errorf("exiftool failed: %w", waitErr)
	}

	if wErr != nil {
		return fmt.Errorf("failed to write exiftool config: %w", wErr)
	}

	// s.log.DebugContext(ctx, "exiftool success",
	// 	"targetFile", targetFile,
	// 	"metadata", metadataToWrite,
	// 	"output", out.String())
	return nil
}
