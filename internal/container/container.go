// meta1v is a command-line tool for viewing and manipulating metadata for Canon EOS-1V files of the EFD format.
// Copyright (C) 2026  Matt F
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
