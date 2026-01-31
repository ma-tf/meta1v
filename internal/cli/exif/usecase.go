package exif

import (
	"context"
	"errors"
	"fmt"

	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/ma-tf/meta1v/internal/service/exif"
	"github.com/ma-tf/meta1v/pkg/records"
)

var (
	ErrFailedToInterpretEFD = errors.New("failed to interpret EFD file")
	ErrDuplicateFrameNumber = errors.New("duplicate frame number in EFD file")
	ErrFrameNumberNotFound  = errors.New("frame number not found in EFD file")
)

type exportUseCase struct {
	efdService  efd.Service
	exifService exif.Service
}

func NewUseCase(
	efdService efd.Service,
	exifService exif.Service,
) UseCase {
	return exportUseCase{
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
	root, err := uc.efdService.RecordsFromFile(ctx, efdFile)
	if err != nil {
		return errors.Join(ErrFailedToInterpretEFD, err)
	}

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

	err = uc.exifService.WriteEXIF(ctx, efrm, targetFile, strict)
	if err != nil {
		return fmt.Errorf("write exif failed: %w", err)
	}

	return nil
}
