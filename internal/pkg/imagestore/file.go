package imagestore

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io/ioutil"
)

type FileStore struct {
}

func NewFileStore() FileStore {
	return FileStore{}
}

func (f FileStore) GetImage(_ context.Context, target string) (image.Image, error) {
	buf, err := ioutil.ReadFile(target)
	if err != nil {
		return nil, err
	}

	return jpeg.Decode(bytes.NewReader(buf))
}
