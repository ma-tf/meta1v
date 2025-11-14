package roll

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
		Use:   "roll <filename>",
		Short: "Prints all roll information to stdout.",
		Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

			if errD := dr.Display(); errD != nil {
				return fmt.Errorf("failed display file: %w", errD)
			}

			return nil
		},
	}
}
