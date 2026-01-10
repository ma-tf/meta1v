package customfunctions

import (
	"log/slog"

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

	uc := NewListUseCase(
		ctr.EFDService,
		ctr.DisplayableRollFactory,
		ctr.DisplayService,
	)

	cmd.AddCommand(list.NewCommand(log, uc))

	return cmd
}
