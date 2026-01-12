//go:generate mockgen -destination=./mocks/builder_mock.go -package=display_test github.com/ma-tf/meta1v/internal/service/display FrameMetadataBuilder,ExposureSettingsBuilder,CameraModesBuilder,CustomFunctionsBuilder,ThumbnailBuilder,DisplayableFrameBuilder
package display

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"

	"github.com/ma-tf/meta1v/pkg/domain"
	"github.com/ma-tf/meta1v/pkg/records"
)

type builder struct {
	log   *slog.Logger
	efrm  records.EFRM
	frame DisplayableFrame
	err   error
}

type FrameMetadataBuilder interface {
	WithFrameMetadata(
		ctx context.Context,
		r records.EFRM,
	) ExposureSettingsBuilder
}

type ExposureSettingsBuilder interface {
	WithExposureSettings(
		ctx context.Context,
		strict bool,
	) CameraModesBuilder
}

type CameraModesBuilder interface {
	WithCameraModesAndFlashInfo(
		ctx context.Context,
		strict bool,
	) CustomFunctionsBuilder
}

type CustomFunctionsBuilder interface {
	WithCustomFunctionsAndFocusPoints(
		ctx context.Context,
		strict bool,
	) ThumbnailBuilder
}

type ThumbnailBuilder interface {
	WithThumbnail(
		ctx context.Context,
		t *DisplayableThumbnail,
	) DisplayableFrameBuilder
}

type DisplayableFrameBuilder interface {
	Build() (DisplayableFrame, error)
}

func NewFrameBuilder(
	log *slog.Logger,
) FrameMetadataBuilder {
	return &builder{
		log:   log,
		efrm:  records.EFRM{},     //nolint:exhaustruct // will be built step by step
		frame: DisplayableFrame{}, //nolint:exhaustruct // will be built step by step
		err:   nil,
	}
}

func (fb *builder) WithFrameMetadata(
	ctx context.Context,
	r records.EFRM,
) ExposureSettingsBuilder {
	fb.efrm = r

	if fb.frame.FilmID, fb.err = domain.NewFilmID(r.CodeA, r.CodeB); fb.err != nil {
		return fb
	}

	fb.frame.FilmLoadedAt, fb.err = domain.NewDateTime(
		r.RollYear, r.RollMonth, r.RollDay,
		r.RollHour, r.RollMinute, r.RollSecond)
	if fb.err != nil {
		return fb
	}

	fb.frame.BatteryLoadedAt, fb.err = domain.NewDateTime(
		r.BatteryYear,
		r.BatteryMonth,
		r.BatteryDay,
		r.BatteryHour,
		r.BatteryMinute,
		r.BatterySecond,
	)
	if fb.err != nil {
		return fb
	}

	fb.frame.TakenAt, fb.err = domain.NewDateTime(
		r.Year, r.Month, r.Day,
		r.Hour, r.Minute, r.Second)
	if fb.err != nil {
		return fb
	}

	fb.frame.FrameNumber = uint(r.FrameNumber)
	fb.frame.Remarks = domain.NewRemarks(r.Remarks)
	fb.frame.UserModifiedRecord = r.IsModifiedRecord != 0

	fb.log.DebugContext(ctx, "parsed frame metadata",
		slog.Any("FilmID", fb.frame.FilmID),
		slog.Any("FilmLoadedAt", fb.frame.FilmLoadedAt),
		slog.Any("BatteryLoadedAt", fb.frame.BatteryLoadedAt),
		slog.Any("TakenAt", fb.frame.TakenAt),
		slog.Any("FrameNumber", fb.frame.FrameNumber),
		slog.Any("Remarks", fb.frame.Remarks),
		slog.Any("UserModifiedRecord", fb.frame.UserModifiedRecord),
	)

	return fb
}

func (fb *builder) WithExposureSettings(
	ctx context.Context,
	strict bool,
) CameraModesBuilder {
	if fb.err != nil {
		return fb
	}

	if fb.frame.MaxAperture, fb.err = domain.NewAv(fb.efrm.MaxAperture, strict); fb.err != nil {
		return fb
	}

	if fb.frame.MaxAperture != "" && fb.frame.MaxAperture != "00" {
		fb.frame.MaxAperture = "f/" + fb.frame.MaxAperture
	}

	if fb.frame.Tv, fb.err = domain.NewTv(fb.efrm.Tv, strict); fb.err != nil {
		return fb
	}

	if fb.frame.Tv == "Bulb" {
		if fb.frame.BulbExposureTime, fb.err = domain.NewBulbExposureTime(fb.efrm.BulbExposureTime); fb.err != nil {
			return fb
		}
	}

	if fb.frame.Av, fb.err = domain.NewAv(fb.efrm.Av, strict); fb.err != nil {
		return fb
	}

	if fb.frame.Av != "" && fb.frame.Av != "00" {
		fb.frame.Av = "f/" + fb.frame.Av
	}

	if fb.frame.FocalLength = domain.NewFocalLength(fb.efrm.FocalLength); fb.frame.FocalLength != "" {
		fb.frame.FocalLength += "mm"
	}

	fb.frame.IsoDX = domain.NewIso(fb.efrm.IsoDX)
	fb.frame.IsoM = domain.NewIso(fb.efrm.IsoM)

	if fb.frame.ExposureCompensation, fb.err = domain.NewExposureCompensation(fb.efrm.ExposureCompenation, strict); fb.err != nil {
		return fb
	}

	if fb.frame.MultipleExposure, fb.err = domain.NewMultipleExposure(fb.efrm.MultipleExposure); fb.err != nil {
		return fb
	}

	fb.log.DebugContext(ctx, "parsed exposure settings",
		slog.Any("MaxAperture", fb.frame.MaxAperture),
		slog.Any("Tv", fb.frame.Tv),
		slog.Any("BulbExposureTime", fb.frame.BulbExposureTime),
		slog.Any("Av", fb.frame.Av),
		slog.Any("FocalLength", fb.frame.FocalLength),
		slog.Any("IsoDX", fb.frame.IsoDX),
		slog.Any("IsoM", fb.frame.IsoM),
		slog.Any("ExposureCompensation", fb.frame.ExposureCompensation),
		slog.Any("MultipleExposure", fb.frame.MultipleExposure))

	return fb
}

func (fb *builder) WithCameraModesAndFlashInfo(
	ctx context.Context,
	strict bool,
) CustomFunctionsBuilder {
	if fb.err != nil {
		return fb
	}

	if fb.frame.FlashExposureComp, fb.err = domain.NewExposureCompensation(fb.efrm.FlashExposureCompensation, strict); fb.err != nil {
		return fb
	}

	if fb.frame.FlashMode, fb.err = domain.NewFlashMode(fb.efrm.FlashMode); fb.err != nil {
		return fb
	}

	if fb.frame.MeteringMode, fb.err = domain.NewMeteringMode(fb.efrm.MeteringMode); fb.err != nil {
		return fb
	}

	if fb.frame.ShootingMode, fb.err = domain.NewShootingMode(fb.efrm.ShootingMode); fb.err != nil {
		return fb
	}

	if fb.frame.FilmAdvanceMode, fb.err = domain.NewFilmAdvanceMode(fb.efrm.FilmAdvanceMode); fb.err != nil {
		return fb
	}

	if fb.frame.AFMode, fb.err = domain.NewAutoFocusMode(fb.efrm.AFMode); fb.err != nil {
		return fb
	}

	fb.log.DebugContext(ctx, "parsed camera modes and flash info",
		slog.Any("FlashExposureComp", fb.frame.FlashExposureComp),
		slog.Any("FlashMode", fb.frame.FlashMode),
		slog.Any("MeteringMode", fb.frame.MeteringMode),
		slog.Any("ShootingMode", fb.frame.ShootingMode),
		slog.Any("FilmAdvanceMode", fb.frame.FilmAdvanceMode),
		slog.Any("AFMode", fb.frame.AFMode),
	)

	return fb
}

func (fb *builder) WithCustomFunctionsAndFocusPoints(
	ctx context.Context,
	strict bool,
) ThumbnailBuilder {
	if fb.err != nil {
		return fb
	}

	if fb.frame.CustomFunctions, fb.err = NewCustomFunctions(fb.efrm, strict); fb.err != nil {
		return fb
	}

	fb.frame.FocusingPoints = DisplayableFocusPoints{
		Selection: uint(fb.efrm.FocusingPoint),
		Points: [8]byte{
			fb.efrm.FocusPoints1,
			fb.efrm.FocusPoints2,
			fb.efrm.FocusPoints3,
			fb.efrm.FocusPoints4,
			fb.efrm.FocusPoints5,
			fb.efrm.FocusPoints6,
			fb.efrm.FocusPoints7,
			fb.efrm.FocusPoints8,
		},
	}

	fb.log.DebugContext(ctx, "parsed custom functions and focus points",
		slog.Any("CustomFunctions", fb.frame.CustomFunctions),
		slog.Any("FocusingPoints", fb.frame.FocusingPoints),
	)

	return fb
}

func (fb *builder) WithThumbnail(
	ctx context.Context,
	t *DisplayableThumbnail,
) DisplayableFrameBuilder {
	if fb.err != nil {
		return fb
	}

	fb.frame.Thumbnail = t

	fb.log.DebugContext(ctx, "parsed thumbnail") // no printing because it's big

	return fb
}

func (fb *builder) Build() (DisplayableFrame, error) {
	return fb.frame, fb.err
}

// --------------------------

var ErrInvalidCustomFunction = errors.New("invalid custom function")

//nolint:gochecknoglobals,mnd // not exported anyway, magic numbers are defined by Canon manual
var cfMaxRanges = map[int]byte{
	0:  1,
	1:  3,
	2:  1,
	3:  1,
	4:  3,
	5:  3,
	6:  2,
	7:  2,
	8:  2,
	9:  3,
	10: 3,
	11: 3,
	12: 1,
	13: 3,
	14: 1,
	15: 1,
	16: 1,
	17: 2,
	18: 2,
	19: 5,
}

type DisplayableCustomFunctions [20]string

func NewCustomFunctions(
	r records.EFRM,
	strict bool,
) (DisplayableCustomFunctions, error) {
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

	values := [20]string{}

	for i, cf := range cfs {
		if cf != math.MaxUint8 && (cf < cfMin || cf > cfMaxRanges[i]) {
			if strict {
				return DisplayableCustomFunctions{}, fmt.Errorf(
					"%w %d: out of range (%d-%d): %d",
					ErrInvalidCustomFunction,
					i,
					cfMin,
					cfMaxRanges[i],
					cf,
				)
			}
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
