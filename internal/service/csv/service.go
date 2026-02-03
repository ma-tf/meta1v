//go:generate mockgen -destination=./mocks/service_mock.go -package=csv_test github.com/ma-tf/meta1v/internal/service/csv Service
package csv

import (
	"fmt"
	"io"
	"strings"

	"github.com/ma-tf/meta1v/internal/service/display"
)

var _ Service = &service{}

type Service interface {
	ExportRoll(w io.Writer, r display.DisplayableRoll) error
	ExportFrames(w io.Writer, f display.DisplayableRoll) error
	ExportCustomFunctions(w io.Writer, f display.DisplayableRoll) error
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) ExportRoll(w io.Writer, r display.DisplayableRoll) error {
	var b strings.Builder
	b.WriteString(
		"FILM ID,FIRST ROW,FRAMES PER ROW,TITLE,FILM LOADED AT,FRAME COUNT,ISO (DX),REMARKS\n",
	)

	_, err := fmt.Fprintf(&b, "%s,%s,%s,%s,%s,%s,%s,%s",
		r.FilmID,
		r.FirstRow,
		r.PerRow,
		r.Title,
		r.FilmLoadedDate,
		r.FrameCount,
		r.IsoDX,
		r.Remarks,
	)
	if err != nil {
		return fmt.Errorf("failed to write roll header: %w", err)
	}

	if _, err = w.Write([]byte(b.String())); err != nil {
		return fmt.Errorf("failed to write roll header: %w", err)
	}

	return nil
}

func (s *service) ExportFrames(w io.Writer, f display.DisplayableRoll) error {
	var b strings.Builder
	b.WriteString(
		"FILM ID,FILM LOADED AT,FRAME NUMBER,ISO (DX),FOCAL LENGTH,MAX APERTURE,Tv,Av,ISO (M),EXPOSURE COMPENSATION,FLASH EXPOSURE COMPENSATION,FLASH MODE,METERING MODE,SHOOTING MODE,FILM ADVANCE  MODE,AUTOFOCUS MODE,BULB EXPSOSURE TIME,TAKEN AT,MULTIPLE EXPOSURE,BATTERY LOADED AT,REMARKS,USER MODIFIED RECORD\n",
	)

	for _, frame := range f.Frames {
		_, err := fmt.Fprintf(
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
			frame.FlashExposureComp,
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
		if err != nil {
			return fmt.Errorf(
				"failed to write frame %d: %w",
				frame.FrameNumber,
				err,
			)
		}
	}

	if _, err := w.Write([]byte(b.String())); err != nil {
		return fmt.Errorf("failed to write frames: %w", err)
	}

	return nil
}

func (s *service) ExportCustomFunctions(
	_ io.Writer,
	_ display.DisplayableRoll,
) error {
	return nil
}
