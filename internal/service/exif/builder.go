package exif

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/ma-tf/meta1v/pkg/domain"
	"github.com/ma-tf/meta1v/pkg/records"
)

func transformAperture(av uint32, strict bool) (string, error) {
	avValue, err := domain.NewAv(av, strict)
	if err != nil {
		return "", fmt.Errorf("failed to parse aperture: %w", err)
	}

	if avValue == "00" {
		return "", nil
	}

	return string(avValue), nil
}

// transformMaxAperture converts the max aperture value to EXIF APEX format.
func transformMaxAperture(maxAperture uint32, strict bool) (string, error) {
	maxAv, err := domain.NewAv(maxAperture, strict)
	if err != nil {
		return "", fmt.Errorf("failed to parse max aperture: %w", err)
	}

	if maxAv == "" || maxAv == "00" {
		return "", nil
	}

	f, parseErr := strconv.ParseFloat(string(maxAv), 64)
	if parseErr != nil {
		return "", fmt.Errorf(
			"failed to parse max aperture float: %w",
			parseErr,
		)
	}

	const apexConst = 2

	apexMaxAv := apexConst * math.Log2(f)

	return fmt.Sprintf("%.1f", apexMaxAv), nil
}

// transformExposureTime converts the exposure time value to EXIF format.
func transformExposureTime(
	tv int32,
	bulbTime uint32,
	strict bool,
) (string, error) {
	tvValue, err := domain.NewTv(tv, strict)
	if err != nil {
		return "", fmt.Errorf("failed to parse exposure time: %w", err)
	}

	switch {
	case tvValue == "Bulb":
		_, bulbErr := domain.NewBulbExposureTime(bulbTime)
		if bulbErr != nil {
			return "", fmt.Errorf("failed to parse bulb exposure time: %w",
				bulbErr)
		}

		return strconv.FormatUint(uint64(bulbTime), 10), nil
	case tv > 0:
		return strings.TrimSuffix(string(tvValue), "\""), nil
	default:
		return string(tvValue), nil
	}
}

// transformFrameToExif converts frame record data to EXIF metadata structure.
func transformFrameToExif(
	efrm records.EFRM,
	strict bool,
) (*exifMappedFrame, error) {
	frame := &exifMappedFrame{} //nolint:exhaustruct // fields populated below

	// Transform aperture values
	fNumber, err := transformAperture(efrm.Av, strict)
	if err != nil {
		return nil, err
	}

	frame.FNumber = fNumber

	maxAperture, err := transformMaxAperture(efrm.MaxAperture, strict)
	if err != nil {
		return nil, err
	}

	frame.MaxAperture = maxAperture

	// Transform exposure time
	exposureTime, err := transformExposureTime(
		efrm.Tv, efrm.BulbExposureTime, strict)
	if err != nil {
		return nil, err
	}

	frame.ExposureTime = exposureTime

	// Transform focal length
	fl := domain.NewFocalLength(efrm.FocalLength)
	frame.FocalLength = string(fl)

	// Transform ISO
	iso := string(domain.NewIso(efrm.IsoM))
	if iso == "" {
		iso = string(domain.NewIso(efrm.IsoDX))
	}

	frame.Iso = iso

	// Transform remarks
	frame.DcDescription = string(domain.NewRemarks(efrm.Remarks))

	return frame, nil
}

type exifMappedFrame struct {
	DcDescription string

	FNumber      string
	MaxAperture  string
	FocalLength  string
	ExposureTime string
	Iso          string
}

func (emf *exifMappedFrame) FormatAsArgFile() string {
	var builder strings.Builder

	// Helper function to append tag only if the value is not empty
	appendTag := func(tag string, value string) {
		if value != "" {
			// Write the tag assignment, followed by a newline separator
			// ExifTool expects: -TAGNAME="VALUE"
			// The \n is the argument separator for -@ -
			fmt.Fprintf(&builder, "-%s=%s\n", tag, value)
		}
	}

	appendTag("XMP-dc:description", emf.DcDescription)
	appendTag("XMP-exif:FNumber", emf.FNumber)
	appendTag("XMP-exif:MaxApertureValue", emf.MaxAperture)
	appendTag("XMP-exif:FocalLength", emf.FocalLength)
	appendTag("XMP-exif:ExposureTime", emf.ExposureTime)
	appendTag("XMP-exif:ISO", emf.Iso)

	return builder.String()
}
