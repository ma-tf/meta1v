package roll

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/cli/roll/export"
	"github.com/ma-tf/meta1v/internal/cli/roll/list"
	"github.com/ma-tf/meta1v/internal/service/csv"
	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/ma-tf/meta1v/internal/service/osfs"
)

const permission = 0o666 // rw-rw-rw-

var (
	ErrFailedToReadFile         = errors.New("failed to read file for roll")
	ErrFailedToParseFile        = errors.New("failed to parse file for roll")
	ErrFailedToCreateOutputFile = errors.New(
		"failed to create output file for roll",
	)
	ErrFailedToExport = errors.New("failed to export roll to CSV")
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
	uc.log.InfoContext(ctx, "starting roll list",
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

	uc.log.DebugContext(ctx, "displayable roll created",
		slog.String("film_id", string(dr.FilmID)))

	uc.displayService.DisplayRoll(ctx, os.Stdout, dr)

	uc.log.InfoContext(ctx, "roll list completed successfully")

	return nil
}

type exportUseCase struct {
	log                    *slog.Logger
	efdService             efd.Service
	displayableRollFactory display.DisplayableRollFactory
	csvService             csv.Service
	fs                     osfs.FileSystem
}

func NewExportUseCase(
	log *slog.Logger,
	efdService efd.Service,
	displayableRollFactory display.DisplayableRollFactory,
	csvService csv.Service,
	fs osfs.FileSystem,
) export.UseCase {
	return exportUseCase{
		log:                    log,
		efdService:             efdService,
		displayableRollFactory: displayableRollFactory,
		csvService:             csvService,
		fs:                     fs,
	}
}

func (uc exportUseCase) Export(
	ctx context.Context,
	efdFile string,
	outputFile *string,
	strict bool,
	force bool,
) error {
	uc.log.InfoContext(ctx, "starting roll export",
		slog.String("efd_file", efdFile),
		slog.Any("output_file", outputFile),
		slog.Bool("strict", strict),
		slog.Bool("force", force))

	records, err := uc.efdService.RecordsFromFile(ctx, efdFile)
	if err != nil {
		return fmt.Errorf("%w %q: %w", ErrFailedToReadFile, efdFile, err)
	}

	uc.log.DebugContext(ctx, "efd file read",
		slog.Int("frame_count", len(records.EFRMs)))

	dr, err := uc.displayableRollFactory.Create(ctx, records, strict)
	if err != nil {
		return fmt.Errorf("%w %q: %w", ErrFailedToParseFile, efdFile, err)
	}

	uc.log.DebugContext(ctx, "displayable roll created",
		slog.String("film_id", string(dr.FilmID)))

	var writer osfs.File = os.Stdout

	if outputFile != nil {
		flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
		if force {
			flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		}

		writer, err = uc.fs.OpenFile(*outputFile, flags, permission)
		if err != nil {
			if !force && errors.Is(err, os.ErrExist) {
				return cli.ErrOutputFileAlreadyExists
			}

			return fmt.Errorf("%w %q: %w",
				ErrFailedToCreateOutputFile, *outputFile, err)
		}

		defer writer.Close()

		uc.log.DebugContext(ctx, "output file opened",
			slog.String("file", *outputFile))
	}

	if err = uc.csvService.ExportRoll(ctx, writer, dr); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToExport, err)
	}

	uc.log.InfoContext(ctx, "roll export completed successfully")

	return nil
}
