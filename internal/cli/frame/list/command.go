//go:generate mockgen -destination=./mocks/usecase_mock.go -package=list_test github.com/ma-tf/meta1v/internal/cli/frame/list UseCase

// Package list provides the CLI command for listing frame information from EFD files.
package list

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/spf13/cobra"
)

// UseCase defines the business logic for listing frame information from EFD files.
type UseCase interface {
	// List reads an EFD file and prints frame information in a human-readable format.
	List(ctx context.Context, filename string, strict bool) error
}

func NewCommand(log *slog.Logger, uc UseCase) *cobra.Command {
	return &cobra.Command{
		Use:   "list <filename>",
		Short: "Display frame information in human-readable format",
		Long: `Display detailed information about frames including exposure settings (Tv, Av, ISO), 
exposure compensation, focus points, custom functions, and user-provided remarks.`,
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
