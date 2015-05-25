package main

import (
	"image"
	"image/draw"
)

func ScaleImage(src image.Image, w int, h int) *image.NRGBA {

	nrgba := image.NewNRGBA(src.Bounds())
	draw.Draw(nrgba, nrgba.Bounds(), src, image.ZP, draw.Src)

	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	sb := nrgba.Bounds()

	xr := int((sb.Dx() << 16) / w)
	yr := int((sb.Dy() << 16) / h)

	ch := 4 //R, G, B, A channals

	for i := 0; i < (h * ch); i += ch {
		for j := 0; j < (w * ch); j += ch {
			x := ((j * xr) >> 16)
			y := ((i * yr) >> 16)

			dst.Pix[(i*w)+j+0] = nrgba.Pix[(y*sb.Dx())+x+0]
			dst.Pix[(i*w)+j+1] = nrgba.Pix[(y*sb.Dx())+x+1]
			dst.Pix[(i*w)+j+2] = nrgba.Pix[(y*sb.Dx())+x+2]
			dst.Pix[(i*w)+j+3] = nrgba.Pix[(y*sb.Dx())+x+3]
		}
	}
	return dst
}
