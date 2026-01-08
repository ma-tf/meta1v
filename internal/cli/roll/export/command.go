package export

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/spf13/cobra"
)

const (
	requiredArgs      = 2
	argsMissingTarget = 1
	argsMissingEFD    = 0
)

//go:generate mockgen -destination=./mocks/usecase_mock.go -package=export_test github.com/ma-tf/meta1v/internal/cli/roll/export UseCase
type UseCase interface {
	Export(
		ctx context.Context,
		efdFile string,
		outputFile string,
		strict bool,
	) error
}

func NewCommand(log *slog.Logger, uc UseCase) *cobra.Command {
	return &cobra.Command{
		Use: "export <efd_file> <target_file>",
		Args: func(_ *cobra.Command, args []string) error {
			switch len(args) {
			case argsMissingEFD:
				return cli.ErrEFDFileMustBeProvided
			case argsMissingTarget:
				return cli.ErrTargetFileMustBeSpecified
			case requiredArgs:
				return nil
			default:
				return cli.ErrTooManyArguments
			}
		},
		Short: "Export roll information in csv format to stdout or specified file.",
		Long: `Information about the film roll, including film ID, title, load date,
frame count, ISO and user provided remarks.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			strict, err := cmd.Flags().GetBool("strict")
			if err != nil {
				return fmt.Errorf("failed to get strict flag: %w", err)
			}

			log.DebugContext(ctx, "arguments:",
				slog.String("efd_file", args[0]),
				slog.String("target_file", args[1]),
				slog.Bool("strict", strict),
			)

			return uc.Export(ctx, args[0], args[1], strict)
		},
	}
}
