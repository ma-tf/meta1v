//go:generate mockgen -destination=./mocks/builder_mock.go -package=exif_test github.com/ma-tf/meta1v/internal/service/exif Builder
package exif

import (
	"errors"
	"fmt"
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
	ErrParseMaxApertureFloat = errors.New(
		"failed to parse max aperture float value",
	)
	ErrParseExposureTimeValue = errors.New(
		"failed to parse exposure time value",
	)
	ErrParseBulbExposureTime = errors.New(
		"failed to parse bulb exposure time value",
	)
	ErrParseFlashModeValue = errors.New("failed to parse flash mode value")
)

type Builder interface {
	Build(efrm records.EFRM, strict bool) (map[string]string, error)
}

type builder struct{}

func NewExifBuilder() Builder {
	return builder{}
}

func (b builder) Build(
	efrm records.EFRM,
	strict bool,
) (map[string]string, error) {
	metadata := map[string]string{}

	frameMeta, err := b.withFrameMetadata(efrm)
	if err != nil {
		return nil, err
	}

	exposure, err := b.withExposureSettings(efrm, strict)
	if err != nil {
		return nil, err
	}

	flashSettings, err := b.withFlashSettings(efrm, strict)
	if err != nil {
		return nil, err
	}

	cameraModes, err := b.withCameraModes(efrm, strict)
	if err != nil {
		return nil, err
	}

	for k, v := range frameMeta {
		if v != "" {
			metadata[k] = v
		}
	}

	for k, v := range exposure {
		if v != "" {
			metadata[k] = v
		}
	}

	for k, v := range flashSettings {
		if v != "" {
			metadata[k] = v
		}
	}

	for k, v := range cameraModes {
		if v != "" {
			metadata[k] = v
		}
	}

	return metadata, nil
}

func (b builder) withFrameMetadata(
	efrm records.EFRM,
) (map[string]string, error) {
	result := make(map[string]string)

	if remarks := string(domain.NewRemarks(efrm.Remarks)); remarks != "" {
		result["EXIF:UserComment"] = remarks
	}

	frameDatetime, err := domain.NewDateTime(
		efrm.Year, efrm.Month, efrm.Day,
		efrm.Hour, efrm.Minute, efrm.Second,
	)
	if err != nil {
		return nil, errors.Join(ErrInvalidCaptureDate, err)
	} else if frameDatetime != "" {
		result["EXIF:DateTimeOriginal"] = string(frameDatetime)
	}

	batteryDatetime, err := domain.NewDateTime(
		efrm.BatteryYear, efrm.BatteryMonth, efrm.BatteryDay,
		efrm.BatteryHour, efrm.BatteryMinute, efrm.BatterySecond,
	)
	if err != nil {
		return nil, errors.Join(ErrInvalidBatteryLoadedDate, err)
	} else if batteryDatetime != "" {
		result["XMP-AnalogueData:BatteryLoadedDate"] = string(batteryDatetime)
	}

	rollDatetime, err := domain.NewDateTime(
		efrm.RollYear, efrm.RollMonth, efrm.RollDay,
		efrm.RollHour, efrm.RollMinute, efrm.RollSecond,
	)
	if err != nil {
		return nil, errors.Join(ErrInvalidFilmLoadedDate, err)
	} else if rollDatetime != "" {
		result["XMP-AnalogueData:FilmLoadedDate"] = string(rollDatetime)
	}

	return result, nil
}

func (b builder) processApertureValue(av uint32, strict bool) (string, error) {
	avValue, err := domain.NewAv(av, strict)
	if err != nil {
		return "", fmt.Errorf("aperture conversion failed: %w", err)
	}

	if avValue == "" || avValue == "00" {
		return "", nil
	}

	return string(avValue), nil
}

func (b builder) processExposureTime(
	tv int32,
	bulbTime uint32,
	strict bool,
) (string, error) {
	tvValue, err := domain.NewTv(tv, strict)
	if err != nil {
		return "", fmt.Errorf("exposure time conversion failed: %w", err)
	}

	switch {
	case tvValue == "Bulb":
		_, bulbErr := domain.NewBulbExposureTime(bulbTime)
		if bulbErr != nil {
			return "",
				fmt.Errorf("bulb exposure time conversion failed: %w", bulbErr)
		}

		return strconv.FormatUint(uint64(bulbTime), 10), nil
	case tv > 0:
		return strings.TrimSuffix(string(tvValue), "\""), nil
	default:
		return string(tvValue), nil
	}
}

func (b builder) withExposureSettings(
	efrm records.EFRM,
	strict bool,
) (map[string]string, error) {
	result := make(map[string]string)

	fNumber, err := b.processApertureValue(efrm.Av, strict)
	if err != nil {
		return nil, errors.Join(ErrParseApertureValue, err)
	} else if fNumber != "" {
		result["XMP-exif:FNumber"] = fNumber
	}

	maxAv, err := b.processApertureValue(efrm.MaxAperture, strict)
	if err != nil {
		return nil, errors.Join(ErrParseMaxApertureValue, err)
	} else if maxAv != "" {
		f, parseErr := strconv.ParseFloat(maxAv, 64)
		if parseErr != nil {
			return nil, errors.Join(ErrParseMaxApertureFloat, parseErr)
		}

		const apexConst = 2.0

		mav := fmt.Sprintf("%.1f", apexConst*math.Log2(f))
		result["XMP-exif:MaxApertureValue"] = mav
	}

	exposureTime, err := b.processExposureTime(
		efrm.Tv, efrm.BulbExposureTime, strict,
	)
	if err != nil {
		return nil, errors.Join(ErrParseExposureTimeValue, err)
	} else if exposureTime != "" {
		result["XMP-exif:ExposureTime"] = exposureTime
	}

	if focalLength := domain.NewFocalLength(efrm.FocalLength); focalLength != "" {
		result["XMP-exif:FocalLength"] = string(focalLength)
	}

	if isoDX := string(domain.NewIso(efrm.IsoDX)); isoDX != "" {
		result["XMP-exif:ISO"], result["XMP-AnalogueData:FilmISO"] = isoDX, isoDX
	}

	if isoM := string(domain.NewIso(efrm.IsoM)); isoM != "" {
		result["XMP-exif:ISO"], result["XMP-AnalogueData:ManualISO"] = isoM, isoM
	}

	ec, err := domain.NewExposureCompensation(efrm.ExposureCompenation, strict)
	if err != nil {
		return nil, errors.Join(ErrInvalidExposureCompensation, err)
	} else if ec != "" {
		result["EXIF:ExposureCompensation"] = string(ec)
	}

	return result, nil
}

func (b builder) withFlashSettings(
	efrm records.EFRM,
	strict bool,
) (map[string]string, error) {
	result := make(map[string]string)

	fec, err := domain.NewExposureCompensation(
		efrm.FlashExposureCompensation,
		strict,
	)
	if err != nil {
		return nil, errors.Join(ErrInvalidFlashExposureComp, err)
	} else if fec != "" {
		result["EXIF:FlashExposureComp"] = string(fec)
	}

	flashMode, err := domain.NewFlashMode(efrm.FlashMode)
	if err != nil {
		return nil, errors.Join(
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
			result["EXIF:Flash"] = flashBitmask
		}

		result["XMP-AnalogueData:FlashMode"] = string(flashMode)
	}

	return result, nil
}

func (b builder) withCameraModes(
	efrm records.EFRM,
	_ bool,
) (map[string]string, error) {
	result := make(map[string]string)

	mm, err := domain.NewMeteringMode(efrm.MeteringMode)
	if err != nil {
		return nil, errors.Join(ErrInvalidMeteringMode, err)
	} else if mm != "" {
		result["EXIF:MeteringMode"] = string(mm)
	}

	sm, err := domain.NewShootingMode(efrm.ShootingMode)
	if err != nil {
		return nil, errors.Join(ErrInvalidShootingMode, err)
	} else if sm != "" {
		result["XMP-AnalogueData:ShootingMode"] = string(sm)
	}

	afm, err := domain.NewAutoFocusMode(efrm.AFMode)
	if err != nil {
		return nil, errors.Join(ErrInvalidAutoFocusMode, err)
	} else if afm != "" {
		result["XMP-AnalogueData:AFMode"] = string(afm)
	}

	fam, err := domain.NewFilmAdvanceMode(efrm.FilmAdvanceMode)
	if err != nil {
		return nil, errors.Join(ErrInvalidFilmAdvanceMode, err)
	} else if fam != "" {
		result["XMP-AnalogueData:FilmAdvanceMode"] = string(fam)
	}

	me, err := domain.NewMultipleExposure(efrm.MultipleExposure)
	if err != nil {
		return nil, errors.Join(ErrInvalidMultipleExposure, err)
	} else if me != "" {
		result["XMP-AnalogueData:MultipleExposure"] = string(me)
	}

	return result, nil
}
