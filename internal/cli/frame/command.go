// Package frame provides business logic for listing and exporting frame information from EFD files.
package frame

import (
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli/frame/export"
	"github.com/ma-tf/meta1v/internal/cli/frame/list"
	"github.com/ma-tf/meta1v/internal/container"
	"github.com/spf13/cobra"
)

func NewCommand(log *slog.Logger, ctr *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frame <command>",
		Short: "Frame information for a specified file.",
		Long: `Information about the frames on the roll, including Tv, Av, ISO,
exposure compensation, focus points, custom functions, and more.`,
		Aliases: []string{"f"},
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

	cmd.AddCommand(list.NewCommand(log, listUseCase))
	cmd.AddCommand(export.NewCommand(log, exportUseCase))

	return cmd
}
