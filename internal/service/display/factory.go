//go:generate mockgen -destination=./mocks/factory_mock.go -package=display_test github.com/ma-tf/meta1v/internal/service/display DisplayableRollFactory

package display

import (
	"fmt"

	"github.com/ma-tf/meta1v/pkg/domain"
	"github.com/ma-tf/meta1v/pkg/records"
)

var _ DisplayableRollFactory = &factory{}

type factory struct{}

type DisplayableRollFactory interface {
	Create(r records.Root) (DisplayableRoll, error)
}

func NewDisplayableRollFactory() DisplayableRollFactory {
	return &factory{}
}

func (f *factory) Create(
	r records.Root,
) (DisplayableRoll, error) {
	fid, err := domain.NewFilmID(r.EFDF.CodeA, r.EFDF.CodeB)
	if err != nil {
		return DisplayableRoll{},
			fmt.Errorf("failed to parse roll data: %w", err)
	}

	filmLoadedDate, err := domain.NewDateTime(
		r.EFDF.Year, r.EFDF.Month, r.EFDF.Day,
		r.EFDF.Hour, r.EFDF.Minute, r.EFDF.Second)
	if err != nil {
		return DisplayableRoll{},
			fmt.Errorf("failed to parse roll data: %w", err)
	}

	thumbnails, err := getThumbnails(r)
	if err != nil {
		return DisplayableRoll{}, err
	}

	// r.EFRMs != rr.Exposures ∵ multiple exposures? untested with real world frames
	frames, err := getFrames(r, thumbnails)
	if err != nil {
		return DisplayableRoll{}, err
	}

	return DisplayableRoll{
		FilmID:         fid,
		FirstRow:       uint(r.EFDF.PerRow - r.EFDF.FirstRow),
		PerRow:         uint(r.EFDF.PerRow),
		Title:          domain.NewTitle(r.EFDF.Title),
		FilmLoadedDate: filmLoadedDate,
		FrameCount:     uint(r.EFDF.FrameCount),
		IsoDX:          domain.NewIso(r.EFDF.IsoDX),
		Remarks:        domain.NewRemarks(r.EFDF.Remarks),
		Frames:         frames,
	}, nil
}
