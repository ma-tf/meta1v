package roll

import (
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli/roll/export"
	"github.com/ma-tf/meta1v/internal/cli/roll/list"
	"github.com/ma-tf/meta1v/internal/service/csv"
	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/ma-tf/meta1v/internal/service/osfs"
	"github.com/ma-tf/meta1v/pkg/records"
	"github.com/spf13/cobra"
)

func NewCommand(log *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roll <command>",
		Short: "Roll information for a specified file.",
		Long: `Information about the film roll, including film ID, title, load date,
frame count, ISO and user provided remarks.`,
		Aliases: []string{"r"},
	}

	fs := osfs.NewFileSystem()
	builder := efd.NewRootBuilder(log)
	parser := efd.NewParser(log, records.NewDefaultThumbnailFactory())

	listUseCase := NewListUseCase(
		efd.NewService(
			log,
			builder,
			parser,
			fs,
		),
		display.NewDisplayableRollFactory(
			display.NewFrameBuilder(false),
		),
		display.NewService(),
	)

	exportUseCase := NewExportUseCase(
		efd.NewService(
			log,
			builder,
			parser,
			fs,
		),
		csv.NewService(),
	)

	cmd.AddCommand(list.NewCommand(log, listUseCase))
	cmd.AddCommand(export.NewCommand(log, exportUseCase))

	return cmd
}
