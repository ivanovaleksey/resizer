package app

import (
	"context"
	"image"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/ivanovaleksey/resizer/internal/pkg/resizer"
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

func (a *Application) Init() error {
	a.handler = chi.ServerBaseContext(a.ctx, a.initRouter())

	service, err := resizer.NewService(a.logger)
	if err != nil {
		return errors.Wrap(err, "can't create service")
	}
	a.resizeService = service

	return nil
}

func (a *Application) initRouter() http.Handler {
	r := chi.NewRouter()

	r.Route("/image", func(r chi.Router) {
		r.Get("/resize", a.resizeImage)
	})

	return r
}
