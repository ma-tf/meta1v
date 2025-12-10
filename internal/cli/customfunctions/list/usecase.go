package list

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
)

var (
	ErrFailedToReadFile  = errors.New("failed read file for custom functions")
	ErrFailedToParseFile = errors.New(
		"failed to parse file for custom functions",
	)
)

type CustomFunctionsListUseCase struct {
	efdService             efd.Service
	displayableRollFactory display.DisplayableRollFactory
	displayService         display.Service
}

func NewCustomFunctionsListUseCase(
	efdService efd.Service,
	displayableRollFactory display.DisplayableRollFactory,
	displayService display.Service,
) CustomFunctionsListUseCase {
	return CustomFunctionsListUseCase{
		efdService:             efdService,
		displayableRollFactory: displayableRollFactory,
		displayService:         displayService,
	}
}

func (uc CustomFunctionsListUseCase) DisplayCustomFunctions(
	ctx context.Context,
	r io.Reader,
) error {
	records, err := uc.efdService.RecordsFromFile(ctx, r)
	if err != nil {
		return errors.Join(ErrFailedToReadFile, err)
	}

	dr, err := uc.displayableRollFactory.Create(records)
	if err != nil {
		return errors.Join(ErrFailedToParseFile, err)
	}

	uc.displayService.DisplayCustomFunctions(os.Stdout, dr)

	return nil
}
