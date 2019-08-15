package resizer

import (
	"image"

	"github.com/disintegration/imaging"
)

type Resizer struct {
}

func (r Resizer) Resize(img image.Image, params Params) (image.Image, error) {
	return imaging.Resize(img, params.Width, params.Height, imaging.Lanczos), nil
}


