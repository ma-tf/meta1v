package records_test

import (
	"image"
	"reflect"
	"testing"

	"github.com/ma-tf/meta1v/internal/records"
)

func Test_ThumbnailFactory_NewRGBA(t *testing.T) {
	t.Parallel()

	factory := records.NewDefaultThumbnailFactory()

	rect := image.Rect(0, 0, 1, 1)
	result := factory.NewRGBA(rect)

	if result == nil {
		t.Fatal("expected non-nil RGBA")
	}

	if !reflect.DeepEqual(result.Bounds(), rect) {
		t.Errorf("unexpected bounds: got %v, want %v", result.Bounds(), rect)
	}
}
