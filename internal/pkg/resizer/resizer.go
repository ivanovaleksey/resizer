package resizer

import (
	"context"
	"image"

	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Resizer struct {
	logger        *zap.Logger
	imageProvider ImageProvider
}

type ImageProvider interface {
	GetImage(ctx context.Context, target string) (image.Image, error)
}

func NewResizer(logger *zap.Logger, provider ImageProvider) Resizer {
	return Resizer{
		logger:        logger,
		imageProvider: provider,
	}
}

func (r Resizer) Resize(ctx context.Context, target string, params Params) (image.Image, error) {
	img, err := r.imageProvider.GetImage(ctx, target)
	if err != nil {
		return nil, errors.Wrap(err, "can't get image")
	}

	image := make(chan image.Image, 1)
	go func() {
		defer close(image)
		image <- imaging.Resize(img, params.Width, params.Height, imaging.Lanczos)
	}()

	select {
	case i := <-image:
		return i, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
