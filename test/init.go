package test

import (
	"bytes"
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func RootDir(t *testing.T, level int) string {
	dir, err := os.Getwd()
	require.NoError(t, err)
	for i := 0; i < level; i++ {
		dir = path.Dir(dir)
	}
	return dir
}

func SampleImage(t *testing.T, level int) image.Image {
	const imagePath = "test/testdata/nature.jpg"

	file, err := ioutil.ReadFile(path.Join(RootDir(t, level), imagePath))
	require.NoError(t, err)
	img, err := jpeg.Decode(bytes.NewReader(file))
	require.NoError(t, err)
	return img
}
