package list

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/spf13/cobra"
)

//go:generate mockgen -destination=./mocks/usecase_mock.go -package=list_test github.com/ma-tf/meta1v/internal/cli/focusingpoints/list UseCase
type UseCase interface {
	List(ctx context.Context, filename string, strict bool) error
}

func NewCommand(log *slog.Logger, uc UseCase) *cobra.Command {
	return &cobra.Command{
		Use:   "ls <filename>",
		Short: "Print a grid of focusing points grid used by the frames for a specified file.",
		Long: `Print a rendered grid of focusing points used by the auto focus when
taking a photograph to stdout.

For the setting focusing points on the camera, check the Canon EOS-1V manual.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if len(args) != 1 {
				return cli.ErrEFDFileMustBeProvided
			}

			strict, err := cmd.Flags().GetBool("strict")
			if err != nil {
				return fmt.Errorf("failed to get strict flag: %w", err)
			}

			log.DebugContext(ctx, "arguments:",
				slog.String("filename", args[0]),
				slog.Bool("strict", strict),
			)

			return uc.List(ctx, args[0], strict)
		},
	}
}
