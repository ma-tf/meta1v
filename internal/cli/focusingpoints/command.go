// Package focusingpoints provides business logic for displaying focusing point grids from EFD files.
package focusingpoints

import (
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli/focusingpoints/list"
	"github.com/ma-tf/meta1v/internal/container"
	"github.com/spf13/cobra"
)

func NewCommand(log *slog.Logger, ctr *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "focusingpoints <command>",
		Short: "Focusing points grid used by the frames for a specified file.",
		Long: `A rendered grid of focusing points used by the auto focus when
taking a photograph.

For the setting focusing points on the camera, check the Canon EOS-1V manual.`,
		Aliases: []string{"fp"},
	}

	uc := NewListUseCase(
		ctr.EFDService,
		ctr.DisplayableRollFactory,
		ctr.DisplayService,
	)

	cmd.AddCommand(list.NewCommand(log, uc))

	return cmd
}
