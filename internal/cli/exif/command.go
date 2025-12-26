package exif

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/ma-tf/meta1v/internal/service/exif"
	"github.com/ma-tf/meta1v/internal/service/osfs"
	"github.com/spf13/cobra"
)

func NewCommand(log *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exif <efd_file> <frame_number> <target_file>",
		Short: "Use the specified EFD file to write EXIF data to the target file.",
		RunE: func(command *cobra.Command, args []string) error {
			ctx := command.Context()

			log.DebugContext(ctx, "exif arguments:",
				slog.String("efd_file", args[0]),
				slog.String("frame_number", args[1]),
				slog.String("target_file", args[2]))

			const requiredArgs = 3
			if len(args) != requiredArgs {
				return cli.ErrNoFilenameProvided
			}

			frame, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid specified frame number: %w", err)
			}

			uc := NewUseCase(
				efd.NewService(
					log,
					efd.NewRootBuilder(log),
					osfs.NewFileSystem(),
				),
				exif.NewService(log),
			)

			return uc.ExportExif(ctx, args[0], frame /*args[2]*/)
		},
	}

	return cmd
}

// https://github.com/thetestspecimen/film-exif
// https://analogexif.sourceforge.net/help/analogexif-xmp.php
