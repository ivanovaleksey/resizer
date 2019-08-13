package app

import (
	"context"
	"image"
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/ivanovaleksey/resizer/internal/pkg/imagestore"
	"github.com/ivanovaleksey/resizer/internal/pkg/resizer"
)

type Application struct {
	ctx     context.Context
	logger  *zap.Logger
	handler http.Handler
	resizer Resizer
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

	imageProvider := imagestore.NewHTTPStore()
	a.resizer = resizer.NewResizer(a.logger, imageProvider)

	return nil
}

func (a *Application) initRouter() http.Handler {
	r := chi.NewRouter()

	r.Route("/image", func(r chi.Router) {
		r.Get("/resize", a.resizeImage)
	})

	return r
}
