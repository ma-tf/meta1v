package list

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/spf13/cobra"
)

func NewCommand(log *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "ls <filename>",
		Short: "Print custom functions used by the frames for a specified file.",
		Long: `Print a table of the custom functions used by the frames.

For the meaning of each custom function and its respective value, check the
Canon EOS-1V manual.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if len(args) != 1 {
				return cli.ErrNoFilenameProvided
			}

			log.DebugContext(ctx, "arguments:",
				slog.String("filename", args[0]))

			file, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()

			log.DebugContext(ctx, "opened file:",
				slog.String("filename", args[0]))

			uc := NewCustomFunctionsListUseCase(
				efd.NewService(log),
				display.NewDisplayableRollFactory(),
				display.NewService(),
			)

			return uc.DisplayCustomFunctions(ctx, file)
		},
	}
}
