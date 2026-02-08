package domain_test

import (
	"testing"

	"github.com/ma-tf/meta1v/pkg/domain"
)

func Test_MapProvider(t *testing.T) {
	t.Parallel()

	provider := domain.NewMapProvider()

	type testCase struct {
		name           string
		fut            func() (string, bool)
		expectedResult string
	}

	tests := []testCase{
		{
			name: "shutter speeds loaded",
			fut: func() (string, bool) {
				v, ok := provider.GetTv(100)

				return string(v), ok
			},
			expectedResult: `1"`,
		},
		{
			name: "aperture values loaded",
			fut: func() (string, bool) {
				v, ok := provider.GetAv(100)

				return string(v), ok
			},
			expectedResult: "1.0",
		},
		{
			name: "exposure compensations loaded",
			fut: func() (string, bool) {
				v, ok := provider.GetExposureCompensation(100)

				return string(v), ok
			},
			expectedResult: "+1.0",
		},
		{
			name: "flash modes loaded",
			fut: func() (string, bool) {
				v, ok := provider.GetFlashMode(1)

				return string(v), ok
			},
			expectedResult: "ON",
		},
		{
			name: "metering modes loaded",
			fut: func() (string, bool) {
				v, ok := provider.GetMeteringMode(0)

				return string(v), ok
			},
			expectedResult: "Evaluative",
		},
		{
			name: "shooting modes loaded",
			fut: func() (string, bool) {
				v, ok := provider.GetShootingMode(1)

				return string(v), ok
			},
			expectedResult: "Program AE",
		},
		{
			name: "film advance modes loaded",
			fut: func() (string, bool) {
				v, ok := provider.GetFilmAdvanceMode(10)

				return string(v), ok
			},
			expectedResult: "Single-frame",
		},
		{
			name: "auto focus modes loaded",
			fut: func() (string, bool) {
				v, ok := provider.GetAutoFocusMode(1)

				return string(v), ok
			},
			expectedResult: "One-Shot AF",
		},
		{
			name: "multiple exposures loaded",
			fut: func() (string, bool) {
				v, ok := provider.GetMultipleExposure(1)

				return string(v), ok
			},
			expectedResult: "ON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			val, ok := tt.fut()
			if !ok {
				t.Errorf("%s: expected value to exist", tt.name)
			}

			if val != tt.expectedResult {
				t.Errorf("%s: got %q, want %q", tt.name, val, tt.expectedResult)
			}
		})
	}
}
