package display

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/ma-tf/meta1v/pkg/records"
)

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

type DisplayableDatetime string

func NewDateTime(
	year uint16,
	month,
	day,
	hour,
	minute,
	second uint8,
) (DisplayableDatetime, error) {
	if year == math.MaxUint16 && month&day&hour&minute&second == math.MaxUint8 {
		return "", nil
	}

	// manual says year limit is 2000 - 2099, I won't validate this for now
	rawDate := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
		year, month, day, hour, minute, second)

	t, err := time.Parse(time.DateTime, rawDate) // performs validation
	if err != nil || t.IsZero() {
		return "", errors.Join(ErrInvalidFilmLoadDate, err)
	}

	return DisplayableDatetime(t.Format(time.DateTime)), nil
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

func NewTv(tv int32) (Tv, error) {
	if tv == -1 {
		return "", nil
	}

	val, ok := tvs[tv]
	if !ok {
		return "", fmt.Errorf(
			"%w: Tv %d is invalid. Please check valid Tv values",
			ErrInvalidTv,
			tv,
		)
	}

	return val, nil
}

type Av string

func NewAv(av uint32) (Av, error) {
	if av == math.MaxUint32 {
		return "", nil
	}

	val, ok := avs[av]
	if !ok {
		return "", fmt.Errorf(
			"%w: Av %d is invalid. Please check valid Av values",
			ErrInvalidAv,
			av,
		)
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

func NewExposureCompensation(ec int32) (ExposureCompenation, error) {
	if ec == -1 {
		return "", nil
	}

	val, ok := exposureComps[ec]
	if !ok {
		return "", fmt.Errorf(
			"%w: exposure compensation %d is invalid. "+
				"Please check valid exposure compensation values",
			ErrUnknownExposureComp, ec,
		)
	}

	return val, nil
}

type FlashMode string

func NewFlashMode(fm uint32) (FlashMode, error) {
	val, ok := flashModes[fm]
	if !ok {
		return "", ErrUnknownFlashMode
	}

	return val, nil
}

type MeteringMode string

func NewMeteringMode(mm uint32) (MeteringMode, error) {
	val, ok := meteringModes[mm]
	if !ok {
		return "", ErrUnknownMeteringMode
	}

	return val, nil
}

type ShootingMode string

func NewShootingMode(sm uint32) (ShootingMode, error) {
	val, ok := shootingModes[sm]
	if !ok {
		return "", ErrUnknownShootingMode
	}

	return val, nil
}

type FilmAdvanceMode string

func NewFilmAdvanceMode(fam uint32) (FilmAdvanceMode, error) {
	val, ok := filmAdvanceModes[fam]
	if !ok {
		return "", ErrUnknownFilmAdvanceMode
	}

	return val, nil
}

type AutoFocusMode string

func NewAutoFocusMode(afm uint32) (AutoFocusMode, error) {
	val, ok := afModes[afm]
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
	val, ok := multipleExposures[me]
	if !ok {
		return "", ErrUnknownMultipleExposure
	}

	return val, nil
}

type DisplayableCustomFunctions []string

func NewCustomFunctions(r records.EFRM) (DisplayableCustomFunctions, error) {
	const cfMin = 0

	cfs := [20]byte{
		r.CustomFunction0,
		r.CustomFunction1,
		r.CustomFunction2,
		r.CustomFunction3,
		r.CustomFunction4,
		r.CustomFunction5,
		r.CustomFunction6,
		r.CustomFunction7,
		r.CustomFunction8,
		r.CustomFunction9,
		r.CustomFunction10,
		r.CustomFunction11,
		r.CustomFunction12,
		r.CustomFunction13,
		r.CustomFunction14,
		r.CustomFunction15,
		r.CustomFunction16,
		r.CustomFunction17,
		r.CustomFunction18,
		r.CustomFunction19,
	}

	values := make([]string, len(cfs))
	for i, cf := range cfs {
		if cf != math.MaxUint8 && (cf < cfMin || cf > cfMaxRanges[i]) {
			return DisplayableCustomFunctions{}, fmt.Errorf(
				"%w %d: out of range (%d-%d): %d",
				ErrInvalidCustomFunction,
				i,
				cfMin,
				cfMaxRanges[i],
				cf,
			)
		}

		if cf == math.MaxUint8 {
			values[i] = " "
		} else {
			values[i] = strconv.Itoa(int(cf))
		}
	}

	return values, nil
}

type DisplayableFocusPoints struct {
	Selection uint
	Points    [8]byte
}
