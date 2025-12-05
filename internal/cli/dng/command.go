package dng

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/ma-tf/meta1v/internal/cli"
	"github.com/ma-tf/meta1v/internal/service/dng"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/spf13/cobra"
)

func NewCommand(log *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dng <efd_file> <frame_number> <dng_file>",
		Short: "Export EXIF data of a frame to a dng file for a specified file.",
		RunE: func(command *cobra.Command, args []string) error {
			ctx := command.Context()

			log.DebugContext(ctx, "arguments:",
				slog.String("efd_file", args[0]),
				slog.String("frame_number", args[1]),
				slog.String("dng_file", args[2]))

			const requiredArgs = 3
			if len(args) != requiredArgs {
				return cli.ErrNoFilenameProvided
			}

			file, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()

			rs := efd.NewService()

			records, err := rs.RecordsFromFile(file)
			if err != nil {
				return fmt.Errorf("failed to interpret file content: %w", err)
			}

			fn, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid specified frame number: %w", err)
			}

			s := dng.NewService(log, records, fn)

			target := "./test_files/20251011_Japan 1_0.dng"

			err = s.WriteEXIF(ctx, target)
			if err != nil {
				return fmt.Errorf("write exif failed: %w", err)
			}

			return nil
		},
	}

	return cmd
}

// https://github.com/thetestspecimen/film-exif
// https://analogexif.sourceforge.net/help/analogexif-xmp.php
