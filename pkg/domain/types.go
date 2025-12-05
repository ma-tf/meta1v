package domain

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

//nolint:gochecknoglobals // global defaultMaps is acceptable here
var defaultMaps = NewMapProvider()

type FilmID string

func NewFilmID(prefix, suffix uint32) (FilmID, error) {
	if prefix == math.MaxUint32 && suffix == math.MaxUint32 {
		return "", nil
	}

	const maxA, maxB = 99, 999
	if prefix > maxA {
		return "", fmt.Errorf("%w: %d", ErrPrefixOutOfRange, prefix)
	}

	if suffix > maxB {
		return "", fmt.Errorf("%w: %d", ErrSuffixOutOfRange, suffix)
	}

	s := fmt.Sprintf("%02d-%03d", prefix, suffix)

	return FilmID(s), nil
}

type ValidatedDatetime string

func NewDateTime(
	year uint16,
	month,
	day,
	hour,
	minute,
	second uint8,
) (ValidatedDatetime, error) {
	if year == math.MaxUint16 && month&day&hour&minute&second == math.MaxUint8 {
		return "", nil
	}

	// manual says year limit is 2000 - 2099, I won't validate this for now
	rawDate := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
		year, month, day, hour, minute, second)

	t, err := time.Parse(time.DateTime, rawDate) // performs format validation
	if err != nil || t.IsZero() {
		return "", errors.Join(ErrInvalidFilmLoadDate, err)
	}

	return ValidatedDatetime(t.Format(time.DateTime)), nil
}

type Title string

func NewTitle(t [64]byte) Title {
	return Title(t[:bytes.IndexByte(t[:], 0)])
}

type Remarks string

func NewRemarks(r [256]byte) Remarks {
	return Remarks(r[:bytes.IndexByte(r[:], 0)])
}

type FocalLength string

func NewFocalLength(fl uint32) FocalLength {
	if fl == math.MaxUint32 {
		return ""
	}

	return FocalLength(fmt.Sprintf("%dmm", fl))
}

type Tv string

func NewTv(tv int32, strict bool) (Tv, error) {
	if tv == -1 {
		return "", nil
	}

	val, ok := defaultMaps.GetTv(tv)
	if !ok {
		if strict {
			return "", fmt.Errorf(
				"%w: Tv %d is invalid. Please check valid Tv values",
				ErrInvalidTv,
				tv,
			)
		}

		val = Tv(strconv.Itoa(int(tv)))
	}

	return val, nil
}

type Av string

func NewAv(av uint32, strict bool) (Av, error) {
	if av == math.MaxUint32 {
		return "", nil
	}

	val, ok := defaultMaps.GetAv(av)
	if !ok {
		if strict {
			return "", fmt.Errorf(
				"%w: Av %d is invalid. Please check valid Av values",
				ErrInvalidAv,
				av,
			)
		}

		val = Av(strconv.Itoa(int(av)))
	}

	return val, nil
}

type Iso string

func NewIso(iso uint32) Iso {
	if iso == math.MaxUint32 {
		return ""
	}

	return Iso(strconv.FormatUint(uint64(iso), 10))
}

type ExposureCompenation string

func NewExposureCompensation(
	ec int32,
	strict bool,
) (ExposureCompenation, error) {
	if ec == -1 {
		return "", nil
	}

	val, ok := defaultMaps.GetExposureCompenation(ec)
	if !ok {
		if strict {
			return "", fmt.Errorf(
				"%w: exposure compensation %d is invalid. "+
					"Please check valid exposure compensation values",
				ErrUnknownExposureComp, ec,
			)
		}

		prefix := "+"
		if ec&1 == 1 {
			prefix = "-"
		}

		const divisor = 10

		val = ExposureCompenation(
			prefix + fmt.Sprintf("%.1f", float64(ec)/divisor),
		)
	}

	return val, nil
}

type FlashMode string

func NewFlashMode(fm uint32) (FlashMode, error) {
	val, ok := defaultMaps.GetFlashMode(fm)
	if !ok {
		return "", ErrUnknownFlashMode
	}

	return val, nil
}

type MeteringMode string

func NewMeteringMode(mm uint32) (MeteringMode, error) {
	val, ok := defaultMaps.GetMeteringMode(mm)
	if !ok {
		return "", ErrUnknownMeteringMode
	}

	return val, nil
}

type ShootingMode string

func NewShootingMode(sm uint32) (ShootingMode, error) {
	val, ok := defaultMaps.GetShootingMode(sm)
	if !ok {
		return "", ErrUnknownShootingMode
	}

	return val, nil
}

type FilmAdvanceMode string

func NewFilmAdvanceMode(fam uint32) (FilmAdvanceMode, error) {
	val, ok := defaultMaps.GetFilmAdvanceMode(fam)
	if !ok {
		return "", ErrUnknownFilmAdvanceMode
	}

	return val, nil
}

type AutoFocusMode string

func NewAutoFocusMode(afm uint32) (AutoFocusMode, error) {
	val, ok := defaultMaps.GetAutoFocusMode(afm)
	if !ok {
		return "", ErrUnknownAutoFocusMode
	}

	return val, nil
}

type BulbExposureTime string

func NewBulbExposureTime(bd uint32) (BulbExposureTime, error) {
	if bd == 0 { // manual says limits are 1 sec. - 18 hours
		return "", ErrInvalidBulbTime
	}

	const minutesInHour, secondsInMinute = 60, 60

	d := time.Duration(bd) * time.Second
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % minutesInHour
	seconds := int(d.Seconds()) % secondsInMinute
	s := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)

	return BulbExposureTime(s), nil
}

type MultipleExposure string

func NewMultipleExposure(me uint32) (MultipleExposure, error) {
	val, ok := defaultMaps.GetMultipleExposure(me)
	if !ok {
		return "", ErrUnknownMultipleExposure
	}

	return val, nil
}
