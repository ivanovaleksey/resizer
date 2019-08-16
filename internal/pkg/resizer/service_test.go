package resizer

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io/ioutil"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ivanovaleksey/resizer/internal/pkg/resizer/cache"
	"github.com/ivanovaleksey/resizer/internal/pkg/resizer/mocks"
)

func TestResizer_Resize(t *testing.T) {
	url := "http://example.com/1.jpg"
	params := Params{Width: 500, Height: 300}
	entity := cache.Entity{
		URL:    url,
		Width:  params.Width,
		Height: params.Height,
	}

	file, err := ioutil.ReadFile("testdata/nature.jpg")
	require.NoError(t, err)
	imgCfg, err := jpeg.DecodeConfig(bytes.NewReader(file))
	require.NoError(t, err)
	require.Equal(t, 2560, imgCfg.Width)
	require.Equal(t, 1920, imgCfg.Height)
	srcImage, err := jpeg.Decode(bytes.NewReader(file))
	require.NoError(t, err)

	t.Run("with value in cache", func(t *testing.T) {
		ctx := context.Background()

		cacheMock := &mocks.CacheProvider{}
		cacheMock.On("Get", entity).Return(srcImage, nil)

		resizer, err := NewService(zap.NewNop(), WithCacheProvider(cacheMock))
		require.NoError(t, err)

		out, err := resizer.Resize(ctx, url, params)

		require.NoError(t, err)
		assert.Equal(t, srcImage, out)
		cacheMock.AssertExpectations(t)
	})

	t.Run("without value in cache", func(t *testing.T) {
		cacheMock := &mocks.CacheProvider{}
		cacheMock.On("Get", entity).Return(nil, errors.New("not found"))

		t.Run("when unable to get image", func(t *testing.T) {
			ctx := context.Background()

			imageProvider := &mocks.ImageProvider{}
			resizer, err := NewService(zap.NewNop(), WithImageProvider(imageProvider), WithCacheProvider(cacheMock))
			require.NoError(t, err)

			imageErr := errors.New("some error")
			imageProvider.On("GetImage", ctx, url).Return(nil, imageErr)

			out, err := resizer.Resize(ctx, url, params)

			require.Error(t, err)
			assert.EqualError(t, err, "can't get image: some error")
			assert.Nil(t, out)
			imageProvider.AssertExpectations(t)
			cacheMock.AssertExpectations(t)
		})

		t.Run("when able to get image", func(t *testing.T) {
			t.Run("with long resizing", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				defer cancel()

				imageProvider := &mocks.ImageProvider{}
				opts := []ServiceOption{
					WithImageProvider(imageProvider),
					WithImageResizer(sleepyResizer{timeout: 1 * time.Second}),
					WithCacheProvider(cacheMock),
				}
				resizer, err := NewService(zap.NewNop(), opts...)
				require.NoError(t, err)

				imageProvider.On("GetImage", ctx, url).Return(srcImage, nil)

				out, err := resizer.Resize(ctx, url, params)

				require.Error(t, err)
				require.Error(t, ctx.Err())
				assert.EqualError(t, err, ctx.Err().Error())
				assert.Nil(t, out)
				imageProvider.AssertExpectations(t)
				cacheMock.AssertExpectations(t)
			})

			t.Run("with fast resizing", func(t *testing.T) {
				ctx := context.Background()

				imageProvider := &mocks.ImageProvider{}
				resizer, err := NewService(zap.NewNop(), WithImageProvider(imageProvider))
				require.NoError(t, err)

				imageProvider.On("GetImage", ctx, url).Return(srcImage, nil)

				out, err := resizer.Resize(ctx, url, params)

				require.NoError(t, err)
				require.NotNil(t, out)

				buf := bytes.NewBuffer(nil)
				err = jpeg.Encode(buf, out, nil)
				require.NoError(t, err)

				outCfg, err := jpeg.DecodeConfig(buf)
				require.NoError(t, err)
				assert.Equal(t, 500, outCfg.Width)
				assert.Equal(t, 300, outCfg.Height)
				imageProvider.AssertExpectations(t)
				cacheMock.AssertExpectations(t)
			})
		})
	})
}

type sleepyResizer struct {
	timeout time.Duration
	Resizer
}

func (s sleepyResizer) Resize(img image.Image, params Params) (image.Image, error) {
	<-time.After(s.timeout)
	return s.Resizer.Resize(img, params)
}
