package thumbnail

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/ma-tf/meta1v/internal/cli/thumbnail/list"
	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
)

var (
	ErrFailedToReadFile  = errors.New("failed to read file for thumbnails")
	ErrFailedToParseFile = errors.New("failed to parse file for thumbnails")
)

type usecase struct {
	log                    *slog.Logger
	efdService             efd.Service
	displayableRollFactory display.DisplayableRollFactory
	displayService         display.Service
}

func NewThumbnailListUseCase(
	log *slog.Logger,
	efdService efd.Service,
	displayableRollFactory display.DisplayableRollFactory,
	displayService display.Service,
) list.UseCase {
	return usecase{
		log:                    log,
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
	uc.log.InfoContext(ctx, "starting thumbnail display",
		slog.String("file", filename),
		slog.Bool("strict", strict))

	records, err := uc.efdService.RecordsFromFile(ctx, filename)
	if err != nil {
		return fmt.Errorf("%w %q: %w", ErrFailedToReadFile, filename, err)
	}

	uc.log.DebugContext(ctx, "efd file read",
		slog.Int("frame_count", len(records.EFRMs)))

	dr, err := uc.displayableRollFactory.Create(ctx, records, strict)
	if err != nil {
		return fmt.Errorf("%w %q: %w", ErrFailedToParseFile, filename, err)
	}

	uc.log.DebugContext(ctx, "displayable thumbnails created",
		slog.Int("frame_count", len(dr.Frames)))

	uc.displayService.DisplayThumbnails(ctx, os.Stdout, dr)

	uc.log.InfoContext(ctx, "thumbnail display completed successfully")

	return nil
}
