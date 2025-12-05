package dng

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

var ErrMultipleFrames = errors.New("multiple frames with same number found")

type dngBuilder struct {
	strict      bool
	efrm        records.EFRM
	frameNumber int
	frame       exifMappedFrame
	err         error
}

type step1FrameBuilder interface {
	WithAvs() step2FrameBuilder
}

type step2FrameBuilder interface {
	WithTv() step3FrameBuilder
}

type step3FrameBuilder interface {
	WithFocalLengthAndIsoAndRemarks() *dngBuilder
}

func newDNGBuilder(
	r records.Root,
	frameNumber int,
	strict bool,
) step1FrameBuilder {
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

			efrm = &e
		}
	}

	return &dngBuilder{
		strict:      strict,
		efrm:        r.EFRMs[frameNumber],
		frameNumber: frameNumber,
		frame:       exifMappedFrame{}, //nolint:exhaustruct // will be built step by step
		err:         err,
	}
}

func (b *dngBuilder) WithAvs() step2FrameBuilder {
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

func (b *dngBuilder) WithTv() step3FrameBuilder {
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

func (b *dngBuilder) WithFocalLengthAndIsoAndRemarks() *dngBuilder {
	f := b.efrm

	b.frame.DcDescription = string(domain.NewRemarks(f.Remarks))

	fl := domain.NewFocalLength(f.FocalLength)
	b.frame.FocalLength = strings.TrimSuffix(string(fl), "mm")

	iso := string(domain.NewIso(f.IsoM))
	if iso == "" {
		iso = string(domain.NewIso(f.IsoDX))
	}

	b.frame.Iso = iso

	return b
}

func (b *dngBuilder) Build() (Exportable, error) {
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

	AuxImageNumber string
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
