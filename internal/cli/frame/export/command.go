//go:generate mockgen -destination=./mocks/usecase_mock.go -package=export_test github.com/ma-tf/meta1v/internal/cli/frame/export UseCase

// Package export provides the CLI command for exporting frame information to CSV format.
package export

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/spf13/cobra"
)

const (
	minArgs         = 1
	maxArgs         = 2
	targetFileIndex = 1
)

// UseCase defines the business logic for exporting frame information from EFD files.
type UseCase interface {
	// Export reads an EFD file and exports frame information in CSV format to stdout or a specified file.
	Export(
		ctx context.Context,
		efdFile string,
		outputFile *string,
		strict bool,
		force bool,
	) error
}

func NewCommand(log *slog.Logger, uc UseCase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <efd_file> [target_file]",
		Args:  cobra.RangeArgs(minArgs, maxArgs),
		Short: "Export frame information in csv format to stdout or specified file.",
		Long: `Information about the frame, including frame number, exposure
settings, and user provided remarks.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			strict, err := cmd.Flags().GetBool("strict")
			if err != nil {
				return errors.Join(cli.ErrFailedToGetStrictFlag, err)
			}

			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return errors.Join(cli.ErrFailedToGetForceFlag, err)
			}

			var targetFile *string
			if len(args) == maxArgs {
				targetFile = &args[targetFileIndex]
			}

			if force && targetFile == nil {
				return cli.ErrForceFlagRequiresTargetFile
			}

			log.DebugContext(ctx, "arguments:",
				slog.String("efd_file", args[0]),
				slog.Any("target_file", targetFile),
				slog.Bool("strict", strict),
				slog.Bool("force", force),
			)

			return uc.Export(ctx, args[0], targetFile, strict, force)
		},
	}

	cmd.Flags().BoolP("force", "F", false, "overwrite output file if it exists")

	return cmd
}
