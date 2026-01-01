//go:generate mockgen -destination=./mocks/factory_mock.go -package=display_test github.com/ma-tf/meta1v/internal/service/display DisplayableRollFactory

package display

import (
	"bytes"
	"errors"
	"fmt"
	"math"

	"github.com/ma-tf/meta1v/pkg/domain"
	"github.com/ma-tf/meta1v/pkg/records"
	"github.com/qeesung/image2ascii/convert"
)

var (
	ErrFailedToParseRollData = errors.New("failed to parse roll data")
	ErrFailedToBuildFrame    = errors.New("failed to build frame")
)

type factory struct {
	frameBuilder FrameMetadataBuilder
}

type DisplayableRollFactory interface {
	Create(r records.Root) (DisplayableRoll, error)
}

func NewDisplayableRollFactory(
	frameBuilder FrameMetadataBuilder,
) DisplayableRollFactory {
	return &factory{
		frameBuilder: frameBuilder,
	}
}

func (f *factory) Create(
	r records.Root,
) (DisplayableRoll, error) {
	fid, err := domain.NewFilmID(r.EFDF.CodeA, r.EFDF.CodeB)
	if err != nil {
		return DisplayableRoll{},
			errors.Join(ErrFailedToParseRollData, err)
	}

	firstRow, err := domain.NewFirstRow(r.EFDF.FirstRow, r.EFDF.PerRow)
	if err != nil {
		return DisplayableRoll{},
			errors.Join(ErrFailedToParseRollData, err)
	}

	filmLoadedDate, err := domain.NewDateTime(
		r.EFDF.Year, r.EFDF.Month, r.EFDF.Day,
		r.EFDF.Hour, r.EFDF.Minute, r.EFDF.Second)
	if err != nil {
		return DisplayableRoll{},
			errors.Join(ErrFailedToParseRollData, err)
	}

	thumbnails, err := f.getThumbnails(r)
	if err != nil {
		return DisplayableRoll{}, err
	}

	// r.EFRMs != rr.Exposures ∵ multiple exposures? untested with real world frames
	frames, err := f.getFrames(r, thumbnails)
	if err != nil {
		return DisplayableRoll{}, err
	}

	return DisplayableRoll{
		FilmID:         fid,
		FirstRow:       firstRow,
		PerRow:         domain.NewPerRow(r.EFDF.PerRow),
		Title:          domain.NewTitle(r.EFDF.Title),
		FilmLoadedDate: filmLoadedDate,
		FrameCount:     domain.NewFrameCount(r.EFDF.FrameCount),
		IsoDX:          domain.NewIso(r.EFDF.IsoDX),
		Remarks:        domain.NewRemarks(r.EFDF.Remarks),
		Frames:         frames,
	}, nil
}

func (f *factory) getFrames(
	r records.Root,
	thumbnails map[uint16]*DisplayableThumbnail,
) ([]DisplayableFrame, error) {
	frames := make([]DisplayableFrame, 0, len(r.EFRMs))
	for i, frame := range r.EFRMs {
		idx := i + 1
		if idx < 0 || idx > math.MaxUint16 {
			return nil,
				fmt.Errorf("%w: index %d", ErrFrameIndexOutOfRange, i+1)
		}

		var pt *DisplayableThumbnail
		if t, ok := thumbnails[uint16(idx)]; ok {
			pt = t
		}

		framePF, errPF := f.frameBuilder.
			WithFrameMetadata(frame).
			WithExposureSettings().
			WithCameraModesAndFlashInfo().
			WithCustomFunctionsAndFocusPoints().
			WithThumbnail(pt).
			Build()
		if errPF != nil {
			return nil, errors.Join(ErrFailedToBuildFrame, errPF)
		}

		frames = append(frames, framePF)
	}

	return frames, nil
}

func (f *factory) getThumbnails(
	r records.Root,
) (map[uint16]*DisplayableThumbnail, error) {
	const heightRatio = 2

	thumbnails := make(map[uint16]*DisplayableThumbnail, len(r.EFTPs))
	for _, eftp := range r.EFTPs {
		filepath := string(eftp.Filepath[:bytes.IndexByte(eftp.Filepath[:], 0)])

		options := convert.DefaultOptions
		options.FixedWidth = int(eftp.Width)

		options.FixedHeight = int(eftp.Height / heightRatio)

		ascii := convert.NewImageConverter().
			Image2ASCIIString(eftp.Thumbnail, &options)

		if thumbnails[eftp.Index] != nil {
			return nil, fmt.Errorf("%w: frame number %d",
				ErrMultipleThumbnailsForFrame, eftp.Index)
		}

		thumbnails[eftp.Index] = &DisplayableThumbnail{
			Thumbnail: ascii,
			Filepath:  filepath,
		}
	}

	return thumbnails, nil
}
