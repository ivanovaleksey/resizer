package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"

	"github.com/ivanovaleksey/resizer/internal/pkg/app"
)

func main() {
	var cfg app.Config
	flag.IntVar(&cfg.ImageProvider, "image_provider", 1, "1 - http, 2 - file")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("can't create logger")
	}
	defer logger.Sync()

	application := app.NewApp(ctx, logger)
	if err := application.Init(cfg); err != nil {
		logger.Fatal("can't init application")
	}

	srv := http.Server{
		Addr:         ":80",
		Handler:      application.Handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	shutdown := make(chan struct{})
	go func(ctx context.Context) {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig

		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		logger.Debug("shutting down")
		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("can't shut down server", zap.Error(err))
		}
		close(shutdown)
	}(ctx)

	logger.Debug("server started")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logger.Error("server error", zap.Error(err))
	}

	logger.Debug("waiting for shut down")
	<-shutdown
	logger.Debug("shut down, exit")
}
