package focusingpoints

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/ma-tf/meta1v/internal/cli/focusingpoints/list"
	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
)

var (
	ErrFailedToReadFile  = errors.New("failed to read file for focusing points")
	ErrFailedToParseFile = errors.New(
		"failed to parse file for focusing points",
	)
)

type listUseCase struct {
	efdService             efd.Service
	displayableRollFactory display.DisplayableRollFactory
	displayService         display.Service
}

func NewListUseCase(
	efdService efd.Service,
	displayableRollFactory display.DisplayableRollFactory,
	displayService display.Service,
) list.UseCase {
	return listUseCase{
		efdService:             efdService,
		displayableRollFactory: displayableRollFactory,
		displayService:         displayService,
	}
}

func (uc listUseCase) List(
	ctx context.Context,
	filename string,
	strict bool,
) error {
	records, err := uc.efdService.RecordsFromFile(ctx, filename)
	if err != nil {
		return fmt.Errorf("%w %q: %w", ErrFailedToReadFile, filename, err)
	}

	dr, err := uc.displayableRollFactory.Create(ctx, records, strict)
	if err != nil {
		return fmt.Errorf("%w %q: %w", ErrFailedToParseFile, filename, err)
	}

	uc.displayService.DisplayFocusingPoints(os.Stdout, dr)

	return nil
}
