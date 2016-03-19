package glu

import (
	"fmt"
	"gopkg.in/qml.v1/gl/es2"
	"gopkg.in/qml.v1/gl/glbase"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
)

const gamma = 2.2

// Global texture id number
var textureId int32

func nextId() int32 {
	n := textureId
	textureId++
	return n
}

// Texture interface type
type Texture interface {
	Id() int32
	Activate()
	Dims() []int
}

type Texture2D struct{ *textureBase }

// NewTexture2D creates a new 2D opengl texture.
func NewTexture2D(wrap int32) Texture2D {
	t := &textureBase{
		id:  nextId(),
		typ: GL.TEXTURE_2D,
		tex: gl.GenTextures(1),
	}
	gl.BindTexture(GL.TEXTURE_2D, t.tex[0])
	gl.TexParameteri(GL.TEXTURE_2D, GL.TEXTURE_WRAP_S, wrap)
	gl.TexParameteri(GL.TEXTURE_2D, GL.TEXTURE_WRAP_T, wrap)
	gl.TexParameteri(GL.TEXTURE_2D, GL.TEXTURE_MIN_FILTER, GL.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(GL.TEXTURE_2D, GL.TEXTURE_MAG_FILTER, GL.LINEAR)
	gl.BindTexture(t.typ, 0)
	CheckError()
	return Texture2D{t}
}

// SetImage loads an image, if srgba is set then it is converted to linear RGB space.
func (t Texture2D) SetImage(img image.Image, srgba bool) Texture2D {
	gl.BindTexture(GL.TEXTURE_2D, t.tex[0])
	pix := ToNRGBA(img, srgba)
	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	t.dims = []int{dx, dy}
	gl.TexImage2D(GL.TEXTURE_2D, 0, GL.RGBA, dx, dy, 0, GL.RGBA, GL.UNSIGNED_BYTE, pix)
	gl.GenerateMipmap(GL.TEXTURE_2D)
	gl.BindTexture(GL.TEXTURE_2D, 0)
	CheckError()
	return t
}

type TextureCube struct{ *textureBase }

// NewTextureCube creates a new cubemap texture.
func NewTextureCube() TextureCube {
	t := &textureBase{
		id:  nextId(),
		typ: GL.TEXTURE_CUBE_MAP,
		tex: gl.GenTextures(1),
	}
	gl.BindTexture(t.typ, t.tex[0])
	gl.TexParameteri(t.typ, GL.TEXTURE_WRAP_S, GL.CLAMP_TO_EDGE)
	gl.TexParameteri(t.typ, GL.TEXTURE_WRAP_T, GL.CLAMP_TO_EDGE)
	gl.TexParameteri(t.typ, GL.TEXTURE_MIN_FILTER, GL.LINEAR)
	gl.TexParameteri(t.typ, GL.TEXTURE_MAG_FILTER, GL.LINEAR)
	gl.BindTexture(t.typ, 0)
	CheckError()
	return TextureCube{t}
}

// SetImage loads an image, if srgba is set then it is converted to linear RGB space.
// The index is the number of the image in the cubemap.
func (t TextureCube) SetImage(img image.Image, srgba bool, index int) TextureCube {
	gl.BindTexture(GL.TEXTURE_CUBE_MAP, t.tex[0])
	pix := ToNRGBA(img, srgba)
	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	t.dims = []int{dx, dy}
	target := GL.TEXTURE_CUBE_MAP_POSITIVE_X + glbase.Enum(index)
	gl.TexImage2D(target, 0, GL.RGBA, dx, dy, 0, GL.RGBA, GL.UNSIGNED_BYTE, pix)
	gl.BindTexture(GL.TEXTURE_CUBE_MAP, 0)
	CheckError()
	return t
}

type Texture3D struct{ *textureBase }

// NewTexture3D creates a 3D texture mapping
func NewTexture3D() Texture3D {
	t := &textureBase{
		id:  nextId(),
		typ: GL.TEXTURE_2D,
		tex: gl.GenTextures(1),
	}
	gl.BindTexture(GL.TEXTURE_2D, t.tex[0])
	gl.TexParameteri(GL.TEXTURE_2D, GL.TEXTURE_WRAP_S, GL.CLAMP_TO_EDGE)
	gl.TexParameteri(GL.TEXTURE_2D, GL.TEXTURE_WRAP_T, GL.CLAMP_TO_EDGE)
	gl.TexParameteri(GL.TEXTURE_2D, GL.TEXTURE_MIN_FILTER, GL.NEAREST)
	gl.TexParameteri(GL.TEXTURE_2D, GL.TEXTURE_MAG_FILTER, GL.NEAREST)
	gl.BindTexture(t.typ, 0)
	CheckError()
	return Texture3D{t}
}

// SetImage loads an image, if srgba is set then it is converted to linear RGB space
// If dims is required to set the x,y,z mapping.
func (t Texture3D) SetImage(img image.Image, srgba bool, dims []int) Texture3D {
	gl.BindTexture(GL.TEXTURE_2D, t.tex[0])
	pix := ToNRGBA(img, srgba)
	t.dims = dims
	gl.TexImage2D(GL.TEXTURE_2D, 0, GL.RGBA, dims[0], dims[1]*dims[2], 0, GL.RGBA, GL.UNSIGNED_BYTE, pix)
	gl.BindTexture(GL.TEXTURE_2D, 0)
	CheckError()
	return t
}

// base type for all textures
type textureBase struct {
	id   int32
	typ  glbase.Enum
	tex  []glbase.Texture
	dims []int
}

func (t *textureBase) Id() int32 {
	return t.id
}

func (t *textureBase) Activate() {
	gl.ActiveTexture(GL.TEXTURE0 + glbase.Enum(t.id))
	gl.BindTexture(t.typ, t.tex[0])
	if Debug {
		CheckError()
	}
}

func (t *textureBase) Dims() []int {
	return t.dims
}

// PNGImage loads a PNG format image from a file
func PNGImage(file string) (image.Image, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// Convert an image from SRGBA to linear NRGBA colour space, returns pixel data
func ToNRGBA(img image.Image, hasGamma bool) []uint8 {
	if hasGamma == false {
		switch t := img.(type) {
		case *image.NRGBA:
			return t.Pix
		case *image.RGBA:
			return t.Pix
		}
	}
	bounds := img.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	if hasGamma {
		fmt.Println("converting image to linear RGB space")
		img = &toNRGBA{img}
	}
	draw.Draw(dst, bounds, img, image.ZP, draw.Src)
	return dst.Pix
}

// image converter to linearise image which is in srgba format
type toNRGBA struct {
	image.Image
}

func (t *toNRGBA) ColorModel() color.Model {
	return color.NRGBAModel
}

func (t *toNRGBA) At(x, y int) color.Color {
	ir, ig, ib, ia := t.Image.At(x, y).RGBA()
	if ia == 0 {
		return color.NRGBA{}
	}
	fa := float64(ia)
	return color.NRGBA{
		R: uint8(0xff * math.Pow(float64(ir)/fa, gamma)),
		G: uint8(0xff * math.Pow(float64(ig)/fa, gamma)),
		B: uint8(0xff * math.Pow(float64(ib)/fa, gamma)),
		A: uint8(ia >> 8),
	}
}
