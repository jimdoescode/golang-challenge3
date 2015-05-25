package main

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestColorAverage(t *testing.T) {
	checkered := image.NewNRGBA(image.Rect(0, 0, 64, 64))

	draw.Draw(checkered, image.Rect(0, 0, 32, 32), image.Black, image.ZP, draw.Src)
	draw.Draw(checkered, image.Rect(0, 32, 32, 64), image.White, image.ZP, draw.Src)

	draw.Draw(checkered, image.Rect(32, 32, 64, 64), image.Black, image.ZP, draw.Src)
	draw.Draw(checkered, image.Rect(32, 0, 64, 32), image.White, image.ZP, draw.Src)

	c := AverageImageColor(checkered)
	r32, g32, b32, _ := c.RGBA()

	y := uint8(127)
	r := uint8(r32 / 0x101)
	g := uint8(g32 / 0x101)
	b := uint8(b32 / 0x101)

	if r != y || g != y || b != y {
		t.Fatalf("Expecting RGB values of %d, found R(%d) G(%d) B(%d)", y, r, g, b)
	}

	white := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	draw.Draw(white, image.Rect(0, 0, 64, 64), image.White, image.ZP, draw.Src)

	c = AverageImageColor(white)
	r32, g32, b32, _ = c.RGBA()

	y = uint8(255)
	r = uint8(r32 / 0x101)
	g = uint8(g32 / 0x101)
	b = uint8(b32 / 0x101)

	if r != y || g != y || b != y {
		t.Fatalf("Expecting white RGB values of %d, found R(%d) G(%d) B(%d)", y, r, g, b)
	}
}

func TestImageResize(t *testing.T) {
	checkered := image.NewNRGBA(image.Rect(0, 0, 64, 64))

	draw.Draw(checkered, image.Rect(0, 0, 32, 32), image.Black, image.ZP, draw.Src)
	draw.Draw(checkered, image.Rect(0, 32, 32, 64), image.White, image.ZP, draw.Src)

	draw.Draw(checkered, image.Rect(32, 32, 64, 64), image.Black, image.ZP, draw.Src)
	draw.Draw(checkered, image.Rect(32, 0, 64, 32), image.White, image.ZP, draw.Src)

	i := ScaleImage(checkered, 32, 32)

	if i.Bounds().Dx() != 32 || i.Bounds().Dy() != 32 {
		t.Fatalf("Expecting 32 by 32 image. Found %d by %d image.", i.Bounds().Dx(), i.Bounds().Dy())
	}

	b1 := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	w1 := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	b2 := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	w2 := image.NewNRGBA(image.Rect(0, 0, 16, 16))

	draw.Draw(b1, b1.Bounds(), i, image.ZP, draw.Src)
	draw.Draw(w1, w1.Bounds(), i, image.Pt(16, 0), draw.Src)
	draw.Draw(b2, b2.Bounds(), i, image.Pt(16, 16), draw.Src)
	draw.Draw(w2, w2.Bounds(), i, image.Pt(0, 16), draw.Src)

	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 0}

	if !cc(b1, black) || !cc(b2, black) || !cc(w1, white) || !cc(w2, white) {
		t.Fatal("Did not find correct colors on scaled image.")
	}
}

func cc(i image.Image, c color.Color) bool {
	r, g, b, _ := c.RGBA()
	for x := i.Bounds().Min.X; x < i.Bounds().Max.X; x++ {
		for y := i.Bounds().Min.Y; y < i.Bounds().Max.Y; y++ {
			rp, gp, bp, _ := i.At(x, y).RGBA()
			if rp != r || gp != g || bp != b {
				return false
			}
		}
	}

	return true
}
