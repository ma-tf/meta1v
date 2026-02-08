//go:generate mockgen -destination=./mocks/service_mock.go -package=csv_test github.com/ma-tf/meta1v/internal/service/csv Service

// Package csv provides CSV export functionality for Canon EFD metadata.
//
// It converts displayable roll and frame data into CSV format suitable for
// spreadsheet applications.
package csv

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ma-tf/meta1v/internal/service/display"
)

var (
	ErrFailedToBufferRollHeader = errors.New("failed to buffer roll header")
	ErrFailedToWriteRollHeader  = errors.New(
		"failed to write out roll header",
	)
	ErrFailedToWriteFrames          = errors.New("failed to write frames")
	ErrFailedToWriteCustomFunctions = errors.New(
		"failed to write custom functions",
	)
)

// Service provides CSV export operations for film roll metadata.
type Service interface {
	// ExportRoll writes roll-level metadata (film ID, title, ISO, etc.) as CSV.
	ExportRoll(w io.Writer, r display.DisplayableRoll) error

	// ExportFrames writes detailed frame-by-frame metadata as CSV.
	ExportFrames(w io.Writer, f display.DisplayableRoll) error

	// ExportCustomFunctions writes custom function settings for each frame as CSV.
	ExportCustomFunctions(w io.Writer, cf display.DisplayableRoll) error
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) ExportRoll(w io.Writer, r display.DisplayableRoll) error {
	var b strings.Builder

	_, _ = b.WriteString(
		"FILM ID,FIRST ROW,FRAMES PER ROW,TITLE,FILM LOADED AT,FRAME COUNT,ISO (DX),REMARKS\n",
	)

	_, _ = fmt.Fprintf(&b, "%s,%s,%s,%s,%s,%s,%s,%s\n",
		r.FilmID,
		r.FirstRow,
		r.PerRow,
		r.Title,
		r.FilmLoadedDate,
		r.FrameCount,
		r.IsoDX,
		r.Remarks)

	if _, err := w.Write([]byte(b.String())); err != nil {
		return errors.Join(ErrFailedToWriteRollHeader, err)
	}

	return nil
}

func (s *service) ExportFrames(w io.Writer, f display.DisplayableRoll) error {
	var b strings.Builder

	_, _ = b.WriteString(
		"FILM ID,FILM LOADED AT,FRAME NUMBER,ISO (DX),FOCAL LENGTH,MAX APERTURE,Tv,Av,ISO (M),EXPOSURE COMPENSATION,FLASH EXPOSURE COMPENSATION,FLASH MODE,METERING MODE,SHOOTING MODE,FILM ADVANCE  MODE,AUTOFOCUS MODE,BULB EXPSOSURE TIME,TAKEN AT,MULTIPLE EXPOSURE,BATTERY LOADED AT,REMARKS,USER MODIFIED RECORD\n",
	)

	for _, frame := range f.Frames {
		_, _ = fmt.Fprintf(
			&b,
			"%s,%s,%d,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%v\n",
			frame.FilmID,
			frame.FilmLoadedAt,
			frame.FrameNumber,
			frame.IsoDX,
			frame.FocalLength,
			frame.MaxAperture,
			frame.Tv,
			frame.Av,
			frame.IsoM,
			frame.ExposureCompensation,
			frame.FlashExposureCompensation,
			frame.FlashMode,
			frame.MeteringMode,
			frame.ShootingMode,
			frame.FilmAdvanceMode,
			frame.AFMode,
			frame.BulbExposureTime,
			frame.TakenAt,
			frame.MultipleExposure,
			frame.BatteryLoadedAt,
			frame.Remarks,
			frame.UserModifiedRecord,
		)
	}

	if _, err := w.Write([]byte(b.String())); err != nil {
		return errors.Join(ErrFailedToWriteFrames, err)
	}

	return nil
}

func (s *service) ExportCustomFunctions(
	w io.Writer,
	cf display.DisplayableRoll,
) error {
	var b strings.Builder

	_, _ = b.WriteString(
		"FILM ID,FRAME NO.,C.Fn-1,C.Fn-2,C.Fn-3,C.Fn-4,C.Fn-5,C.Fn-6,C.Fn-7,C.Fn-8,C.Fn-9,C.Fn-10,C.Fn-11,C.Fn-12,C.Fn-13,C.Fn-14,C.Fn-15,C.Fn-16,C.Fn-17,C.Fn-18,C.Fn-19,C.Fn-20\n",
	)

	for _, frame := range cf.Frames {
		_, _ = fmt.Fprintf(&b, "%s,%d,%s\n",
			frame.FilmID,
			frame.FrameNumber,
			strings.Join(frame.CustomFunctions[:], ","))
	}

	if _, err := w.Write([]byte(b.String())); err != nil {
		return errors.Join(ErrFailedToWriteCustomFunctions, err)
	}

	return nil
}
