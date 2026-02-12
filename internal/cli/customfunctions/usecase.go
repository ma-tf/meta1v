// meta1v is a command-line tool for viewing and manipulating metadata for Canon EOS-1V files of the EFD format.
// Copyright (C) 2026  Matt F
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package customfunctions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/cli/customfunctions/export"
	"github.com/ma-tf/meta1v/internal/cli/customfunctions/ls"
	"github.com/ma-tf/meta1v/internal/service/csvexport"
	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/ma-tf/meta1v/internal/service/osfs"
)

const permission = 0o666 // rw-rw-rw-

var (
	ErrFailedToReadFile  = errors.New("failed read file for custom functions")
	ErrFailedToParseFile = errors.New(
		"failed to parse file for custom functions",
	)
	ErrFailedToDisplay = errors.New(
		"failed to display custom functions",
	)
	ErrFailedToCreateOutputFile = errors.New(
		"failed to create output file for custom functions",
	)
	ErrFailedToWriteCSV = errors.New("failed to write custom functions to csv")
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
) ls.UseCase {
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
	uc.log.InfoContext(ctx, "starting custom functions list",
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

	uc.log.DebugContext(ctx, "displayable custom functions created",
		slog.Int("frame_count", len(dr.Frames)))

	err = uc.displayService.DisplayCustomFunctions(ctx, os.Stdout, dr)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToDisplay, err)
	}

	uc.log.InfoContext(ctx, "custom functions list completed successfully")

	return nil
}

type exportUseCase struct {
	log                    *slog.Logger
	efdService             efd.Service
	displayableRollFactory display.DisplayableRollFactory
	csvService             csvexport.Service
	fs                     osfs.FileSystem
}

func NewExportUseCase(
	log *slog.Logger,
	efdService efd.Service,
	displayableRollFactory display.DisplayableRollFactory,
	csvService csvexport.Service,
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
	uc.log.InfoContext(ctx, "starting custom functions export",
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

	uc.log.DebugContext(ctx, "displayable custom functions created",
		slog.Int("frame_count", len(dr.Frames)))

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

	if err = uc.csvService.ExportCustomFunctions(ctx, writer, dr); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToWriteCSV, err)
	}

	uc.log.InfoContext(ctx, "custom functions export completed successfully")

	return nil
}
