// meta1v is a command-line tool for viewing and manipulating metadata for Canon EOS-1V files of the EFD format.
// Copyright (C) 2026  Matt F
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

//go:generate mockgen -destination=./mocks/service_mock.go -package=csv_test github.com/ma-tf/meta1v/internal/service/csv Service

// Package csv provides CSV export functionality for Canon EFD metadata.
//
// It converts displayable roll and frame data into CSV format suitable for
// spreadsheet applications.
package csv

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
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
	ExportRoll(
		ctx context.Context,
		w io.Writer,
		r display.DisplayableRoll,
	) error

	// ExportFrames writes detailed frame-by-frame metadata as CSV.
	ExportFrames(
		ctx context.Context,
		w io.Writer,
		f display.DisplayableRoll,
	) error

	// ExportCustomFunctions writes custom function settings for each frame as CSV.
	ExportCustomFunctions(
		ctx context.Context,
		w io.Writer,
		cf display.DisplayableRoll,
	) error
}

type service struct {
	log *slog.Logger
}

func NewService(log *slog.Logger) Service {
	return &service{
		log: log,
	}
}

func (s *service) ExportRoll(
	ctx context.Context,
	w io.Writer,
	r display.DisplayableRoll,
) error {
	s.log.InfoContext(ctx, "exporting roll to csv",
		slog.String("film_id", string(r.FilmID)))

	var b strings.Builder

	_, _ = b.WriteString(
		"FILM ID,FIRST ROW,FRAMES PER ROW,TITLE,FILM LOADED AT,FRAME COUNT,ISO (DX),REMARKS\n",
	)

	s.log.DebugContext(ctx, "csv headers written")

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

	s.log.InfoContext(ctx, "roll csv export completed",
		slog.Int("bytes_written", b.Len()))

	return nil
}

func (s *service) ExportFrames(
	ctx context.Context,
	w io.Writer,
	f display.DisplayableRoll,
) error {
	s.log.InfoContext(ctx, "exporting frames to csv",
		slog.String("film_id", string(f.FilmID)),
		slog.Int("frame_count", len(f.Frames)))

	var b strings.Builder

	_, _ = b.WriteString(
		"FILM ID,FILM LOADED AT,FRAME NUMBER,ISO (DX),FOCAL LENGTH,MAX APERTURE,Tv,Av,ISO (M),EXPOSURE COMPENSATION,FLASH EXPOSURE COMPENSATION,FLASH MODE,METERING MODE,SHOOTING MODE,FILM ADVANCE  MODE,AUTOFOCUS MODE,BULB EXPSOSURE TIME,TAKEN AT,MULTIPLE EXPOSURE,BATTERY LOADED AT,REMARKS,USER MODIFIED RECORD\n",
	)

	s.log.DebugContext(ctx, "csv headers written")

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

	s.log.DebugContext(ctx, "frame data written",
		slog.Int("frame_count", len(f.Frames)))

	if _, err := w.Write([]byte(b.String())); err != nil {
		return errors.Join(ErrFailedToWriteFrames, err)
	}

	s.log.InfoContext(ctx, "frames csv export completed",
		slog.Int("bytes_written", b.Len()),
		slog.Int("frame_count", len(f.Frames)))

	return nil
}

func (s *service) ExportCustomFunctions(
	ctx context.Context,
	w io.Writer,
	cf display.DisplayableRoll,
) error {
	s.log.InfoContext(ctx, "exporting custom functions to csv",
		slog.String("film_id", string(cf.FilmID)),
		slog.Int("frame_count", len(cf.Frames)))

	var b strings.Builder

	_, _ = b.WriteString(
		"FILM ID,FRAME NO.,C.Fn-1,C.Fn-2,C.Fn-3,C.Fn-4,C.Fn-5,C.Fn-6,C.Fn-7,C.Fn-8,C.Fn-9,C.Fn-10,C.Fn-11,C.Fn-12,C.Fn-13,C.Fn-14,C.Fn-15,C.Fn-16,C.Fn-17,C.Fn-18,C.Fn-19,C.Fn-20\n",
	)

	s.log.DebugContext(ctx, "csv headers written")

	for _, frame := range cf.Frames {
		_, _ = fmt.Fprintf(&b, "%s,%d,%s\n",
			frame.FilmID,
			frame.FrameNumber,
			strings.Join(frame.CustomFunctions[:], ","))
	}

	s.log.DebugContext(ctx, "custom functions data written",
		slog.Int("frame_count", len(cf.Frames)))

	if _, err := w.Write([]byte(b.String())); err != nil {
		return errors.Join(ErrFailedToWriteCustomFunctions, err)
	}

	s.log.InfoContext(ctx, "custom functions csv export completed",
		slog.Int("bytes_written", b.Len()),
		slog.Int("frame_count", len(cf.Frames)))

	return nil
}
