package efd_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log/slog"
	"testing"

	"github.com/ma-tf/meta1v/internal/service/efd"
	"github.com/ma-tf/meta1v/pkg/records"
	"go.uber.org/mock/gomock"
)

func buildEFDFRaw(record records.EFDF) records.Raw {
	buf := &bytes.Buffer{}
	_ = binary.Write(buf, binary.LittleEndian, record)
	data := buf.Bytes()
	return records.Raw{
		Magic:  [4]byte{'E', 'F', 'D', 'F'},
		Length: uint64(len(data)),
		Data:   data,
	}
}

//nolint:exhaustruct // for testcase struct literals
func Test_RootBuilder(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}

			return a
		},
	}))

	type testcase struct {
		name           string
		recordsToAdd   []records.Raw
		expectedError  error
		expectedResult records.Root
	}

	tests := []testcase{
		{
			name: "successful build with one EFDF record",
			recordsToAdd: []records.Raw{
				buildEFDFRaw(records.EFDF{Title: [64]byte{'t', 'i', 't', 'l', 'e'}}),
			},
			expectedError: nil,
		},
		{
			name: "error on multiple EFDF records",
			recordsToAdd: []records.Raw{
				buildEFDFRaw(records.EFDF{Title: [64]byte{'t', 'i', 't', 'l', 'e'}}),
				buildEFDFRaw(records.EFDF{Title: [64]byte{'o', 't', 'h', 'e', 'r'}}),
			},
			expectedError: efd.ErrMultipleEFDFRecords,
		},
		// {
		// 	name:          "error on missing EFDF record",
		// 	recordsToAdd:  []records.Raw{},
		// 	expectedError: efd.ErrMissingEFDFRecord,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := t.Context()

			builder := efd.NewRootBuilder(logger)

			for _, record := range tt.recordsToAdd {
				err := builder.AddRecord(ctx, record)
				if err != nil {
					t.Fatalf("failed to add record: %v", err)
				}
			}

			_, err := builder.Build()

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedError)
				}

				if !errors.Is(err, tt.expectedError) {
					t.Fatalf(
						"expected error %v, got %v",
						tt.expectedError,
						err,
					)
				}

				return
			}

			// if result != tt.expectedResult {
			// 	t.Fatalf("expected result %v, got %v", tt.expectedResult, result)
			// }

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
