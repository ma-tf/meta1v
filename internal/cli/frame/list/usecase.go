package list

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
)

type FrameListUseCase struct {
	efdService             efd.Service
	displayableRollFactory display.DisplayableRollFactory
	displayService         display.Service
}

func NewFrameListUseCase(
	efdService efd.Service,
	displayableRollFactory display.DisplayableRollFactory,
	displayService display.Service,
) FrameListUseCase {
	return FrameListUseCase{
		efdService:             efdService,
		displayableRollFactory: displayableRollFactory,
		displayService:         displayService,
	}
}

func (uc FrameListUseCase) DisplayFrames(
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

	if err = uc.displayService.DisplayFrames(os.Stdout, dr); err != nil {
		return fmt.Errorf("failed to display frame: %w", err)
	}

	return nil
}
