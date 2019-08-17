package singleflight

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io/ioutil"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ivanovaleksey/resizer/internal/pkg/cache"
)

func TestSingleFlight_GetImage(t *testing.T) {
	const goroutinesCount = 10000

	url := "http://example.com/1.jpg"
	file, err := ioutil.ReadFile("../resizer/testdata/nature.jpg")
	require.NoError(t, err)
	srcImage, err := jpeg.Decode(bytes.NewReader(file))
	require.NoError(t, err)

	t.Run("it waits for in-flight calls", func(t *testing.T) {
		ctx := context.Background()
		imageProvider := &imageProviderWithCounter{img: srcImage}
		opts := []Option{
			WithCacheProvider(&simpleImageCache{m: make(map[cache.Entity]image.Image)}),
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
	})

	t.Run("it serves from cache", func(t *testing.T) {
		ctx := context.Background()
		imageProvider := &imageProviderWithCounter{img: srcImage}
		opts := []Option{
			WithCacheProvider(&simpleImageCache{m: make(map[cache.Entity]image.Image)}),
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
	})
}

type imageProviderWithCounter struct {
	counter int32 // atomic access
	img     image.Image
}

func (i *imageProviderWithCounter) GetImage(ctx context.Context, target string) (image.Image, error) {
	time.Sleep(100 * time.Millisecond)
	atomic.AddInt32(&i.counter, 1)
	return i.img, nil
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
