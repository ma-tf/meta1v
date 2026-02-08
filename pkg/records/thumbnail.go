//go:generate mockgen -destination=./mocks/thumbnail_factory_mock.go -package=records_test github.com/ma-tf/meta1v/pkg/records ThumbnailFactory
package records

import "image"

// ThumbnailFactory creates RGBA image buffers for thumbnail data.
// This interface allows for test mocking and alternative image allocation strategies.
type ThumbnailFactory interface {
	// NewRGBA creates a new RGBA image with the given dimensions.
	NewRGBA(rect image.Rectangle) *image.RGBA
}

type thumbnailFactory struct{}

func (thumbnailFactory) NewRGBA(rect image.Rectangle) *image.RGBA {
	return image.NewRGBA(rect)
}

func NewDefaultThumbnailFactory() ThumbnailFactory {
	return thumbnailFactory{}
}
