package roll

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/cli/roll/export"
	"github.com/ma-tf/meta1v/internal/cli/roll/list"
	"github.com/ma-tf/meta1v/internal/service/csv"
	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/ma-tf/meta1v/internal/service/osfs"
)

var (
	ErrFailedToReadFile         = errors.New("failed read file for roll")
	ErrFailedToParseFile        = errors.New("failed to parse file for roll")
	ErrFailedToCreateOutputFile = errors.New(
		"failed to create output file for roll",
	)
	ErrFailedToExport = errors.New("failed to export roll to CSV")
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

	uc.displayService.DisplayRoll(os.Stdout, dr)

	return nil
}

type exportUseCase struct {
	efdService             efd.Service
	displayableRollFactory display.DisplayableRollFactory
	csvService             csv.Service
	fs                     osfs.FileSystem
}

func NewExportUseCase(
	efdService efd.Service,
	displayableRollFactory display.DisplayableRollFactory,
	csvService csv.Service,
	fs osfs.FileSystem,
) export.UseCase {
	return exportUseCase{
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
	records, err := uc.efdService.RecordsFromFile(ctx, efdFile)
	if err != nil {
		return fmt.Errorf("%w %q: %w", ErrFailedToReadFile, efdFile, err)
	}

	dr, err := uc.displayableRollFactory.Create(ctx, records, strict)
	if err != nil {
		return fmt.Errorf("%w %q: %w", ErrFailedToParseFile, efdFile, err)
	}

	var writer osfs.File = os.Stdout

	if outputFile != nil {
		flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
		if force {
			flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		}

		const permission = 0o666 // rw-rw-rw-
		if writer, err = uc.fs.OpenFile(*outputFile, flags, permission); err != nil {
			if !force && errors.Is(err, os.ErrExist) {
				return cli.ErrOutputFileAlreadyExists
			}

			return fmt.Errorf(
				"%w %q: %w",
				ErrFailedToCreateOutputFile,
				*outputFile,
				err,
			)
		}

		defer writer.Close()
	}

	if err = uc.csvService.ExportRoll(writer, dr); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToExport, err)
	}

	return nil
}
