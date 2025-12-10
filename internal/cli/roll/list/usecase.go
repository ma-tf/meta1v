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
	ErrFailedToReadFile  = errors.New("failed read file for roll")
	ErrFailedToParseFile = errors.New("failed to parse file for roll")
)

type RollListUseCase struct {
	efdService             efd.Service
	displayableRollFactory display.DisplayableRollFactory
	displayService         display.Service
}

func NewRollListUseCase(
	efdService efd.Service,
	displayableRollFactory display.DisplayableRollFactory,
	displayService display.Service,
) RollListUseCase {
	return RollListUseCase{
		efdService:             efdService,
		displayableRollFactory: displayableRollFactory,
		displayService:         displayService,
	}
}

func (uc RollListUseCase) DisplayRoll(ctx context.Context, r io.Reader) error {
	records, err := uc.efdService.RecordsFromFile(ctx, r)
	if err != nil {
		return errors.Join(ErrFailedToReadFile, err)
	}

	dr, err := uc.displayableRollFactory.Create(records)
	if err != nil {
		return errors.Join(ErrFailedToParseFile, err)
	}

	uc.displayService.DisplayRoll(os.Stdout, dr)

	return nil
}
