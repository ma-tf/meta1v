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
		Short: "List or export custom function settings from EFD files",
		Long: `Display or export custom function settings used by the frames.

For the meaning of each custom function and its respective value, refer to the 
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
