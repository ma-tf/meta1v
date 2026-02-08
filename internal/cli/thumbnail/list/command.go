//go:generate mockgen -destination=./mocks/usecase_mock.go -package=list_test github.com/ma-tf/meta1v/internal/cli/thumbnail/list UseCase

// Package list provides the CLI command for displaying embedded thumbnails from EFD files.
package list

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/spf13/cobra"
)

// UseCase defines the business logic for displaying embedded thumbnails from EFD files.
type UseCase interface {
	// DisplayThumbnails reads an EFD file and displays embedded thumbnails as ASCII art to stdout.
	DisplayThumbnails(ctx context.Context, filename string, strict bool) error
}

func NewCommand(log *slog.Logger, uc UseCase) *cobra.Command {
	return &cobra.Command{
		Use:   "list <filename>",
		Short: "Prints embedded thumbnails as ascii to stdout.",
		Long: `Information about the thumbnail, including the path, as well as the
thumbnail converted to ascii.`,
		Aliases: []string{"ls"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			strict, err := cmd.Flags().GetBool("strict")
			if err != nil {
				return errors.Join(cli.ErrFailedToGetStrictFlag, err)
			}

			log.DebugContext(ctx, "arguments:",
				slog.String("filename", args[0]))

			return uc.DisplayThumbnails(ctx, args[0], strict)
		},
	}
}
