package display_test

import (
	"errors"
	"math"
	"testing"

	"github.com/ma-tf/meta1v/internal/service/display"
	"github.com/ma-tf/meta1v/pkg/domain"
	"github.com/ma-tf/meta1v/pkg/records"
)

//nolint:exhaustruct // only partial is needed
func Test_FrameBuilder_WithFrameMetadata(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		frame          records.EFRM
		strict         bool
		expectedResult display.DisplayableFrame
		expectedError  error
	}

	tests := []testcase{
		{
			name: "invalid film id",
			frame: records.EFRM{
				CodeA: 100,
				CodeB: 1000,
			},
			expectedError: domain.ErrPrefixOutOfRange,
		},
		{
			name: "invalid film loaded date",
			frame: records.EFRM{
				RollYear:  99,
				RollMonth: 0,
			},
			expectedError: domain.ErrInvalidDateTime,
		},
		{
			name: "invalid battery load date",
			frame: records.EFRM{
				RollYear:     math.MaxUint16,
				RollMonth:    math.MaxUint8,
				RollDay:      math.MaxUint8,
				RollHour:     math.MaxUint8,
				RollMinute:   math.MaxUint8,
				RollSecond:   math.MaxUint8,
				BatteryYear:  99,
				BatteryMonth: 0,
			},
			expectedError: domain.ErrInvalidDateTime,
		},
		{
			name: "invalid capture date",
			frame: records.EFRM{
				RollYear:      math.MaxUint16,
				RollMonth:     math.MaxUint8,
				RollDay:       math.MaxUint8,
				RollHour:      math.MaxUint8,
				RollMinute:    math.MaxUint8,
				RollSecond:    math.MaxUint8,
				BatteryYear:   math.MaxUint16,
				BatteryMonth:  math.MaxUint8,
				BatteryDay:    math.MaxUint8,
				BatteryHour:   math.MaxUint8,
				BatteryMinute: math.MaxUint8,
				BatterySecond: math.MaxUint8,
				Year:          99,
				Month:         0,
			},
			expectedError: domain.ErrInvalidDateTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			frameBuilder := display.NewFrameBuilder(newTestLogger())

			result, err := frameBuilder.
				WithFrameMetadata(ctx, tt.frame).
				WithExposureSettings(ctx, tt.strict).
				WithCameraModesAndFlashInfo(ctx, tt.strict).
				WithCustomFunctionsAndFocusPoints(ctx, tt.strict).
				WithThumbnail(ctx, nil).
				Build()

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedError)
				}

				if !errors.Is(err, tt.expectedError) {
					t.Fatalf("expected error %v, got %v", tt.expectedError, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expectedResult {
				t.Fatalf("expected result %v, got %v",
					tt.expectedResult, result)
			}
		})
	}
}

//nolint:exhaustruct // only partial is needed
func Test_FrameBuilder_WithExposureSettings(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		frame          records.EFRM
		strict         bool
		expectedResult display.DisplayableFrame
		expectedError  error
	}

	tests := []testcase{
		{
			name: "invalid max aperture",
			frame: records.EFRM{
				RollYear:            math.MaxUint16,
				RollMonth:           math.MaxUint8,
				RollDay:             math.MaxUint8,
				RollHour:            math.MaxUint8,
				RollMinute:          math.MaxUint8,
				RollSecond:          math.MaxUint8,
				BatteryYear:         math.MaxUint16,
				BatteryMonth:        math.MaxUint8,
				BatteryDay:          math.MaxUint8,
				BatteryHour:         math.MaxUint8,
				BatteryMinute:       math.MaxUint8,
				BatterySecond:       math.MaxUint8,
				Year:                math.MaxUint16,
				Month:               math.MaxUint8,
				Day:                 math.MaxUint8,
				Hour:                math.MaxUint8,
				Minute:              math.MaxUint8,
				Second:              math.MaxUint8,
				MaxAperture:         10,
				Tv:                  -1,
				Av:                  0,
				ExposureCompenation: 0,
				MultipleExposure:    0,
				FlashMode:           0,
				MeteringMode:        0,
				ShootingMode:        0,
				FilmAdvanceMode:     10,
				AFMode:              0,
			},
			strict:        true,
			expectedError: domain.ErrInvalidAv,
		},
		{
			name: "invalid shutter speed",
			frame: records.EFRM{
				RollYear:            math.MaxUint16,
				RollMonth:           math.MaxUint8,
				RollDay:             math.MaxUint8,
				RollHour:            math.MaxUint8,
				RollMinute:          math.MaxUint8,
				RollSecond:          math.MaxUint8,
				BatteryYear:         math.MaxUint16,
				BatteryMonth:        math.MaxUint8,
				BatteryDay:          math.MaxUint8,
				BatteryHour:         math.MaxUint8,
				BatteryMinute:       math.MaxUint8,
				BatterySecond:       math.MaxUint8,
				Year:                math.MaxUint16,
				Month:               math.MaxUint8,
				Day:                 math.MaxUint8,
				Hour:                math.MaxUint8,
				Minute:              math.MaxUint8,
				Second:              math.MaxUint8,
				Tv:                  1,
				Av:                  0,
				ExposureCompenation: 0,
				MultipleExposure:    0,
				FlashMode:           0,
				MeteringMode:        0,
				ShootingMode:        0,
				FilmAdvanceMode:     10,
				AFMode:              0,
			},
			strict:        true,
			expectedError: domain.ErrInvalidTv,
		},
		{
			name: "invalid bulb shutter speed",
			frame: records.EFRM{
				RollYear:         math.MaxUint16,
				RollMonth:        math.MaxUint8,
				RollDay:          math.MaxUint8,
				RollHour:         math.MaxUint8,
				RollMinute:       math.MaxUint8,
				RollSecond:       math.MaxUint8,
				BatteryYear:      math.MaxUint16,
				BatteryMonth:     math.MaxUint8,
				BatteryDay:       math.MaxUint8,
				BatteryHour:      math.MaxUint8,
				BatteryMinute:    math.MaxUint8,
				BatterySecond:    math.MaxUint8,
				Year:             math.MaxUint16,
				Month:            math.MaxUint8,
				Day:              math.MaxUint8,
				Hour:             math.MaxUint8,
				Minute:           math.MaxUint8,
				Second:           math.MaxUint8,
				Tv:               2130706432, // bulb magic number
				BulbExposureTime: 0,
			},
			expectedError: domain.ErrInvalidBulbTime,
		},
		{
			name: "invalid aperture",
			frame: records.EFRM{
				RollYear:            math.MaxUint16,
				RollMonth:           math.MaxUint8,
				RollDay:             math.MaxUint8,
				RollHour:            math.MaxUint8,
				RollMinute:          math.MaxUint8,
				RollSecond:          math.MaxUint8,
				BatteryYear:         math.MaxUint16,
				BatteryMonth:        math.MaxUint8,
				BatteryDay:          math.MaxUint8,
				BatteryHour:         math.MaxUint8,
				BatteryMinute:       math.MaxUint8,
				BatterySecond:       math.MaxUint8,
				Year:                math.MaxUint16,
				Month:               math.MaxUint8,
				Day:                 math.MaxUint8,
				Hour:                math.MaxUint8,
				Minute:              math.MaxUint8,
				Second:              math.MaxUint8,
				Tv:                  -1,
				Av:                  10,
				ExposureCompenation: 0,
				MultipleExposure:    0,
				FlashMode:           0,
				MeteringMode:        0,
				ShootingMode:        0,
				FilmAdvanceMode:     10,
				AFMode:              0,
			},
			strict:        true,
			expectedError: domain.ErrInvalidAv,
		},
		{
			name: "invalid exposure compensation",
			frame: records.EFRM{
				RollYear:            math.MaxUint16,
				RollMonth:           math.MaxUint8,
				RollDay:             math.MaxUint8,
				RollHour:            math.MaxUint8,
				RollMinute:          math.MaxUint8,
				RollSecond:          math.MaxUint8,
				BatteryYear:         math.MaxUint16,
				BatteryMonth:        math.MaxUint8,
				BatteryDay:          math.MaxUint8,
				BatteryHour:         math.MaxUint8,
				BatteryMinute:       math.MaxUint8,
				BatterySecond:       math.MaxUint8,
				Year:                math.MaxUint16,
				Month:               math.MaxUint8,
				Day:                 math.MaxUint8,
				Hour:                math.MaxUint8,
				Minute:              math.MaxUint8,
				Second:              math.MaxUint8,
				Tv:                  -1,
				Av:                  0,
				ExposureCompenation: 1,
				MultipleExposure:    0,
				FlashMode:           0,
				MeteringMode:        0,
				ShootingMode:        0,
				FilmAdvanceMode:     10,
				AFMode:              0,
			},
			strict:        true,
			expectedError: domain.ErrUnknownExposureComp,
		},
		{
			name: "invalid multiple exposure value",
			frame: records.EFRM{
				RollYear:            math.MaxUint16,
				RollMonth:           math.MaxUint8,
				RollDay:             math.MaxUint8,
				RollHour:            math.MaxUint8,
				RollMinute:          math.MaxUint8,
				RollSecond:          math.MaxUint8,
				BatteryYear:         math.MaxUint16,
				BatteryMonth:        math.MaxUint8,
				BatteryDay:          math.MaxUint8,
				BatteryHour:         math.MaxUint8,
				BatteryMinute:       math.MaxUint8,
				BatterySecond:       math.MaxUint8,
				Year:                math.MaxUint16,
				Month:               math.MaxUint8,
				Day:                 math.MaxUint8,
				Hour:                math.MaxUint8,
				Minute:              math.MaxUint8,
				Second:              math.MaxUint8,
				Tv:                  -1,
				Av:                  0,
				ExposureCompenation: 0,
				MultipleExposure:    2,
			},
			expectedError: domain.ErrUnknownMultipleExposure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			frameBuilder := display.NewFrameBuilder(newTestLogger())

			result, err := frameBuilder.
				WithFrameMetadata(ctx, tt.frame).
				WithExposureSettings(ctx, tt.strict).
				WithCameraModesAndFlashInfo(ctx, false).
				WithCustomFunctionsAndFocusPoints(ctx, false).
				WithThumbnail(ctx, nil).
				Build()

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedError)
				}

				if !errors.Is(err, tt.expectedError) {
					t.Fatalf("expected error %v, got %v", tt.expectedError, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expectedResult {
				t.Fatalf("expected result %v, got %v",
					tt.expectedResult, result)
			}
		})
	}
}

//nolint:exhaustruct // only partial is needed
func Test_FrameBuilder_WithCameraModesAndFlashInfo(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		frame          records.EFRM
		strict         bool
		expectedResult display.DisplayableFrame
		expectedError  error
	}

	tests := []testcase{
		{
			name: "invalid flash exposure compensation",
			frame: records.EFRM{
				RollYear:                  math.MaxUint16,
				RollMonth:                 math.MaxUint8,
				RollDay:                   math.MaxUint8,
				RollHour:                  math.MaxUint8,
				RollMinute:                math.MaxUint8,
				RollSecond:                math.MaxUint8,
				BatteryYear:               math.MaxUint16,
				BatteryMonth:              math.MaxUint8,
				BatteryDay:                math.MaxUint8,
				BatteryHour:               math.MaxUint8,
				BatteryMinute:             math.MaxUint8,
				BatterySecond:             math.MaxUint8,
				Year:                      math.MaxUint16,
				Month:                     math.MaxUint8,
				Day:                       math.MaxUint8,
				Hour:                      math.MaxUint8,
				Minute:                    math.MaxUint8,
				Second:                    math.MaxUint8,
				Tv:                        -1,
				Av:                        0,
				ExposureCompenation:       0,
				MultipleExposure:          0,
				FlashExposureCompensation: 1,
				FlashMode:                 0,
				MeteringMode:              0,
				ShootingMode:              0,
				FilmAdvanceMode:           10,
				AFMode:                    0,
			},
			strict:        true,
			expectedError: domain.ErrUnknownExposureComp,
		},
		{
			name: "invalid flash exposure compensation",
			frame: records.EFRM{
				RollYear:                  math.MaxUint16,
				RollMonth:                 math.MaxUint8,
				RollDay:                   math.MaxUint8,
				RollHour:                  math.MaxUint8,
				RollMinute:                math.MaxUint8,
				RollSecond:                math.MaxUint8,
				BatteryYear:               math.MaxUint16,
				BatteryMonth:              math.MaxUint8,
				BatteryDay:                math.MaxUint8,
				BatteryHour:               math.MaxUint8,
				BatteryMinute:             math.MaxUint8,
				BatterySecond:             math.MaxUint8,
				Year:                      math.MaxUint16,
				Month:                     math.MaxUint8,
				Day:                       math.MaxUint8,
				Hour:                      math.MaxUint8,
				Minute:                    math.MaxUint8,
				Second:                    math.MaxUint8,
				Tv:                        -1,
				Av:                        0,
				ExposureCompenation:       0,
				MultipleExposure:          0,
				FlashExposureCompensation: 0,
				FlashMode:                 2,
			},
			expectedError: domain.ErrUnknownFlashMode,
		},
		{
			name: "invalid metering mode",
			frame: records.EFRM{
				RollYear:                  math.MaxUint16,
				RollMonth:                 math.MaxUint8,
				RollDay:                   math.MaxUint8,
				RollHour:                  math.MaxUint8,
				RollMinute:                math.MaxUint8,
				RollSecond:                math.MaxUint8,
				BatteryYear:               math.MaxUint16,
				BatteryMonth:              math.MaxUint8,
				BatteryDay:                math.MaxUint8,
				BatteryHour:               math.MaxUint8,
				BatteryMinute:             math.MaxUint8,
				BatterySecond:             math.MaxUint8,
				Year:                      math.MaxUint16,
				Month:                     math.MaxUint8,
				Day:                       math.MaxUint8,
				Hour:                      math.MaxUint8,
				Minute:                    math.MaxUint8,
				Second:                    math.MaxUint8,
				Tv:                        -1,
				Av:                        0,
				ExposureCompenation:       0,
				MultipleExposure:          0,
				FlashExposureCompensation: 0,
				FlashMode:                 0,
				MeteringMode:              4,
			},
			expectedError: domain.ErrUnknownMeteringMode,
		},
		{
			name: "invalid shooting mode",
			frame: records.EFRM{
				RollYear:                  math.MaxUint16,
				RollMonth:                 math.MaxUint8,
				RollDay:                   math.MaxUint8,
				RollHour:                  math.MaxUint8,
				RollMinute:                math.MaxUint8,
				RollSecond:                math.MaxUint8,
				BatteryYear:               math.MaxUint16,
				BatteryMonth:              math.MaxUint8,
				BatteryDay:                math.MaxUint8,
				BatteryHour:               math.MaxUint8,
				BatteryMinute:             math.MaxUint8,
				BatterySecond:             math.MaxUint8,
				Year:                      math.MaxUint16,
				Month:                     math.MaxUint8,
				Day:                       math.MaxUint8,
				Hour:                      math.MaxUint8,
				Minute:                    math.MaxUint8,
				Second:                    math.MaxUint8,
				Tv:                        -1,
				Av:                        0,
				ExposureCompenation:       0,
				MultipleExposure:          0,
				FlashExposureCompensation: 0,
				FlashMode:                 0,
				MeteringMode:              0,
				ShootingMode:              6,
			},
			expectedError: domain.ErrUnknownShootingMode,
		},
		{
			name: "invalid film advance mode",
			frame: records.EFRM{
				RollYear:                  math.MaxUint16,
				RollMonth:                 math.MaxUint8,
				RollDay:                   math.MaxUint8,
				RollHour:                  math.MaxUint8,
				RollMinute:                math.MaxUint8,
				RollSecond:                math.MaxUint8,
				BatteryYear:               math.MaxUint16,
				BatteryMonth:              math.MaxUint8,
				BatteryDay:                math.MaxUint8,
				BatteryHour:               math.MaxUint8,
				BatteryMinute:             math.MaxUint8,
				BatterySecond:             math.MaxUint8,
				Year:                      math.MaxUint16,
				Month:                     math.MaxUint8,
				Day:                       math.MaxUint8,
				Hour:                      math.MaxUint8,
				Minute:                    math.MaxUint8,
				Second:                    math.MaxUint8,
				Tv:                        -1,
				Av:                        0,
				ExposureCompenation:       0,
				MultipleExposure:          0,
				FlashExposureCompensation: 0,
				FlashMode:                 0,
				MeteringMode:              0,
				ShootingMode:              0,
				FilmAdvanceMode:           0,
			},
			expectedError: domain.ErrUnknownFilmAdvanceMode,
		},
		{
			name: "invalid auto focus mode",
			frame: records.EFRM{
				RollYear:                  math.MaxUint16,
				RollMonth:                 math.MaxUint8,
				RollDay:                   math.MaxUint8,
				RollHour:                  math.MaxUint8,
				RollMinute:                math.MaxUint8,
				RollSecond:                math.MaxUint8,
				BatteryYear:               math.MaxUint16,
				BatteryMonth:              math.MaxUint8,
				BatteryDay:                math.MaxUint8,
				BatteryHour:               math.MaxUint8,
				BatteryMinute:             math.MaxUint8,
				BatterySecond:             math.MaxUint8,
				Year:                      math.MaxUint16,
				Month:                     math.MaxUint8,
				Day:                       math.MaxUint8,
				Hour:                      math.MaxUint8,
				Minute:                    math.MaxUint8,
				Second:                    math.MaxUint8,
				Tv:                        -1,
				Av:                        0,
				ExposureCompenation:       0,
				MultipleExposure:          0,
				FlashExposureCompensation: 0,
				FlashMode:                 0,
				MeteringMode:              0,
				ShootingMode:              0,
				FilmAdvanceMode:           10,
				AFMode:                    3,
			},
			expectedError: domain.ErrUnknownAutoFocusMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			frameBuilder := display.NewFrameBuilder(newTestLogger())

			result, err := frameBuilder.
				WithFrameMetadata(ctx, tt.frame).
				WithExposureSettings(ctx, false).
				WithCameraModesAndFlashInfo(ctx, tt.strict).
				WithCustomFunctionsAndFocusPoints(ctx, false).
				WithThumbnail(ctx, nil).
				Build()

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedError)
				}

				if !errors.Is(err, tt.expectedError) {
					t.Fatalf("expected error %v, got %v", tt.expectedError, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expectedResult {
				t.Fatalf("expected result %v, got %v",
					tt.expectedResult, result)
			}
		})
	}
}

//nolint:exhaustruct // only partial is needed
func Test_FrameBuilder_CustomFunctionsAndSuccess(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		frame          records.EFRM
		strict         bool
		expectedResult display.DisplayableFrame
		expectedError  error
	}

	tests := []testcase{
		{
			name: "invalid custom functions",
			frame: records.EFRM{
				RollYear:                  math.MaxUint16,
				RollMonth:                 math.MaxUint8,
				RollDay:                   math.MaxUint8,
				RollHour:                  math.MaxUint8,
				RollMinute:                math.MaxUint8,
				RollSecond:                math.MaxUint8,
				BatteryYear:               math.MaxUint16,
				BatteryMonth:              math.MaxUint8,
				BatteryDay:                math.MaxUint8,
				BatteryHour:               math.MaxUint8,
				BatteryMinute:             math.MaxUint8,
				BatterySecond:             math.MaxUint8,
				Year:                      math.MaxUint16,
				Month:                     math.MaxUint8,
				Day:                       math.MaxUint8,
				Hour:                      math.MaxUint8,
				Minute:                    math.MaxUint8,
				Second:                    math.MaxUint8,
				Tv:                        -1,
				Av:                        0,
				ExposureCompenation:       0,
				MultipleExposure:          0,
				FlashExposureCompensation: 0,
				FlashMode:                 0,
				MeteringMode:              0,
				ShootingMode:              0,
				FilmAdvanceMode:           10,
				AFMode:                    0,
				CustomFunction0:           math.MaxUint8 - 1,
			},
			strict:        true,
			expectedError: display.ErrInvalidCustomFunction,
		},
		{
			name: "successful build",
			frame: records.EFRM{
				CodeA:                     math.MaxUint32,
				CodeB:                     math.MaxUint32,
				FocalLength:               math.MaxUint32,
				RollYear:                  math.MaxUint16,
				RollMonth:                 math.MaxUint8,
				RollDay:                   math.MaxUint8,
				RollHour:                  math.MaxUint8,
				RollMinute:                math.MaxUint8,
				RollSecond:                math.MaxUint8,
				BatteryYear:               math.MaxUint16,
				BatteryMonth:              math.MaxUint8,
				BatteryDay:                math.MaxUint8,
				BatteryHour:               math.MaxUint8,
				BatteryMinute:             math.MaxUint8,
				BatterySecond:             math.MaxUint8,
				Year:                      math.MaxUint16,
				Month:                     math.MaxUint8,
				Day:                       math.MaxUint8,
				Hour:                      math.MaxUint8,
				Minute:                    math.MaxUint8,
				Second:                    math.MaxUint8,
				Tv:                        -1,
				IsoM:                      math.MaxUint32,
				IsoDX:                     math.MaxUint32,
				MaxAperture:               math.MaxUint32,
				Av:                        math.MaxUint32,
				ExposureCompenation:       -1,
				MultipleExposure:          99,
				FlashExposureCompensation: -1,
				FlashMode:                 99,
				MeteringMode:              99,
				ShootingMode:              99,
				FilmAdvanceMode:           99,
				AFMode:                    99,
				CustomFunction0:           math.MaxUint8,
				CustomFunction1:           1,
				CustomFunction2:           math.MaxUint8,
				CustomFunction3:           math.MaxUint8,
				CustomFunction4:           math.MaxUint8,
				CustomFunction5:           math.MaxUint8,
				CustomFunction6:           math.MaxUint8,
				CustomFunction7:           math.MaxUint8,
				CustomFunction8:           math.MaxUint8,
				CustomFunction9:           math.MaxUint8,
				CustomFunction10:          math.MaxUint8,
				CustomFunction11:          math.MaxUint8,
				CustomFunction12:          math.MaxUint8,
				CustomFunction13:          math.MaxUint8,
				CustomFunction14:          math.MaxUint8,
				CustomFunction15:          math.MaxUint8,
				CustomFunction16:          math.MaxUint8,
				CustomFunction17:          math.MaxUint8,
				CustomFunction18:          math.MaxUint8,
				CustomFunction19:          math.MaxUint8,
			},
			expectedResult: display.DisplayableFrame{
				CustomFunctions: display.DisplayableCustomFunctions{
					0:  " ",
					1:  "1",
					2:  " ",
					3:  " ",
					4:  " ",
					5:  " ",
					6:  " ",
					7:  " ",
					8:  " ",
					9:  " ",
					10: " ",
					11: " ",
					12: " ",
					13: " ",
					14: " ",
					15: " ",
					16: " ",
					17: " ",
					18: " ",
					19: " ",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			frameBuilder := display.NewFrameBuilder(newTestLogger())

			result, err := frameBuilder.
				WithFrameMetadata(ctx, tt.frame).
				WithExposureSettings(ctx, false).
				WithCameraModesAndFlashInfo(ctx, false).
				WithCustomFunctionsAndFocusPoints(ctx, tt.strict).
				WithThumbnail(ctx, nil).
				Build()

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedError)
				}

				if !errors.Is(err, tt.expectedError) {
					t.Fatalf("expected error %v, got %v", tt.expectedError, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expectedResult {
				t.Fatalf("expected result %v, got %v",
					tt.expectedResult, result)
			}
		})
	}
}
