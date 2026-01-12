package exif

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ma-tf/meta1v/pkg/records"
)

type Service interface {
	WriteEXIF(
		ctx context.Context,
		efrm records.EFRM,
		targetFile string,
		strict bool,
	) error
}

type service struct {
	log     *slog.Logger
	runner  ToolRunner
	builder Builder
}

// WriteEXIF runs exiftool with a user-defined config. It avoids shells and
// temporary files by streaming the config over an anonymous pipe and passing
// the read end as fd 3 to the child process (accessible as /proc/self/fd/3).
func (s service) WriteEXIF(
	ctx context.Context,
	efrm records.EFRM,
	targetFile string,
	strict bool,
) error {
	data, err := s.builder.Build(efrm, strict)
	if err != nil {
		return fmt.Errorf("failed to build exif data: %w", err)
	}

	var args strings.Builder

	for tag, value := range data {
		if value != "" {
			fmt.Fprintf(&args, "-%s=%s\n", tag, value)
		}
	}

	err = s.runner.Run(ctx, targetFile, args.String())
	if err != nil {
		return fmt.Errorf("failed to run exiftool: %w", err)
	}

	return nil
}
