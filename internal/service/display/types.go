// Package display provides formatting and rendering services for Canon EFD metadata.
//
// This package transforms validated domain types into human-readable structures
// suitable for console output, CSV export, and other display formats.
package display

import "github.com/ma-tf/meta1v/pkg/domain"

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
