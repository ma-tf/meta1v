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

package csvexport_test

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/ma-tf/meta1v/internal/domain"
	"github.com/ma-tf/meta1v/internal/service/csvexport"
	"github.com/ma-tf/meta1v/internal/service/display"
)

var errExample = errors.New("example error")

type failWriter struct{}

func (fw *failWriter) Write(_ []byte) (int, error) {
	return 0, errExample
}

//nolint:exhaustruct // only partial is needed
func newTestLogger() *slog.Logger {
	buf := &bytes.Buffer{}

	return slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}

			return a
		},
	}))
}

//nolint:exhaustruct // only partial is needed
func Test_ExportRoll_Error(t *testing.T) {
	t.Parallel()

	dr := display.DisplayableRoll{}
	writer := &failWriter{}
	expectedError := csvexport.ErrFailedToWriteRollHeader

	ctx := t.Context()

	err := csvexport.NewService(newTestLogger()).ExportRoll(ctx, writer, dr)

	if !errors.Is(err, expectedError) {
		t.Errorf(
			"unexpected error: got %v, want %v",
			err,
			expectedError,
		)
	}
}

//nolint:exhaustruct // only partial is needed
func Test_ExportRoll_Success(t *testing.T) {
	t.Parallel()

	dr := display.DisplayableRoll{
		FilmID:         "AAA-BB",
		FirstRow:       "2024-01-01",
		PerRow:         "5",
		Title:          "My Film Roll",
		FilmLoadedDate: "2024-01-01T12:00:00Z",
		FrameCount:     "36",
		IsoDX:          "200",
		Remarks:        "This is a test roll.",
	}
	writer := &bytes.Buffer{}
	expectedOutput := []byte(
		`FILM ID,FIRST ROW,FRAMES PER ROW,TITLE,FILM LOADED AT,FRAME COUNT,ISO (DX),REMARKS
AAA-BB,2024-01-01,5,My Film Roll,2024-01-01T12:00:00Z,36,200,This is a test roll.
`,
	)

	ctx := t.Context()

	svc := csvexport.NewService(newTestLogger())

	err := svc.ExportRoll(ctx, writer, dr)
	if err != nil {
		t.Errorf("unexpected error: got %v, want %v", err, nil)
	}

	if !bytes.Equal(writer.Bytes(), expectedOutput) {
		t.Errorf(
			"unexpected output: got %s, want %s",
			writer.String(),
			string(expectedOutput),
		)
	}
}

//nolint:exhaustruct // only partial is needed
func Test_ExportFrames_Error(t *testing.T) {
	t.Parallel()

	dr := display.DisplayableRoll{
		Frames: []display.DisplayableFrame{
			{},
		},
	}
	writer := &failWriter{}
	expectedError := csvexport.ErrFailedToWriteFrames

	ctx := t.Context()

	err := csvexport.NewService(newTestLogger()).ExportFrames(ctx, writer, dr)

	if !errors.Is(err, expectedError) {
		t.Errorf(
			"unexpected error: got %v, want %v",
			err,
			expectedError,
		)
	}
}

//nolint:exhaustruct // only partial is needed
func Test_ExportFrames_Success(t *testing.T) {
	t.Parallel()

	dr := display.DisplayableRoll{
		Frames: []display.DisplayableFrame{
			{
				FilmID:                    "AAA-BB",
				FilmLoadedAt:              "2024-01-01T12:00:00Z",
				FrameNumber:               1,
				IsoDX:                     "200",
				FocalLength:               "50mm",
				MaxAperture:               "f/1.8",
				Tv:                        "1/125",
				Av:                        "f/1.8",
				IsoM:                      "200",
				ExposureCompensation:      "+0.3",
				FlashExposureCompensation: "+0.7",
				FlashMode:                 "On",
				MeteringMode:              "Evaluative",
				ShootingMode:              "Manual",
				FilmAdvanceMode:           "Single Frame",
				AFMode:                    "One-Shot AF",
				BulbExposureTime:          "",
				TakenAt:                   "2024-01-01T12:00:00Z",
				MultipleExposure:          "No",
				BatteryLoadedAt:           "2024-01-01T11:00:00Z",
				Remarks:                   "This is a test frame.",
				UserModifiedRecord:        true,
			},
		},
	}
	writer := &bytes.Buffer{}
	expectedOutput := []byte(
		`FILM ID,FILM LOADED AT,FRAME NUMBER,ISO (DX),FOCAL LENGTH,MAX APERTURE,Tv,Av,ISO (M),EXPOSURE COMPENSATION,FLASH EXPOSURE COMPENSATION,FLASH MODE,METERING MODE,SHOOTING MODE,FILM ADVANCE  MODE,AUTOFOCUS MODE,BULB EXPSOSURE TIME,TAKEN AT,MULTIPLE EXPOSURE,BATTERY LOADED AT,REMARKS,USER MODIFIED RECORD
AAA-BB,2024-01-01T12:00:00Z,1,200,50mm,f/1.8,1/125,f/1.8,200,+0.3,+0.7,On,Evaluative,Manual,Single Frame,One-Shot AF,,2024-01-01T12:00:00Z,No,2024-01-01T11:00:00Z,This is a test frame.,true
`,
	)

	ctx := t.Context()

	svc := csvexport.NewService(newTestLogger())

	err := svc.ExportFrames(ctx, writer, dr)
	if err != nil {
		t.Errorf("unexpected error: got %v, want %v", err, nil)
	}

	if !bytes.Equal(writer.Bytes(), expectedOutput) {
		t.Errorf(
			"unexpected output: got %s, want %s",
			writer.String(),
			string(expectedOutput),
		)
	}
}

//nolint:exhaustruct // only partial is needed
func Test_ExportCustomFunctions_Error(t *testing.T) {
	t.Parallel()

	dr := display.DisplayableRoll{
		Frames: []display.DisplayableFrame{
			{},
		},
	}
	writer := &failWriter{}
	expectedError := csvexport.ErrFailedToWriteCustomFunctions

	ctx := t.Context()

	svc := csvexport.NewService(newTestLogger())

	err := svc.ExportCustomFunctions(ctx, writer, dr)

	if !errors.Is(err, expectedError) {
		t.Errorf(
			"unexpected error: got %v, want %v",
			err,
			expectedError,
		)
	}
}

//nolint:exhaustruct // only partial is needed
func Test_ExportCustomFunctions_Success(t *testing.T) {
	t.Parallel()

	dr := display.DisplayableRoll{
		Frames: []display.DisplayableFrame{
			{
				FilmID:      "AAA-BB",
				FrameNumber: 1,
				CustomFunctions: domain.CustomFunctions([]string{
					"1",
					"1",
					"1",
					"1",
					"1",
					"1",
					"1",
					"1",
					"1",
					"1",
					"1",
					"1",
					"1",
					"1",
					"1",
					"1",
					"1",
					"1",
					"1",
					"1",
				}),
			},
		},
	}
	writer := &bytes.Buffer{}
	expectedOutput := []byte(
		`FILM ID,FRAME NO.,C.Fn-1,C.Fn-2,C.Fn-3,C.Fn-4,C.Fn-5,C.Fn-6,C.Fn-7,C.Fn-8,C.Fn-9,C.Fn-10,C.Fn-11,C.Fn-12,C.Fn-13,C.Fn-14,C.Fn-15,C.Fn-16,C.Fn-17,C.Fn-18,C.Fn-19,C.Fn-20
AAA-BB,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1
`,
	)

	ctx := t.Context()

	svc := csvexport.NewService(newTestLogger())

	err := svc.ExportCustomFunctions(ctx, writer, dr)
	if err != nil {
		t.Errorf("unexpected error: got %v, want %v", err, nil)
	}

	if !bytes.Equal(writer.Bytes(), expectedOutput) {
		t.Errorf(
			"unexpected output: got %s, want %s",
			writer.String(),
			string(expectedOutput),
		)
	}
}
