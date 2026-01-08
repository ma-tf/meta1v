package list

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/spf13/cobra"
)

//go:generate mockgen -destination=./mocks/usecase_mock.go -package=list_test github.com/ma-tf/meta1v/internal/cli/customfunctions/list UseCase
type UseCase interface {
	List(
		ctx context.Context,
		filename string,
		strict bool,
	) error
}

func NewCommand(log *slog.Logger, uc UseCase) *cobra.Command {
	return &cobra.Command{
		Use:   "ls <filename>",
		Short: "Print custom functions used by the frames for a specified file.",
		Long: `Print a table of the custom functions used by the frames.

For the meaning of each custom function and its respective value, check the
Canon EOS-1V manual.`,
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
