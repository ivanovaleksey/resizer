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

	"github.com/ivanovaleksey/resizer/internal/pkg/resizer/mocks"
)

func TestResizer_Resize(t *testing.T) {
	url := "http://example.com/1.jpg"
	params := Params{Width: 500, Height: 300}

	file, err := ioutil.ReadFile("testdata/nature.jpg")
	require.NoError(t, err)
	imgCfg, err := jpeg.DecodeConfig(bytes.NewReader(file))
	require.NoError(t, err)
	require.Equal(t, 2560, imgCfg.Width)
	require.Equal(t, 1920, imgCfg.Height)
	img, err := jpeg.Decode(bytes.NewReader(file))
	require.NoError(t, err)

	t.Run("when unable to get image", func(t *testing.T) {
		ctx := context.Background()

		imageProvider := &mocks.ImageProvider{}
		resizer := NewService(zap.NewNop(), WithImageProvider(imageProvider))

		imageErr := errors.New("some error")
		imageProvider.On("GetImage", ctx, url).Return(nil, imageErr)

		out, err := resizer.Resize(ctx, url, params)

		require.Error(t, err)
		assert.EqualError(t, err, "can't get image: some error")
		assert.Nil(t, out)
	})

	t.Run("when able to get image", func(t *testing.T) {
		t.Run("with long resizing", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			imageProvider := &mocks.ImageProvider{}
			opts := []ServiceOption{
				WithImageProvider(imageProvider),
				WithImageResizer(sleepyResizer{timeout: 1 * time.Second}),
			}
			resizer := NewService(zap.NewNop(), opts...)

			imageProvider.On("GetImage", ctx, url).Return(img, nil)

			out, err := resizer.Resize(ctx, url, params)

			require.Error(t, err)
			assert.EqualError(t, err, ctx.Err().Error())
			assert.Nil(t, out)
		})

		t.Run("with fast resizing", func(t *testing.T) {
			ctx := context.Background()

			imageProvider := &mocks.ImageProvider{}
			resizer := NewService(zap.NewNop(), WithImageProvider(imageProvider))

			imageProvider.On("GetImage", ctx, url).Return(img, nil)

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
