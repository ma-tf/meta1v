package customfunctions

import (
	"context"
	"errors"
	"os"

	"github.com/ma-tf/meta1v/internal/cli/customfunctions/list"
	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
)

var (
	ErrFailedToReadFile  = errors.New("failed read file for custom functions")
	ErrFailedToParseFile = errors.New(
		"failed to parse file for custom functions",
	)
	ErrFailedToDisplay = errors.New("failed to display custom functions")
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
) error {
	records, err := uc.efdService.RecordsFromFile(ctx, filename)
	if err != nil {
		return errors.Join(ErrFailedToReadFile, err)
	}

	dr, err := uc.displayableRollFactory.Create(ctx, records)
	if err != nil {
		return errors.Join(ErrFailedToParseFile, err)
	}

	if err = uc.displayService.DisplayCustomFunctions(os.Stdout, dr); err != nil {
		return errors.Join(ErrFailedToDisplay, err)
	}

	return nil
}
