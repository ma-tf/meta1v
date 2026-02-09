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

package efd_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ma-tf/meta1v/internal/records"
	"github.com/ma-tf/meta1v/internal/service/efd"
	"go.uber.org/mock/gomock"
)

//nolint:exhaustruct // only partial is needed
func Test_RootBuilder(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name           string
		efdfsToAdd     []records.EFDF
		efrmsToAdd     []records.EFRM
		eftpsToAdd     []records.EFTP
		expectedError  error
		expectedResult records.Root
	}

	tests := []testcase{
		{
			name: "error on multiple EFDF records",
			efdfsToAdd: []records.EFDF{
				{Title: [64]byte{'t', 'i', 't', 'l', 'e'}},
				{Title: [64]byte{'o', 't', 'h', 'e', 'r'}},
			},
			expectedError: efd.ErrMultipleEFDFRecords,
		},
		{
			name:          "error on missing EFDF record",
			efdfsToAdd:    []records.EFDF{},
			expectedError: efd.ErrMissingEFDFRecord,
		},
		{
			name: "successful build with EFDF, EFRM and EFTP records",
			efdfsToAdd: []records.EFDF{
				{Title: [64]byte{'t', 'i', 't', 'l', 'e'}},
			},
			efrmsToAdd: []records.EFRM{
				{Remarks: [256]byte{'r', 'e', 'm', 'a', 'r', 'k', 's'}},
			},
			eftpsToAdd: []records.EFTP{
				{Index: 1},
			},
			expectedError: nil,
			expectedResult: records.Root{
				EFDF: records.EFDF{Title: [64]byte{'t', 'i', 't', 'l', 'e'}},
				EFRMs: []records.EFRM{
					{Remarks: [256]byte{'r', 'e', 'm', 'a', 'r', 'k', 's'}},
				},
				EFTPs: []records.EFTP{
					{Index: 1},
				},
			},
		},
	}

	addRecordsAndBuild := func(ctx context.Context, builder efd.RootBuilder, tc testcase) (records.Root, error) {
		for _, record := range tc.efdfsToAdd {
			if err := builder.AddEFDF(ctx, record); err != nil {
				return records.Root{}, err
			}
		}

		for _, record := range tc.efrmsToAdd {
			builder.AddEFRM(ctx, record)
		}

		for _, record := range tc.eftpsToAdd {
			builder.AddEFTP(ctx, record)
		}

		return builder.Build()
	}

	assertExpectedError := func(t *testing.T, got, expected error) {
		t.Helper()

		if expected == nil {
			return
		}

		if got == nil || !errors.Is(got, expected) {
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

			builder := efd.NewRootBuilder(newTestLogger())
			result, err := addRecordsAndBuild(ctx, builder, tt)

			if tt.expectedError != nil {
				assertExpectedError(t, err, tt.expectedError)

				return
			}

			assertExpectedResult(t, result, tt.expectedResult)
		})
	}
}
