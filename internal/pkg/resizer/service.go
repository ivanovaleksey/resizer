package resizer

import (
	"context"
	"image"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/ivanovaleksey/resizer/internal/pkg/imagestore"
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

func NewService(logger *zap.Logger, opts ...ServiceOption) Service {
	s := Service{
		logger:        logger,
		imageProvider: imagestore.NewHTTPStore(),
		imageResizer:  Resizer{},
	}

	for _, opt := range opts {
		opt(&s)
	}

	return s
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
		out, err := r.imageResizer.Resize(img, params)
		if err != nil {
			resize <- resizeResult{err: err}
			return
		}
		resize <- resizeResult{ok: out}
	}()

	select {
	case result := <-resize:
		return result.ok, result.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
