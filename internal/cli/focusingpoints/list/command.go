//go:generate mockgen -destination=./mocks/usecase_mock.go -package=list_test github.com/ma-tf/meta1v/internal/cli/focusingpoints/list UseCase

// Package list provides the CLI command for displaying focusing point grids from EFD files.
package list

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/spf13/cobra"
)

// UseCase defines the business logic for listing focusing point grids from EFD files.
type UseCase interface {
	// List reads an EFD file and prints a grid of focusing points used by the frames in a human-readable format.
	List(ctx context.Context, filename string, strict bool) error
}

func NewCommand(log *slog.Logger, uc UseCase) *cobra.Command {
	return &cobra.Command{
		Use:   "list <filename>",
		Short: "Display autofocus point grids in human-readable format",
		Long: `Display rendered grids of autofocus points used when capturing each photograph.

For setting autofocus points on the camera, refer to the Canon EOS-1V manual.`,
		Aliases: []string{"ls"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			strict, err := cmd.Flags().GetBool("strict")
			if err != nil {
				return errors.Join(cli.ErrFailedToGetStrictFlag, err)
			}

			log.DebugContext(ctx, "arguments:",
				slog.String("filename", args[0]),
				slog.Bool("strict", strict),
			)

			return uc.List(ctx, args[0], strict)
		},
	}
}
