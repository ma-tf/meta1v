package exif

import (
	"context"
	"fmt"

	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/ma-tf/meta1v/internal/service/exif"
)

type UseCase struct {
	efdService  efd.Service
	exifService exif.Service
}

func NewUseCase(
	efdService efd.Service,
	exifService exif.Service,
) UseCase {
	return UseCase{
		efdService:  efdService,
		exifService: exifService,
	}
}

func (uc UseCase) ExportExif(
	ctx context.Context,
	filename string,
	frame int,
) error {
	records, err := uc.efdService.RecordsFromFile(ctx, filename)
	if err != nil {
		return fmt.Errorf("failed to interpret file content: %w", err)
	}

	target := "./test_files/20251011_Japan 1_0.dng"

	err = uc.exifService.WriteEXIF(ctx, records, frame, target)
	if err != nil {
		return fmt.Errorf("write exif failed: %w", err)
	}

	return nil
}
