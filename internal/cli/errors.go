package cli

import "errors"

var (
	ErrEFDFileMustBeProvided      = errors.New("efd file must be specified")
	ErrFailedToOpenFile           = errors.New("failed to open specified file")
	ErrTargetFileMustBeSpecified  = errors.New("target file must be specified")
	ErrFrameNumberMustBeSpecified = errors.New("frame number must be specified")
	ErrTooManyArguments           = errors.New("too many arguments provided")
	ErrFailedToGetStrictFlag      = errors.New("failed to get strict flag")
)
