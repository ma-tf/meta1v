package roll

import (
	"github.com/ma-tf/meta1v/internal/cli/roll/export"
	"github.com/ma-tf/meta1v/internal/cli/roll/list"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roll <command>",
		Short: "Roll information for a specified file.",
		Long: `Information about the film roll, including film ID, title, load date,
frame count, ISO and user provided remarks.`,
		Aliases: []string{"r"},
	}

	cmd.AddCommand(list.NewCommand())
	cmd.AddCommand(export.NewCommand())

	return cmd
}
