//go:generate mockgen -destination=./mocks/service_mock.go -package=csv_test github.com/ma-tf/meta1v/internal/service/csv Service
package csv

var _ Service = &service{}

type Service interface {
	ExportRoll()
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) ExportRoll() {
	// Implementation goes here
}
