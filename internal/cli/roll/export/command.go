package export

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/spf13/cobra"
)

func NewCommand(log *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "export [filename]",
		Short: "Export roll information in csv format to stdout or specified file.",
		Long: `Information about the film roll, including film ID, title, load date,
frame count, ISO and user provided remarks.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if len(args) != 1 {
				return cli.ErrNoFilenameProvided
			}

			log.DebugContext(ctx, "arguments:",
				slog.String("filename", args[0]),
			)

			file, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()

			uc := NewRollExportUseCase(
				efd.NewService(log),
			)

			return uc.Export(ctx, file)
		},
	}
}
