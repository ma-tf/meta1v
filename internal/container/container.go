package container

import (
	"log/slog"

	"github.com/ma-tf/meta1v/internal/service/csv"
	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/ma-tf/meta1v/internal/service/exif"
	"github.com/ma-tf/meta1v/internal/service/osfs"
	"github.com/ma-tf/meta1v/pkg/records"
)

type Container struct {
	Logger                 *slog.Logger
	FileSystem             osfs.FileSystem
	EFDService             efd.Service
	DisplayService         display.Service
	DisplayableRollFactory display.DisplayableRollFactory
	CSVService             csv.Service
	ExifServiceFactory     exif.ServiceFactory
}

func New(logger *slog.Logger) *Container {
	fs := osfs.NewFileSystem()
	thumbnailFactory := records.NewDefaultThumbnailFactory()
	frameBuilder := display.NewFrameBuilder(logger)

	return &Container{
		Logger:     logger,
		FileSystem: fs,
		EFDService: efd.NewService(
			logger,
			efd.NewRootBuilder(logger),
			efd.NewReader(logger, thumbnailFactory),
			fs,
		),
		DisplayService:         display.NewService(),
		DisplayableRollFactory: display.NewDisplayableRollFactory(frameBuilder),
		CSVService:             csv.NewService(),
		ExifServiceFactory: exif.NewServiceFactory(
			logger,
			exif.NewExifToolRunner(fs),
		),
	}
}
