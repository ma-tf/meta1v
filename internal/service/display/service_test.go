package display_test

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/ma-tf/meta1v/internal/domain"
	"github.com/ma-tf/meta1v/internal/service/display"
)

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
func Test_DisplayRoll(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		roll           display.DisplayableRoll
		expectedOutput []byte
	}

	//nolint:golines // long lines for literal console output
	tests := []testcase{
		{
			name: "empty roll",
			roll: display.DisplayableRoll{},
			expectedOutput: []byte(`FILM ID  FIRST ROW FRAMES PER ROW TITLE                FILM LOADED AT       FRAME COUNT ISO (DX) REMARKS                       
-------------------------------------------------------------------------------------------------------------------------------
                                                                                                                               
`,
			),
		},
		{
			name: "populated roll",
			roll: display.DisplayableRoll{
				FilmID:         "12-ABC",
				FirstRow:       "3",
				PerRow:         "5",
				Title:          "My Film",
				FrameCount:     "10",
				FilmLoadedDate: "2023-05-15 14:30:00",
				IsoDX:          "200",
				Remarks:        "Sample remarks",
			},
			expectedOutput: []byte(`FILM ID  FIRST ROW FRAMES PER ROW TITLE                FILM LOADED AT       FRAME COUNT ISO (DX) REMARKS                       
-------------------------------------------------------------------------------------------------------------------------------
12-ABC   3         5              My Film              2023-05-15 14:30:00  10          200      Sample remarks                
`,
			),
		},
		{
			name: "populated roll with truncation",
			roll: display.DisplayableRoll{
				FilmID:         "12-ABC",
				FirstRow:       "3",
				PerRow:         "5",
				Title:          "63-character title 1234567890123456789012345678901234567890123456789012345678",
				FrameCount:     "10",
				FilmLoadedDate: "2023-05-15 14:30:00",
				IsoDX:          "200",
				Remarks:        "Sample remarks",
			},
			expectedOutput: []byte(`FILM ID  FIRST ROW FRAMES PER ROW TITLE                FILM LOADED AT       FRAME COUNT ISO (DX) REMARKS                       
-------------------------------------------------------------------------------------------------------------------------------
12-ABC   3         5              63-character titl... 2023-05-15 14:30:00  10          200      Sample remarks                
`,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			svc := display.NewService(newTestLogger())

			var b bytes.Buffer
			svc.DisplayRoll(ctx, &b, tt.roll)

			if !bytes.Equal(b.Bytes(), tt.expectedOutput) {
				t.Errorf("unexpected output:\n got:\n%s\nwant:\n%s",
					b.Bytes(),
					tt.expectedOutput,
				)
			}
		})
	}
}

func newCustomFunctions() domain.CustomFunctions {
	return domain.CustomFunctions{
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
	}
}

//nolint:exhaustruct // only partial is needed
func Test_DisplayCustomFunctions(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		roll           display.DisplayableRoll
		expectedError  error
		expectedOutput []byte
	}

	//nolint:golines // long lines for literal console output
	tests := []testcase{
		{
			name: "full custom functions",
			roll: display.DisplayableRoll{
				Frames: []display.DisplayableFrame{
					{
						CustomFunctions: newCustomFunctions(),
					},
				},
			},
			expectedOutput: []byte(`FILM ID  FRAME NO. #  1  2  3  4  5  6  7  8  9  10 11 12 13 14 15 16 17 18 19 20
---------------------------------------------------------------------------------
         0            1  1  1  1  1  1  1  1  1  1  1  1  1  1  1  1  1  1  1  1 
`,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			svc := display.NewService(newTestLogger())

			var b bytes.Buffer

			err := svc.DisplayCustomFunctions(ctx, &b, tt.roll)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error but got none")

					return
				}

				if !errors.Is(err, tt.expectedError) {
					t.Errorf("unexpected error: %v", err)
				}

				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)

				return
			}

			if !bytes.Equal(b.Bytes(), tt.expectedOutput) {
				t.Errorf("unexpected output:\n got:\n%s\nwant:\n%s",
					b.Bytes(),
					tt.expectedOutput,
				)
			}
		})
	}
}

//nolint:exhaustruct // only partial is needed
func Test_DisplayFocusingPoints(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		roll           display.DisplayableRoll
		expectedOutput []byte
	}

	tests := []testcase{
		{
			name: "successfully print focusing points",
			roll: display.DisplayableRoll{
				Frames: []display.DisplayableFrame{
					{
						FocusingPoints: `focusing
points`,
					},
				},
			},
			expectedOutput: []byte(`FILM ID  FRAME NO. FOCUSING POINTS      
----------------------------------------
         0         focusing
                   points

`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			svc := display.NewService(newTestLogger())

			var b bytes.Buffer
			svc.DisplayFocusingPoints(ctx, &b, tt.roll)

			if !bytes.Equal(b.Bytes(), tt.expectedOutput) {
				t.Errorf("unexpected output:\n got:\n%s\nwant:\n%s",
					b.Bytes(),
					tt.expectedOutput,
				)
			}
		})
	}
}

//nolint:exhaustruct // only partial is needed
func Test_DisplayFrame(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		frame          display.DisplayableFrame
		expectedOutput []byte
	}

	//nolint:golines // long lines for literal console output
	tests := []testcase{
		{
			name:  "empty frame",
			frame: display.DisplayableFrame{},
			expectedOutput: []byte(`FILM ID  FRAME NO. FILM LOADED AT       ISO (DX) FOCAL LENGTH MAX APERTURE TV      AV      ISO (M) EXPOSURE COMP.  FLASH EXPOSURE COMP. FLASH MODE      METERING MODE   SHOOTING MODE   FILM ADVANCE MODE AF MODE      BULB EXPOSURE TIME   TAKEN AT             MULTIPLE EXPOSURE    BATTERY LOADED AT    REMARKS                       
-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
         0                                                                                                                                                                                                                                                                                                                               
`,
			),
		},
		{
			name: "populated frame",
			frame: display.DisplayableFrame{
				FrameNumber:               1,
				FocalLength:               "50mm",
				MaxAperture:               "f/1.8",
				Tv:                        "1/125s",
				Av:                        "f/2.0",
				IsoM:                      "200",
				ExposureCompensation:      "+0.3",
				FlashExposureCompensation: "0",
				FilmLoadedAt:              "2023-05-15 14:30:00",
				TakenAt:                   "2023-05-15 14:30:00",
				FlashMode:                 "Auto",
				FilmAdvanceMode:           "Manual",
			},
			expectedOutput: []byte(`FILM ID  FRAME NO. FILM LOADED AT       ISO (DX) FOCAL LENGTH MAX APERTURE TV      AV      ISO (M) EXPOSURE COMP.  FLASH EXPOSURE COMP. FLASH MODE      METERING MODE   SHOOTING MODE   FILM ADVANCE MODE AF MODE      BULB EXPOSURE TIME   TAKEN AT             MULTIPLE EXPOSURE    BATTERY LOADED AT    REMARKS                       
-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
         1         2023-05-15 14:30:00           50mm         f/1.8        1/125s  f/2.0   200     +0.3            0                    Auto                                            Manual                                              2023-05-15 14:30:00                                                                          
`,
			),
		},
		{
			name: "frame with user modification",
			frame: display.DisplayableFrame{
				UserModifiedRecord:        true,
				FrameNumber:               1,
				FocalLength:               "50mm",
				MaxAperture:               "f/1.8",
				Tv:                        "1/125s",
				Av:                        "f/2.0",
				IsoM:                      "200",
				ExposureCompensation:      "+0.3",
				FlashExposureCompensation: "0",
				FilmLoadedAt:              "2023-05-15 14:30:00",
				TakenAt:                   "2023-05-15 14:30:00",
				FlashMode:                 "Auto",
				FilmAdvanceMode:           "Manual",
			},
			expectedOutput: []byte(`FILM ID  FRAME NO. FILM LOADED AT       ISO (DX) FOCAL LENGTH MAX APERTURE TV      AV      ISO (M) EXPOSURE COMP.  FLASH EXPOSURE COMP. FLASH MODE      METERING MODE   SHOOTING MODE   FILM ADVANCE MODE AF MODE      BULB EXPOSURE TIME   TAKEN AT             MULTIPLE EXPOSURE    BATTERY LOADED AT    REMARKS                       
-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
         1*        2023-05-15 14:30:00           50mm         f/1.8        1/125s  f/2.0   200     +0.3            0                    Auto                                            Manual                                              2023-05-15 14:30:00                                                                          
`,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			svc := display.NewService(newTestLogger())

			var b bytes.Buffer
			svc.DisplayFrames(ctx, &b, display.DisplayableRoll{
				Frames: []display.DisplayableFrame{tt.frame},
			})

			if !bytes.Equal(b.Bytes(), tt.expectedOutput) {
				t.Errorf("unexpected output:\n got:\n%s\nwant:\n%s",
					b.Bytes(),
					tt.expectedOutput,
				)
			}
		})
	}
}

//nolint:exhaustruct // only partial is needed
func Test_DisplayThumbnail(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		thumbnail      display.DisplayableRoll
		expectedOutput []byte
	}

	//nolint:golines // long lines for literal console output
	tests := []testcase{
		{
			name: "no thumbnail",
			thumbnail: display.DisplayableRoll{
				Frames: []display.DisplayableFrame{
					{
						FilmID:      "12-345",
						FrameNumber: 1,
					},
				},
			},
			expectedOutput: []byte(`FILM ID  FRAME NO. IMAGE FILE                                                       THUMBNAIL                                                       
----------------------------------------------------------------------------------------------------------------------------------------------------
`),
		},
		{
			name: "thumbnail exists",
			thumbnail: display.DisplayableRoll{
				Frames: []display.DisplayableFrame{
					{
						FilmID:      "12-345",
						FrameNumber: 1,
						Thumbnail: &display.DisplayableThumbnail{
							Filepath:  "thumb.jpg",
							Thumbnail: "qwerty", // image already in ascii form
						},
					},
				},
			},
			expectedOutput: []byte(`FILM ID  FRAME NO. IMAGE FILE                                                       THUMBNAIL                                                       
----------------------------------------------------------------------------------------------------------------------------------------------------
12-345   1         thumb.jpg                                                        qwerty
                                                         `),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			svc := display.NewService(newTestLogger())

			var b bytes.Buffer
			svc.DisplayThumbnails(ctx, &b, tt.thumbnail)

			if !bytes.Equal(b.Bytes(), tt.expectedOutput) {
				t.Errorf("unexpected output:\n got:\n%s\nwant:\n%s",
					b.Bytes(),
					tt.expectedOutput,
				)
			}
		})
	}
}
