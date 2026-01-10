package exif

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/ma-tf/meta1v/pkg/records"
)

const exiftoolConfig = `%Image::ExifTool::UserDefined = (
        'Image::ExifTool::XMP::Main' => {
            AnalogueData => {
                SubDirectory => {
                    TagTable => 'Image::ExifTool::UserDefined::AnalogueData',
                },
            },
        },
    );
    %Image::ExifTool::UserDefined::AnalogueData = (
        GROUPS => { 0 => 'XMP', 1 => 'XMP-AnalogueData', 2 => 'Film' },
        NAMESPACE => { 'AnalogueData' => 'https://filmgra.in/AnalogueData/1.0/' },
        WRITABLE => 'string',
        FilmMaker => { },
        FilmName => { },
        FilmFormat => { },
        FilmDevelopProcess => { },
        FilmDeveloper => { },
        FilmProcessLab => { },
        FilmScanner => { },
        LensFilter => { Groups => { 2 => 'Camera' } },
    );
    1;
    `

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

func NewService(log *slog.Logger) Service {
	return service{
		log: log,
	}
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
		WithFocalLengthAndIsoAndRemarks().
		Build()
	if err != nil {
		return fmt.Errorf("failed to build exportable data: %w", err)
	}

	return s.runExifTool(
		ctx,
		exiftoolConfig,
		emf.GetMetadataToWrite(),
		frameNumber,
		targetFile,
	)
}

// runExifTool creates an anonymous pipe to pass the exiftool config via fd 3
// to the child process and streams the metadata on stdin.
func (s service) runExifTool(ctx context.Context,
	cfg string,
	metadataToWrite string,
	frameNumber int,
	targetFile string,
) error {
	const timeout = 3 * time.Minute

	ctx, cancel := context.WithTimeout(ctx, timeout) // move this up call chain?
	defer cancel()

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
	go func() {
		defer w.Close()

		select {
		case <-ctx.Done():
			return
		default:
			_, _ = w.WriteString(cfg)
		}
	}()

	if err = cmd.Wait(); err != nil {
		s.log.ErrorContext(ctx, "exiftool execution failed",
			"frameNumber", frameNumber,
			"targetFile", targetFile,
			"output", out.String())

		return fmt.Errorf("exiftool failed: %w", err)
	}

	s.log.DebugContext(ctx, "exiftool success",
		"targetFile", targetFile,
		"metadata", metadataToWrite,
		"output", out.String())

	return nil
}
