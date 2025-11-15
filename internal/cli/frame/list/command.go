package list

import (
	"fmt"
	"os"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <filename>",
		Short: "Prints frame information to stdout.",
		Long: `Information about the frames on the roll, including Tv, Av, ISO,
exposure compensation, focus points, custom functions, and more.`,
		Aliases: []string{"ls"},
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cli.ErrNoFilenameProvided
			}

			file, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()

			rs := efd.NewService()

			records, err := rs.RecordsFromFile(file)
			if err != nil {
				return fmt.Errorf("failed read file: %w", err)
			}

			dr, err := display.NewDisplayableRoll(records)
			if err != nil {
				return fmt.Errorf("failed parse file: %w", err)
			}

			if err = dr.DisplayFrames(); err != nil {
				return fmt.Errorf("failed to display frame: %w", err)
			}

			return nil
		},
	}
}
