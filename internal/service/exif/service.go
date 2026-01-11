package exif

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ma-tf/meta1v/pkg/records"
)

type Service interface {
	WriteEXIF(
		ctx context.Context,
		r records.Root,
		frameNumber int,
		targetFile string,
	) error
}

type service struct {
	log    *slog.Logger
	runner ToolRunner
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
	exifData, err := newExifBuilder(r, frameNumber).
		WithAvs().
		WithTv().
		WithFocalLength().
		WithIso().
		WithRemarks().
		Build()
	if err != nil {
		return fmt.Errorf("failed to build exportable data: %w", err)
	}

	err = s.runner.Run(ctx, targetFile, exifData.FormatAsArgFile())
	if err != nil {
		return fmt.Errorf("failed to run exiftool: %w", err)
	}

	return nil
}
