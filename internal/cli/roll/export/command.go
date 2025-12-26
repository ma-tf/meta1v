package export

import (
	"context"
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/spf13/cobra"
)

//go:generate mockgen -destination=./mocks/usecase_mock.go -package=export_test github.com/ma-tf/meta1v/internal/cli/roll/export UseCase
type UseCase interface {
	Export(ctx context.Context, filename string) error
}

func NewCommand(log *slog.Logger, uc UseCase) *cobra.Command {
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

			return uc.Export(ctx, args[0])
		},
	}
}
