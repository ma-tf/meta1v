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

// Package frame provides business logic for listing and exporting frame information from EFD files.
package frame

import (
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli/frame/export"
	"github.com/ma-tf/meta1v/internal/cli/frame/ls"
	"github.com/ma-tf/meta1v/internal/container"
	"github.com/spf13/cobra"
)

func NewCommand(log *slog.Logger, ctr *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frame <command>",
		Short: "List or export frame information from EFD files",
		Long: `Display or export detailed information about frames on the roll, including 
exposure settings (Tv, Av, ISO), exposure compensation, focus points, custom functions, 
and user-provided remarks.`,
		Aliases: []string{"f"},
	}

	listUseCase := NewListUseCase(
		log,
		ctr.EFDService,
		ctr.DisplayableRollFactory,
		ctr.DisplayService,
	)

	exportUseCase := NewExportUseCase(
		log,
		ctr.EFDService,
		ctr.DisplayableRollFactory,
		ctr.CSVService,
		ctr.FileSystem,
	)

	cmd.AddCommand(ls.NewCommand(log, listUseCase))
	cmd.AddCommand(export.NewCommand(log, exportUseCase))

	return cmd
}
