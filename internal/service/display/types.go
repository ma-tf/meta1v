package display

import "github.com/ma-tf/meta1v/pkg/domain"

type DisplayableRoll struct {
	FilmID         domain.FilmID
	FirstRow       domain.FirstRow
	PerRow         domain.PerRow
	Title          domain.Title
	FilmLoadedDate domain.ValidatedDatetime
	FrameCount     domain.FrameCount
	IsoDX          domain.Iso
	Remarks        domain.Remarks // film name, location, push/pull, etc.

	Frames []DisplayableFrame
}

type DisplayableFrame struct {
	FrameNumber  uint
	FilmID       domain.FilmID
	FilmLoadedAt domain.ValidatedDatetime
	IsoDX        domain.Iso

	UserModifiedRecord bool

	FocalLength domain.FocalLength
	MaxAperture domain.Av
	Tv          domain.Tv
	Av          domain.Av
	IsoM        domain.Iso

	ExposureCompensation domain.ExposureCompenation
	FlashExposureComp    domain.ExposureCompenation
	FlashMode            domain.FlashMode
	MeteringMode         domain.MeteringMode
	ShootingMode         domain.ShootingMode

	FilmAdvanceMode  domain.FilmAdvanceMode
	AFMode           domain.AutoFocusMode
	BulbExposureTime domain.BulbExposureTime
	TakenAt          domain.ValidatedDatetime

	MultipleExposure domain.MultipleExposure
	BatteryLoadedAt  domain.ValidatedDatetime

	CustomFunctions domain.CustomFunctions
	Remarks         domain.Remarks

	FocusingPoints domain.FocusPoints

	Thumbnail *DisplayableThumbnail
}

type DisplayableThumbnail struct {
	Thumbnail string
	Filepath  string
}
