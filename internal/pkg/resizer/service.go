package resizer

import (
	"context"
	"image"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Service struct {
	logger        *zap.Logger
	imageProvider ImageProvider
	imageResizer  ImageResizer
}

type ImageProvider interface {
	GetImage(ctx context.Context, target string) (image.Image, error)
}

type ImageResizer interface {
	Resize(image.Image, Params) (image.Image, error)
}

func NewService(opts ...ServiceOption) (Service, error) {
	s := Service{
		logger:        zap.NewNop(),
		imageProvider: dummyImageProvider{},
		imageResizer:  dummyResizer{},
	}

	for _, opt := range opts {
		opt(&s)
	}

	return s, nil
}

func (r Service) Resize(ctx context.Context, target string, params Params) (image.Image, error) {
	img, err := r.imageProvider.GetImage(ctx, target)
	if err != nil {
		return nil, errors.Wrap(err, "can't get image")
	}

	type resizeResult struct {
		ok  image.Image
		err error
	}

	resize := make(chan resizeResult, 1)
	go func() {
		defer close(resize)
		var result resizeResult
		result.ok, result.err = r.imageResizer.Resize(img, params)
		resize <- result
	}()

	select {
	case result := <-resize:
		if err := result.err; err != nil {
			return nil, err
		}
		return result.ok, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
