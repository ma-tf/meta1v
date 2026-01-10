package exif

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/ma-tf/meta1v/pkg/domain"
	"github.com/ma-tf/meta1v/pkg/records"
)

var (
	ErrMultipleFrames = errors.New("multiple frames with same number found")
	ErrFrameNotFound  = errors.New("frame not found")
)

type exifBuilder struct {
	efrm        records.EFRM
	frameNumber int
	frame       exifMappedFrame
	err         error
}

func newExifBuilder(
	r records.Root,
	frameNumber int,
) *exifBuilder {
	var (
		efrm *records.EFRM
		err  error
	)

	for _, e := range r.EFRMs {
		if int(e.FrameNumber) == frameNumber {
			if efrm != nil {
				err = fmt.Errorf("%w: frame number %d",
					ErrMultipleFrames, frameNumber)

				break
			}

			e := e // Create a copy to avoid pointer issues
			efrm = &e
		}
	}

	if efrm == nil && err == nil {
		err = fmt.Errorf("%w: frame number %d", ErrFrameNotFound, frameNumber)
	}

	var eframValue records.EFRM
	if efrm != nil {
		eframValue = *efrm
	}

	return &exifBuilder{
		efrm:        eframValue,
		frameNumber: frameNumber,
		frame:       exifMappedFrame{}, //nolint:exhaustruct // will be built step by step
		err:         err,
	}
}

func (b *exifBuilder) WithAvs() *exifBuilder {
	if b.err != nil {
		return b
	}

	var (
		av    domain.Av
		maxAv domain.Av
	)

	av, b.err = domain.NewAv(b.efrm.Av, false)
	if b.err != nil {
		return b
	}

	if av == "00" {
		b.frame.FNumber = ""
	} else {
		b.frame.FNumber = strings.TrimPrefix(string(av), "f/")
	}

	maxAv, b.err = domain.NewAv(b.efrm.MaxAperture, false)
	if b.err != nil {
		return b
	}

	if maxAv == "" || maxAv == "00" {
		b.frame.MaxAperture = ""
	} else {
		var f float64

		mav := strings.TrimPrefix(string(maxAv), "f/")

		f, b.err = strconv.ParseFloat(mav, 64)
		if b.err != nil {
			return b
		}

		const apexConst = 2

		apexMaxAv := apexConst * math.Log2(f)
		b.frame.MaxAperture = fmt.Sprintf("%.1f", apexMaxAv)
	}

	return b
}

func (b *exifBuilder) WithTv() *exifBuilder {
	if b.err != nil {
		return b
	}

	var (
		tv               domain.Tv
		bulbExposureTime domain.BulbExposureTime
	)

	tv, b.err = domain.NewTv(b.efrm.Tv, false)
	if b.err != nil {
		return b
	}

	switch {
	case tv == "Bulb":
		bulbExposureTime, b.err = domain.NewBulbExposureTime(
			b.efrm.BulbExposureTime,
		)
		if b.err != nil {
			return b
		}

		t, err := time.Parse(time.TimeOnly, string(bulbExposureTime))
		if err != nil {
			b.err = err

			return b
		}

		total := t.Hour()*3600 + t.Minute()*60 + t.Second()
		b.frame.ExposureTime = strconv.Itoa(total)
	case b.efrm.Tv > 0:
		b.frame.ExposureTime = strings.TrimSuffix(string(tv), "\"")
	default:
		b.frame.ExposureTime = string(tv)
	}

	return b
}

func (b *exifBuilder) WithFocalLength() *exifBuilder {
	if b.err != nil {
		return b
	}

	fl := domain.NewFocalLength(b.efrm.FocalLength)
	b.frame.FocalLength = strings.TrimSuffix(string(fl), "mm")

	return b
}

func (b *exifBuilder) WithIso() *exifBuilder {
	if b.err != nil {
		return b
	}

	iso := string(domain.NewIso(b.efrm.IsoM))
	if iso == "" {
		iso = string(domain.NewIso(b.efrm.IsoDX))
	}

	b.frame.Iso = iso

	return b
}

func (b *exifBuilder) WithRemarks() *exifBuilder {
	if b.err != nil {
		return b
	}

	b.frame.DcDescription = string(domain.NewRemarks(b.efrm.Remarks))

	return b
}

func (b *exifBuilder) Build() (Exportable, error) {
	return &b.frame, b.err
}

type Exportable interface {
	GetMetadataToWrite() string
}

type exifMappedFrame struct {
	DcDescription string

	FNumber      string
	MaxAperture  string
	FocalLength  string
	ExposureTime string
	Iso          string
}

func (emf *exifMappedFrame) GetMetadataToWrite() string {
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
