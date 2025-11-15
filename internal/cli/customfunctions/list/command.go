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
		Use:   "ls <filename>",
		Short: "Print custom functions used by the frames for a specified file.",
		Long: `Print a table of the custom functions used by the frames.

For the meaning of each custom function and its respective value, check the
Canon EOS-1V manual.`,
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

			dr.DisplayCustomFunctions()

			return nil
		},
	}
}
