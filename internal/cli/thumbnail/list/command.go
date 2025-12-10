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
		Short: "Prints embedded thumbnails as ascii to stdout.",
		Long: `Information about the thumbnail, including the path, as well as the
thumbnail converted to ascii.`,
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

			uc := NewThumbnailListUseCase(
				efd.NewService(log),
				display.NewDisplayableRollFactory(),
				display.NewService(),
			)

			return uc.DisplayThumbnails(ctx, file)
		},
	}
}
