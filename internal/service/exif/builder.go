//go:generate mockgen -destination=./mocks/builder_mock.go -package=exif_test github.com/ma-tf/meta1v/internal/service/exif Builder
package exif

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"

	"github.com/ma-tf/meta1v/pkg/domain"
	"github.com/ma-tf/meta1v/pkg/records"
)

var (
	ErrInvalidCaptureDate          = errors.New("invalid capture date")
	ErrInvalidBatteryLoadedDate    = errors.New("invalid battery loaded date")
	ErrInvalidFilmLoadedDate       = errors.New("invalid film loaded date")
	ErrInvalidExposureCompensation = errors.New("invalid exposure compensation")
	ErrInvalidFlashExposureComp    = errors.New(
		"invalid flash exposure compensation",
	)
	ErrInvalidFlashMode        = errors.New("invalid flash mode")
	ErrInvalidMeteringMode     = errors.New("invalid metering mode")
	ErrInvalidShootingMode     = errors.New("invalid shooting mode")
	ErrInvalidAutoFocusMode    = errors.New("invalid auto focus mode")
	ErrInvalidFilmAdvanceMode  = errors.New("invalid film advance mode")
	ErrInvalidMultipleExposure = errors.New("invalid multiple exposure")
	ErrParseApertureValue      = errors.New("failed to parse aperture value")
	ErrParseMaxApertureValue   = errors.New(
		"failed to parse max aperture value",
	)
	ErrParseExposureTimeValue = errors.New(
		"failed to parse exposure time value",
	)
	ErrParseBulbExposureTime = errors.New(
		"failed to parse bulb exposure time value",
	)
	ErrParseFlashModeValue = errors.New("failed to parse flash mode value")
)

const (
	TagUserComment          = "EXIF:UserComment"
	TagDateTimeOriginal     = "EXIF:DateTimeOriginal"
	TagExposureCompensation = "EXIF:ExposureCompensation"
	TagFlashExposureComp    = "EXIF:FlashExposureComp"
	TagFlash                = "EXIF:Flash"
	TagMeteringMode         = "EXIF:MeteringMode"

	TagFNumber          = "XMP-exif:FNumber"
	TagMaxApertureValue = "XMP-exif:MaxApertureValue"
	TagExposureTime     = "XMP-exif:ExposureTime"
	TagFocalLength      = "XMP-exif:FocalLength"
	TagISO              = "XMP-exif:ISO"

	TagBatteryLoadedDate = "XMP-AnalogueData:BatteryLoadedDate"
	TagFilmLoadedDate    = "XMP-AnalogueData:FilmLoadedDate"
	TagFilmISO           = "XMP-AnalogueData:FilmISO"
	TagManualISO         = "XMP-AnalogueData:ManualISO"
	TagFlashMode         = "XMP-AnalogueData:FlashMode"
	TagShootingMode      = "XMP-AnalogueData:ShootingMode"
	TagAFMode            = "XMP-AnalogueData:AFMode"
	TagFilmAdvanceMode   = "XMP-AnalogueData:FilmAdvanceMode"
	TagMultipleExposure  = "XMP-AnalogueData:MultipleExposure"

	metadataCapacity = 21
)

// Builder constructs EXIF metadata tag mappings from Canon EFD frame records.
type Builder interface {
	// Build converts an EFRM record into a map of EXIF tag names to values.
	// The strict parameter controls whether unknown metadata values cause errors.
	Build(efrm records.EFRM, strict bool) (map[string]string, error)
}

type builder struct {
	log *slog.Logger
}

func NewExifBuilder(log *slog.Logger) Builder {
	return &builder{log: log}
}

func (b *builder) Build(
	efrm records.EFRM,
	strict bool,
) (map[string]string, error) {
	metadata := make(map[string]string, metadataCapacity)

	if err := b.withFrameMetadata(metadata, efrm); err != nil {
		return nil, err
	}

	if err := b.withExposureSettings(metadata, efrm, strict); err != nil {
		return nil, err
	}

	if err := b.withFlashSettings(metadata, efrm, strict); err != nil {
		return nil, err
	}

	if err := b.withCameraModes(metadata, efrm); err != nil {
		return nil, err
	}

	return metadata, nil
}

func (b *builder) withFrameMetadata(
	metadata map[string]string,
	efrm records.EFRM,
) error {
	if remarks := string(domain.NewRemarks(efrm.Remarks)); remarks != "" {
		metadata[TagUserComment] = remarks
	}

	frameDatetime, err := domain.NewDateTime(
		efrm.Year, efrm.Month, efrm.Day,
		efrm.Hour, efrm.Minute, efrm.Second,
	)
	if err != nil {
		return errors.Join(ErrInvalidCaptureDate, err)
	} else if frameDatetime != "" {
		metadata[TagDateTimeOriginal] = string(frameDatetime)
	}

	batteryDatetime, err := domain.NewDateTime(
		efrm.BatteryYear, efrm.BatteryMonth, efrm.BatteryDay,
		efrm.BatteryHour, efrm.BatteryMinute, efrm.BatterySecond,
	)
	if err != nil {
		return errors.Join(ErrInvalidBatteryLoadedDate, err)
	} else if batteryDatetime != "" {
		metadata[TagBatteryLoadedDate] = string(batteryDatetime)
	}

	rollDatetime, err := domain.NewDateTime(
		efrm.RollYear, efrm.RollMonth, efrm.RollDay,
		efrm.RollHour, efrm.RollMinute, efrm.RollSecond,
	)
	if err != nil {
		return errors.Join(ErrInvalidFilmLoadedDate, err)
	} else if rollDatetime != "" {
		metadata[TagFilmLoadedDate] = string(rollDatetime)
	}

	return nil
}

func (b *builder) withExposureSettings(
	metadata map[string]string,
	efrm records.EFRM,
	strict bool,
) error {
	fNumber, err := b.processApertureValue(efrm.Av, strict)
	if err != nil {
		return errors.Join(ErrParseApertureValue, err)
	} else if fNumber != "" {
		metadata[TagFNumber] = fNumber
	}

	maxAv, err := b.processApertureValue(efrm.MaxAperture, strict)
	if err != nil {
		return errors.Join(ErrParseMaxApertureValue, err)
	} else if maxAv != "" {
		const apexConst = 2.0

		f, _ := strconv.ParseFloat(maxAv, 64)
		mav := fmt.Sprintf("%.1f", apexConst*math.Log2(f))
		metadata[TagMaxApertureValue] = mav
	}

	exposureTime, err := b.processExposureTime(
		efrm.Tv, efrm.BulbExposureTime, strict,
	)
	if err != nil {
		return err
	} else if exposureTime != "" {
		metadata[TagExposureTime] = exposureTime
	}

	focalLength := domain.NewFocalLength(efrm.FocalLength)
	if focalLength != "" {
		metadata[TagFocalLength] = string(focalLength)
	}

	if isoDX := string(domain.NewIso(efrm.IsoDX)); isoDX != "" {
		metadata[TagISO], metadata[TagFilmISO] = isoDX, isoDX
	}

	if isoM := string(domain.NewIso(efrm.IsoM)); isoM != "" {
		metadata[TagISO], metadata[TagManualISO] = isoM, isoM
	}

	ec, err := domain.NewExposureCompensation(efrm.ExposureCompensation, strict)
	if err != nil {
		return errors.Join(ErrInvalidExposureCompensation, err)
	} else if ec != "" {
		metadata[TagExposureCompensation] = string(ec)
	}

	return nil
}

func (b *builder) processApertureValue(av uint32, strict bool) (string, error) {
	avValue, err := domain.NewAv(av, strict)
	if err != nil {
		return "", err //nolint:wrapcheck // propagated and wrapped by caller
	}

	if avValue == "" || avValue == "00" {
		return "", nil
	}

	return string(avValue), nil
}

func (b *builder) processExposureTime(
	tv int32,
	bulbTime uint32,
	strict bool,
) (string, error) {
	tvValue, err := domain.NewTv(tv, strict)
	if err != nil {
		return "", errors.Join(ErrParseExposureTimeValue, err)
	}

	switch {
	case tvValue == "Bulb":
		_, bulbErr := domain.NewBulbExposureTime(bulbTime)
		if bulbErr != nil {
			return "", errors.Join(ErrParseBulbExposureTime, bulbErr)
		}

		return strconv.FormatUint(uint64(bulbTime), 10), nil
	case tv > 0:
		return strings.TrimSuffix(string(tvValue), "\""), nil
	default:
		return string(tvValue), nil
	}
}

func (b *builder) withFlashSettings(
	metadata map[string]string,
	efrm records.EFRM,
	strict bool,
) error {
	fec, err := domain.NewExposureCompensation(
		efrm.FlashExposureCompensation,
		strict,
	)
	if err != nil {
		return errors.Join(ErrInvalidFlashExposureComp, err)
	} else if fec != "" {
		metadata[TagFlashExposureComp] = string(fec)
	}

	flashMode, err := domain.NewFlashMode(efrm.FlashMode)
	if err != nil {
		return errors.Join(
			ErrInvalidFlashMode, ErrParseFlashModeValue, err,
		)
	} else if flashMode != "" {
		var flashBitmask string

		switch flashMode {
		case "OFF":
			flashBitmask = "0" // 0x00 - Flash did not fire
		case "ON":
			flashBitmask = "1" // 0x01 - Flash fired
		case "Manual flash":
			flashBitmask = "9" // 0x09 - Flash fired, forced on
		case "TTL autoflash", "A-TTL", "E-TTL":
			flashBitmask = "25" // 0x19 - Flash fired, auto mode
		}

		if flashBitmask != "" {
			metadata[TagFlash] = flashBitmask
		}

		metadata[TagFlashMode] = string(flashMode)
	}

	return nil
}

func (b *builder) withCameraModes(
	metadata map[string]string,
	efrm records.EFRM,
) error {
	mm, err := domain.NewMeteringMode(efrm.MeteringMode)
	if err != nil {
		return errors.Join(ErrInvalidMeteringMode, err)
	} else if mm != "" {
		metadata[TagMeteringMode] = string(mm)
	}

	sm, err := domain.NewShootingMode(efrm.ShootingMode)
	if err != nil {
		return errors.Join(ErrInvalidShootingMode, err)
	} else if sm != "" {
		metadata[TagShootingMode] = string(sm)
	}

	afm, err := domain.NewAutoFocusMode(efrm.AFMode)
	if err != nil {
		return errors.Join(ErrInvalidAutoFocusMode, err)
	} else if afm != "" {
		metadata[TagAFMode] = string(afm)
	}

	fam, err := domain.NewFilmAdvanceMode(efrm.FilmAdvanceMode)
	if err != nil {
		return errors.Join(ErrInvalidFilmAdvanceMode, err)
	} else if fam != "" {
		metadata[TagFilmAdvanceMode] = string(fam)
	}

	me, err := domain.NewMultipleExposure(efrm.MultipleExposure)
	if err != nil {
		return errors.Join(ErrInvalidMultipleExposure, err)
	} else if me != "" {
		metadata[TagMultipleExposure] = string(me)
	}

	return nil
}
