package thumbnail

import (
	"context"
	"errors"
	"fmt"
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

	uc.displayService.DisplayThumbnails(os.Stdout, dr)

	return nil
}
