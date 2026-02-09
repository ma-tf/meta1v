// Package container provides dependency injection for meta1v services.
//
// It wires together all the services, repositories, and infrastructure components
// needed by the application, making them available through a single Container struct.
package container

import (
	"log/slog"

	"github.com/ma-tf/meta1v/internal/records"
	"github.com/ma-tf/meta1v/internal/service/csv"
	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/ma-tf/meta1v/internal/service/exif"
	"github.com/ma-tf/meta1v/internal/service/osexec"
	"github.com/ma-tf/meta1v/internal/service/osfs"
)

// Container holds all application dependencies and services.
// It provides a centralized location for dependency management and injection.
type Container struct {
	Logger                 *slog.Logger
	FileSystem             osfs.FileSystem
	LookPath               osexec.LookPath
	EFDService             efd.Service
	DisplayService         display.Service
	DisplayableRollFactory display.DisplayableRollFactory
	CSVService             csv.Service
	ExifService            exif.Service
}

// New creates and initializes a Container with all required services and dependencies.
func New(logger *slog.Logger, lookPath osexec.LookPath) *Container {
	fs := osfs.NewFileSystem()
	thumbnailFactory := records.NewDefaultThumbnailFactory()
	frameBuilder := display.NewFrameBuilder(logger)

	return &Container{
		Logger:     logger,
		FileSystem: fs,
		LookPath:   lookPath,
		EFDService: efd.NewService(
			logger,
			efd.NewRootBuilder(logger),
			efd.NewReader(logger, thumbnailFactory),
			fs,
		),
		DisplayService:         display.NewService(logger),
		DisplayableRollFactory: display.NewDisplayableRollFactory(frameBuilder),
		CSVService:             csv.NewService(logger),
		ExifService: exif.NewService(
			logger,
			exif.NewExifToolRunner(
				fs,
				exif.NewExiftoolCommandFactory(lookPath),
			),
			exif.NewExifBuilder(logger),
		),
	}
}
