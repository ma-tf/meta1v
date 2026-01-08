//go:generate mockgen -destination=./mocks/service_mock.go -package=csv_test github.com/ma-tf/meta1v/internal/service/csv Service
package csv

import (
	"fmt"
	"io"
	"strings"

	"github.com/ma-tf/meta1v/internal/service/display"
)

var _ Service = &service{}

type Service interface {
	ExportRoll(w io.Writer, r display.DisplayableRoll) error
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) ExportRoll(w io.Writer, r display.DisplayableRoll) error {
	var b strings.Builder
	b.WriteString(
		"FILM ID,FIRST ROW,FRAMES PER ROW,TITLE,FILM LOADED AT,FRAME COUNT,ISO (DX),REMARKS\n",
	)

	_, err := fmt.Fprintf(&b, "%s,%s,%s,%s,%s,%s,%s,%s",
		r.FilmID,
		r.FirstRow,
		r.PerRow,
		r.Title,
		r.FilmLoadedDate,
		r.FrameCount,
		r.IsoDX,
		r.Remarks,
	)
	if err != nil {
		return fmt.Errorf("failed to write roll header: %w", err)
	}

	if _, err = w.Write([]byte(b.String())); err != nil {
		return fmt.Errorf("failed to write roll header: %w", err)
	}

	return nil
}
