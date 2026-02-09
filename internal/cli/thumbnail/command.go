// Package thumbnail provides business logic for displaying embedded thumbnails from EFD files.
package thumbnail

import (
	"log/slog"

	"github.com/ma-tf/meta1v/internal/cli/thumbnail/list"
	"github.com/ma-tf/meta1v/internal/container"
	"github.com/spf13/cobra"
)

func NewCommand(log *slog.Logger, ctr *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "thumbnail <command>",
		Short: "Display embedded thumbnail images from EFD files",
		Long: `Display embedded thumbnail images as ASCII art, including the file path and 
rendered ASCII representation.`,
		Aliases: []string{"t", "thumb"},
	}

	uc := NewThumbnailListUseCase(
		log,
		ctr.EFDService,
		ctr.DisplayableRollFactory,
		ctr.DisplayService,
	)

	cmd.AddCommand(list.NewCommand(log, uc))

	return cmd
}
