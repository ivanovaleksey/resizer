package app

import (
	"context"
	"image"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-http-utils/etag"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/ivanovaleksey/resizer/internal/pkg/cache"
	"github.com/ivanovaleksey/resizer/internal/pkg/imagestore"
	"github.com/ivanovaleksey/resizer/internal/pkg/resizer"
	"github.com/ivanovaleksey/resizer/internal/pkg/singleflight"
)

type Application struct {
	ctx           context.Context
	logger        *zap.Logger
	handler       http.Handler
	resizeService Resizer
}

type Resizer interface {
	Resize(ctx context.Context, target string, params resizer.Params) (image.Image, error)
}

func NewApp(ctx context.Context, logger *zap.Logger) *Application {
	return &Application{
		ctx:    ctx,
		logger: logger,
	}
}

func (a *Application) Handler() http.Handler {
	return a.handler
}

func (a *Application) Init(cfg Config) error {
	a.handler = chi.ServerBaseContext(a.ctx, a.initRouter())

	imageProvider, err := a.initImageProvider(cfg)
	if err != nil {
		return errors.Wrap(err, "can't create image provider")
	}

	opts := []resizer.ServiceOption{
		resizer.WithLogger(a.logger),
		resizer.WithImageProvider(imageProvider),
		resizer.WithImageResizer(resizer.NewResizer()),
	}
	service, err := resizer.NewService(opts...)
	if err != nil {
		return errors.Wrap(err, "can't create service")
	}
	a.resizeService = service

	return nil
}

func (a *Application) initRouter() http.Handler {
	r := chi.NewRouter()

	r.Route("/image", func(r chi.Router) {
		r.Get("/resize", etag.Handler(http.HandlerFunc(a.ResizeImage), false).ServeHTTP)
	})

	return r
}

func (a *Application) initImageProvider(cfg Config) (resizer.ImageProvider, error) {
	imageCache, err := cache.NewCache()
	if err != nil {
		return nil, errors.Wrap(err, "can't create cache")
	}
	var imageProvider resizer.ImageProvider
	switch cfg.ImageProvider {
	case ImageProviderHTTP:
		imageProvider = imagestore.NewHTTPStore()
	case ImageProviderFile:
		imageProvider = imagestore.NewFileStore()
	default:
		return nil, errors.New("unknown image provider")
	}

	opts := []singleflight.Option{
		singleflight.WithLogger(a.logger),
		singleflight.WithCacheProvider(imageCache),
		singleflight.WithImageProvider(imageProvider),
	}
	return singleflight.NewSingleFlight(opts...), nil
}
