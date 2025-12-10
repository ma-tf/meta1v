package focusingpoints

import (
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli/focusingpoints/list"
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

	cmd.AddCommand(list.NewCommand(log))

	return cmd
}
