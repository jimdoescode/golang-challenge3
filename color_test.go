package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"testing"
)

func eq(c1, c2 uint32) bool {
	m := 2 //Precision loss

	return c1 == c2 || int(math.Abs(float64(c1)-float64(c2))) <= m
}

func TestXYZWhiteConversion(t *testing.T) {
	c := color.White
	xyz := XYZModel.Convert(c)
	r, g, b, a := xyz.RGBA()
	cr, cg, cb, ca := c.RGBA()

	if !eq(r, cr) || !eq(g, cg) || !eq(b, cb) || !eq(a, ca) {
		t.Fatalf("Expecting color RGBA(%d, %d, %d, %d) got RGBA(%d, %d, %d, %d)", cr, cg, cb, ca, r, g, b, a)
	}
}

func TestXYZBlackConversion(t *testing.T) {
	c := color.Black
	xyz := XYZModel.Convert(c)
	r, g, b, a := xyz.RGBA()
	cr, cg, cb, ca := c.RGBA()

	if !eq(r, cr) || !eq(g, cg) || !eq(b, cb) || !eq(a, ca) {
		t.Fatalf("Expecting color RGBA(%d, %d, %d, %d) got RGBA(%d, %d, %d, %d)", cr, cg, cb, ca, r, g, b, a)
	}
}

func TestXYZColorConversion(t *testing.T) {
	c := color.RGBA{186, 85, 211, 255}
	xyz := XYZModel.Convert(c).(XYZ)
	r, g, b, a := xyz.RGBA()
	cr, cg, cb, ca := c.RGBA()

	if !eq(r, cr) || !eq(g, cg) || !eq(b, cb) || !eq(a, ca) {
		t.Fatalf("Expecting color RGBA(%d, %d, %d, %d) got RGBA(%d, %d, %d, %d)", cr, cg, cb, ca, r, g, b, a)
	}
}

func TestLUVColorConversion(t *testing.T) {
	c := XYZModel.Convert(color.RGBA{186, 85, 211, 255}).(XYZ)
	luv := LUVModel.Convert(c).(LUV)
	r, g, b, a := luv.RGBA()
	cr, cg, cb, ca := c.RGBA()

	if !eq(r, cr) || !eq(g, cg) || !eq(b, cb) || !eq(a, ca) {
		t.Fatalf("Expecting color RGBA(%d, %d, %d, %d) got RGBA(%d, %d, %d, %d)", cr, cg, cb, ca, r, g, b, a)
	}
}

func TestLCHColorConversion(t *testing.T) {
	c := color.RGBA{186, 85, 211, 255}
	lch := LCHModel.Convert(c)
	r, g, b, a := lch.RGBA()
	cr, cg, cb, ca := c.RGBA()

	if !eq(r, cr) || !eq(g, cg) || !eq(b, cb) || !eq(a, ca) {
		t.Fatalf("Expecting color RGBA(%d, %d, %d, %d) got RGBA(%d, %d, %d, %d)", cr, cg, cb, ca, r, g, b, a)
	}
}

func TestLCHColorDistance(t *testing.T) {
	lch1 := LCHModel.Convert(color.Black).(LCH)
	lch2 := LCHModel.Convert(color.White).(LCH)

	b2w := lch1.Distance(lch2)
	w2b := lch2.Distance(lch1)

	if b2w != w2b {
		t.Fatalf("Black to white distance (%d) does not match white to black distance (%d)", b2w, w2b)
	}

	if !eq(uint32(b2w), 100) {
		t.Fatalf("Black to white distance should be ~100, %f returned", b2w)
	}

	lch1 = LCHModel.Convert(color.RGBA{186, 85, 211, 255}).(LCH)
	lch2 = LCHModel.Convert(color.RGBA{186, 85, 211, 255}).(LCH)
	c2c := lch1.Distance(lch2)

	if c2c != 0 && !math.IsNaN(c2c) {
		t.Fatalf("Expecting duplicate color distance of 0 got %f instead", c2c)
	}

	//These values were causing NaN distance calc due to some bad math.
	lch1 = LCH{99.586813, 9.924174, 1.957027}
	lch2 = LCH{66.873655, 6.755820, 1.369488}
	c2c = lch1.Distance(lch2)

	if math.IsNaN(c2c) {
		t.Fatalf("Unexpected NaN value for standard LCH values.")
	}
}

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
