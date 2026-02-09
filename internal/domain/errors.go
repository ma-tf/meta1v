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

package domain

import "errors"

var (
	ErrMultipleThumbnailsForFrame = errors.New("frame has multiple thumbnails")
	ErrFrameIndexOutOfRange       = errors.New("frame index out of range")

	ErrPrefixOutOfRange          = errors.New("prefix out of range (0-99)")
	ErrSuffixOutOfRange          = errors.New("suffix out of range (0-999)")
	ErrFirstRowGreaterThanPerRow = errors.New(
		"first row cannot be greater than per row",
	)
	ErrInvalidDateTime     = errors.New("invalid date time")
	ErrInvalidAv           = errors.New("invalid Av value")
	ErrInvalidTv           = errors.New("invalid Tv value")
	ErrUnknownExposureComp = errors.New(
		"unknown exposure compensation value",
	)
	ErrUnknownFlashMode        = errors.New("unknown flash mode")
	ErrUnknownMeteringMode     = errors.New("unknown metering mode")
	ErrUnknownShootingMode     = errors.New("unknown shooting mode")
	ErrUnknownFilmAdvanceMode  = errors.New("unknown film advance mode")
	ErrUnknownAutoFocusMode    = errors.New("unknown auto focus mode")
	ErrInvalidBulbTime         = errors.New("invalid bulb exposure time")
	ErrUnknownMultipleExposure = errors.New("unknown multiple exposure value")
	ErrInvalidCustomFunction   = errors.New("invalid custom function")
)
