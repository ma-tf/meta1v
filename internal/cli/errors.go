package cli

import "errors"

var (
	ErrFailedToOpenFile        = errors.New("failed to open specified file")
	ErrFailedToGetStrictFlag   = errors.New("failed to get strict flag")
	ErrOutputFileAlreadyExists = errors.New(
		"output file already exists, use --force/-F to overwrite",
	)
	ErrFailedToGetForceFlag        = errors.New("failed to get force flag")
	ErrForceFlagRequiresTargetFile = errors.New(
		"--force/-F flag can only be used when exporting to a file",
	)
)
