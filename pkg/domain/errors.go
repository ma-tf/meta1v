package domain

import "errors"

var (
	ErrMultipleThumbnailsForFrame = errors.New("frame has multiple thumbnails")
	ErrFrameIndexOutOfRange       = errors.New("frame index out of range")

	ErrPrefixOutOfRange    = errors.New("prefix out of range (0-99)")
	ErrSuffixOutOfRange    = errors.New("suffix out of range (0-999)")
	ErrInvalidFilmLoadDate = errors.New("invalid film loaded date time")
	ErrInvalidAv           = errors.New("invalid Av value")
	ErrInvalidTv           = errors.New("invalid Tv value")
	ErrUnknownExposureComp = errors.New(
		"unknown exposure compensation value",
	)
	ErrUnknownFlashMode        = errors.New("unknown flash mode")
	ErrUnknownMeteringMode     = errors.New("unknown metering mode")
	ErrUnknownShootingMode     = errors.New("unknown shooting mode")
	ErrUnknownFilmAdvanceMode  = errors.New("unknown film advance mode")
	ErrUnknownAutoFocusMode    = errors.New("unknown auto focus mode")
	ErrInvalidBulbTime         = errors.New("invalid bulb exposure time")
	ErrUnknownMultipleExposure = errors.New("unknown multiple exposure value")
)
