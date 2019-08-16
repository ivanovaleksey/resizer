package app

import (
	"bytes"
	"image/jpeg"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/ivanovaleksey/resizer/internal/pkg/resizer"
)

func (a *Application) resizeImage(w http.ResponseWriter, r *http.Request) {
	const (
		urlParamName    = "url"
		widthParamName  = "width"
		heightParamName = "height"
	)

	a.logger.Debug("resize image: start")

	ctx := r.Context()

	imageURL := r.URL.Query().Get(urlParamName)

	imageWidth, err := strconv.Atoi(r.URL.Query().Get(widthParamName))
	if err != nil {
		a.logger.Error("can't parse width", zap.Error(err))
		http.Error(w, "invalid width", http.StatusUnprocessableEntity)
		return
	}

	imageHeight, err := strconv.Atoi(r.URL.Query().Get(heightParamName))
	if err != nil {
		a.logger.Error("can't parse height", zap.Error(err))
		http.Error(w, "invalid height", http.StatusUnprocessableEntity)
		return
	}

	params := resizer.Params{Width: imageWidth, Height: imageHeight}
	image, err := a.resizeService.Resize(ctx, imageURL, params)
	if err != nil {
		a.logger.Error("can't resize image", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	buf := &bytes.Buffer{}
	err = jpeg.Encode(buf, image, nil)
	if err != nil {
		a.logger.Error("can't write image", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	if _, err := w.Write(buf.Bytes()); err != nil {
		a.logger.Error("can't write response", zap.Error(err))
		return
	}

	a.logger.Debug("resize image: done")
}
