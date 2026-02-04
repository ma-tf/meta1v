//go:generate mockgen -destination=./mocks/builder_mock.go -package=display_test github.com/ma-tf/meta1v/internal/service/display Builder
package display

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ma-tf/meta1v/pkg/domain"
	"github.com/ma-tf/meta1v/pkg/records"
)

var (
	ErrInvalidFilmID               = errors.New("invalid film ID")
	ErrInvalidFilmLoadedDate       = errors.New("invalid film loaded date")
	ErrInvalidBatteryLoadedDate    = errors.New("invalid battery loaded date")
	ErrInvalidCaptureDate          = errors.New("invalid capture date")
	ErrInvalidMaxAperture          = errors.New("invalid max aperture")
	ErrInvalidShutterSpeed         = errors.New("invalid shutter speed")
	ErrInvalidBulbExposureTime     = errors.New("invalid bulb exposure time")
	ErrInvalidAperture             = errors.New("invalid aperture")
	ErrInvalidExposureCompensation = errors.New("invalid exposure compensation")
	ErrInvalidMultipleExposure     = errors.New("invalid multiple exposure")
	ErrInvalidFlashExposureComp    = errors.New(
		"invalid flash exposure compensation",
	)
	ErrInvalidFlashMode       = errors.New("invalid flash mode")
	ErrInvalidMeteringMode    = errors.New("invalid metering mode")
	ErrInvalidShootingMode    = errors.New("invalid shooting mode")
	ErrInvalidFilmAdvanceMode = errors.New("invalid film advance mode")
	ErrInvalidAutoFocusMode   = errors.New("invalid auto focus mode")
)

type builder struct {
	log *slog.Logger
}

type Builder interface {
	Build(
		ctx context.Context,
		efrm records.EFRM,
		thumbnail *DisplayableThumbnail,
		strict bool,
	) (DisplayableFrame, error)
}

func NewFrameBuilder(log *slog.Logger) Builder {
	return &builder{log: log}
}

func (b *builder) Build(
	ctx context.Context,
	efrm records.EFRM,
	thumbnail *DisplayableThumbnail,
	strict bool,
) (DisplayableFrame, error) {
	var frame DisplayableFrame

	if err := b.withFrameMetadata(&frame, efrm); err != nil {
		return DisplayableFrame{}, err
	}

	b.log.DebugContext(ctx, "parsed frame metadata",
		slog.String("filmID", string(frame.FilmID)),
		slog.String("filmLoadedAt", string(frame.FilmLoadedAt)),
		slog.String("batteryLoadedAt", string(frame.BatteryLoadedAt)),
		slog.String("takenAt", string(frame.TakenAt)),
		slog.Uint64("frameNumber", uint64(frame.FrameNumber)),
		slog.Bool("userModified", frame.UserModifiedRecord),
	)

	if err := b.withExposureSettings(&frame, efrm, strict); err != nil {
		return DisplayableFrame{}, err
	}

	b.log.DebugContext(ctx, "parsed exposure settings",
		slog.String("maxAperture", string(frame.MaxAperture)),
		slog.String("tv", string(frame.Tv)),
		slog.String("av", string(frame.Av)),
		slog.String("focalLength", string(frame.FocalLength)),
	)

	if err := b.withCameraModesAndFlashInfo(&frame, efrm, strict); err != nil {
		return DisplayableFrame{}, err
	}

	b.log.DebugContext(ctx, "parsed camera modes and flash info",
		slog.String("flashMode", string(frame.FlashMode)),
		slog.String("meteringMode", string(frame.MeteringMode)),
		slog.String("shootingMode", string(frame.ShootingMode)),
		slog.String("afMode", string(frame.AFMode)),
	)

	if err := b.withCustomFunctionsAndFocusPoints(&frame, efrm, strict); err != nil {
		return DisplayableFrame{}, err
	}

	b.log.DebugContext(ctx, "parsed custom functions and focus points",
		slog.Any("customFunctions", frame.CustomFunctions),
		slog.Group("focusingPoints",
			slog.Uint64("selection", uint64(frame.FocusingPoints.Selection)),
			slog.Any("points", frame.FocusingPoints.Points),
		),
	)

	frame.Thumbnail = thumbnail

	return frame, nil
}

func (b *builder) withFrameMetadata(
	frame *DisplayableFrame,
	efrm records.EFRM,
) error {
	filmID, err := domain.NewFilmID(efrm.CodeA, efrm.CodeB)
	if err != nil {
		return errors.Join(ErrInvalidFilmID, err)
	}

	filmLoadedAt, err := domain.NewDateTime(
		efrm.RollYear, efrm.RollMonth, efrm.RollDay,
		efrm.RollHour, efrm.RollMinute, efrm.RollSecond)
	if err != nil {
		return errors.Join(ErrInvalidFilmLoadedDate, err)
	}

	batteryLoadedAt, err := domain.NewDateTime(
		efrm.BatteryYear, efrm.BatteryMonth, efrm.BatteryDay,
		efrm.BatteryHour, efrm.BatteryMinute, efrm.BatterySecond,
	)
	if err != nil {
		return errors.Join(ErrInvalidBatteryLoadedDate, err)
	}

	takenAt, err := domain.NewDateTime(
		efrm.Year, efrm.Month, efrm.Day,
		efrm.Hour, efrm.Minute, efrm.Second)
	if err != nil {
		return errors.Join(ErrInvalidCaptureDate, err)
	}

	frame.FilmID = filmID
	frame.FilmLoadedAt = filmLoadedAt
	frame.BatteryLoadedAt = batteryLoadedAt
	frame.TakenAt = takenAt
	frame.FrameNumber = uint(efrm.FrameNumber)
	frame.Remarks = domain.NewRemarks(efrm.Remarks)
	frame.UserModifiedRecord = efrm.IsModifiedRecord != 0

	return nil
}

func (b *builder) withExposureSettings(
	frame *DisplayableFrame,
	efrm records.EFRM,
	strict bool,
) error {
	maxAperture, err := formatAperture(efrm.MaxAperture, strict)
	if err != nil {
		return errors.Join(ErrInvalidMaxAperture, err)
	}

	tv, err := domain.NewTv(efrm.Tv, strict)
	if err != nil {
		return errors.Join(ErrInvalidShutterSpeed, err)
	}

	var bulbExposureTime domain.BulbExposureTime
	if tv == "Bulb" {
		if bulbExposureTime, err = domain.NewBulbExposureTime(efrm.BulbExposureTime); err != nil {
			return errors.Join(ErrInvalidBulbExposureTime, err)
		}
	}

	av, err := formatAperture(efrm.Av, strict)
	if err != nil {
		return errors.Join(ErrInvalidAperture, err)
	}

	focalLength := domain.NewFocalLength(efrm.FocalLength)
	if focalLength != "" {
		focalLength += "mm"
	}

	exposureCompensation, err := domain.NewExposureCompensation(
		efrm.ExposureCompensation,
		strict,
	)
	if err != nil {
		return errors.Join(ErrInvalidExposureCompensation, err)
	}

	multipleExposure, err := domain.NewMultipleExposure(efrm.MultipleExposure)
	if err != nil {
		return errors.Join(ErrInvalidMultipleExposure, err)
	}

	frame.MaxAperture = maxAperture
	frame.Tv = tv
	frame.BulbExposureTime = bulbExposureTime
	frame.Av = av
	frame.FocalLength = focalLength
	frame.IsoDX = domain.NewIso(efrm.IsoDX)
	frame.IsoM = domain.NewIso(efrm.IsoM)
	frame.ExposureCompensation = exposureCompensation
	frame.MultipleExposure = multipleExposure

	return nil
}

func (b *builder) withCameraModesAndFlashInfo(
	frame *DisplayableFrame,
	efrm records.EFRM,
	strict bool,
) error {
	flashExposureComp, err := domain.NewExposureCompensation(
		efrm.FlashExposureCompensation,
		strict,
	)
	if err != nil {
		return errors.Join(ErrInvalidFlashExposureComp, err)
	}

	flashMode, err := domain.NewFlashMode(efrm.FlashMode)
	if err != nil {
		return errors.Join(ErrInvalidFlashMode, err)
	}

	meteringMode, err := domain.NewMeteringMode(efrm.MeteringMode)
	if err != nil {
		return errors.Join(ErrInvalidMeteringMode, err)
	}

	shootingMode, err := domain.NewShootingMode(efrm.ShootingMode)
	if err != nil {
		return errors.Join(ErrInvalidShootingMode, err)
	}

	filmAdvanceMode, err := domain.NewFilmAdvanceMode(efrm.FilmAdvanceMode)
	if err != nil {
		return errors.Join(ErrInvalidFilmAdvanceMode, err)
	}

	afMode, err := domain.NewAutoFocusMode(efrm.AFMode)
	if err != nil {
		return errors.Join(ErrInvalidAutoFocusMode, err)
	}

	frame.FlashExposureComp = flashExposureComp
	frame.FlashMode = flashMode
	frame.MeteringMode = meteringMode
	frame.ShootingMode = shootingMode
	frame.FilmAdvanceMode = filmAdvanceMode
	frame.AFMode = afMode

	return nil
}

func (b *builder) withCustomFunctionsAndFocusPoints(
	frame *DisplayableFrame,
	efrm records.EFRM,
	strict bool,
) error {
	cfs := [20]byte{
		efrm.CustomFunction0, efrm.CustomFunction1, efrm.CustomFunction2, efrm.CustomFunction3,
		efrm.CustomFunction4, efrm.CustomFunction5, efrm.CustomFunction6, efrm.CustomFunction7,
		efrm.CustomFunction8, efrm.CustomFunction9, efrm.CustomFunction10, efrm.CustomFunction11,
		efrm.CustomFunction12, efrm.CustomFunction13, efrm.CustomFunction14, efrm.CustomFunction15,
		efrm.CustomFunction16, efrm.CustomFunction17, efrm.CustomFunction18, efrm.CustomFunction19,
	}

	customFunctions, err := domain.NewCustomFunctions(cfs, strict)
	if err != nil {
		return fmt.Errorf(
			"failed to parse custom functions %v: %w",
			cfs,
			err,
		)
	}

	focusingPoints := domain.NewFocusPoints(
		efrm.FocusingPoint,
		[8]byte{
			efrm.FocusPoints1,
			efrm.FocusPoints2,
			efrm.FocusPoints3,
			efrm.FocusPoints4,
			efrm.FocusPoints5,
			efrm.FocusPoints6,
			efrm.FocusPoints7,
			efrm.FocusPoints8,
		},
	)

	frame.CustomFunctions = customFunctions
	frame.FocusingPoints = focusingPoints

	return nil
}

func formatAperture(av uint32, strict bool) (domain.Av, error) {
	result, err := domain.NewAv(av, strict)
	if err != nil {
		return "", err //nolint:wrapcheck // wrapped at call sites with context-specific errors
	}

	if result != "" && result != "00" {
		result = "f/" + result
	}

	return result, nil
}
