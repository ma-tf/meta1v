package frame

import (
	"github.com/ma-tf/meta1v/internal/cli/frame/list"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frame <command>",
		Short: "Frame information for a specified file.",
		Long: `Information about the frames on the roll, including Tv, Av, ISO,
exposure compensation, focus points, custom functions, and more.`,
		Aliases: []string{"f"},
	}

	cmd.AddCommand(list.NewCommand())

	return cmd
}
