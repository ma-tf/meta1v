package thumbnail

import (
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli/thumbnail/list"
	"github.com/spf13/cobra"
)

func NewCommand(log *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "thumbnail <command>",
		Short: "Prints embedded thumbnails as ascii to stdout.",
		Long: `Information about the thumbnail, including the path, as well as the
thumbnail converted to ascii.`,
		Aliases: []string{"t", "thumb"},
	}

	cmd.AddCommand(list.NewCommand(log))

	return cmd
}
