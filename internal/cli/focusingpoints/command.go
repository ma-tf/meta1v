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

// Package focusingpoints provides business logic for displaying focusing point grids from EFD files.
package focusingpoints

import (
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli/focusingpoints/ls"
	"github.com/ma-tf/meta1v/internal/container"
	"github.com/spf13/cobra"
)

func NewCommand(log *slog.Logger, ctr *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "focusingpoints <command>",
		Short: "Display autofocus point grids from EFD files",
		Long: `Display rendered grids of autofocus points used when capturing each photograph.

For setting autofocus points on the camera, refer to the Canon EOS-1V manual.`,
		Aliases: []string{"fp"},
	}

	uc := NewListUseCase(
		log,
		ctr.EFDService,
		ctr.DisplayableRollFactory,
		ctr.DisplayService,
	)

	cmd.AddCommand(ls.NewCommand(log, uc))

	return cmd
}
