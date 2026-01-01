package customfunctions

import (
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli/customfunctions/list"
	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/ma-tf/meta1v/internal/service/osfs"
	"github.com/ma-tf/meta1v/pkg/records"
	"github.com/spf13/cobra"
)

func NewCommand(log *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "customfunctions <command>",
		Short: "Custom functions used by the frames for a specified file.",
		Long: `Information about the custom functions used by the frames.

For the meaning of each custom function and its respective value, check the
Canon EOS-1V manual.`,
		Aliases: []string{"cf"},
	}

	uc := NewListUseCase(
		efd.NewService(
			log,
			efd.NewRootBuilder(log),
			efd.NewParser(log, records.NewDefaultThumbnailFactory()),
			osfs.NewFileSystem(),
		),
		display.NewDisplayableRollFactory(
			display.NewFrameBuilder(false),
		),
		display.NewService(),
	)

	cmd.AddCommand(list.NewCommand(log, uc))

	return cmd
}
