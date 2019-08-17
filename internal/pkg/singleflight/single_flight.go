package singleflight

import (
	"context"
	"image"
	"sync"

	"github.com/cespare/xxhash"
	"go.uber.org/zap"

	"github.com/ivanovaleksey/resizer/internal/pkg/cache"
)

const bucketsCount = 256

type SingleFlight struct {
	locks   [bucketsCount]sync.Mutex
	buckets [bucketsCount]map[cache.Entity]*Entry

	logger        *zap.Logger
	cache         CacheProvider
	imageProvider ImageProvider
}

type CacheProvider interface {
	Get(cache.Entity) (image.Image, error)
	Set(cache.Entity, image.Image) error
}

type ImageProvider interface {
	GetImage(ctx context.Context, target string) (image.Image, error)
}

func NewSingleFlight(opts ...Option) *SingleFlight {
	s := &SingleFlight{
		logger:        zap.NewNop(),
		cache:         dummyCacheProvider{},
		imageProvider: dummyImageProvider{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type Entry struct {
	ok    image.Image
	err   error
	ready chan struct{}
}

func (s *SingleFlight) GetImage(ctx context.Context, target string) (image.Image, error) {
	e := cache.Entity(target)

	value, err := s.cache.Get(e)
	if value != nil {
		s.logger.Debug("cache hit")
		return value, nil
	}

	if err == cache.ErrCacheMiss {
		s.logger.Debug("cache miss")
	}
	if err != nil && err != cache.ErrCacheMiss {
		s.logger.Error("can't get cache", zap.Error(err), zap.String("key", e.Key()))
	}

	idx := xxhash.Sum64([]byte(e)) % bucketsCount

	lock := &s.locks[idx]
	lock.Lock()

	if s.buckets[idx] == nil {
		s.buckets[idx] = make(map[cache.Entity]*Entry)
	}
	bucket := s.buckets[idx]

	entry, ok := bucket[e]
	if !ok {
		entry = &Entry{ready: make(chan struct{})}
		bucket[e] = entry
		lock.Unlock()

		entry.ok, entry.err = s.imageProvider.GetImage(ctx, target)
		close(entry.ready)
		if entry.err == nil {
			if err := s.cache.Set(e, entry.ok); err != nil {
				s.logger.Error("can't set cache", zap.Error(err), zap.String("key", e.Key()))
			}
		}

		lock.Lock()
		delete(bucket, e)
		lock.Unlock()
	} else {
		lock.Unlock()
		select {
		case <-entry.ready:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if err := entry.err; err != nil {
		return nil, err
	}

	return entry.ok, nil
}
