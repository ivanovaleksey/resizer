package cache

import (
	"bytes"
	"image"
	"image/jpeg"
	"time"

	"github.com/allegro/bigcache"
)

const (
	TTL     = 3600 * time.Second
	MaxSize = 1024 // in MB
)

type Cache struct {
	inner *bigcache.BigCache
}

func NewCache() (Cache, error) {
	cfg := bigcache.DefaultConfig(TTL)
	cfg.HardMaxCacheSize = MaxSize
	cfg.CleanWindow = 1 * time.Second

	cache, err := bigcache.NewBigCache(cfg)
	if err != nil {
		return Cache{}, nil
	}

	return Cache{inner: cache}, nil
}

func (c Cache) Get(entity Entity) (image.Image, error) {
	value, err := c.inner.Get(entity.Key())
	if err == bigcache.ErrEntryNotFound {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}

	img, err := jpeg.Decode(bytes.NewReader(value))
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (c Cache) Set(entity Entity, value image.Image) error {
	buf := bytes.NewBuffer(nil)
	err := jpeg.Encode(buf, value, nil)
	if err != nil {
		return err
	}

	return c.inner.Set(entity.Key(), buf.Bytes())
}
