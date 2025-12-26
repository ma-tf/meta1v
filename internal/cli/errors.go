package cli

import "errors"

var (
	ErrNoFilenameProvided = errors.New("filename must be specified")
	ErrFailedToOpenFile   = errors.New("failed to open specified file")
)
