package domain_test

import (
	"errors"
	"math"
	"testing"

	"github.com/ma-tf/meta1v/pkg/domain"
)

//nolint:exhaustruct // only partial needed
func Test_NewFilmID(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		prefix         uint32
		suffix         uint32
		expectedResult domain.FilmID
		expectedError  error
	}

	tests := []testcase{
		{
			name:           "valid film ID",
			prefix:         12,
			suffix:         345,
			expectedResult: "12-345",
		},
		{
			name:           "max values return empty string",
			prefix:         math.MaxUint32,
			suffix:         math.MaxUint32,
			expectedResult: "",
		},
		{
			name:          "prefix out of range",
			prefix:        100,
			suffix:        0,
			expectedError: domain.ErrPrefixOutOfRange,
		},
		{
			name:          "suffix out of range",
			prefix:        0,
			suffix:        1000,
			expectedError: domain.ErrSuffixOutOfRange,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			id, err := domain.NewFilmID(tc.prefix, tc.suffix)

			if id != tc.expectedResult {
				t.Errorf("expected ID %q, got %q", tc.expectedResult, id)
			}

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

//nolint:exhaustruct // only partial needed
func Test_NewFirstRow(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		firstRow       uint8
		perRow         uint8
		expectedResult domain.FirstRow
		expectedError  error
	}

	tests := []testcase{
		{
			name:           "valid first row",
			firstRow:       2,
			perRow:         36,
			expectedResult: "34",
		},
		{
			name:          "first row greater than per row",
			firstRow:      40,
			perRow:        36,
			expectedError: domain.ErrFirstRowGreaterThanPerRow,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fr, err := domain.NewFirstRow(tc.firstRow, tc.perRow)

			if fr != tc.expectedResult {
				t.Errorf("expected FirstRow %q, got %q", tc.expectedResult, fr)
			}

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

func Test_NewPerRow(t *testing.T) {
	t.Parallel()

	perRow := uint8(36)
	expected := domain.PerRow("36")

	result := domain.NewPerRow(perRow)

	if result != expected {
		t.Errorf("expected PerRow %q, got %q", expected, result)
	}
}

func Test_NewFrameCount(t *testing.T) {
	t.Parallel()

	fc := uint32(24)
	expected := domain.FrameCount("24")

	result := domain.NewFrameCount(fc)

	if result != expected {
		t.Errorf("expected FrameCount %q, got %q", expected, result)
	}
}

//nolint:exhaustruct // only partial needed
func Test_NewDateTime(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		year           uint16
		month          uint8
		day            uint8
		hour           uint8
		minute         uint8
		second         uint8
		expectedResult domain.ValidatedDatetime
		expectedError  error
	}

	tests := []testcase{
		{
			name:           "valid datetime",
			year:           2023,
			month:          10,
			day:            5,
			hour:           14,
			minute:         30,
			second:         0,
			expectedResult: "2023-10-05 14:30:00",
		},
		{
			name:           "max values return empty string",
			year:           math.MaxUint16,
			month:          math.MaxUint8,
			day:            math.MaxUint8,
			hour:           math.MaxUint8,
			minute:         math.MaxUint8,
			second:         math.MaxUint8,
			expectedResult: "",
		},
		{
			name:          "invalid datetime format",
			year:          2023,
			month:         13,
			day:           32,
			hour:          25,
			minute:        61,
			second:        61,
			expectedError: domain.ErrInvalidDateTime,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dt, err := domain.NewDateTime(
				tc.year, tc.month, tc.day, tc.hour, tc.minute, tc.second)

			if dt != tc.expectedResult {
				t.Errorf("expected Datetime %q, got %q", tc.expectedResult, dt)
			}

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

func Test_NewTitle(t *testing.T) {
	t.Parallel()

	var rawTitle [64]byte
	copy(rawTitle[:], "Sample Title")

	expected := domain.Title("Sample Title")
	result := domain.NewTitle(rawTitle)

	if result != expected {
		t.Errorf("expected Title %q, got %q", expected, result)
	}
}

func Test_NewRemarks(t *testing.T) {
	t.Parallel()

	var rawRemarks [256]byte
	copy(rawRemarks[:], "These are sample remarks for testing.")

	expected := domain.Remarks("These are sample remarks for testing.")
	result := domain.NewRemarks(rawRemarks)

	if result != expected {
		t.Errorf("expected Remarks %q, got %q", expected, result)
	}
}

func Test_NewFocalLength(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		focalLength    uint32
		expectedResult domain.FocalLength
	}

	tests := []testcase{
		{
			name:           "valid focal length",
			focalLength:    50,
			expectedResult: "50",
		},
		{
			name:           "max value returns empty string",
			focalLength:    math.MaxUint32,
			expectedResult: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fl := domain.NewFocalLength(tc.focalLength)

			if fl != tc.expectedResult {
				t.Errorf("expected FocalLength %q, got %q",
					tc.expectedResult, fl)
			}
		})
	}
}

//nolint:exhaustruct // only partial needed
func Test_NewTv(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		tv             int32
		strict         bool
		expectedResult domain.Tv
		expectedError  error
	}

	tests := []testcase{
		{
			name:           "valid Tv value",
			tv:             -10000,
			strict:         true,
			expectedResult: "1/100",
		},
		{
			name:          "invalid Tv value with strict mode",
			tv:            9999,
			strict:        true,
			expectedError: domain.ErrInvalidTv,
		},
		{
			name:           "invalid Tv value without strict mode",
			tv:             9999,
			strict:         false,
			expectedResult: "9999",
		},
		{
			name:           "-1 returns empty string",
			tv:             -1,
			strict:         true,
			expectedResult: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tv, err := domain.NewTv(tc.tv, tc.strict)

			if tv != tc.expectedResult {
				t.Errorf("expected Tv %q, got %q", tc.expectedResult, tv)
			}

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

//nolint:exhaustruct // only partial needed
func Test_NewAv(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		aperture       uint32
		strict         bool
		expectedResult domain.Av
		expectedError  error
	}

	tests := []testcase{
		{
			name:           "valid Av value",
			aperture:       280,
			strict:         true,
			expectedResult: "2.8",
		},
		{
			name:          "invalid Av value with strict mode",
			aperture:      999,
			strict:        true,
			expectedError: domain.ErrInvalidAv,
		},
		{
			name:           "invalid Av value without strict mode",
			aperture:       999,
			strict:         false,
			expectedResult: "999",
		},
		{
			name:           "max value returns empty string",
			aperture:       math.MaxUint32,
			strict:         true,
			expectedResult: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			av, err := domain.NewAv(tc.aperture, tc.strict)

			if av != tc.expectedResult {
				t.Errorf("expected Av %q, got %q", tc.expectedResult, av)
			}

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

func Test_NewIso(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		iso            uint32
		expectedResult domain.Iso
	}

	tests := []testcase{
		{
			name:           "valid ISO value",
			iso:            200,
			expectedResult: "200",
		},
		{
			name:           "max value returns empty string",
			iso:            math.MaxUint32,
			expectedResult: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			iso := domain.NewIso(tc.iso)

			if iso != tc.expectedResult {
				t.Errorf("expected Iso %q, got %q", tc.expectedResult, iso)
			}
		})
	}
}

//nolint:exhaustruct // only partial needed
func Test_NewExposureCompenation(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		ec             int32
		strict         bool
		expectedResult domain.ExposureCompenation
		expectedError  error
	}

	tests := []testcase{
		{
			name:           "valid Exposure Compensation value",
			ec:             230,
			strict:         true,
			expectedResult: "+2.3",
		},
		{
			name:          "invalid Exposure Compensation value with strict mode",
			ec:            99,
			strict:        true,
			expectedError: domain.ErrUnknownExposureComp,
		},
		{
			name:           "invalid positive Exposure Compensation value without strict mode",
			ec:             99,
			strict:         false,
			expectedResult: "+9.9",
		},
		{
			name:           "invalid negative Exposure Compensation value without strict mode",
			ec:             -99,
			strict:         false,
			expectedResult: "-9.9",
		},
		{
			name:           "-1 returns empty string",
			ec:             -1,
			strict:         true,
			expectedResult: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ec, err := domain.NewExposureCompensation(tc.ec, tc.strict)

			if ec != tc.expectedResult {
				t.Errorf("expected ExposureCompenation %q, got %q",
					tc.expectedResult, ec)
			}

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

//nolint:exhaustruct // only partial needed
func Test_NewFlashMode(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		fm             uint32
		expectedResult domain.FlashMode
		expectedError  error
	}

	tests := []testcase{
		{
			name:           "valid Flash Mode value",
			fm:             1,
			expectedResult: "ON",
		},
		{
			name:          "invalid Flash Mode value",
			fm:            999,
			expectedError: domain.ErrUnknownFlashMode,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fm, err := domain.NewFlashMode(tc.fm)

			if fm != tc.expectedResult {
				t.Errorf("expected FlashMode %q, got %q", tc.expectedResult, fm)
			}

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

//nolint:exhaustruct // only partial needed
func Test_NewMeteringMode(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		mm             uint32
		expectedResult domain.MeteringMode
		expectedError  error
	}

	tests := []testcase{
		{
			name:           "valid Metering Mode value",
			mm:             0,
			expectedResult: "Evaluative",
		},
		{
			name:          "invalid Metering Mode value",
			mm:            999,
			expectedError: domain.ErrUnknownMeteringMode,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mm, err := domain.NewMeteringMode(tc.mm)

			if mm != tc.expectedResult {
				t.Errorf("expected MeteringMode %q, got %q",
					tc.expectedResult, mm)
			}

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

//nolint:exhaustruct // only partial needed
func Test_NewShootingMode(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		sm             uint32
		expectedResult domain.ShootingMode
		expectedError  error
	}

	tests := []testcase{
		{
			name:           "valid Shooting Mode value",
			sm:             1,
			expectedResult: "Program AE",
		},
		{
			name:          "invalid Shooting Mode value",
			sm:            999,
			expectedError: domain.ErrUnknownShootingMode,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sm, err := domain.NewShootingMode(tc.sm)

			if sm != tc.expectedResult {
				t.Errorf("expected ShootingMode %q, got %q",
					tc.expectedResult, sm)
			}

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

//nolint:exhaustruct // only partial needed
func Test_NewFilmAdvanceMode(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		fam            uint32
		expectedResult domain.FilmAdvanceMode
		expectedError  error
	}

	tests := []testcase{
		{
			name:           "valid Film Advance Mode value",
			fam:            10,
			expectedResult: "Single-frame",
		},
		{
			name:          "invalid Film Advance Mode value",
			fam:           999,
			expectedError: domain.ErrUnknownFilmAdvanceMode,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fam, err := domain.NewFilmAdvanceMode(tc.fam)

			if fam != tc.expectedResult {
				t.Errorf("expected FilmAdvanceMode %q, got %q",
					tc.expectedResult, fam)
			}

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v",
					tc.expectedError, err)
			}
		})
	}
}

//nolint:exhaustruct // only partial needed
func Test_NewAutoFocusMode(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		afm            uint32
		expectedResult domain.AutoFocusMode
		expectedError  error
	}

	tests := []testcase{
		{
			name:           "valid Auto Focus Mode value",
			afm:            1,
			expectedResult: "One-Shot AF",
		},
		{
			name:          "invalid Auto Focus Mode value",
			afm:           999,
			expectedError: domain.ErrUnknownAutoFocusMode,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			afm, err := domain.NewAutoFocusMode(tc.afm)

			if afm != tc.expectedResult {
				t.Errorf("expected AutoFocusMode %q, got %q",
					tc.expectedResult, afm)
			}

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v",
					tc.expectedError, err)
			}
		})
	}
}

//nolint:exhaustruct // only partial needed
func Test_NewBulbExposureTime(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		bd             uint32
		expectedResult domain.BulbExposureTime
		expectedError  error
	}

	tests := []testcase{
		{
			name:           "valid Bulb Exposure Time value",
			bd:             3601,
			expectedResult: "01:00:01",
		},
		{
			name:          "invalid Bulb Exposure Time value",
			bd:            0,
			expectedError: domain.ErrInvalidBulbTime,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			bd, err := domain.NewBulbExposureTime(tc.bd)

			if bd != tc.expectedResult {
				t.Errorf("expected BulbExposureTime %q, got %q",
					tc.expectedResult, bd)
			}

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v",
					tc.expectedError, err)
			}
		})
	}
}

//nolint:exhaustruct // only partial needed
func Test_NewMultipleExposure(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		me             uint32
		expectedResult domain.MultipleExposure
		expectedError  error
	}

	tests := []testcase{
		{
			name:           "valid Multiple Exposure value",
			me:             1,
			expectedResult: "ON",
		},
		{
			name:          "invalid Multiple Exposure value",
			me:            999,
			expectedError: domain.ErrUnknownMultipleExposure,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			me, err := domain.NewMultipleExposure(tc.me)

			if me != tc.expectedResult {
				t.Errorf("expected MultipleExposure %q, got %q",
					tc.expectedResult, me)
			}

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v",
					tc.expectedError, err)
			}
		})
	}
}

//nolint:exhaustruct // only partial needed
func Test_NewCustomFunctions(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		cfs            [20]byte
		strict         bool
		expectedResult domain.CustomFunctions
		expectedError  error
	}

	tests := []testcase{
		{
			name: "valid Custom Functions",
			cfs: [20]byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
			strict: true,
			expectedResult: domain.CustomFunctions{
				"0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
				"0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
			},
		},
		{
			name: "CustomFunctions with max value set to space",
			cfs: [20]byte{
				math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8,
				math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8,
				math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8,
				math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8,
				math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8,
			},
			strict: true,
			expectedResult: domain.CustomFunctions{
				" ", " ", " ", " ", " ", " ", " ", " ", " ", " ",
				" ", " ", " ", " ", " ", " ", " ", " ", " ", " ",
			},
		},
		{
			name: "invalid CustomFunctions with strict mode",
			cfs: [20]byte{
				254, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
			strict:        true,
			expectedError: domain.ErrInvalidCustomFunction,
		},
		{
			name: "invalid CustomFunctions without strict mode",
			cfs: [20]byte{
				254, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
			strict: false,
			expectedResult: domain.CustomFunctions{
				"254", "0", "0", "0", "0", "0", "0", "0", "0", "0",
				"0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfs, err := domain.NewCustomFunctions(tc.cfs, tc.strict)

			if cfs != tc.expectedResult {
				t.Errorf("expected CustomFunctions %v, got %v",
					tc.expectedResult, cfs)
			}

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v",
					tc.expectedError, err)
			}
		})
	}
}
