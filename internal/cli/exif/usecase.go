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
	ErrExifServiceUnavailable = errors.New("exif service unavailable")
	ErrFailedToInterpretEFD   = errors.New("failed to interpret EFD file")
	ErrDuplicateFrameNumber   = errors.New("duplicate frame number in EFD file")
	ErrFrameNumberNotFound    = errors.New("frame number not found in EFD file")
)

type exportUseCase struct {
	efdService         efd.Service
	exifServiceFactory exif.ServiceFactory
}

func NewUseCase(
	efdService efd.Service,
	exifServiceFactory exif.ServiceFactory,
) UseCase {
	return exportUseCase{
		efdService:         efdService,
		exifServiceFactory: exifServiceFactory,
	}
}

func (uc exportUseCase) ExportExif(
	ctx context.Context,
	efdFile string,
	frame int,
	targetFile string,
) error {
	exifService, err := uc.exifServiceFactory.Create()
	if err != nil {
		return errors.Join(ErrExifServiceUnavailable, err)
	}

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

	err = exifService.WriteEXIF(ctx, efrm, targetFile)
	if err != nil {
		return fmt.Errorf("write exif failed: %w", err)
	}

	return nil
}
