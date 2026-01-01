package efd_test

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"testing"

	"github.com/ma-tf/meta1v/internal/service/efd"
	efd_test "github.com/ma-tf/meta1v/internal/service/efd/mocks"
	osfs_test "github.com/ma-tf/meta1v/internal/service/osfs/mocks"
	"github.com/ma-tf/meta1v/pkg/records"
	"go.uber.org/mock/gomock"
)

var errExample = errors.New("example error")

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

//nolint:exhaustruct // for records
func Test_RecordsFromFile(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name     string
		filename string
		expect   func(
			mockFileSystem osfs_test.MockFileSystem,
			mockParser efd_test.MockParser,
			mockRootBuilder efd_test.MockRootBuilder,
			tt testcase,
			ctrl *gomock.Controller,
		)
		expectedResult *records.Root
		expectedError  error
	}

	tests := []testcase{
		{
			name:     "file does not exist",
			filename: "nonexistent.efd",
			expect: func(
				mockFileSystem osfs_test.MockFileSystem,
				_ efd_test.MockParser,
				_ efd_test.MockRootBuilder,
				tt testcase,
				_ *gomock.Controller,
			) {
				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(nil, errExample)
			},
			expectedError: efd.ErrFailedToOpenFile,
		},
		{
			name:     "failed to read record",
			filename: "failed_to_read.efd",
			expect: func(
				mockFileSystem osfs_test.MockFileSystem,
				mockParser efd_test.MockParser,
				_ efd_test.MockRootBuilder,
				tt testcase,
				ctrl *gomock.Controller,
			) {
				file := osfs_test.NewMockFile(ctrl)
				file.EXPECT().
					Close().
					Return(nil)

				mockParser.EXPECT().
					ParseRaw(gomock.Any(), file).
					Return(records.Raw{}, errExample)

				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(file, nil)
			},
			expectedError: efd.ErrFailedToReadRecord,
		},
		{
			name:     "failed to build root record",
			filename: "failed_to_build.efd",
			expect: func(
				mockFileSystem osfs_test.MockFileSystem,
				mockParser efd_test.MockParser,
				mockRootBuilder efd_test.MockRootBuilder,
				tt testcase,
				ctrl *gomock.Controller,
			) {
				file := osfs_test.NewMockFile(ctrl)
				file.EXPECT().
					Close().
					Return(nil)

				mockParser.EXPECT().
					ParseRaw(gomock.Any(), file).
					Return(records.Raw{}, io.EOF)

				mockRootBuilder.EXPECT().
					Build().
					Return(records.Root{}, errExample)

				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(file, nil)
			},
			expectedError: efd.ErrFailedToBuildRoot,
		},
		{
			name: "successfully parsed EFD file",
			expect: func(
				mockFileSystem osfs_test.MockFileSystem,
				mockParser efd_test.MockParser,
				mockRootBuilder efd_test.MockRootBuilder,
				tt testcase,
				ctrl *gomock.Controller,
			) {
				file := osfs_test.NewMockFile(ctrl)
				file.EXPECT().
					Close().
					Return(nil)

				efdfRaw, efrmRaw, eftpRaw := []byte{}, []byte{}, []byte{}

				mockParser.EXPECT().
					ParseRaw(gomock.Any(), file).
					Return(records.Raw{
						Magic:  [4]byte{'E', 'F', 'D', 'F'},
						Length: uint64(len(efdfRaw)),
						Data:   efdfRaw,
					}, nil)

				mockParser.EXPECT().
					ParseRaw(gomock.Any(), file).
					Return(records.Raw{
						Magic:  [4]byte{'E', 'F', 'R', 'M'},
						Length: uint64(len(efrmRaw)),
						Data:   efrmRaw,
					}, nil)

				mockParser.EXPECT().
					ParseRaw(gomock.Any(), file).
					Return(records.Raw{
						Magic:  [4]byte{'E', 'F', 'T', 'P'},
						Length: uint64(len(eftpRaw)),
						Data:   eftpRaw,
					}, nil)

				mockParser.EXPECT().
					ParseRaw(gomock.Any(), file).
					Return(records.Raw{}, io.EOF)

				efdf := records.EFDF{Title: [64]byte{'t', 'i', 't', 'l', 'e'}}
				mockParser.EXPECT().
					ParseEFDF(gomock.Any(), efdfRaw).
					Return(efdf, nil)

				efrm := records.EFRM{FrameNumber: 1}
				mockParser.EXPECT().
					ParseEFRM(gomock.Any(), efrmRaw).
					Return(efrm, nil)

				eftp := records.EFTP{Width: 100, Height: 100}
				mockParser.EXPECT().
					ParseEFTP(gomock.Any(), eftpRaw).
					Return(eftp, nil)

				mockRootBuilder.EXPECT().
					AddEFDF(gomock.Any(), efdf).
					Return(nil)

				mockRootBuilder.EXPECT().
					AddEFRM(gomock.Any(), efrm)

				mockRootBuilder.EXPECT().
					AddEFTP(gomock.Any(), eftp)

				mockRootBuilder.EXPECT().
					Build().
					Return(records.Root{
						EFDF:  efdf,
						EFRMs: []records.EFRM{efrm},
						EFTPs: []records.EFTP{eftp},
					}, nil)

				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(file, nil)
			},
			expectedResult: &records.Root{
				EFDF:  records.EFDF{Title: [64]byte{'t', 'i', 't', 'l', 'e'}},
				EFRMs: []records.EFRM{{FrameNumber: 1}},
				EFTPs: []records.EFTP{{Width: 100, Height: 100}},
			},
			expectedError: nil,
		},
	}

	assertExpectedError := func(t *testing.T, got, expected error) {
		t.Helper()

		if expected == nil {
			return
		}

		if got == nil {
			t.Fatalf("expected error %v, got nil", expected)
		}

		if !errors.Is(got, expected) {
			t.Fatalf("expected error %v, got %v", expected, got)
		}
	}

	assertExpectedResult := func(t *testing.T, got, expected records.Root) {
		t.Helper()

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("expected result %v, got %v", expected, got)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := t.Context()

			mockFileSystem := osfs_test.NewMockFileSystem(ctrl)
			mockParser := efd_test.NewMockParser(ctrl)
			mockRootBuilder := efd_test.NewMockRootBuilder(ctrl)

			if tt.expect != nil {
				tt.expect(
					*mockFileSystem,
					*mockParser,
					*mockRootBuilder,
					tt,
					ctrl,
				)
			}

			svc := efd.NewService(
				newTestLogger(),
				mockRootBuilder,
				mockParser,
				mockFileSystem,
			)

			result, err := svc.RecordsFromFile(
				ctx,
				tt.filename,
			)

			if tt.expectedError != nil {
				assertExpectedError(t, err, tt.expectedError)

				return
			}

			if tt.expectedResult == nil {
				t.Fatalf("test case missing expectedResult")
			}

			assertExpectedResult(t, result, *tt.expectedResult)
		})
	}
}

//nolint:exhaustruct // for records
func Test_RecordsFromFile_ProcessRecord(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name     string
		filename string
		expect   func(
			mockFileSystem osfs_test.MockFileSystem,
			mockParser efd_test.MockParser,
			mockRootBuilder efd_test.MockRootBuilder,
			tt testcase,
			ctrl *gomock.Controller,
		)
		expectedError error
	}

	tests := []testcase{
		{
			name:     "failed to parse EFDF record",
			filename: "failed_to_parse_efdf.efd",
			expect: func(
				mockFileSystem osfs_test.MockFileSystem,
				mockParser efd_test.MockParser,
				_ efd_test.MockRootBuilder,
				tt testcase,
				ctrl *gomock.Controller,
			) {
				file := osfs_test.NewMockFile(ctrl)
				file.EXPECT().
					Close().
					Return(nil)

				efdfRaw := []byte{}
				mockParser.EXPECT().
					ParseRaw(gomock.Any(), file).
					Return(records.Raw{
						Magic:  [4]byte{'E', 'F', 'D', 'F'},
						Length: uint64(len(efdfRaw)),
						Data:   efdfRaw,
					}, nil)

				mockParser.EXPECT().
					ParseEFDF(gomock.Any(), efdfRaw).
					Return(records.EFDF{}, errExample)

				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(file, nil)
			},
			expectedError: efd.ErrFailedToAddRecord,
		},
		{
			name:     "failed to add efdf to builder",
			filename: "failed_to_add_efdf.efd",
			expect: func(
				mockFileSystem osfs_test.MockFileSystem,
				mockParser efd_test.MockParser,
				mockRootBuilder efd_test.MockRootBuilder,
				tt testcase,
				ctrl *gomock.Controller,
			) {
				file := osfs_test.NewMockFile(ctrl)
				file.EXPECT().
					Close().
					Return(nil)

				efdfRaw := []byte{}
				mockParser.EXPECT().
					ParseRaw(gomock.Any(), file).
					Return(records.Raw{
						Magic:  [4]byte{'E', 'F', 'D', 'F'},
						Length: uint64(len(efdfRaw)),
						Data:   efdfRaw,
					}, nil)

				efdf := records.EFDF{Title: [64]byte{'t', 'i', 't', 'l', 'e'}}
				mockParser.EXPECT().
					ParseEFDF(gomock.Any(), efdfRaw).
					Return(efdf, nil)

				mockRootBuilder.EXPECT().
					AddEFDF(gomock.Any(), efdf).
					Return(errExample)

				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(file, nil)
			},
			expectedError: efd.ErrFailedToAddRecord,
		},
		{
			name:     "failed to parse EFRM record",
			filename: "failed_to_parse_efrm.efd",
			expect: func(
				mockFileSystem osfs_test.MockFileSystem,
				mockParser efd_test.MockParser,
				_ efd_test.MockRootBuilder,
				tt testcase,
				ctrl *gomock.Controller,
			) {
				file := osfs_test.NewMockFile(ctrl)
				file.EXPECT().
					Close().
					Return(nil)

				efrmRaw := []byte{}
				mockParser.EXPECT().
					ParseRaw(gomock.Any(), file).
					Return(records.Raw{
						Magic:  [4]byte{'E', 'F', 'R', 'M'},
						Length: uint64(len(efrmRaw)),
						Data:   efrmRaw,
					}, nil)

				mockParser.EXPECT().
					ParseEFRM(gomock.Any(), efrmRaw).
					Return(records.EFRM{}, errExample)

				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(file, nil)
			},
			expectedError: efd.ErrFailedToAddRecord,
		},
		{
			name:     "failed to parse EFTP record",
			filename: "failed_to_parse_eftp.efd",
			expect: func(
				mockFileSystem osfs_test.MockFileSystem,
				mockParser efd_test.MockParser,
				_ efd_test.MockRootBuilder,
				tt testcase,
				ctrl *gomock.Controller,
			) {
				file := osfs_test.NewMockFile(ctrl)
				file.EXPECT().
					Close().
					Return(nil)

				eftpRaw := []byte{}
				mockParser.EXPECT().
					ParseRaw(gomock.Any(), file).
					Return(records.Raw{
						Magic:  [4]byte{'E', 'F', 'T', 'P'},
						Length: uint64(len(eftpRaw)),
						Data:   eftpRaw,
					}, nil)

				mockParser.EXPECT().
					ParseEFTP(gomock.Any(), eftpRaw).
					Return(records.EFTP{}, errExample)

				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(file, nil)
			},
			expectedError: efd.ErrFailedToAddRecord,
		},
		{
			name: "unknown record magic number",
			expect: func(
				mockFileSystem osfs_test.MockFileSystem,
				mockParser efd_test.MockParser,
				_ efd_test.MockRootBuilder,
				tt testcase,
				ctrl *gomock.Controller,
			) {
				file := osfs_test.NewMockFile(ctrl)
				file.EXPECT().
					Close().
					Return(nil)

				mockParser.EXPECT().
					ParseRaw(gomock.Any(), file).
					Return(records.Raw{
						Magic:  [4]byte{'X', 'Y', 'Z', 'W'},
						Length: 0,
						Data:   []byte{},
					}, nil)

				mockFileSystem.EXPECT().
					Open(tt.filename).
					Return(file, nil)
			},
			expectedError: efd.ErrUnknownRecordType,
		},
	}

	assertExpectedError := func(t *testing.T, got, expected error) {
		t.Helper()

		if expected == nil {
			return
		}

		if got == nil {
			t.Fatalf("expected error %v, got nil", expected)
		}

		if !errors.Is(got, expected) {
			t.Fatalf("expected error %v, got %v", expected, got)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := t.Context()

			mockFileSystem := osfs_test.NewMockFileSystem(ctrl)
			mockParser := efd_test.NewMockParser(ctrl)
			mockRootBuilder := efd_test.NewMockRootBuilder(ctrl)

			if tt.expect != nil {
				tt.expect(
					*mockFileSystem,
					*mockParser,
					*mockRootBuilder,
					tt,
					ctrl,
				)
			}

			svc := efd.NewService(
				newTestLogger(),
				mockRootBuilder,
				mockParser,
				mockFileSystem,
			)

			_, err := svc.RecordsFromFile(
				ctx,
				tt.filename,
			)

			if tt.expectedError != nil {
				assertExpectedError(t, err, tt.expectedError)

				return
			}

			t.Fatalf("expected error %v, got nil", tt.expectedError)
		})
	}
}
