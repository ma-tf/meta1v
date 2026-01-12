package exif

import (
	"errors"
	"log/slog"
	"os/exec"
)

var ErrExiftoolNotFound = errors.New("exiftool binary not found in PATH")

// ServiceFactory creates exif Service instances after validating dependencies.
type ServiceFactory struct {
	log     *slog.Logger
	runner  ToolRunner
	builder Builder
}

// NewServiceFactory returns a new ServiceFactory.
func NewServiceFactory(
	log *slog.Logger,
	runner ToolRunner,
	builder Builder,
) ServiceFactory {
	return ServiceFactory{log: log, runner: runner, builder: builder}
}

// Create validates that exiftool is available and returns a new Service.
func (f ServiceFactory) Create() (Service, error) {
	if _, err := exec.LookPath("exiftool"); err != nil {
		return nil, errors.Join(ErrExiftoolNotFound, err)
	}

	return &service{
		log:     f.log,
		runner:  f.runner,
		builder: f.builder,
	}, nil
}
