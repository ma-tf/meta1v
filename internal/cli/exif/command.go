//go:generate mockgen -destination=./mocks/usecase_mock.go -package=exif_test github.com/ma-tf/meta1v/internal/cli/exif UseCase

// Package exif provides the CLI command for writing EXIF metadata to image files.
package exif

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/spf13/cobra"
)

const requiredArgsCount = 3

var ErrInvalidFrameNumber = errors.New("invalid specified frame number")

// UseCase defines the business logic for exporting EXIF metadata from EFD files.
type UseCase interface {
	// ExportExif writes EXIF metadata from a specific frame to a target image file.
	ExportExif(
		ctx context.Context,
		efdFile string,
		frame int,
		targetFile string,
		strict bool,
	) error
}

func NewCommand(log *slog.Logger, uc UseCase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exif <efd_file> <frame_number> <target_file>",
		Short: "Write EXIF metadata from EFD file to target image file",
		Long: `Extract exposure metadata (Tv, Av, ISO, exposure compensation) from a specific 
frame in an EFD file and write it as EXIF data to a target image file.`,
		Args: cobra.ExactArgs(requiredArgsCount),
		RunE: func(command *cobra.Command, args []string) error {
			ctx := command.Context()

			strict, err := command.Flags().GetBool("strict")
			if err != nil {
				return errors.Join(cli.ErrFailedToGetStrictFlag, err)
			}

			log.DebugContext(ctx, "exif arguments:",
				slog.String("efd_file", args[0]),
				slog.String("frame_number", args[1]),
				slog.String("target_file", args[2]),
				slog.Bool("strict", strict))

			frame, err := strconv.Atoi(args[1])
			if err != nil {
				return errors.Join(ErrInvalidFrameNumber, err)
			}

			return uc.ExportExif(ctx, args[0], frame, args[2], strict)
		},
	}

	return cmd
}

// https://github.com/thetestspecimen/film-exif
// https://analogexif.sourceforge.net/help/analogexif-xmp.php
