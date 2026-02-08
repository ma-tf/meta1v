//go:generate mockgen -destination=./mocks/builder_mock.go -package=display_test github.com/ma-tf/meta1v/internal/service/display Builder
package display

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"

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
	ErrInvalidCustomFunctions = errors.New("failed to parse custom functions")
)

const (
	ansiRed   = "\033[31m"
	ansiGray  = "\033[30m"
	ansiReset = "\033[0m"

	emptyBox  = "\u25AF"
	filledBox = "\u25AE"

	redFilledBox = ansiRed + filledBox + ansiReset + " "
	redEmptyBox  = ansiRed + emptyBox + ansiReset + " "
	greyBox      = emptyBox + " "

	// - Manual focus is a single redEmptyBox among greyBoxes.
	// - AI Servo AF is redEmptyBoxes on the edge. with greyBoxes in the interior.
	// - One-Shot AF has redFilledBoxes for active focus points, redEmptyBoxes for inactive edge points,
	//   and greyBoxes for inactive interior points.

	emptyFocusPointsGrid = "    " + ansiGray + "\u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF\n" +
		" \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF\n" +
		"\u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF\n" +
		" \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF\n" +
		"    \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF \u25AF" + ansiReset + "\n"

	// focusPointBitCounts defines the number of bits used in each byte of focus point data.
	// Canon EOS cameras use a 45-point AF grid arranged in 5 rows. The 8 bytes map to:
	//
	//		[0]: Row 0 top - 7 bits
	//		[1]: Row 1 right - 2 bits,  [2]: Row 1 left - 8 bits
	//		[3]: Row 2 right - 3 bits,  [4]: Row 2 left - 8 bits
	//		[5]: Row 3 right - 2 bits,  [6]: Row 3 left - 8 bits
	//		[7]: Row 4 bottom - 7 bits

	topOrBottomRowBits = 7
	leftSegmentBits    = 8

	topBits                         = topOrBottomRowBits
	topRightBits, topLeftBits       = 2, leftSegmentBits
	middleRightBits, middleLeftBits = 3, leftSegmentBits
	bottomRightBits, bottomLeftBits = 2, leftSegmentBits
	bottomBits                      = topOrBottomRowBits

	leftmostBitMask = byte(0b10000000)
)

//nolint:gochecknoglobals // package-level constant for focus point rendering
var focusPointBitCounts = [8]int{
	topBits,
	topRightBits, topLeftBits,
	middleRightBits, middleLeftBits,
	bottomRightBits, bottomLeftBits,
	bottomBits,
}

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
		slog.String("focusingPoints", string(frame.FocusingPoints)),
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
		return wrapFrameError(ErrInvalidFilmID, err, efrm.FrameNumber)
	}

	filmLoadedAt, err := domain.NewDateTime(
		efrm.RollYear, efrm.RollMonth, efrm.RollDay,
		efrm.RollHour, efrm.RollMinute, efrm.RollSecond)
	if err != nil {
		return wrapFrameError(ErrInvalidFilmLoadedDate, err, efrm.FrameNumber)
	}

	batteryLoadedAt, err := domain.NewDateTime(
		efrm.BatteryYear, efrm.BatteryMonth, efrm.BatteryDay,
		efrm.BatteryHour, efrm.BatteryMinute, efrm.BatterySecond,
	)
	if err != nil {
		return wrapFrameError(
			ErrInvalidBatteryLoadedDate, err, efrm.FrameNumber)
	}

	takenAt, err := domain.NewDateTime(
		efrm.Year, efrm.Month, efrm.Day,
		efrm.Hour, efrm.Minute, efrm.Second)
	if err != nil {
		return wrapFrameError(ErrInvalidCaptureDate, err, efrm.FrameNumber)
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
		return wrapFrameError(ErrInvalidMaxAperture, err, efrm.FrameNumber)
	}

	tv, err := domain.NewTv(efrm.Tv, strict)
	if err != nil {
		return wrapFrameError(ErrInvalidShutterSpeed, err, efrm.FrameNumber)
	}

	var bulbExposureTime domain.BulbExposureTime
	if tv == "Bulb" {
		if bulbExposureTime, err = domain.NewBulbExposureTime(efrm.BulbExposureTime); err != nil {
			return wrapFrameError(
				ErrInvalidBulbExposureTime, err, efrm.FrameNumber)
		}
	}

	av, err := formatAperture(efrm.Av, strict)
	if err != nil {
		return wrapFrameError(ErrInvalidAperture, err, efrm.FrameNumber)
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
		return wrapFrameError(
			ErrInvalidExposureCompensation, err, efrm.FrameNumber)
	}

	multipleExposure, err := domain.NewMultipleExposure(efrm.MultipleExposure)
	if err != nil {
		return wrapFrameError(ErrInvalidMultipleExposure, err, efrm.FrameNumber)
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
		return wrapFrameError(
			ErrInvalidFlashExposureComp, err, efrm.FrameNumber)
	}

	flashMode, err := domain.NewFlashMode(efrm.FlashMode)
	if err != nil {
		return wrapFrameError(ErrInvalidFlashMode, err, efrm.FrameNumber)
	}

	meteringMode, err := domain.NewMeteringMode(efrm.MeteringMode)
	if err != nil {
		return wrapFrameError(ErrInvalidMeteringMode, err, efrm.FrameNumber)
	}

	shootingMode, err := domain.NewShootingMode(efrm.ShootingMode)
	if err != nil {
		return wrapFrameError(ErrInvalidShootingMode, err, efrm.FrameNumber)
	}

	filmAdvanceMode, err := domain.NewFilmAdvanceMode(efrm.FilmAdvanceMode)
	if err != nil {
		return wrapFrameError(ErrInvalidFilmAdvanceMode, err, efrm.FrameNumber)
	}

	afMode, err := domain.NewAutoFocusMode(efrm.AFMode)
	if err != nil {
		return wrapFrameError(ErrInvalidAutoFocusMode, err, efrm.FrameNumber)
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
		return wrapFrameError(
			fmt.Errorf("%w %q", ErrInvalidCustomFunctions, cfs),
			err,
			efrm.FrameNumber,
		)
	}

	rawFocusPointsBytes := [8]byte{
		efrm.FocusPoints1,
		efrm.FocusPoints2,
		efrm.FocusPoints3,
		efrm.FocusPoints4,
		efrm.FocusPoints5,
		efrm.FocusPoints6,
		efrm.FocusPoints7,
		efrm.FocusPoints8,
	}

	focusingPoints := b.formatFocusPoints(
		efrm.FocusingPoint,
		rawFocusPointsBytes,
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

func renderFocusPointByte(pointsByte byte, bitCount int) string {
	var grid strings.Builder

	isTopOrBottomRow := bitCount == topOrBottomRowBits
	isLeftSegment := bitCount == leftSegmentBits

	for bitIndex, mask := 0, leftmostBitMask; bitIndex < bitCount; bitIndex, mask = bitIndex+1, mask>>1 {
		isFirstOrLastBit := (isLeftSegment && bitIndex == 0) ||
			(!isLeftSegment && bitIndex == bitCount-1)
		isEdge := isTopOrBottomRow || isFirstOrLastBit

		isFocusPointActive := pointsByte&mask != 0

		switch {
		case isFocusPointActive:
			grid.WriteString(redFilledBox) // Active focus point
		case isEdge:
			grid.WriteString(redEmptyBox) // Edge position, not active
		default:
			grid.WriteString(greyBox) // Interior position, not active
		}
	}

	return grid.String()
}

func (b *builder) formatFocusPoints(
	selection uint32,
	points [8]byte,
) DisplayableFocusPoints {
	if selection == math.MaxUint32 {
		return DisplayableFocusPoints(emptyFocusPointsGrid)
	}

	renderedSegments := make([]string, len(focusPointBitCounts))
	for i, bitCount := range focusPointBitCounts {
		renderedSegments[i] = renderFocusPointByte(points[i], bitCount)
	}

	// Assemble the 5-row grid layout:
	// Row 0:     segment[0]              (7 points, indented)
	// Row 1:  segment[2] + segment[1]    (10 points: 8 left + 2 right)
	// Row 2:  segment[4] + segment[3]    (11 points: 8 left + 3 right)
	// Row 3:  segment[6] + segment[5]    (10 points: 8 left + 2 right)
	// Row 4:     segment[7]              (7 points, indented)
	focusPointGrid := "    " + renderedSegments[0] + "\n" +
		" " + renderedSegments[2] + renderedSegments[1] + "\n" +
		renderedSegments[4] + renderedSegments[3] + "\n" +
		" " + renderedSegments[6] + renderedSegments[5] + "\n" +
		"    " + renderedSegments[7] + "\n"

	return DisplayableFocusPoints(focusPointGrid)
}

func wrapFrameError(baseErr, frameErr error, frameNumber uint32) error {
	return fmt.Errorf("%w in frame %d: %w", baseErr, frameNumber, frameErr)
}
