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
	"runtime"
	"sync"
)

type ImageConvert int

const (
	NoConvert ImageConvert = iota
	SRGBToLinear
	BumpToNormal
)

var Threads int

func init() {
	Threads = runtime.NumCPU()
}

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
	fmt.Println("convert image", path.Base(name))
	conv := NewConverter()
	switch mode {
	case NoConvert:
		conv.Add(draw.Src)
	case SRGBToLinear:
		conv.Add(ColorFilter(Ungamma))
	case BumpToNormal:
		conv.Add(BlurFilter{Radius: 1.5, Clamp: true}, SobelFilter{Strength: 1.25, Clamp: true})
	default:
		panic("unknown conversion mode!")
	}
	dst := conv.Apply(src)
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
type Converter struct {
	filters []draw.Drawer
}

func NewConverter() *Converter {
	return &Converter{filters: []draw.Drawer{}}
}

func (c *Converter) Add(filter ...draw.Drawer) {
	c.filters = append(c.filters, filter...)
}

func (c *Converter) Apply(src image.Image) *image.NRGBA {
	var dst *image.NRGBA
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	yslice := max(height/Threads, 16)
	for _, filter := range c.filters {
		dst = image.NewNRGBA(bounds)
		// parallelise across sections of the image
		var wg sync.WaitGroup
		ypos := 0
		for ypos < height {
			wg.Add(1)
			go func(f draw.Drawer, y0 int) {
				defer wg.Done()
				y1 := min(y0+yslice, height)
				f.Draw(dst, image.Rect(0, y0, width, y1), src, image.Pt(0, y0))
			}(filter, ypos)
			ypos += yslice
		}
		wg.Wait()
		src = dst
	}
	return dst
}

// Convert from SRGB to linear color
func Ungamma(x float64) float64 { return math.Pow(x, 2.2) }

// Apply filter to RGB color of each pixel
type ColorFilter func(float64) float64

func (f ColorFilter) Draw(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point) {
	draw.Draw(dst, r, doColorFilter{Image: src, ColorFilter: f}, sp, draw.Src)
}

type doColorFilter struct {
	image.Image
	ColorFilter
}

func (f doColorFilter) ColorModel() color.Model { return color.NRGBAModel }

func (f doColorFilter) At(x, y int) color.Color {
	var c color.NRGBA
	ir, ig, ib, ia := f.Image.At(x, y).RGBA()
	if ia != 0 {
		fa := float64(ia)
		c.R = uint8(0xff * f.ColorFilter(float64(ir)/fa))
		c.G = uint8(0xff * f.ColorFilter(float64(ig)/fa))
		c.B = uint8(0xff * f.ColorFilter(float64(ib)/fa))
		c.A = uint8(ia >> 8)
	}
	return c
}

// Apply sobel filter and convert x offset to red channel and y offset to green channel
type SobelFilter struct {
	Strength float64
	Clamp    bool
}

func (f SobelFilter) Draw(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point) {
	draw.Draw(dst, r, doSobelFilter{Image: src, SobelFilter: f}, sp, draw.Src)
}

type doSobelFilter struct {
	image.Image
	SobelFilter
}

func (f doSobelFilter) ColorModel() color.Model { return color.NRGBAModel }

func (f doSobelFilter) At(x, y int) color.Color {
	t := intensity(f.Image, x, y-1, f.Clamp)
	b := intensity(f.Image, x, y+1, f.Clamp)
	l := intensity(f.Image, x-1, y, f.Clamp)
	r := intensity(f.Image, x+1, y, f.Clamp)
	tl := intensity(f.Image, x-1, y-1, f.Clamp)
	tr := intensity(f.Image, x+1, y-1, f.Clamp)
	bl := intensity(f.Image, x-1, y+1, f.Clamp)
	br := intensity(f.Image, x+1, y+1, f.Clamp)
	dx := (tl + 2*l + bl) - (tr + 2*r + br)
	dy := (bl + 2*b + br) - (tl + 2*t + tr)
	dz := 1 / f.Strength
	norm := 1 / math.Sqrt(dx*dx+dy*dy+dz*dz)
	return color.NRGBA{
		R: uint8(0xff * (0.5 + 0.5*dx*norm)),
		G: uint8(0xff * (0.5 + 0.5*dy*norm)),
		B: uint8(0xff * (0.5 + 0.5*dz*norm)),
		A: 0xff,
	}
}

// Apply Gaussian blur to the image
type BlurFilter struct {
	Radius float64
	Clamp  bool
}

// Box blur as per http://blog.ivank.net/fastest-gaussian-blur.html, algorithm 2
func (f BlurFilter) Draw(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point) {
	for _, box := range boxesForGauss(f.Radius, 3) {
		draw.Draw(dst, r, boxBlur{Image: src, BlurFilter: f, rs: box}, sp, draw.Src)
	}
}

func boxesForGauss(sigma float64, n int) []int {
	wIdeal := math.Sqrt((12*sigma*sigma/float64(n) + 1))
	wl := math.Floor(wIdeal)
	if int(wl)%2 == 0 {
		wl--
	}
	wu := wl + 2
	nf := float64(n)
	mIdeal := (12*sigma*sigma - nf*wl*wl - 4*nf*wl - 3*nf) / (-4*wl - 4)
	m := int(math.Floor(mIdeal + 0.5))
	sizes := make([]int, n)
	for i := range sizes {
		if i < m {
			sizes[i] = int((wl - 1) / 2)
		} else {
			sizes[i] = int((wu - 1) / 2)
		}
	}
	return sizes
}

type boxBlur struct {
	image.Image
	BlurFilter
	rs int
}

func (f boxBlur) ColorModel() color.Model { return color.Gray16Model }

func (f boxBlur) At(x, y int) color.Color {
	val := 0.0
	for iy := y - f.rs; iy < y+f.rs+1; iy++ {
		for ix := x - f.rs; ix < x+f.rs+1; ix++ {
			val += intensity(f.Image, ix, iy, f.Clamp)
		}
	}
	val /= float64((2*f.rs + 1) * (2*f.rs + 1))
	return color.Gray16{uint16(0xffff * val)}
}

func intensity(img image.Image, x, y int, applyClamp bool) float64 {
	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	if applyClamp {
		x = clamp(x, 0, dx-1)
		y = clamp(y, 0, dy-1)
	} else {
		x = (x + dx) % dx
		y = (y + dy) % dy
	}
	ir, ig, ib, ia := img.At(x, y).RGBA()
	return float64(ir+ig+ib) / float64(3*ia)
}

func clamp(x, min, max int) int {
	if x < min {
		x = min
	}
	if x > max {
		x = max
	}
	return x
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
