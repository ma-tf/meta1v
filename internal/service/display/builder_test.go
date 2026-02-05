package display_test

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/pkg/domain"
	"github.com/ma-tf/meta1v/pkg/records"
)

//nolint:exhaustruct // only partial needed
func Test_FrameBuilder_FrameMetadata(t *testing.T) {
	t.Parallel()

	validBaseFrame := func() records.EFRM {
		return records.EFRM{
			CodeA:                     12,
			CodeB:                     34,
			RollYear:                  2023,
			RollMonth:                 5,
			RollDay:                   15,
			RollHour:                  10,
			RollMinute:                30,
			RollSecond:                45,
			BatteryYear:               2023,
			BatteryMonth:              5,
			BatteryDay:                15,
			BatteryHour:               9,
			BatteryMinute:             15,
			BatterySecond:             0,
			Year:                      2023,
			Month:                     5,
			Day:                       15,
			Hour:                      10,
			Minute:                    45,
			Second:                    30,
			MaxAperture:               280,
			Tv:                        100,
			Av:                        280,
			ExposureCompensation:      100,
			MultipleExposure:          1,
			FlashExposureCompensation: 100,
			FlashMode:                 1,
			MeteringMode:              1,
			ShootingMode:              1,
			FilmAdvanceMode:           99,
			AFMode:                    1,
		}
	}

	assertError := func(t *testing.T, got, want error) {
		t.Helper()

		if got == nil {
			t.Fatalf("expected error %v, got nil", want)
		}

		if !errors.Is(got, want) {
			t.Fatalf("expected error %v, got %v", want, got)
		}
	}

	type testcase struct {
		name          string
		frame         records.EFRM
		expectedError error
	}

	tests := []testcase{
		{
			name: "invalid film id",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.CodeA = 100
				f.CodeB = 1000

				return f
			}(),
			expectedError: display.ErrInvalidFilmID,
		},
		{
			name: "invalid film loaded date",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.RollYear = 99
				f.RollMonth = 0

				return f
			}(),
			expectedError: display.ErrInvalidFilmLoadedDate,
		},
		{
			name: "invalid battery load date",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.BatteryYear = 99
				f.BatteryMonth = 0

				return f
			}(),
			expectedError: display.ErrInvalidBatteryLoadedDate,
		},
		{
			name: "invalid capture date",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.Year = 99
				f.Month = 0

				return f
			}(),
			expectedError: display.ErrInvalidCaptureDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			frameBuilder := display.NewFrameBuilder(newTestLogger())

			_, err := frameBuilder.Build(ctx, tt.frame, nil, false)

			assertError(t, err, tt.expectedError)
		})
	}
}

//nolint:exhaustruct // only partial needed
func Test_FrameBuilder_ExposureSettings(t *testing.T) {
	t.Parallel()

	validBaseFrame := func() records.EFRM {
		return records.EFRM{
			CodeA:                     12,
			CodeB:                     34,
			RollYear:                  2023,
			RollMonth:                 5,
			RollDay:                   15,
			RollHour:                  10,
			RollMinute:                30,
			RollSecond:                45,
			BatteryYear:               2023,
			BatteryMonth:              5,
			BatteryDay:                15,
			BatteryHour:               9,
			BatteryMinute:             15,
			BatterySecond:             0,
			Year:                      2023,
			Month:                     5,
			Day:                       15,
			Hour:                      10,
			Minute:                    45,
			Second:                    30,
			MaxAperture:               280,
			Tv:                        100,
			Av:                        280,
			ExposureCompensation:      100,
			MultipleExposure:          1,
			FlashExposureCompensation: 100,
			FlashMode:                 1,
			MeteringMode:              1,
			ShootingMode:              1,
			FilmAdvanceMode:           99,
			AFMode:                    1,
		}
	}

	assertError := func(t *testing.T, got, want error) {
		t.Helper()

		if got == nil {
			t.Fatalf("expected error %v, got nil", want)
		}

		if !errors.Is(got, want) {
			t.Fatalf("expected error %v, got %v", want, got)
		}
	}

	type testcase struct {
		name          string
		frame         records.EFRM
		strict        bool
		expectedError error
	}

	tests := []testcase{
		{
			name: "invalid max aperture in strict mode",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.MaxAperture = 101

				return f
			}(),
			strict:        true,
			expectedError: display.ErrInvalidMaxAperture,
		},
		{
			name: "invalid tv in strict mode",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.Tv = 101

				return f
			}(),
			strict:        true,
			expectedError: display.ErrInvalidShutterSpeed,
		},
		{
			name: "invalid bulb exposure time",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.Tv = 2130706432 // Bulb
				f.BulbExposureTime = 0

				return f
			}(),
			expectedError: display.ErrInvalidBulbExposureTime,
		},
		{
			name: "invalid av in strict mode",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.Av = 101

				return f
			}(),
			strict:        true,
			expectedError: display.ErrInvalidAperture,
		},
		{
			name: "invalid exposure compensation in strict mode",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.ExposureCompensation = 101

				return f
			}(),
			strict:        true,
			expectedError: display.ErrInvalidExposureCompensation,
		},
		{
			name: "invalid multiple exposure in strict mode",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.MultipleExposure = 101

				return f
			}(),
			strict:        true,
			expectedError: display.ErrInvalidMultipleExposure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			frameBuilder := display.NewFrameBuilder(newTestLogger())

			_, err := frameBuilder.Build(ctx, tt.frame, nil, tt.strict)

			assertError(t, err, tt.expectedError)
		})
	}
}

//nolint:exhaustruct // only partial needed
func Test_FrameBuilder_CameraModesAndFlash(t *testing.T) {
	t.Parallel()

	validBaseFrame := func() records.EFRM {
		return records.EFRM{
			CodeA:                     12,
			CodeB:                     34,
			RollYear:                  2023,
			RollMonth:                 5,
			RollDay:                   15,
			RollHour:                  10,
			RollMinute:                30,
			RollSecond:                45,
			BatteryYear:               2023,
			BatteryMonth:              5,
			BatteryDay:                15,
			BatteryHour:               9,
			BatteryMinute:             15,
			BatterySecond:             0,
			Year:                      2023,
			Month:                     5,
			Day:                       15,
			Hour:                      10,
			Minute:                    45,
			Second:                    30,
			MaxAperture:               280,
			Tv:                        100,
			Av:                        280,
			ExposureCompensation:      100,
			MultipleExposure:          1,
			FlashExposureCompensation: 100,
			FlashMode:                 1,
			MeteringMode:              1,
			ShootingMode:              1,
			FilmAdvanceMode:           99,
			AFMode:                    1,
		}
	}

	assertError := func(t *testing.T, got, want error) {
		t.Helper()

		if got == nil {
			t.Fatalf("expected error %v, got nil", want)
		}

		if !errors.Is(got, want) {
			t.Fatalf("expected error %v, got %v", want, got)
		}
	}

	type testcase struct {
		name          string
		frame         records.EFRM
		strict        bool
		expectedError error
	}

	tests := []testcase{
		{
			name: "invalid flash exposure compensation in strict mode",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.FlashExposureCompensation = 101

				return f
			}(),
			strict:        true,
			expectedError: display.ErrInvalidFlashExposureComp,
		},
		{
			name: "invalid flash mode in strict mode",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.FlashMode = 101

				return f
			}(),
			strict:        true,
			expectedError: display.ErrInvalidFlashMode,
		},
		{
			name: "invalid metering mode in strict mode",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.MeteringMode = 101

				return f
			}(),
			strict:        true,
			expectedError: display.ErrInvalidMeteringMode,
		},
		{
			name: "invalid shooting mode in strict mode",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.ShootingMode = 101

				return f
			}(),
			strict:        true,
			expectedError: display.ErrInvalidShootingMode,
		},
		{
			name: "invalid film advance mode in strict mode",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.FilmAdvanceMode = 101

				return f
			}(),
			strict:        true,
			expectedError: display.ErrInvalidFilmAdvanceMode,
		},
		{
			name: "invalid auto focus mode in strict mode",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.AFMode = 101

				return f
			}(),
			strict:        true,
			expectedError: display.ErrInvalidAutoFocusMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			frameBuilder := display.NewFrameBuilder(newTestLogger())

			_, err := frameBuilder.Build(ctx, tt.frame, nil, tt.strict)

			assertError(t, err, tt.expectedError)
		})
	}
}

//nolint:exhaustruct // only partial needed
func Test_FrameBuilder_CustomFunctionsAndFocus(t *testing.T) {
	t.Parallel()

	const esc = "\x1b"

	ansi := func(s string) display.DisplayableFocusPoints {
		return display.DisplayableFocusPoints(
			strings.ReplaceAll(s, "<ESC>", esc),
		)
	}

	validBaseFrame := func() records.EFRM {
		return records.EFRM{
			CodeA:                     12,
			CodeB:                     34,
			RollYear:                  2023,
			RollMonth:                 5,
			RollDay:                   15,
			RollHour:                  10,
			RollMinute:                30,
			RollSecond:                45,
			BatteryYear:               2023,
			BatteryMonth:              5,
			BatteryDay:                15,
			BatteryHour:               9,
			BatteryMinute:             15,
			BatterySecond:             0,
			Year:                      2023,
			Month:                     5,
			Day:                       15,
			Hour:                      10,
			Minute:                    45,
			Second:                    30,
			MaxAperture:               280,
			Tv:                        100,
			Av:                        280,
			ExposureCompensation:      100,
			MultipleExposure:          1,
			FlashExposureCompensation: 100,
			FlashMode:                 1,
			MeteringMode:              1,
			ShootingMode:              1,
			FilmAdvanceMode:           99,
			AFMode:                    1,
		}
	}

	assertError := func(t *testing.T, got, want error) {
		t.Helper()

		if got == nil {
			t.Fatalf("expected error\n%v, got nil", want)
		}

		if !errors.Is(got, want) {
			t.Fatalf("expected error %v, got %v", want, got)
		}
	}

	assertResult := func(t *testing.T, got, want display.DisplayableFrame, err error) {
		t.Helper()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != want {
			t.Fatalf("expected result\n%+v\n, got\n%+v\n", want, got)
		}
	}

	type testcase struct {
		name           string
		frame          records.EFRM
		strict         bool
		expectedError  error
		expectedResult display.DisplayableFrame
	}

	tests := []testcase{
		{
			name: "invalid custom function in strict mode",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.CustomFunction0 = 100

				return f
			}(),
			strict:        true,
			expectedError: domain.ErrInvalidCustomFunction,
		},
		{
			name: "no focusing points",
			frame: func() records.EFRM {
				f := validBaseFrame()

				return f
			}(),
			expectedResult: display.DisplayableFrame{
				FrameNumber:          0,
				FilmID:               "12-034",
				FilmLoadedAt:         "2023-05-15 10:30:45",
				BatteryLoadedAt:      "2023-05-15 09:15:00",
				TakenAt:              "2023-05-15 10:45:30",
				MaxAperture:          "f/2.8",
				Tv:                   "1\"",
				Av:                   "f/2.8",
				FocalLength:          "0mm",
				IsoDX:                "0",
				IsoM:                 "0",
				ExposureCompensation: "+1.0",
				MultipleExposure:     "ON",
				FlashExposureComp:    "+1.0",
				FlashMode:            "ON",
				MeteringMode:         "Center averaging",
				ShootingMode:         "Program AE",
				AFMode:               "One-Shot AF",
				CustomFunctions: [20]string{
					"0", "0", "0", "0", "0",
					"0", "0", "0", "0", "0",
					"0", "0", "0", "0", "0",
					"0", "0", "0", "0", "0",
				},
				//nolint:staticcheck // would make it even less readable
				FocusingPoints: ansi(
					`    [31m▯[0m [31m▯[0m [31m▯[0m [31m▯[0m [31m▯[0m [31m▯[0m [31m▯[0m 
 [31m▯[0m ▯ ▯ ▯ ▯ ▯ ▯ ▯ ▯ [31m▯[0m 
[31m▯[0m ▯ ▯ ▯ ▯ ▯ ▯ ▯ ▯ ▯ [31m▯[0m 
 [31m▯[0m ▯ ▯ ▯ ▯ ▯ ▯ ▯ ▯ [31m▯[0m 
    [31m▯[0m [31m▯[0m [31m▯[0m [31m▯[0m [31m▯[0m [31m▯[0m [31m▯[0m 
`,
				),
			},
		},
		{
			name: "empty focusing points",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.FocusingPoint = math.MaxUint32

				return f
			}(),
			expectedResult: display.DisplayableFrame{
				FrameNumber:          0,
				FilmID:               "12-034",
				FilmLoadedAt:         "2023-05-15 10:30:45",
				BatteryLoadedAt:      "2023-05-15 09:15:00",
				TakenAt:              "2023-05-15 10:45:30",
				MaxAperture:          "f/2.8",
				Tv:                   "1\"",
				Av:                   "f/2.8",
				FocalLength:          "0mm",
				IsoDX:                "0",
				IsoM:                 "0",
				ExposureCompensation: "+1.0",
				MultipleExposure:     "ON",
				FlashExposureComp:    "+1.0",
				FlashMode:            "ON",
				MeteringMode:         "Center averaging",
				ShootingMode:         "Program AE",
				AFMode:               "One-Shot AF",
				CustomFunctions: [20]string{
					"0", "0", "0", "0", "0",
					"0", "0", "0", "0", "0",
					"0", "0", "0", "0", "0",
					"0", "0", "0", "0", "0",
				},
				//nolint:staticcheck // would make it even less readable
				FocusingPoints: ansi(`    [30m▯ ▯ ▯ ▯ ▯ ▯ ▯
 ▯ ▯ ▯ ▯ ▯ ▯ ▯ ▯ ▯ ▯
▯ ▯ ▯ ▯ ▯ ▯ ▯ ▯ ▯ ▯ ▯
 ▯ ▯ ▯ ▯ ▯ ▯ ▯ ▯ ▯ ▯
    ▯ ▯ ▯ ▯ ▯ ▯ ▯[0m
`),
			},
		},
		{
			name: "all focusing points selected with red box",
			frame: func() records.EFRM {
				f := validBaseFrame()
				f.FocusingPoint = 1
				f.FocusPoints1 = 0b11111111
				f.FocusPoints2 = 0b11111111
				f.FocusPoints3 = 0b11111111
				f.FocusPoints4 = 0b11111111
				f.FocusPoints5 = 0b11111111
				f.FocusPoints6 = 0b11111111
				f.FocusPoints7 = 0b11111111
				f.FocusPoints8 = 0b11111111

				return f
			}(),
			expectedResult: display.DisplayableFrame{
				FrameNumber:          0,
				FilmID:               "12-034",
				FilmLoadedAt:         "2023-05-15 10:30:45",
				BatteryLoadedAt:      "2023-05-15 09:15:00",
				TakenAt:              "2023-05-15 10:45:30",
				MaxAperture:          "f/2.8",
				Tv:                   "1\"",
				Av:                   "f/2.8",
				FocalLength:          "0mm",
				IsoDX:                "0",
				IsoM:                 "0",
				ExposureCompensation: "+1.0",
				MultipleExposure:     "ON",
				FlashExposureComp:    "+1.0",
				FlashMode:            "ON",
				MeteringMode:         "Center averaging",
				ShootingMode:         "Program AE",
				AFMode:               "One-Shot AF",
				CustomFunctions: [20]string{
					"0", "0", "0", "0", "0",
					"0", "0", "0", "0", "0",
					"0", "0", "0", "0", "0",
					"0", "0", "0", "0", "0",
				},
				//nolint:staticcheck // would make it even less readable
				FocusingPoints: ansi(
					`    [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m 
 [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m 
[31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m 
 [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m 
    [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m [31m▮[0m 
`,
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			frameBuilder := display.NewFrameBuilder(newTestLogger())

			result, err := frameBuilder.Build(ctx, tt.frame, nil, tt.strict)

			if tt.expectedError != nil {
				assertError(t, err, tt.expectedError)

				return
			}

			assertResult(t, result, tt.expectedResult, err)
		})
	}
}
