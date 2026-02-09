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

	cmd.AddCommand(list.NewCommand(log, listUseCase))
	cmd.AddCommand(export.NewCommand(log, exportUseCase))

	return cmd
}
