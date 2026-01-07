package frame

import (
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli/frame/list"
	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/ma-tf/meta1v/internal/service/osfs"
	"github.com/ma-tf/meta1v/pkg/records"
	"github.com/spf13/cobra"
)

func NewCommand(log *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frame <command>",
		Short: "Frame information for a specified file.",
		Long: `Information about the frames on the roll, including Tv, Av, ISO,
exposure compensation, focus points, custom functions, and more.`,
		Aliases: []string{"f"},
	}

	uc := NewListUseCase(
		efd.NewService(
			log,
			efd.NewRootBuilder(log),
			efd.NewReader(log, records.NewDefaultThumbnailFactory()),
			osfs.NewFileSystem(),
		),
		display.NewDisplayableRollFactory(
			display.NewFrameBuilder(log, false),
		),
		display.NewService(),
	)

	cmd.AddCommand(list.NewCommand(log, uc))

	return cmd
}
