package imagestore

import (
	"context"
	"image"
	"image/jpeg"
	"net/http"
)

type HTTPStore struct {
	client http.Client
}

func NewHTTPStore() HTTPStore {
	return HTTPStore{
		client: http.Client{},
	}
}

func (d HTTPStore) GetImage(ctx context.Context, url string) (image.Image, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return jpeg.Decode(resp.Body)
}
