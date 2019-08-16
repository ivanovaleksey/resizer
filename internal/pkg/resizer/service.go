package resizer

import (
	"context"
	"image"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/ivanovaleksey/resizer/internal/pkg/imagestore"
	"github.com/ivanovaleksey/resizer/internal/pkg/resizer/cache"
)

type Service struct {
	logger        *zap.Logger
	imageProvider ImageProvider
	imageResizer  ImageResizer
	cache         CacheProvider
}

type ImageProvider interface {
	GetImage(ctx context.Context, target string) (image.Image, error)
}

type ImageResizer interface {
	Resize(image.Image, Params) (image.Image, error)
}

type CacheProvider interface {
	Get(cache.Entity) (image.Image, error)
	Set(cache.Entity, image.Image) error
}

func NewService(logger *zap.Logger, opts ...ServiceOption) (Service, error) {
	ch, err := cache.NewCache()
	if err != nil {
		return Service{}, err
	}
	s := Service{
		logger:        logger,
		imageProvider: imagestore.NewHTTPStore(),
		imageResizer:  Resizer{},
		cache:         ch,
	}

	for _, opt := range opts {
		opt(&s)
	}

	return s, nil
}

func (r Service) Resize(ctx context.Context, target string, params Params) (image.Image, error) {
	entity := cache.Entity{
		URL:    target,
		Width:  params.Width,
		Height: params.Height,
	}
	cachedImage, err := r.cache.Get(entity)
	if err != nil {
		if cache.IsMiss(err) {
			r.logger.Debug("cache miss")
		} else {
			r.logger.Error("can't get from cache", zap.Error(err), zap.String("key", entity.Key()))
		}
	}
	if cachedImage != nil {
		r.logger.Debug("cache hit")
		return cachedImage, nil
	}

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
		if err := result.err; err != nil {
			return nil, err
		}
		if err := r.cache.Set(entity, result.ok); err != nil {
			r.logger.Error("can't set cache", zap.Error(err), zap.String("key", entity.Key()))
		}
		return result.ok, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
