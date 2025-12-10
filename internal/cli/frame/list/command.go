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
		Use:   "list <filename>",
		Short: "Prints frame information to stdout.",
		Long: `Information about the frames on the roll, including Tv, Av, ISO,
exposure compensation, focus points, custom functions, and more.`,
		Aliases: []string{"ls"},
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

			uc := NewFrameListUseCase(
				efd.NewService(log),
				display.NewDisplayableRollFactory(),
				display.NewService(),
			)

			return uc.DisplayFrames(ctx, file)
		},
	}
}
