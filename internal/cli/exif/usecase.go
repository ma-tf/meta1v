package exif

import (
	"context"
	"fmt"

	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/ma-tf/meta1v/internal/service/exif"
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
		return fmt.Errorf("exif service unavailable: %w", err)
	}

	records, err := uc.efdService.RecordsFromFile(ctx, efdFile)
	if err != nil {
		return fmt.Errorf("failed to interpret file content: %w", err)
	}

	err = exifService.WriteEXIF(ctx, records, frame, targetFile)
	if err != nil {
		return fmt.Errorf("write exif failed: %w", err)
	}

	return nil
}
