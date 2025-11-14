package display

import "github.com/ma-tf/meta1v/pkg/records"

type frameBuilder struct {
	efrm  records.EFRM
	frame displayableFrame
	err   error
}

type step1FrameBuilder interface {
	WithBasicInfoAndModes() step2FrameBuilder
}

type step2FrameBuilder interface {
	WithCameraModesAndFlashInfo() step3FrameBuilder
}

type step3FrameBuilder interface {
	WithCustomFunctionsAndFocusPoints() *frameBuilder
}

func newFrameBuilder(
	r records.EFRM,
	t *DisplayableThumbnail,
) step1FrameBuilder {
	fb := &frameBuilder{
		efrm:  r,
		frame: displayableFrame{}, //nolint:exhaustruct // will be built step by step
		err:   nil,
	}

	fb.frame.FilmID, fb.err = NewFilmID(r.CodeA, r.CodeB)
	if fb.err != nil {
		return fb
	}

	fb.frame.FilmLoadedAt, fb.err = NewDateTime(r.Year, r.Month, r.Day,
		r.Hour, r.Minute, r.Second)
	if fb.err != nil {
		return fb
	}

	fb.frame.BatteryLoadedAt, fb.err = NewDateTime(
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

	fb.frame.IsoDX = NewIso(r.IsoDX)
	fb.frame.FrameNumber = uint(r.FrameNumber)
	fb.frame.Remarks = NewRemarks(r.Remarks)
	fb.frame.Thumbnail = t
	fb.frame.UserModifiedRecord = r.IsModifiedRecord != 0

	return fb
}

func (fb *frameBuilder) WithBasicInfoAndModes() step2FrameBuilder {
	if fb.err != nil {
		return fb
	}

	fb.frame.FocalLength = NewFocalLength(fb.efrm.FocalLength)

	fb.frame.MaxAperture, fb.err = NewAv(fb.efrm.MaxAperture)
	if fb.err != nil {
		return fb
	}

	fb.frame.Tv, fb.err = NewTv(fb.efrm.Tv)
	if fb.err != nil {
		return fb
	}

	if fb.frame.Tv == "Bulb" {
		fb.frame.BulbExposureTime, fb.err = NewBulbExposureTime(
			fb.efrm.BulbExposureTime,
		)
		if fb.err != nil {
			return fb
		}
	}

	fb.frame.Av, fb.err = NewAv(fb.efrm.Av)
	if fb.err != nil {
		return fb
	}

	fb.frame.IsoM = NewIso(fb.efrm.IsoM)

	fb.frame.ExposureCompensation, fb.err = NewExposureCompensation(
		fb.efrm.ExposureCompenation,
	)
	if fb.err != nil {
		return fb
	}

	fb.frame.TakenAt, fb.err = NewDateTime(
		fb.efrm.Year,
		fb.efrm.Month,
		fb.efrm.BatteryDay,
		fb.efrm.Hour,
		fb.efrm.Minute,
		fb.efrm.Second,
	)
	if fb.err != nil {
		return fb
	}

	fb.frame.MultipleExposure, fb.err = NewMultipleExposure(
		fb.efrm.MultipleExposure,
	)
	if fb.err != nil {
		return fb
	}

	return fb
}

func (fb *frameBuilder) WithCameraModesAndFlashInfo() step3FrameBuilder {
	if fb.err != nil {
		return fb
	}

	fb.frame.FlashExposureComp, fb.err = NewExposureCompensation(
		fb.efrm.FlashExposureCompensation,
	)
	if fb.err != nil {
		return fb
	}

	fb.frame.FlashMode, fb.err = NewFlashMode(fb.efrm.FlashMode)
	if fb.err != nil {
		return fb
	}

	fb.frame.MeteringMode, fb.err = NewMeteringMode(fb.efrm.MeteringMode)
	if fb.err != nil {
		return fb
	}

	fb.frame.ShootingMode, fb.err = NewShootingMode(fb.efrm.ShootingMode)
	if fb.err != nil {
		return fb
	}

	fb.frame.FilmAdvanceMode, fb.err = NewFilmAdvanceMode(
		fb.efrm.FilmAdvanceMode,
	)
	if fb.err != nil {
		return fb
	}

	fb.frame.AFMode, fb.err = NewAutoFocusMode(fb.efrm.AFMode)
	if fb.err != nil {
		return fb
	}

	return fb
}

func (fb *frameBuilder) WithCustomFunctionsAndFocusPoints() *frameBuilder {
	if fb.err != nil {
		return fb
	}

	fb.frame.CustomFunctions, fb.err = NewCustomFunctions(fb.efrm)
	if fb.err != nil {
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

	return fb
}

func (fb *frameBuilder) Build() (displayableFrame, error) {
	return fb.frame, fb.err
}
