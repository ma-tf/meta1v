// meta1v is a command-line tool for viewing and manipulating metadata for Canon EOS-1V files of the EFD format.
// Copyright (C) 2026  Matt F
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

//go:generate mockgen -destination=./mocks/service_mock.go -package=exif_test github.com/ma-tf/meta1v/internal/service/exif Service

// Package exif provides services for writing EXIF metadata to image files using Canon EFD frame data.
//
// This package builds EXIF tags from frame metadata and executes exiftool to embed
// the metadata into target image files.
package exif

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/ma-tf/meta1v/internal/records"
)

var (
	ErrBuildExifData = errors.New("failed to build exif data")
	ErrRunExifTool   = errors.New("failed to run exiftool")
)

// Service provides operations for writing EXIF metadata to image files from Canon EFD frame records.
type Service interface {
	// WriteEXIF writes EXIF metadata from an EFRM record to the target image file.
	// The strict parameter controls whether unknown metadata values cause errors.
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

func NewService(
	log *slog.Logger,
	runner ToolRunner,
	builder Builder,
) Service {
	return &service{
		log:     log,
		runner:  runner,
		builder: builder,
	}
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
	s.log.InfoContext(ctx, "writing exif data to file",
		slog.String("target_file", targetFile),
		slog.Uint64("frame_number", uint64(efrm.FrameNumber)),
		slog.Bool("strict", strict))

	data, err := s.builder.Build(efrm, strict)
	if err != nil {
		return fmt.Errorf(
			"%w for frame %d: %w",
			ErrBuildExifData,
			efrm.FrameNumber,
			err,
		)
	}

	s.log.DebugContext(ctx, "exif data built",
		slog.Int("tag_count", len(data)))

	keys := make([]string, 0, len(data))
	for tag := range data {
		keys = append(keys, tag)
	}

	sort.Strings(keys)

	var args strings.Builder

	for _, tag := range keys {
		value := data[tag]
		if value != "" {
			fmt.Fprintf(&args, "-%s=%s\n", tag, value)
		}
	}

	s.log.DebugContext(ctx, "running exiftool",
		slog.String("target_file", targetFile))

	err = s.runner.Run(ctx, targetFile, args.String())
	if err != nil {
		return fmt.Errorf("%w on %q: %w", ErrRunExifTool, targetFile, err)
	}

	s.log.InfoContext(ctx, "exif data written successfully",
		slog.String("target_file", targetFile),
		slog.Int("tags_written", len(data)))

	return nil
}
