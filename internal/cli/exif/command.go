package exif

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/spf13/cobra"
)

const (
	requiredArgs           = 3
	argsMissingTarget      = 2
	argsMissingFrameNumber = 1
	argsMissingEFD         = 0
)

var ErrInvalidFrameNumber = errors.New("invalid specified frame number")

//go:generate mockgen -destination=./mocks/usecase_mock.go -package=exif_test github.com/ma-tf/meta1v/internal/cli/exif UseCase
type UseCase interface {
	ExportExif(
		ctx context.Context,
		efdFile string,
		frame int,
		targetFile string,
	) error
}

func NewCommand(log *slog.Logger, uc UseCase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exif <efd_file> <frame_number> <target_file>",
		Short: "Use the specified EFD file to write EXIF data to the target file.",
		Args: func(_ *cobra.Command, args []string) error {
			switch len(args) {
			case argsMissingTarget:
				return cli.ErrTargetFileMustBeSpecified
			case argsMissingFrameNumber:
				return cli.ErrFrameNumberMustBeSpecified
			case argsMissingEFD:
				return cli.ErrEFDFileMustBeProvided
			case requiredArgs:
				return nil
			default:
				return cli.ErrTooManyArguments
			}
		},
		RunE: func(command *cobra.Command, args []string) error {
			ctx := command.Context()

			log.DebugContext(ctx, "exif arguments:",
				slog.String("efd_file", args[0]),
				slog.String("frame_number", args[1]),
				slog.String("target_file", args[2]))

			frame, err := strconv.Atoi(args[1])
			if err != nil {
				return errors.Join(ErrInvalidFrameNumber, err)
			}

			return uc.ExportExif(ctx, args[0], frame, args[2])
		},
	}

	return cmd
}

// https://github.com/thetestspecimen/film-exif
// https://analogexif.sourceforge.net/help/analogexif-xmp.php
