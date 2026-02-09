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

package exif

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ma-tf/meta1v/internal/records"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/ma-tf/meta1v/internal/service/exif"
)

var (
	ErrFailedToInterpretEFD = errors.New("failed to interpret EFD file")
	ErrDuplicateFrameNumber = errors.New("duplicate frame number in EFD file")
	ErrFrameNumberNotFound  = errors.New("frame number not found in EFD file")
	ErrWriteEXIFFailed      = errors.New("failed to write EXIF data")
)

type exportUseCase struct {
	log         *slog.Logger
	efdService  efd.Service
	exifService exif.Service
}

func NewUseCase(
	log *slog.Logger,
	efdService efd.Service,
	exifService exif.Service,
) UseCase {
	return exportUseCase{
		log:         log,
		efdService:  efdService,
		exifService: exifService,
	}
}

func (uc exportUseCase) ExportExif(
	ctx context.Context,
	efdFile string,
	frame int,
	targetFile string,
	strict bool,
) error {
	uc.log.InfoContext(ctx, "starting exif export",
		slog.String("efd_file", efdFile),
		slog.Int("frame", frame),
		slog.String("target_file", targetFile),
		slog.Bool("strict", strict))

	root, err := uc.efdService.RecordsFromFile(ctx, efdFile)
	if err != nil {
		return fmt.Errorf("%w %q: %w", ErrFailedToInterpretEFD, efdFile, err)
	}

	uc.log.DebugContext(ctx, "efd file parsed",
		slog.Int("frame_count", len(root.EFRMs)))

	var (
		efrm  records.EFRM
		found bool
	)

	for _, e := range root.EFRMs {
		if int(e.FrameNumber) == frame {
			if found {
				return fmt.Errorf("%w: frame number %d",
					ErrDuplicateFrameNumber,
					frame,
				)
			}

			efrm = e
			found = true
		}
	}

	if !found {
		return fmt.Errorf("%w: frame number %d",
			ErrFrameNumberNotFound,
			frame,
		)
	}

	uc.log.DebugContext(ctx, "frame located",
		slog.Uint64("frame_number", uint64(efrm.FrameNumber)))

	err = uc.exifService.WriteEXIF(ctx, efrm, targetFile, strict)
	if err != nil {
		return fmt.Errorf("%w on %q: %w", ErrWriteEXIFFailed, targetFile, err)
	}

	uc.log.InfoContext(ctx, "exif export completed successfully",
		slog.String("target_file", targetFile))

	return nil
}
