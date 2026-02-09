package focusingpoints

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	log                    *slog.Logger
	efdService             efd.Service
	displayableRollFactory display.DisplayableRollFactory
	displayService         display.Service
}

func NewListUseCase(
	log *slog.Logger,
	efdService efd.Service,
	displayableRollFactory display.DisplayableRollFactory,
	displayService display.Service,
) list.UseCase {
	return listUseCase{
		log:                    log,
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
	uc.log.InfoContext(ctx, "starting focusing points list",
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

	uc.log.DebugContext(ctx, "displayable focusing points created",
		slog.Int("frame_count", len(dr.Frames)))

	uc.displayService.DisplayFocusingPoints(ctx, os.Stdout, dr)

	uc.log.InfoContext(ctx, "focusing points list completed successfully")

	return nil
}
