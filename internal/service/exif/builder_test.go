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

package exif_test

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/ma-tf/meta1v/internal/records"
	"github.com/ma-tf/meta1v/internal/service/exif"
)

//nolint:exhaustruct // only partial is needed
func Test_Build(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name             string
		frame            records.EFRM
		strict           bool
		expectedMetadata map[string]string
		expectedError    error
	}

	validFrame := func() records.EFRM {
		return records.EFRM{
			Remarks: [256]byte{
				'r', 'e', 'm', 'a', 'r', 'k', 's',
			},
			Year:                      2023,
			Month:                     5,
			Day:                       15,
			Hour:                      14,
			Minute:                    30,
			Second:                    45,
			BatteryYear:               2023,
			BatteryMonth:              5,
			BatteryDay:                10,
			BatteryHour:               9,
			BatteryMinute:             15,
			BatterySecond:             0,
			RollYear:                  2023,
			RollMonth:                 4,
			RollDay:                   20,
			RollHour:                  16,
			RollMinute:                45,
			RollSecond:                30,
			Av:                        280,
			MaxAperture:               280,
			Tv:                        -12500,
			FocalLength:               70,
			IsoDX:                     100,
			IsoM:                      100,
			ExposureCompensation:      30,
			FlashExposureCompensation: 30,
			FlashMode:                 1,
			MeteringMode:              1,
			ShootingMode:              1,
			AFMode:                    1,
			FilmAdvanceMode:           11,
		}
	}

	emptyFrame := func() records.EFRM {
		return records.EFRM{
			Tv:                        -1,
			ExposureCompensation:      -1,
			FlashExposureCompensation: -1,
			FlashMode:                 99,
			MeteringMode:              99,
			ShootingMode:              99,
			AFMode:                    99,
			FilmAdvanceMode:           99,
			MultipleExposure:          99,
			IsoM:                      math.MaxUint32,
			IsoDX:                     math.MaxUint32,
			FocalLength:               math.MaxUint32,
		}
	}

	tests := []testcase{
		{
			name:   "valid strict frame data",
			frame:  validFrame(),
			strict: true,
			expectedMetadata: map[string]string{
				exif.TagDateTimeOriginal:     "2023-05-15 14:30:45",
				exif.TagExposureCompensation: "+0.3",
				exif.TagFlashExposureComp:    "+0.3",
				exif.TagFlash:                "1",
				exif.TagMeteringMode:         "Center averaging",
				exif.TagUserComment:          "remarks",
				exif.TagAFMode:               "One-Shot AF",
				exif.TagBatteryLoadedDate:    "2023-05-10 09:15:00",
				exif.TagFilmAdvanceMode:      "2-sec. self-timer",
				exif.TagFilmISO:              "100",
				exif.TagFilmLoadedDate:       "2023-04-20 16:45:30",
				exif.TagFlashMode:            "ON",
				exif.TagManualISO:            "100",
				exif.TagMultipleExposure:     "OFF",
				exif.TagShootingMode:         "Program AE",
				exif.TagExposureTime:         "1/125",
				exif.TagFNumber:              "2.8",
				exif.TagFocalLength:          "70",
				exif.TagISO:                  "100",
				exif.TagMaxApertureValue:     "3.0",
			},
		},
		{
			name: "valid bulb exposure time",
			frame: func() records.EFRM {
				f := emptyFrame()
				f.Tv = 2130706432
				f.BulbExposureTime = 100

				return f
			}(),
			strict: true,
			expectedMetadata: map[string]string{
				exif.TagExposureTime: "100",
			},
		},
		{
			name: "valid double quoted exposure time",
			frame: func() records.EFRM {
				f := emptyFrame()
				f.Tv = 3000

				return f
			}(),
			strict: true,
			expectedMetadata: map[string]string{
				exif.TagExposureTime: "30",
			},
		},
		{
			name: "valid manual flash mode",
			frame: func() records.EFRM {
				f := emptyFrame()
				f.FlashMode = 10

				return f
			}(),
			strict: true,
			expectedMetadata: map[string]string{
				exif.TagFlash:     "9",
				exif.TagFlashMode: "Manual flash",
			},
		},
		{
			name: "valid auto flash mode",
			frame: func() records.EFRM {
				f := emptyFrame()
				f.FlashMode = 11

				return f
			}(),
			strict: true,
			expectedMetadata: map[string]string{
				exif.TagFlash:     "25",
				exif.TagFlashMode: "TTL autoflash",
			},
		},
	}

	assertError := func(t *testing.T, expected, got error) {
		t.Helper()

		if expected != nil {
			if got == nil || !errors.Is(got, expected) {
				t.Fatalf("expected error %v, got %v", expected, got)
			}

			return
		}

		if got != nil {
			t.Fatalf("unexpected error: %v", got)
		}
	}

	assertResult := func(t *testing.T, expected, got map[string]string) {
		t.Helper()

		if expected != nil {
			if !reflect.DeepEqual(expected, got) {
				t.Fatalf("expected metadata %v, got %v", expected, got)
			}

			return
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := exif.NewExifBuilder(newTestLogger())

			metadata, err := b.Build(tt.frame, tt.strict)

			assertError(t, tt.expectedError, err)
			assertResult(t, tt.expectedMetadata, metadata)
		})
	}
}

//nolint:exhaustruct // only partial is needed
func Test_Build_Fail(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name             string
		frame            records.EFRM
		strict           bool
		expectedMetadata map[string]string
		expectedError    error
	}

	tests := []testcase{
		{
			name: "invalid capture date",
			frame: records.EFRM{
				Year:   2023,
				Month:  13,
				Day:    32,
				Hour:   25,
				Minute: 61,
				Second: 61,
			},
			strict:        true,
			expectedError: exif.ErrInvalidCaptureDate,
		},
		{
			name: "invalid battery loaded date",
			frame: records.EFRM{
				BatteryYear:   2023,
				BatteryMonth:  13,
				BatteryDay:    32,
				BatteryHour:   25,
				BatteryMinute: 61,
				BatterySecond: 61,
			},
			strict:        true,
			expectedError: exif.ErrInvalidBatteryLoadedDate,
		},
		{
			name: "invalid roll loaded date",
			frame: records.EFRM{
				RollYear:   2023,
				RollMonth:  13,
				RollDay:    32,
				RollHour:   25,
				RollMinute: 61,
				RollSecond: 61,
			},
			strict:        true,
			expectedError: exif.ErrInvalidFilmLoadedDate,
		},
		{
			name: "invalid aperture value",
			frame: records.EFRM{
				Av: 12345,
			},
			strict:        true,
			expectedError: exif.ErrParseApertureValue,
		},
		{
			name: "invalid max aperture value",
			frame: records.EFRM{
				MaxAperture: 12345,
			},
			strict:        true,
			expectedError: exif.ErrParseMaxApertureValue,
		},
		{
			name: "invalid exposure time",
			frame: records.EFRM{
				Tv: 12345,
			},
			strict:        true,
			expectedError: exif.ErrParseExposureTimeValue,
		},
		{
			name: "invalid bulb exposure time",
			frame: records.EFRM{
				Tv:               2130706432,
				BulbExposureTime: 0,
			},
			strict:        true,
			expectedError: exif.ErrParseBulbExposureTime,
		},
		{
			name: "invalid exposure compensation",
			frame: records.EFRM{
				Tv:                   -1,
				ExposureCompensation: 12345,
			},
			strict:        true,
			expectedError: exif.ErrInvalidExposureCompensation,
		},
		{
			name: "invalid flash exposure compensation",
			frame: records.EFRM{
				Tv:                        -1,
				FlashExposureCompensation: 12345,
			},
			strict:        true,
			expectedError: exif.ErrInvalidFlashExposureComp,
		},
		{
			name: "invalid flash mode",
			frame: records.EFRM{
				Tv:        -1,
				FlashMode: 12345,
			},
			strict:        true,
			expectedError: exif.ErrInvalidFlashMode,
		},
		{
			name: "invalid metering mode",
			frame: records.EFRM{
				Tv:           -1,
				MeteringMode: 12345,
			},
			strict:        true,
			expectedError: exif.ErrInvalidMeteringMode,
		},
		{
			name: "invalid shooting mode",
			frame: records.EFRM{
				Tv:           -1,
				ShootingMode: 12345,
			},
			strict:        true,
			expectedError: exif.ErrInvalidShootingMode,
		},
		{
			name: "invalid auto focus mode",
			frame: records.EFRM{
				Tv:     -1,
				AFMode: 12345,
			},
			strict:        true,
			expectedError: exif.ErrInvalidAutoFocusMode,
		},
		{
			name: "invalid film advance mode",
			frame: records.EFRM{
				Tv:              -1,
				FilmAdvanceMode: 12345,
			},
			strict:        true,
			expectedError: exif.ErrInvalidFilmAdvanceMode,
		},
		{
			name: "invalid multiple exposure",
			frame: records.EFRM{
				Tv:               -1,
				FilmAdvanceMode:  99,
				MultipleExposure: 12345,
			},
			strict:        true,
			expectedError: exif.ErrInvalidMultipleExposure,
		},
	}

	assertError := func(t *testing.T, expected, got error) {
		t.Helper()

		if expected != nil {
			if got == nil || !errors.Is(got, expected) {
				t.Fatalf("expected error %v, got %v", expected, got)
			}

			return
		}

		if got != nil {
			t.Fatalf("unexpected error: %v", got)
		}
	}

	assertResult := func(t *testing.T, expected, got map[string]string) {
		t.Helper()

		if expected != nil {
			if !reflect.DeepEqual(expected, got) {
				t.Fatalf("expected metadata %v, got %v", expected, got)
			}

			return
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := exif.NewExifBuilder(newTestLogger())

			metadata, err := b.Build(tt.frame, tt.strict)

			assertError(t, tt.expectedError, err)
			assertResult(t, tt.expectedMetadata, metadata)
		})
	}
}
