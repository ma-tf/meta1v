package list

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
)

type ThumbnailListUseCase struct {
	efdService             efd.Service
	displayableRollFactory display.DisplayableRollFactory
	displayService         display.Service
}

func NewThumbnailListUseCase(
	efdService efd.Service,
	displayableRollFactory display.DisplayableRollFactory,
	displayService display.Service,
) ThumbnailListUseCase {
	return ThumbnailListUseCase{
		efdService:             efdService,
		displayableRollFactory: displayableRollFactory,
		displayService:         displayService,
	}
}

func (uc ThumbnailListUseCase) DisplayThumbnails(
	ctx context.Context,
	r io.Reader,
) error {
	records, err := uc.efdService.RecordsFromFile(ctx, r)
	if err != nil {
		return fmt.Errorf("failed read file: %w", err)
	}

	dr, err := uc.displayableRollFactory.Create(records)
	if err != nil {
		return fmt.Errorf("failed parse file: %w", err)
	}

	uc.displayService.DisplayThumbnails(os.Stdout, dr)

	return nil
}
