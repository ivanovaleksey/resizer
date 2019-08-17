package singleflight

import (
	"context"
	"image"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ivanovaleksey/resizer/internal/pkg/cache"
	"github.com/ivanovaleksey/resizer/test"
)

func TestSingleFlight_GetImage(t *testing.T) {
	const (
		goroutinesCount = 10000
		url             = "http://example.com/1.jpg"
	)

	t.Run("it waits for in-flight calls", func(t *testing.T) {
		ctx := context.Background()

		imageCache := &simpleImageCache{m: make(map[cache.Entity]image.Image)}
		imageProvider := newImageProvider(t)
		opts := []Option{
			WithCacheProvider(imageCache),
			WithImageProvider(imageProvider),
		}
		s := NewSingleFlight(opts...)

		var wg sync.WaitGroup
		for i := 0; i < goroutinesCount; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				s.GetImage(ctx, url)
			}()
		}
		wg.Wait()

		assert.EqualValues(t, 1, imageProvider.Counter())
		assert.Equal(t, 1, len(imageCache.m))
	})

	t.Run("it serves from cache", func(t *testing.T) {
		ctx := context.Background()

		imageCache := &simpleImageCache{m: make(map[cache.Entity]image.Image)}
		imageProvider := newImageProvider(t)
		opts := []Option{
			WithCacheProvider(imageCache),
			WithImageProvider(imageProvider),
		}
		s := NewSingleFlight(opts...)

		var wg sync.WaitGroup
		for i := 0; i < goroutinesCount; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				s.GetImage(ctx, url+strconv.Itoa(i))
			}(i)
		}
		wg.Wait()

		for i := 0; i < goroutinesCount; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				s.GetImage(ctx, url+strconv.Itoa(i))
			}(i)
		}
		wg.Wait()

		assert.EqualValues(t, goroutinesCount, imageProvider.Counter())
		assert.Equal(t, goroutinesCount, len(imageCache.m))
	})

	t.Run("it respects deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		imageCache := &simpleImageCache{m: make(map[cache.Entity]image.Image)}
		imageProvider := newImageProvider(t)
		imageProvider.timeout = 300 * time.Millisecond
		opts := []Option{
			WithCacheProvider(imageCache),
			WithImageProvider(imageProvider),
		}
		s := NewSingleFlight(opts...)

		errors := make(chan error, goroutinesCount)
		var wg sync.WaitGroup
		for i := 0; i < goroutinesCount; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := s.GetImage(ctx, url)
				errors <- err
			}()
		}
		wg.Wait()
		close(errors)

		var errorCount int
		for err := range errors {
			assert.Equal(t, ctx.Err(), err)
			errorCount++
		}

		assert.Empty(t, imageCache.m)
		assert.Equal(t, goroutinesCount, errorCount)
	})
}

type imageProviderWithCounter struct {
	counter int32 // atomic access
	img     image.Image
	timeout time.Duration
}

func newImageProvider(t *testing.T) *imageProviderWithCounter {
	return &imageProviderWithCounter{
		img:     test.SampleImage(t, 3),
		timeout: 100 * time.Millisecond,
	}
}

func (i *imageProviderWithCounter) GetImage(ctx context.Context, target string) (image.Image, error) {
	atomic.AddInt32(&i.counter, 1)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(i.timeout):
		return i.img, nil
	}
}

func (i *imageProviderWithCounter) Counter() int32 {
	return atomic.LoadInt32(&i.counter)
}

type simpleImageCache struct {
	m    map[cache.Entity]image.Image
	lock sync.RWMutex
}

func (s *simpleImageCache) Get(key cache.Entity) (image.Image, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	img, ok := s.m[key]
	if !ok {
		return nil, cache.ErrCacheMiss
	}

	return img, nil
}

func (s *simpleImageCache) Set(key cache.Entity, img image.Image) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.m[key] = img
	return nil
}
