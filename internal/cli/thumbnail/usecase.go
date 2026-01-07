package thumbnail

import (
	"context"
	"errors"
	"os"

	"github.com/ma-tf/meta1v/internal/cli/thumbnail/list"
	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
)

var (
	ErrFailedToReadFile  = errors.New("failed read file for thumbnails")
	ErrFailedToParseFile = errors.New("failed to parse file for thumbnails")
)

type usecase struct {
	efdService             efd.Service
	displayableRollFactory display.DisplayableRollFactory
	displayService         display.Service
}

func NewThumbnailListUseCase(
	efdService efd.Service,
	displayableRollFactory display.DisplayableRollFactory,
	displayService display.Service,
) list.UseCase {
	return usecase{
		efdService:             efdService,
		displayableRollFactory: displayableRollFactory,
		displayService:         displayService,
	}
}

func (uc usecase) DisplayThumbnails(
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

	uc.displayService.DisplayThumbnails(os.Stdout, dr)

	return nil
}
