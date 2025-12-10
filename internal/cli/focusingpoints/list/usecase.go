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
	ErrFailedToReadFile  = errors.New("failed read file for focusing points")
	ErrFailedToParseFile = errors.New(
		"failed to parse file for focusing points",
	)
	ErrFailedToList = errors.New("failed to list focusing points")
)

type FocusingPointsListUseCase struct {
	efdService             efd.Service
	displayableRollFactory display.DisplayableRollFactory
	displayService         display.Service
}

func NewFocusingPointsListUseCase(
	efdService efd.Service,
	displayableRollFactory display.DisplayableRollFactory,
	displayService display.Service,
) FocusingPointsListUseCase {
	return FocusingPointsListUseCase{
		efdService:             efdService,
		displayableRollFactory: displayableRollFactory,
		displayService:         displayService,
	}
}

func (uc FocusingPointsListUseCase) DisplayFocusingPoints(
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

	if err = uc.displayService.DisplayFocusingPoints(os.Stdout, dr); err != nil {
		return errors.Join(ErrFailedToList, err)
	}

	return nil
}
