// Package customfunctions provides business logic for listing and exporting custom function settings.
package customfunctions

import (
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli/customfunctions/export"
	"github.com/ma-tf/meta1v/internal/cli/customfunctions/list"
	"github.com/ma-tf/meta1v/internal/container"
	"github.com/spf13/cobra"
)

func NewCommand(log *slog.Logger, ctr *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "customfunctions <command>",
		Short: "Custom functions used by the frames for a specified file.",
		Long: `Information about the custom functions used by the frames.

For the meaning of each custom function and its respective value, check the
Canon EOS-1V manual.`,
		Aliases: []string{"cf"},
	}

	listUseCase := NewListUseCase(
		ctr.EFDService,
		ctr.DisplayableRollFactory,
		ctr.DisplayService,
	)

	exportUseCase := NewExportUseCase(
		ctr.EFDService,
		ctr.DisplayableRollFactory,
		ctr.CSVService,
		ctr.FileSystem,
	)

	cmd.AddCommand(export.NewCommand(log, exportUseCase))
	cmd.AddCommand(list.NewCommand(log, listUseCase))

	return cmd
}
