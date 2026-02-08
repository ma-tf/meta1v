//go:generate mockgen -destination=./mocks/usecase_mock.go -package=export_test github.com/ma-tf/meta1v/internal/cli/customfunctions/export UseCase

// Package export provides the CLI command for exporting custom function settings to CSV format.
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

// UseCase defines the business logic for exporting custom function settings from EFD files.
type UseCase interface {
	// Export reads an EFD file and exports custom function settings in CSV format to stdout or a specified file.
	Export(
		ctx context.Context,
		efdFile string,
		outputFile *string,
		strict bool,
		force bool,
	) error
}

func NewCommand(log *slog.Logger, useCase UseCase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <efd_file> [target_file]",
		Args:  cobra.RangeArgs(minArgs, maxArgs),
		Short: "Export custom functions information in csv format to stdout or specified file.",
		Long: `Information about custom functions, including function ID, name, description,
parameters and user provided remarks.`,
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

			return useCase.Export(ctx, args[0], targetFile, strict, force)
		},
	}

	cmd.Flags().BoolP("force", "F", false, "overwrite output file if it exists")

	return cmd
}
