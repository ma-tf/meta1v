package efd_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"testing"

	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/ma-tf/meta1v/pkg/records"
	records_test "github.com/ma-tf/meta1v/pkg/records/mocks"
	"go.uber.org/mock/gomock"
)

//nolint:exhaustruct // only partial is needed
func Test_Parser_ParseRaw(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		file           []byte
		expectedError  error
		expectedResult records.Raw
	}

	tests := []testcase{
		{
			name:          "error on invalid magic number",
			file:          []byte{0x00, 0x01, 0x02, 0x03},
			expectedError: efd.ErrInvalidRecordMagicNumber,
		},
		{
			name: "error on incomplete record data",
			file: func() []byte {
				buf := &bytes.Buffer{}

				_ = binary.Write(
					buf,
					binary.LittleEndian,
					[8]byte{'E', 'F', 'T', 'P'},
				)
				_ = binary.Write(buf, binary.LittleEndian, uint64(24)) // length
				_ = binary.Write(
					buf,
					binary.LittleEndian,
					[]byte{0x01, 0x02},
				) // incomplete data

				return buf.Bytes()
			}(),
			expectedError: efd.ErrFailedToReadRecord,
		},
		{
			name: "successful parse of valid raw record",
			file: func() []byte {
				buf := &bytes.Buffer{}
				_ = binary.Write(
					buf,
					binary.LittleEndian,
					[8]byte{'E', 'F', 'T', 'P'},
				)
				_ = binary.Write(buf, binary.LittleEndian, uint64(24)) // length
				_ = binary.Write(buf, binary.LittleEndian,
					[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
				)

				return buf.Bytes()
			}(),
			expectedError: nil,
			expectedResult: records.Raw{
				Magic:  [4]byte{'E', 'F', 'T', 'P'},
				Length: 24,
				Data:   []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := t.Context()

			mockThumbnailFactory := records_test.NewMockThumbnailFactory(ctrl)
			parser := efd.NewParser(newTestLogger(), mockThumbnailFactory)

			result, err := parser.ParseRaw(
				ctx,
				bytes.NewReader(tt.file),
			)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v",
					tt.expectedError,
					err,
				)
			}

			if result.Magic != tt.expectedResult.Magic {
				t.Fatalf("expected magic %v, got %v",
					tt.expectedResult.Magic,
					result.Magic,
				)
			}

			if result.Length != tt.expectedResult.Length {
				t.Fatalf("expected length %v, got %v",
					tt.expectedResult.Length,
					result.Length,
				)
			}

			if !bytes.Equal(result.Data, tt.expectedResult.Data) {
				t.Fatalf("expected data %v, got %v",
					tt.expectedResult.Data,
					result.Data,
				)
			}
		})
	}
}

func newEFDF(r records.EFDF) []byte {
	buf := &bytes.Buffer{}
	_ = binary.Write(buf, binary.LittleEndian, r)

	return buf.Bytes()
}

//nolint:exhaustruct // only partial is needed
func Test_Parser_ParseEFDF(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		data           []byte
		expectedError  error
		expectedResult records.EFDF
	}

	tests := []testcase{
		{
			name:          "failed parse of invalid EFDF record",
			data:          []byte{0x00, 0x01, 0x02},
			expectedError: efd.ErrFailedToParseEFDF,
		},
		{
			name: "successful parse of EFDF record",
			data: newEFDF(records.EFDF{
				Title: [64]byte{'t', 'i', 't', 'l', 'e'},
			}),
			expectedError: nil,
			expectedResult: records.EFDF{
				Title: [64]byte{'t', 'i', 't', 'l', 'e'},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := t.Context()

			mockThumbnailFactory := records_test.NewMockThumbnailFactory(ctrl)
			parser := efd.NewParser(newTestLogger(), mockThumbnailFactory)

			result, err := parser.ParseEFDF(
				ctx,
				tt.data,
			)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v",
					tt.expectedError,
					err,
				)
			}

			if result != tt.expectedResult {
				t.Fatalf("expected result %v, got %v",
					tt.expectedResult,
					result,
				)
			}
		})
	}
}

func newEFRM(r records.EFRM) []byte {
	buf := &bytes.Buffer{}
	_ = binary.Write(buf, binary.LittleEndian, r)

	return buf.Bytes()
}

//nolint:exhaustruct // only partial is needed
func Test_Parser_ParseEFRM(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		data           []byte
		expectedError  error
		expectedResult records.EFRM
	}

	tests := []testcase{
		{
			name:          "failed parse of invalid EFRM record",
			data:          []byte{0x00, 0x01, 0x02},
			expectedError: efd.ErrFailedToParseEFRM,
		},
		{
			name: "successful parse of EFRM record",
			data: newEFRM(records.EFRM{
				Remarks: [256]byte{'r', 'e', 'm', 'a', 'r', 'k', 's'},
			}),
			expectedError: nil,
			expectedResult: records.EFRM{
				Remarks: [256]byte{'r', 'e', 'm', 'a', 'r', 'k', 's'},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := t.Context()

			mockThumbnailFactory := records_test.NewMockThumbnailFactory(ctrl)
			parser := efd.NewParser(newTestLogger(), mockThumbnailFactory)

			result, err := parser.ParseEFRM(
				ctx,
				tt.data,
			)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v",
					tt.expectedError,
					err,
				)
			}

			if result != tt.expectedResult {
				t.Fatalf("expected result %v, got %v",
					tt.expectedResult,
					result,
				)
			}
		})
	}
}

//nolint:exhaustruct // only partial is needed
func Test_Parser_ParseEFTP(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name   string
		data   []byte
		expect func(
			mockThumbnailFactory records_test.MockThumbnailFactory,
		)
		expectedError  error
		expectedResult records.EFTP
	}

	exampleThumbnail := image.NewRGBA(
		image.Rect(0, 0, 1, 1),
	)

	tests := []testcase{
		{
			name:          "failed parse of invalid EFTP record",
			data:          []byte{0x00, 0x01, 0x02},
			expectedError: efd.ErrFailedToParseThumbnail,
		},
		{
			name: "failed to parse invalid filepath",
			data: func() []byte {
				buf := &bytes.Buffer{}
				_ = binary.Write(buf, binary.LittleEndian, [16]byte{})
				_ = binary.Write(buf, binary.LittleEndian, []byte{0x01, 0x02})

				return buf.Bytes()
			}(),
			expectedError: efd.ErrFailedToParseThumbnail,
		},
		{
			name: "successful parse of EFTP record",
			data: func() []byte {
				buf := &bytes.Buffer{}
				_ = binary.Write(buf, binary.LittleEndian, [4]byte{})
				_ = binary.Write(buf, binary.LittleEndian, uint16(1)) // width
				_ = binary.Write(buf, binary.LittleEndian, uint16(1)) // height
				_ = binary.Write(buf, binary.LittleEndian, [8]byte{})
				_ = binary.Write(
					buf,
					binary.LittleEndian,
					[256]byte{'f', 'i', 'l', 'e', 'p', 'a', 't', 'h'},
				)
				// _ = binary.Write(buf, binary.LittleEndian, []byte{0x01, 0x02, 0x03})

				return buf.Bytes()
			}(),
			expect: func(
				mockThumbnailFactory records_test.MockThumbnailFactory,
			) {
				mockThumbnailFactory.EXPECT().
					NewRGBA(gomock.Any()).
					Return(exampleThumbnail)
			},
			expectedResult: records.EFTP{
				Width:     1,
				Height:    1,
				Filepath:  [256]byte{'f', 'i', 'l', 'e', 'p', 'a', 't', 'h'},
				Thumbnail: exampleThumbnail,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := t.Context()

			mockThumbnailFactory := records_test.NewMockThumbnailFactory(ctrl)

			if tt.expect != nil {
				tt.expect(*mockThumbnailFactory)
			}

			parser := efd.NewParser(newTestLogger(), mockThumbnailFactory)
			result, err := parser.ParseEFTP(
				ctx,
				tt.data,
			)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v",
					tt.expectedError,
					err,
				)
			}

			if result != tt.expectedResult {
				t.Fatalf("expected result %v, got %v",
					tt.expectedResult,
					result,
				)
			}
		})
	}
}
