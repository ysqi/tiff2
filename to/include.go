package to

import (
	"image"
	"image/jpeg"
	"image/png"
	"io"
)

type convert func(to io.Writer, source image.Image) error

func init() {
	Reg("jpg", saveAsJPEG)
	Reg("png", saveAsPNG)
	Reg("jpeg", saveAsPNG)
}

func saveAsJPEG(to io.Writer, source image.Image) error {
	return jpeg.Encode(to, source, nil)
}

func saveAsPNG(to io.Writer, source image.Image) error {
	return png.Encode(to, source)
}
