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

// Package display provides formatting and rendering services for Canon EFD metadata.
//
// This package transforms validated domain types into human-readable structures
// suitable for console output, CSV export, and other display formats.
package display

import "github.com/ma-tf/meta1v/internal/domain"

// DisplayableRoll represents formatted film roll metadata ready for display or export.
type DisplayableRoll struct {
	FilmID         domain.FilmID
	FirstRow       domain.FirstRow
	PerRow         domain.PerRow
	Title          domain.Title
	FilmLoadedDate domain.ValidatedDatetime
	FrameCount     domain.FrameCount
	IsoDX          domain.Iso
	Remarks        domain.Remarks // film name, location, push/pull, etc.

	Frames []DisplayableFrame
}

// DisplayableFrame represents formatted metadata for a single frame, including
// all exposure settings, camera modes, custom functions, and optional thumbnail.
type DisplayableFrame struct {
	FrameNumber  uint
	FilmID       domain.FilmID
	FilmLoadedAt domain.ValidatedDatetime
	IsoDX        domain.Iso

	UserModifiedRecord bool

	FocalLength domain.FocalLength
	MaxAperture domain.Av
	Tv          domain.Tv
	Av          domain.Av
	IsoM        domain.Iso

	ExposureCompensation      domain.ExposureCompensation
	FlashExposureCompensation domain.ExposureCompensation
	FlashMode                 domain.FlashMode
	MeteringMode              domain.MeteringMode
	ShootingMode              domain.ShootingMode

	FilmAdvanceMode  domain.FilmAdvanceMode
	AFMode           domain.AutoFocusMode
	BulbExposureTime domain.BulbExposureTime
	TakenAt          domain.ValidatedDatetime

	MultipleExposure domain.MultipleExposure
	BatteryLoadedAt  domain.ValidatedDatetime

	CustomFunctions domain.CustomFunctions
	Remarks         domain.Remarks

	FocusingPoints DisplayableFocusPoints

	Thumbnail *DisplayableThumbnail
}

// DisplayableFocusPoints represents rendered ASCII art of the 45-point AF grid.
type DisplayableFocusPoints string

// DisplayableThumbnail represents an ASCII art rendering of a frame's thumbnail image.
type DisplayableThumbnail struct {
	Thumbnail string
	Filepath  string
}
