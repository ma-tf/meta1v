package roll

import (
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli/roll/export"
	"github.com/ma-tf/meta1v/internal/cli/roll/list"
	"github.com/ma-tf/meta1v/internal/container"
	"github.com/spf13/cobra"
)

func NewCommand(log *slog.Logger, ctr *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roll <command>",
		Short: "Roll information for a specified file.",
		Long: `Information about the film roll, including film ID, title, load date,
frame count, ISO and user provided remarks.`,
		Aliases: []string{"r"},
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
