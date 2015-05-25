package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// Scales an image to the specified dimensions using nearest neighbor.
func ScaleImage(src image.Image, w int, h int) *image.NRGBA {

	nrgba := image.NewNRGBA(src.Bounds())
	draw.Draw(nrgba, nrgba.Bounds(), src, image.ZP, draw.Src)

	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	sb := nrgba.Bounds()

	xr := float64(sb.Dx() / w)
	yr := float64(sb.Dy() / h)

	ch := 4 //R, G, B, A channels

	for i := 0; i < (h * ch); i += ch {
		for j := 0; j < (w * ch); j += ch {
			x := int(math.Floor(float64(j) * xr))
			y := int(math.Floor(float64(i) * yr))

			dst.Pix[(i*w)+j+0] = nrgba.Pix[(y*sb.Dx())+x+0]
			dst.Pix[(i*w)+j+1] = nrgba.Pix[(y*sb.Dx())+x+1]
			dst.Pix[(i*w)+j+2] = nrgba.Pix[(y*sb.Dx())+x+2]
			dst.Pix[(i*w)+j+3] = nrgba.Pix[(y*sb.Dx())+x+3]
		}
	}
	return dst
}

// Averages the colors of an image and returns the LCH color of the image
func AverageImageColor(i image.Image) LCH {
	var r, g, b uint32

	bounds := i.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pr, pg, pb, _ := i.At(x, y).RGBA()

			r += pr
			g += pg
			b += pb
		}
	}

	d := uint32(bounds.Dy() * bounds.Dx())

	r /= d
	g /= d
	b /= d

	c := color.NRGBA{uint8(r / 0x101), uint8(g / 0x101), uint8(b / 0x101), 255}

	return LCHModel.Convert(c).(LCH)
}
