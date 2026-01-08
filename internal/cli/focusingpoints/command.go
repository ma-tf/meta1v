package focusingpoints

import (
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli/focusingpoints/list"
	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/ma-tf/meta1v/internal/service/osfs"
	"github.com/ma-tf/meta1v/pkg/records"
	"github.com/spf13/cobra"
)

func NewCommand(log *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "focusingpoints <command>",
		Short: "Focusing points grid used by the frames for a specified file.",
		Long: `A rendered grid of focusing points used by the auto focus when
taking a photograph.

For the setting focusing points on the camera, check the Canon EOS-1V manual.`,
		Aliases: []string{"fp"},
	}

	uc := NewListUseCase(
		efd.NewService(
			log,
			efd.NewRootBuilder(log),
			efd.NewReader(log, records.NewDefaultThumbnailFactory()),
			osfs.NewFileSystem(),
		),
		display.NewDisplayableRollFactory(
			display.NewFrameBuilder(log),
		),
		display.NewService(),
	)

	cmd.AddCommand(list.NewCommand(log, uc))

	return cmd
}
