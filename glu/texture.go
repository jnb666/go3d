package glu

import (
	"fmt"
	"gopkg.in/qml.v1/gl/es2"
	"gopkg.in/qml.v1/gl/glbase"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"path"
	"strings"
)

type ImageFormat int

const (
	UnknownFormat ImageFormat = iota
	PngFormat
	JpegFormat
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

// NewTexture2D creates a new 2D opengl texture. If srgba is set then it is converted to linear RGB space.
// If clamp is set then clamp to edge, else will wrap texture.
func NewTexture2D(clamp, srgba bool) Texture2D {
	t := &textureBase{
		id:    nextId(),
		typ:   GL.TEXTURE_2D,
		tex:   gl.GenTextures(1),
		srgba: srgba,
	}
	gl.BindTexture(GL.TEXTURE_2D, t.tex[0])
	if clamp {
		gl.TexParameteri(GL.TEXTURE_2D, GL.TEXTURE_WRAP_S, GL.CLAMP_TO_EDGE)
		gl.TexParameteri(GL.TEXTURE_2D, GL.TEXTURE_WRAP_T, GL.CLAMP_TO_EDGE)
	} else {
		gl.TexParameteri(GL.TEXTURE_2D, GL.TEXTURE_WRAP_S, GL.REPEAT)
		gl.TexParameteri(GL.TEXTURE_2D, GL.TEXTURE_WRAP_T, GL.REPEAT)
	}
	gl.TexParameteri(GL.TEXTURE_2D, GL.TEXTURE_MIN_FILTER, GL.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(GL.TEXTURE_2D, GL.TEXTURE_MAG_FILTER, GL.LINEAR)
	gl.BindTexture(t.typ, 0)
	CheckError()
	return Texture2D{t}
}

// SetImageFile loads an image from a file
func (t Texture2D) SetImageFile(file string) (Texture2D, error) {
	r, err := os.Open(file)
	if err != nil {
		return t, err
	}
	defer r.Close()
	t.textureBase.name = file
	return t.SetImage(r, GetFormat(file))
}

// SetImage loads an image from an io.Reader
func (t Texture2D) SetImage(r io.Reader, format ImageFormat) (Texture2D, error) {
	pix, bounds, err := getImage(r, format, t.srgba, t.name)
	if err != nil {
		return t, err
	}
	t.dims = []int{bounds.Dx(), bounds.Dy()}
	gl.BindTexture(GL.TEXTURE_2D, t.tex[0])
	gl.TexImage2D(GL.TEXTURE_2D, 0, GL.RGBA, t.dims[0], t.dims[1], 0, GL.RGBA, GL.UNSIGNED_BYTE, pix)
	gl.GenerateMipmap(GL.TEXTURE_2D)
	gl.BindTexture(GL.TEXTURE_2D, 0)
	CheckError()
	return t, nil
}

type TextureCube struct{ *textureBase }

// NewTextureCube creates a new cubemap texture. If srgba is set then it is converted to linear RGB space.
func NewTextureCube(srgba bool) TextureCube {
	t := &textureBase{
		id:    nextId(),
		typ:   GL.TEXTURE_CUBE_MAP,
		tex:   gl.GenTextures(1),
		srgba: srgba,
	}
	gl.BindTexture(t.typ, t.tex[0])
	gl.TexParameteri(t.typ, GL.TEXTURE_WRAP_S, GL.CLAMP_TO_EDGE)
	gl.TexParameteri(t.typ, GL.TEXTURE_WRAP_T, GL.CLAMP_TO_EDGE)
	gl.TexParameteri(t.typ, GL.TEXTURE_MIN_FILTER, GL.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(t.typ, GL.TEXTURE_MAG_FILTER, GL.LINEAR)
	gl.BindTexture(t.typ, 0)
	CheckError()
	return TextureCube{t}
}

// SetImageFile loads an image from a file.  The index is the number of the image in the cubemap.
func (t TextureCube) SetImageFile(file string, index int) (TextureCube, error) {
	r, err := os.Open(file)
	if err != nil {
		return t, err
	}
	defer r.Close()
	t.textureBase.name = file
	return t.SetImage(r, GetFormat(file), index)
}

// SetImage loads an image, if srgba is set then it is converted to linear RGB space.
// The index is the number of the image in the cubemap.
func (t TextureCube) SetImage(r io.Reader, format ImageFormat, index int) (TextureCube, error) {
	pix, bounds, err := getImage(r, format, t.srgba, t.name)
	if err != nil {
		return t, err
	}
	t.dims = []int{bounds.Dx(), bounds.Dy()}
	gl.BindTexture(GL.TEXTURE_CUBE_MAP, t.tex[0])
	target := GL.TEXTURE_CUBE_MAP_POSITIVE_X + glbase.Enum(index)
	gl.TexImage2D(target, 0, GL.RGBA, t.dims[0], t.dims[1], 0, GL.RGBA, GL.UNSIGNED_BYTE, pix)
	gl.GenerateMipmap(GL.TEXTURE_CUBE_MAP)
	gl.BindTexture(GL.TEXTURE_CUBE_MAP, 0)
	CheckError()
	return t, nil
}

type Texture3D struct{ *textureBase }

// NewTexture3D creates a 3D texture mapping.
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

// SetImageFile loads an image from a file
func (t Texture3D) SetImageFile(file string, dims []int) (Texture3D, error) {
	r, err := os.Open(file)
	if err != nil {
		return t, err
	}
	defer r.Close()
	t.textureBase.name = file
	return t.SetImage(r, GetFormat(file), dims)
}

// SetImage loads an image. dims is required to set the x,y,z mapping.
func (t Texture3D) SetImage(r io.Reader, format ImageFormat, dims []int) (Texture3D, error) {
	pix, _, err := getImage(r, format, t.srgba, t.name)
	if err != nil {
		return t, err
	}
	t.dims = dims
	gl.BindTexture(GL.TEXTURE_2D, t.tex[0])
	gl.TexImage2D(GL.TEXTURE_2D, 0, GL.RGBA, dims[0], dims[1]*dims[2], 0, GL.RGBA, GL.UNSIGNED_BYTE, pix)
	gl.BindTexture(GL.TEXTURE_2D, 0)
	CheckError()
	return t, nil
}

// base type for all textures
type textureBase struct {
	id    int32
	typ   glbase.Enum
	tex   []glbase.Texture
	dims  []int
	srgba bool
	name  string
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

// derive image format from file extension
func GetFormat(name string) ImageFormat {
	switch path.Ext(name) {
	case ".png":
		return PngFormat
	case ".jpg", ".jpeg":
		return JpegFormat
	}
	return UnknownFormat
}

// get an image in NRGBA format
func getImage(r io.Reader, format ImageFormat, hasGamma bool, name string) (pix []uint8, bounds image.Rectangle, err error) {
	// did we save a copy before?
	base := strings.Split(path.Base(name), ".")[0]
	tempfile := path.Join(os.TempDir(), base+"_rgb.png")
	if tmp, err := os.Open(tempfile); err == nil {
		r = tmp
		format = PngFormat
		hasGamma = false
	}
	var img image.Image
	switch format {
	case UnknownFormat:
		err = fmt.Errorf("unsupported image format")
	case PngFormat:
		img, err = png.Decode(r)
	case JpegFormat:
		img, err = jpeg.Decode(r)
	}
	if err != nil {
		return
	}
	// shortcut if already in suitable format
	bounds = img.Bounds()
	if hasGamma == false {
		switch t := img.(type) {
		case *image.NRGBA:
			pix = t.Pix
			return
		case *image.RGBA:
			pix = t.Pix
			return
		}
	}
	dst := image.NewNRGBA(bounds)
	if hasGamma {
		fmt.Printf("converting image %s to linear RGB space\n", name)
		img = &ToNRGBA{img}
	}
	draw.Draw(dst, bounds, img, image.ZP, draw.Src)
	pix = dst.Pix
	// save a copy of the output image
	if out, err := os.Create(tempfile); err == nil {
		png.Encode(out, dst)
		out.Close()
	}
	return pix, bounds, nil
}

// image converter to linearise image which is in srgba format
type ToNRGBA struct {
	image.Image
}

func (t *ToNRGBA) ColorModel() color.Model {
	return color.NRGBAModel
}

func (t *ToNRGBA) At(x, y int) color.Color {
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
