package resizer

import (
	"context"
	"image"
)

type dummyImageProvider struct {
}

func (d dummyImageProvider) GetImage(ctx context.Context, target string) (image.Image, error) {
	return nil, nil
}

type dummyResizer struct {
}

func (d dummyResizer) Resize(image.Image, Params) (image.Image, error) {
	return nil, nil
}
