package roll

import (
	"context"
	"errors"
	"os"

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

	dr, err := uc.displayableRollFactory.Create(ctx, records)
	if err != nil {
		return errors.Join(ErrFailedToParseFile, err)
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
	outputFile string,
) error {
	records, err := uc.efdService.RecordsFromFile(ctx, efdFile)
	if err != nil {
		return errors.Join(ErrFailedToReadFile, err)
	}

	dr, err := uc.displayableRollFactory.Create(ctx, records)
	if err != nil {
		return errors.Join(ErrFailedToParseFile, err)
	}

	file, err := uc.fs.Create(outputFile)
	if err != nil {
		return errors.Join(ErrFailedToCreateOutputFile, err)
	}
	defer file.Close()

	return uc.csvService.ExportRoll(file, dr)
}
