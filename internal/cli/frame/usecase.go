package frame

import (
	"context"
	"errors"
	"os"

	"github.com/ma-tf/meta1v/internal/cli/frame/list"
	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
)

var (
	ErrFailedToReadFile  = errors.New("failed read file for frames")
	ErrFailedToParseFile = errors.New(
		"failed to parse file for frames",
	)
	ErrFailedToList = errors.New("failed to list frames")
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

	dr, err := uc.displayableRollFactory.Create(records)
	if err != nil {
		return errors.Join(ErrFailedToParseFile, err)
	}

	if err = uc.displayService.DisplayFrames(os.Stdout, dr); err != nil {
		return errors.Join(ErrFailedToList, err)
	}

	return nil
}
