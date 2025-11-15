package export

import (
	"fmt"
	"os"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "export [filename]",
		Short: "Export roll information in csv format to stdout or specified file.",
		Long: `Information about the film roll, including film ID, title, load date,
frame count, ISO and user provided remarks.`,
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

			_, err = rs.RecordsFromFile(file)
			if err != nil {
				return fmt.Errorf("failed read file: %w", err)
			}

			// export to csv

			return nil
		},
	}
}
