package main

import (
	"image/color"
	"math"
)

const Un = 0.2009
const Vn = 0.461
const Yn = 1.0

const MAX_UINT32 = 65535

func clampUint(v uint32) uint32 {
	if v < 0 {
		v = 0
	} else if v > MAX_UINT32 {
		v = MAX_UINT32
	}

	return v
}

type XYZ struct {
	X, Y, Z float64
}

func (xyz XYZ) RGBA() (uint32, uint32, uint32, uint32) {

	//multiply the xyz matrix by the reverse sRGB matrix. See http://en.wikipedia.org/wiki/SRGB
	r := 3.2406*xyz.X + -1.5372*xyz.Y + -0.4986*xyz.Z
	g := -0.9689*xyz.X + 1.8758*xyz.Y + 0.0415*xyz.Z
	b := 0.0557*xyz.X + -0.2040*xyz.Y + 1.0570*xyz.Z

	return clampUint(uint32(r * MAX_UINT32)), clampUint(uint32(g * MAX_UINT32)), clampUint(uint32(b * MAX_UINT32)), MAX_UINT32
}

var XYZModel color.Model = color.ModelFunc(xyzModel)

func xyzModel(c color.Color) color.Color {
	if _, ok := c.(XYZ); ok {
		return c
	}

	cr, cg, cb, _ := c.RGBA()
	//Normalize RGB to 0-1
	r := float64(cr) / MAX_UINT32
	g := float64(cg) / MAX_UINT32
	b := float64(cb) / MAX_UINT32

	//multiply the rgb matrix by the sRGB matrix. See http://en.wikipedia.org/wiki/SRGB
	return XYZ{
		0.4124*r + 0.3576*g + 0.1805*b,
		0.2126*r + 0.7152*g + 0.0722*b,
		0.0193*r + 0.1192*g + 0.9505*b,
	}
}

type LUV struct {
	L, U, V float64
}

func (luv LUV) RGBA() (uint32, uint32, uint32, uint32) {
	var y float64

	if luv.L <= 8 {
		y = Yn * luv.L * math.Pow(3/29, 3)
	} else {
		y = Yn * math.Pow((luv.L+16)/116, 3)
	}

	u := (luv.U / (13 * luv.L)) + Un
	v := (luv.V / (13 * luv.L)) + Vn

	x := y * ((9 * u) / (4 * v))
	z := y * ((12 - (3 * u) - (20 * v)) / (4 * v))

	xyz := XYZ{x, y, z}

	return xyz.RGBA()
}

var LUVModel color.Model = color.ModelFunc(luvModel)

func luvModel(c color.Color) color.Color {
	if _, ok := c.(LUV); ok {
		return c
	}

	//Make sure we're in XYZ color space
	if _, ok := c.(XYZ); !ok {
		c = XYZModel.Convert(c)
	}

	//For all the magic values used here look at: http://en.wikipedia.org/wiki/CIELUV
	yyn := c.(XYZ).Y / Yn

	var L, u, v float64

	if math.Pow(6/29, 3) >= yyn {
		L = math.Pow(29/3, 3) * yyn
	} else {
		L = 116*math.Cbrt(yyn) - 16
	}

	d := c.(XYZ).X + (15 * c.(XYZ).Y) + (3 * c.(XYZ).Z)

	if d != 0 {
		u = (4 * c.(XYZ).X) / d
		v = (9 * c.(XYZ).Y) / d
	}

	return LUV{
		L,
		13 * L * (u - Un),
		13 * L * (v - Vn),
	}
}

type LCH struct {
	L, C, H float64
}

func (lch LCH) RGBA() (uint32, uint32, uint32, uint32) {
	luv := LUV{
		lch.L,
		lch.C * math.Cos(lch.H),
		lch.C * math.Sin(lch.H),
	}

	return luv.RGBA()
}

// Computes the Deta E (color distance) of two colors
// http://en.wikipedia.org/wiki/Color_difference#CIEDE2000
func (lch1 LCH) Distance(lch2 LCH) float64 {

	hd := lch2.H - lch1.H
	Hp := lch2.H + lch1.H
	if math.Abs(hd) > math.Pi {
		Hp += math.Pi
		if lch2.H <= lch1.H {
			hd += 2 * math.Pi
		} else {
			hd -= 2 * math.Pi
		}
	}

	Hp /= 2

	T := 1 - 0.17*math.Cos(Hp-0.5235987756) + 0.24*math.Cos(2*Hp) + 0.32*math.Cos(3*Hp+0.10471975512) - 0.2*math.Cos(4*Hp-1.0995574288)

	lp := (lch1.L + lch2.L) / 2
	cp := (lch1.C + lch2.C) / 2

	Sh := 1 + 0.015*cp*T
	Sc := 1 + 0.045*cp
	Sl := 1 + (0.015*math.Pow(lp-50, 2))/math.Sqrt(20+math.Pow(lp-50, 2))

	dtheta := 1.0471975512 * math.Exp(-1*math.Pow((Hp-4.799655443)/0.436332313, 2))

	Rc := math.Sqrt(math.Pow(cp, 7) / (math.Pow(cp, 7) + 6103515625))
	Rt := -2 * Rc * math.Sin(dtheta)

	dL := lch2.L - lch1.L
	dC := lch2.C - lch1.C
	dH := math.Sqrt(lch1.C*lch2.C) * math.Sin(hd/2)

	return math.Sqrt(math.Pow(dL/Sl, 2) + math.Pow(dC/Sc, 2) + math.Pow(dH/Sh, 2) + Rt*(dC/Sc)*(dH/Sh))
}

var LCHModel color.Model = color.ModelFunc(lchModel)

func lchModel(c color.Color) color.Color {
	if _, ok := c.(LCH); ok {
		return c
	}

	if _, ok := c.(LUV); !ok {
		c = LUVModel.Convert(c)
	}

	h := math.Atan2(c.(LUV).V, c.(LUV).U)
	circ := 2 * math.Pi

	if h < 0 {
		h += circ
	} else if h >= circ {
		h -= circ
	}

	return LCH{
		c.(LUV).L,
		math.Sqrt(c.(LUV).U*c.(LUV).U + c.(LUV).V*c.(LUV).V),
		h,
	}
}
