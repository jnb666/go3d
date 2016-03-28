// Package img provides image conversion functions.
package img

import (
	"crypto/md5"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"path"
)

type ImageConvert int

const (
	NoConvert ImageConvert = iota
	SRGBToLinear
	BumpToNormal
)

// Get an image pixels in NRGBA format
func Decode(r io.Reader, mode ImageConvert) ([]uint8, image.Rectangle, error) {
	cache, name, tempfile := cachedFile(r, mode)
	if cache != nil {
		r = cache
		mode = NoConvert
	}
	src, _, err := image.Decode(r)
	if err != nil {
		return nil, image.Rectangle{}, err
	}
	bounds := src.Bounds()
	// already in correct format?
	if mode == NoConvert {
		switch t := src.(type) {
		case *image.NRGBA:
			return t.Pix, bounds, nil
		case *image.RGBA:
			return t.Pix, bounds, nil
		}
	}
	// apply conversions
	fmt.Println("convert image", name)
	conv := newConverter()
	switch mode {
	case NoConvert:
		conv.add(draw.Src)
	case SRGBToLinear:
		conv.add(colorFilter{fn: ungamma})
	case BumpToNormal:
		conv.add(blurFilter{radius: 1}, sobelFilter{strength: 1})
	default:
		panic("unknown conversion mode!")
	}
	dst := conv.apply(src)
	// save converted image
	if tempfile != "" {
		if out, err := os.Create(tempfile); err == nil {
			png.Encode(out, dst)
			out.Close()
		}
	}
	return dst.Pix, bounds, nil
}

// do we have a cached copy of the converted image?
func cachedFile(r io.Reader, mode ImageConvert) (*os.File, string, string) {
	f, ok := r.(*os.File)
	if !ok {
		return nil, "", ""
	}
	name := f.Name()
	if !path.IsAbs(name) {
		cwd, _ := os.Getwd()
		name = path.Join(cwd, name)
	}
	tempfile := fmt.Sprintf("%s/%x_%d.png", os.TempDir(), md5.Sum([]byte(name)), mode)
	tmp, err := os.Open(tempfile)
	if err != nil {
		return nil, name, tempfile
	}
	fstat, _ := f.Stat()
	tstat, _ := tmp.Stat()
	if !tstat.ModTime().After(fstat.ModTime()) {
		return nil, name, tempfile
	}
	return tmp, name, ""
}

// Apply a series of image compositing functions
type converter struct {
	filters []draw.Drawer
}

func newConverter() *converter {
	return &converter{filters: []draw.Drawer{}}
}

func (c *converter) add(filter ...draw.Drawer) {
	c.filters = append(c.filters, filter...)
}

func (c *converter) apply(src image.Image) *image.NRGBA {
	var dst *image.NRGBA
	bounds := src.Bounds()
	for _, filter := range c.filters {
		dst = image.NewNRGBA(bounds)
		filter.Draw(dst, bounds, src, image.ZP)
		src = dst
	}
	return dst
}

// Apply filter to RGB color of each pixel
type colorFilter struct{ fn func(float64) float64 }

func ungamma(x float64) float64 {
	return math.Pow(x, 2.2)
}

func (f colorFilter) Draw(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point) {
	draw.Draw(dst, r, doColorFilter{Image: src, colorFilter: f}, sp, draw.Src)
}

type doColorFilter struct {
	image.Image
	colorFilter
}

func (f doColorFilter) ColorModel() color.Model { return color.NRGBAModel }

func (f doColorFilter) At(x, y int) color.Color {
	var c color.NRGBA
	ir, ig, ib, ia := f.Image.At(x, y).RGBA()
	if ia != 0 {
		fa := float64(ia)
		c.R = uint8(0xff * f.fn(float64(ir)/fa))
		c.G = uint8(0xff * f.fn(float64(ig)/fa))
		c.B = uint8(0xff * f.fn(float64(ib)/fa))
		c.A = uint8(ia >> 8)
	}
	return c
}

// Apply sobel filter and convert x offset to red channel and y offset to green channel
type sobelFilter struct{ strength float64 }

func (f sobelFilter) Draw(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point) {
	draw.Draw(dst, r, doSobelFilter{Image: src, sobelFilter: f, dx: r.Dx(), dy: r.Dy()}, sp, draw.Src)
}

type doSobelFilter struct {
	image.Image
	sobelFilter
	dx, dy int
}

func (f doSobelFilter) ColorModel() color.Model { return color.NRGBAModel }

func (f doSobelFilter) At(x, y int) color.Color {
	var c color.NRGBA
	tl, tt, tr := f.level(x-1, y-1), f.level(x, y-1), f.level(x+1, y-1)
	ll, rr := f.level(x-1, y), f.level(x+1, y)
	bl, bb, br := f.level(x-1, y+1), f.level(x, y+1), f.level(x+1, y+1)
	dx := (tr + 2*rr + br) - (tl + 2*ll + bl)
	dy := (bl + 2*bb + br) - (tl + 2*tt + tr)
	dz := 1 / f.strength
	norm := 1 / math.Sqrt(dx*dx+dy*dy+dz*dz)
	c.R = uint8(0xff * (0.5 + 0.5*dx*norm))
	c.G = uint8(0xff * (0.5 + 0.5*dy*norm))
	c.B = uint8(0xff * (0.5 + 0.5*dz*norm))
	c.A = 0xff
	return c
}

func (f doSobelFilter) level(x, y int) float64 {
	ir, ig, ib, ia := f.Image.At((x+f.dx)%f.dx, (y+f.dy)%f.dy).RGBA()
	return float64(ir+ig+ib) / float64(3*ia)
}

// Apply Gaussian blur to the image
type blurFilter struct{ radius float64 }

func (f blurFilter) Draw(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point) {
	rs := int(math.Ceil(f.radius * 2.57))
	draw.Draw(dst, r, doBlurFilter{Image: src, blurFilter: f, rs: rs, dx: r.Dx(), dy: r.Dy()}, sp, draw.Src)
}

type doBlurFilter struct {
	image.Image
	blurFilter
	rs, dx, dy int
}

func (f doBlurFilter) ColorModel() color.Model { return color.NRGBAModel }

func (f doBlurFilter) At(x, y int) color.Color {
	var c color.NRGBA
	val := 0.0
	wsum := 0.0
	rsq := 2 * f.radius * f.radius
	for iy := y - f.rs; iy < y+f.rs+1; iy++ {
		for ix := x - f.rs; ix < x+f.rs+1; ix++ {
			dsq := float64((ix-x)*(ix-x) + (iy-y)*(iy-y))
			weight := math.Exp(-dsq/rsq) / (math.Pi * rsq)
			val += f.level(ix, iy) * weight
			wsum += weight
		}
	}
	val = math.Min(val/wsum, 1)
	c.R = uint8(0xff * val)
	c.G = uint8(0xff * val)
	c.B = uint8(0xff * val)
	c.A = 0xff
	return c
}

func (f doBlurFilter) level(x, y int) float64 {
	ir, ig, ib, ia := f.Image.At((x+f.dx)%f.dx, (y+f.dy)%f.dy).RGBA()
	return float64(ir+ig+ib) / float64(3*ia)
}
