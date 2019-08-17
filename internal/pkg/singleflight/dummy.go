package singleflight

import (
	"context"
	"image"

	"github.com/ivanovaleksey/resizer/internal/pkg/cache"
)

type dummyImageProvider struct {
}

func (d dummyImageProvider) GetImage(ctx context.Context, target string) (image.Image, error) {
	return nil, nil
}

type dummyCacheProvider struct {
}

func (d dummyCacheProvider) Get(cache.Entity) (image.Image, error) {
	return nil, nil
}

func (d dummyCacheProvider) Set(cache.Entity, image.Image) error {
	return nil
}


