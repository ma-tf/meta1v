//go:generate mockgen -destination=./mocks/builder_mock.go -package=exif_test github.com/ma-tf/meta1v/internal/service/exif Builder
package exif

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/ma-tf/meta1v/pkg/domain"
	"github.com/ma-tf/meta1v/pkg/records"
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
	fNumber, err := b.withAperture(efrm.Av, strict)
	if err != nil {
		return nil, err
	}

	maxAperture, err := b.withMaxAperture(efrm.MaxAperture, strict)
	if err != nil {
		return nil, err
	}

	exposureTime, err := b.withExposureTime(
		efrm.Tv,
		efrm.BulbExposureTime,
		strict,
	)
	if err != nil {
		return nil, err
	}

	iso := string(domain.NewIso(efrm.IsoM))
	if iso == "" {
		iso = string(domain.NewIso(efrm.IsoDX))
	}

	return map[string]string{
		"XMP-dc:description":        string(domain.NewRemarks(efrm.Remarks)),
		"XMP-exif:FNumber":          fNumber,
		"XMP-exif:MaxApertureValue": maxAperture,
		"XMP-exif:FocalLength": string(
			domain.NewFocalLength(efrm.FocalLength),
		),
		"XMP-exif:ExposureTime": exposureTime,
		"XMP-exif:ISO":          iso,
	}, nil
}

func (b builder) withAperture(av uint32, strict bool) (string, error) {
	avValue, err := domain.NewAv(av, strict)
	if err != nil {
		return "", fmt.Errorf("failed to parse aperture: %w", err)
	}

	if avValue == "00" {
		return "", nil
	}

	return string(avValue), nil
}

func (b builder) withMaxAperture(
	maxAperture uint32,
	strict bool,
) (string, error) {
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

func (b builder) withExposureTime(
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
