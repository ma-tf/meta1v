package cli

import (
	"errors"
)

var (
	ErrNoFilenameProvided     = errors.New("filename must be specified")
	ErrFailedToOpenFile       = errors.New("failed to open file")
	ErrFailedToReadRecord     = errors.New("failed to read record")
	ErrMultipleEFDFRecords    = errors.New("multiple EFDF records found")
	ErrUnknownRecordType      = errors.New("unknown record type")
	ErrFailedToParseThumbnail = errors.New("failed to parse EFTP thumbnail")
)
