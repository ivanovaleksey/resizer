// +build !race

package app

import (
	"context"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ivanovaleksey/resizer/test"
)

func TestApplication_ResizeImage(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	params := []string{
		"url=" + path.Join(test.RootDir(t, 3), "test/testdata/nature.jpg"),
		"width=500",
		"height=300",
	}

	req := httptest.NewRequest(
		"GET",
		"/image/resize?"+strings.Join(params, "&"),
		nil,
	)

	t.Run("it resizes image", func(t *testing.T) {
		app := NewApp(context.Background(), logger)
		err = app.Init(Config{ImageProvider: ImageProviderFile})
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.ResizeImage)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		cfg, err := jpeg.DecodeConfig(rr.Body)
		require.NoError(t, err)
		assert.Equal(t, 500, cfg.Width)
		assert.Equal(t, 300, cfg.Height)
	})

	t.Run("it supports browser caching", func(t *testing.T) {
		app := NewApp(context.Background(), logger)
		err = app.Init(Config{ImageProvider: ImageProviderFile})
		require.NoError(t, err)

		rr1 := httptest.NewRecorder()
		handler := app.Handler()

		handler.ServeHTTP(rr1, req)
		require.Equal(t, http.StatusOK, rr1.Code)
		etag := rr1.Header().Get("Etag")
		require.NotEmpty(t, etag)
		assert.Equal(t, "max-age=3600", rr1.Header().Get("Cache-Control"))

		req.Header.Set("If-None-Match", etag)
		rr2 := httptest.NewRecorder()
		handler.ServeHTTP(rr2, req)
		//it is because of encode/decode
		assert.Equal(t, http.StatusOK, rr2.Code)
		assert.NotEmpty(t, rr2.Body)

		etag = rr2.Header().Get("Etag")
		req.Header.Set("If-None-Match", etag)
		rr3 := httptest.NewRecorder()
		handler.ServeHTTP(rr3, req)
		assert.Equal(t, http.StatusNotModified, rr3.Code)
		assert.Empty(t, rr3.Body)
	})
}
