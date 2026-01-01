//go:generate mockgen -destination=./mocks/builder_mock.go -package=efd_test github.com/ma-tf/meta1v/internal/service/efd RootBuilder
package efd

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ma-tf/meta1v/pkg/records"
)

var ErrMissingEFDFRecord = errors.New("missing EFDF record")

type RootBuilder interface {
	AddEFDF(ctx context.Context, efdf records.EFDF) error
	AddEFRM(ctx context.Context, efrm records.EFRM)
	AddEFTP(ctx context.Context, eftp records.EFTP)
	Build() (records.Root, error)
}

type rootBuilder struct {
	log   *slog.Logger
	efdf  *records.EFDF
	efrms []records.EFRM
	eftps []records.EFTP
}

func NewRootBuilder(log *slog.Logger) RootBuilder {
	var (
		efrms []records.EFRM
		eftps []records.EFTP
	)

	efrms = make([]records.EFRM, 0)
	eftps = make([]records.EFTP, 0)

	return &rootBuilder{
		log:   log,
		efdf:  nil,
		efrms: efrms,
		eftps: eftps,
	}
}

func (b *rootBuilder) AddEFDF(_ context.Context, efdf records.EFDF) error {
	if b.efdf != nil {
		return ErrMultipleEFDFRecords
	}

	b.efdf = &efdf

	return nil
}

func (b *rootBuilder) AddEFRM(_ context.Context, efrm records.EFRM) {
	b.efrms = append(b.efrms, efrm)
}

func (b *rootBuilder) AddEFTP(_ context.Context, eftp records.EFTP) {
	b.eftps = append(b.eftps, eftp)
}

func (b *rootBuilder) Build() (records.Root, error) {
	if b.efdf == nil {
		return records.Root{}, ErrMissingEFDFRecord
	}

	return records.Root{
		EFDF:  *b.efdf,
		EFRMs: b.efrms,
		EFTPs: b.eftps,
	}, nil
}
