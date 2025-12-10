package export

import (
	"context"
	"fmt"
	"io"

	"github.com/ma-tf/meta1v/internal/service/efd"
)

type RollExportUseCase struct {
	efdService efd.Service
}

func NewRollExportUseCase(efdService efd.Service) RollExportUseCase {
	return RollExportUseCase{
		efdService: efdService,
	}
}

func (uc RollExportUseCase) Export(ctx context.Context, r io.Reader) error {
	_, err := uc.efdService.RecordsFromFile(ctx, r)
	if err != nil {
		return fmt.Errorf("failed read file: %w", err)
	}

	// export to csv

	return nil
}
