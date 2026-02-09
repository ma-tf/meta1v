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

// Package domain provides validated business domain types for Canon EFD metadata.
//
// This package defines strongly-typed wrappers for camera metadata values extracted
// from Canon EFD binary files, including film identifiers, exposure settings,
// camera modes, and timestamps. Each type includes validation logic to ensure
// data integrity and proper formatting.
package domain

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"time"
)

//nolint:gochecknoglobals // global defaultMaps is acceptable here
var defaultMaps = NewMapProvider()

// FilmID represents a validated film roll identifier in the format "XX-YYY"
// where XX is a 2-digit prefix (0-99) and YYY is a 3-digit suffix (0-999).
type FilmID string

// NewFilmID creates a validated FilmID from prefix and suffix values.
// Returns an empty FilmID if both values are math.MaxUint32 (indicating no film ID).
// Returns an error if the prefix is outside 0-99 or suffix is outside 0-999.
func NewFilmID(prefix, suffix uint32) (FilmID, error) {
	if prefix == math.MaxUint32 && suffix == math.MaxUint32 {
		return "", nil
	}

	const maxA, maxB = 99, 999
	if prefix > maxA {
		return "", fmt.Errorf(
			"%w: got %d (valid: 0-99)",
			ErrPrefixOutOfRange,
			prefix,
		)
	}

	if suffix > maxB {
		return "", fmt.Errorf(
			"%w: got %d (valid: 0-999)",
			ErrSuffixOutOfRange,
			suffix,
		)
	}

	s := fmt.Sprintf("%02d-%03d", prefix, suffix)

	return FilmID(s), nil
}

// FirstRow represents the number of frames in the first row of a contact sheet.
// This value is calculated as (perRow - firstRow) to support variable-length first rows.
type FirstRow string

// NewFirstRow creates a FirstRow value from the raw firstRow and perRow byte values.
// Returns an error if firstRow is greater than perRow.
func NewFirstRow(firstRow uint8, perRow uint8) (FirstRow, error) {
	if firstRow > perRow {
		return "", fmt.Errorf("%w (%d > %d)",
			ErrFirstRowGreaterThanPerRow, firstRow, perRow)
	}

	return FirstRow(strconv.Itoa(int(perRow - firstRow))), nil
}

// PerRow represents the number of frames per row in a contact sheet display.
type PerRow string

// NewPerRow creates a PerRow from the raw value.
func NewPerRow(perRow uint8) PerRow {
	return PerRow(strconv.Itoa(int(perRow)))
}

// FrameCount represents the total number of frames in a film roll.
type FrameCount string

// NewFrameCount creates a FrameCount from the raw value.
func NewFrameCount(fc uint32) FrameCount {
	return FrameCount(strconv.FormatUint(uint64(fc), 10))
}

// ValidatedDatetime represents a validated date-time string in the format "YYYY-MM-DD HH:MM:SS".
// Empty string indicates that no date-time was recorded.
type ValidatedDatetime string

// NewDateTime creates a validated datetime string from individual component values.
// Returns an empty string if all values are zero or all are max values (indicating no datetime).
// Returns an error if the values form an invalid date or time.
func NewDateTime(
	year uint16,
	month,
	day,
	hour,
	minute,
	second uint8,
) (ValidatedDatetime, error) {
	if (year == math.MaxUint16 && month&day&hour&minute&second == math.MaxUint8) ||
		(year == 0 && month&day&hour&minute&second == 0) {
		return "", nil
	}

	// manual says year limit is 2000 - 2099, I won't validate this for now
	rawDate := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
		year, month, day, hour, minute, second)

	t, err := time.Parse(time.DateTime, rawDate) // performs format validation
	if err != nil || t.IsZero() {
		return "", fmt.Errorf("%w: %s", ErrInvalidDateTime, rawDate)
	}

	return ValidatedDatetime(t.Format(time.DateTime)), nil
}

// Title represents a null-terminated film roll title string (max 64 bytes).
type Title string

// NewTitle extracts a Title from a 64-byte array, reading up to the first null byte.
func NewTitle(t [64]byte) Title {
	return Title(t[:bytes.IndexByte(t[:], 0)])
}

// Remarks represents null-terminated user remarks about a frame or roll (max 256 bytes).
type Remarks string

// NewRemarks extracts Remarks from a 256-byte array, reading up to the first null byte.
func NewRemarks(r [256]byte) Remarks {
	return Remarks(r[:bytes.IndexByte(r[:], 0)])
}

// FocalLength represents the lens focal length in millimeters.
// Empty string indicates unknown or unavailable focal length.
type FocalLength string

// NewFocalLength creates a FocalLength from the raw value.
// Returns empty string if the value is math.MaxUint32 (indicating unavailable).
func NewFocalLength(fl uint32) FocalLength {
	if fl == math.MaxUint32 {
		return ""
	}

	return FocalLength(strconv.FormatUint(uint64(fl), 10))
}

// Tv represents a shutter speed (Time Value) in human-readable format.
// Examples: "1/500", "2\"", "Bulb". Empty string indicates unavailable.
type Tv string

// NewTv creates a Tv from the raw camera value using the lookup map.
// In strict mode, returns an error for unknown values.
// In non-strict mode, returns the raw numeric value as a string for unknown values.
func NewTv(tv int32, strict bool) (Tv, error) {
	if tv == -1 {
		return "", nil
	}

	val, ok := defaultMaps.GetTv(tv)
	if !ok {
		if strict {
			return "", fmt.Errorf(
				"%w: raw value %d",
				ErrInvalidTv,
				tv,
			)
		}

		val = Tv(strconv.Itoa(int(tv)))
	}

	return val, nil
}

// Av represents an aperture value (Aperture Value) in f-stop notation.
// Examples: "2.8", "5.6", "16". Empty string indicates unavailable.
type Av string

// NewAv creates an Av from the raw camera value using the lookup map.
// In strict mode, returns an error for unknown values.
// In non-strict mode, returns the raw numeric value as a string for unknown values.
func NewAv(av uint32, strict bool) (Av, error) {
	if av == math.MaxUint32 {
		return "", nil
	}

	val, ok := defaultMaps.GetAv(av)
	if !ok {
		if strict {
			return "", fmt.Errorf(
				"%w: raw value %d",
				ErrInvalidAv,
				av,
			)
		}

		val = Av(strconv.Itoa(int(av)))
	}

	return val, nil
}

// Iso represents an ISO sensitivity value (e.g., "100", "400", "1600").
// Empty string indicates unavailable.
type Iso string

// NewIso creates an Iso from the raw camera value.
// Returns empty string if the value is math.MaxUint32 (indicating unavailable).
func NewIso(iso uint32) Iso {
	if iso == math.MaxUint32 {
		return ""
	}

	return Iso(strconv.FormatUint(uint64(iso), 10))
}

// ExposureCompensation represents exposure compensation in stops (e.g., "+1.0", "-0.5", "0.0").
type ExposureCompensation string

// NewExposureCompensation creates an ExposureCompensation from the raw camera value.
// In strict mode, returns an error for unknown values.
// In non-strict mode, calculates the compensation value by dividing by 10.
func NewExposureCompensation(
	ec int32,
	strict bool,
) (ExposureCompensation, error) {
	if ec == -1 {
		return "", nil
	}

	val, ok := defaultMaps.GetExposureCompensation(ec)
	if !ok {
		if strict {
			return "", fmt.Errorf(
				"%w: raw value %d",
				ErrUnknownExposureComp,
				ec,
			)
		}

		var prefix string
		if ec >= 0 {
			prefix = "+"
		}

		const divisor = 10

		val = ExposureCompensation(
			prefix + fmt.Sprintf("%.1f", float64(ec)/divisor),
		)
	}

	return val, nil
}

// FlashMode represents the flash operation mode (e.g., "OFF", "ON", "TTL autoflash").
type FlashMode string

// NewFlashMode creates a FlashMode from the raw camera value using the lookup map.
// Returns an error if the value is not found in the map.
func NewFlashMode(fm uint32) (FlashMode, error) {
	val, ok := defaultMaps.GetFlashMode(fm)
	if !ok {
		return "", fmt.Errorf("%w: raw value %d", ErrUnknownFlashMode, fm)
	}

	return val, nil
}

// MeteringMode represents the light metering mode (e.g., "Evaluative", "Partial", "Center-weighted").
type MeteringMode string

// NewMeteringMode creates a MeteringMode from the raw camera value using the lookup map.
// Returns an error if the value is not found in the map.
func NewMeteringMode(mm uint32) (MeteringMode, error) {
	val, ok := defaultMaps.GetMeteringMode(mm)
	if !ok {
		return "", fmt.Errorf("%w: raw value %d", ErrUnknownMeteringMode, mm)
	}

	return val, nil
}

// ShootingMode represents the camera shooting mode (e.g., "Program", "Av", "Tv", "Manual").
type ShootingMode string

// NewShootingMode creates a ShootingMode from the raw camera value using the lookup map.
// Returns an error if the value is not found in the map.
func NewShootingMode(sm uint32) (ShootingMode, error) {
	val, ok := defaultMaps.GetShootingMode(sm)
	if !ok {
		return "", fmt.Errorf("%w: raw value %d", ErrUnknownShootingMode, sm)
	}

	return val, nil
}

// FilmAdvanceMode represents the film advance mode (e.g., "Single frame", "Continuous").
type FilmAdvanceMode string

// NewFilmAdvanceMode creates a FilmAdvanceMode from the raw camera value using the lookup map.
// Returns an error if the value is not found in the map.
func NewFilmAdvanceMode(fam uint32) (FilmAdvanceMode, error) {
	val, ok := defaultMaps.GetFilmAdvanceMode(fam)
	if !ok {
		return "", fmt.Errorf(
			"%w: raw value %d",
			ErrUnknownFilmAdvanceMode,
			fam,
		)
	}

	return val, nil
}

// AutoFocusMode represents the autofocus mode (e.g., "One-Shot AF", "AI Servo AF", "Manual").
type AutoFocusMode string

// NewAutoFocusMode creates an AutoFocusMode from the raw camera value using the lookup map.
// Returns an error if the value is not found in the map.
func NewAutoFocusMode(afm uint32) (AutoFocusMode, error) {
	val, ok := defaultMaps.GetAutoFocusMode(afm)
	if !ok {
		return "", fmt.Errorf("%w: raw value %d", ErrUnknownAutoFocusMode, afm)
	}

	return val, nil
}

// BulbExposureTime represents the duration of a bulb exposure in "HH:MM:SS" format.
type BulbExposureTime string

// NewBulbExposureTime creates a BulbExposureTime from seconds duration.
// Returns an error if the duration is 0 (minimum is 1 second, maximum is 18 hours per Canon manual).
func NewBulbExposureTime(bd uint32) (BulbExposureTime, error) {
	if bd == 0 { // manual says limits are 1 sec. - 18 hours
		return "", fmt.Errorf(
			"%w: got 0 seconds (minimum: 1)",
			ErrInvalidBulbTime,
		)
	}

	const minutesInHour, secondsInMinute = 60, 60

	d := time.Duration(bd) * time.Second
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % minutesInHour
	seconds := int(d.Seconds()) % secondsInMinute
	s := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)

	return BulbExposureTime(s), nil
}

// MultipleExposure represents the multiple exposure setting (e.g., "Off", "On").
type MultipleExposure string

// NewMultipleExposure creates a MultipleExposure from the raw camera value using the lookup map.
// Returns an error if the value is not found in the map.
func NewMultipleExposure(me uint32) (MultipleExposure, error) {
	val, ok := defaultMaps.GetMultipleExposure(me)
	if !ok {
		return "", fmt.Errorf(
			"%w: raw value %d",
			ErrUnknownMultipleExposure,
			me,
		)
	}

	return val, nil
}

// CustomFunctions represents the 20 custom function settings available on Canon EOS-1V cameras.
// Each element is a string representation of the setting value, or " " if unset.
type CustomFunctions [20]string

// NewCustomFunctions creates a CustomFunctions array from raw byte values.
// In strict mode, validates that each value is within the allowed range for that function.
// Unset values (math.MaxUint8) are represented as a single space " ".
func NewCustomFunctions(cfs [20]byte, strict bool) (CustomFunctions, error) {
	var (
		values   = CustomFunctions{}
		cfLimits = defaultMaps.cfl
	)

	for i := range cfs {
		if cfs[i] == math.MaxUint8 {
			values[i] = " "

			continue
		}

		cfLimit, ok := cfLimits[i]
		if strict && ok && cfs[i] > cfLimit {
			return CustomFunctions{}, fmt.Errorf(
				"%w %d: out of range (0-%d): %d",
				ErrInvalidCustomFunction, i, cfLimit, cfs[i])
		}

		values[i] = strconv.Itoa(int(cfs[i]))
	}

	return values, nil
}
