package glu

import (
	"github.com/jnb666/go3d/img"
	"gopkg.in/qml.v1/gl/es2"
	"gopkg.in/qml.v1/gl/glbase"
	"io"
	"os"
)

// Texture interface type
type Texture interface {
	Activate(id int)
	Dims() []int
}

type Texture2D struct{ *textureBase }

// NewTexture2D creates a new 2D opengl texture. If srgba is set then it is converted to linear RGB space.
// If clamp is set then clamp to edge, else will wrap texture.
func NewTexture2D(clamp bool) Texture2D {
	t := &textureBase{
		typ: GL.TEXTURE_2D,
		tex: gl.GenTextures(1),
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
func (t Texture2D) SetImageFile(file string, conv img.ImageConvert) (Texture2D, error) {
	r, err := os.Open(file)
	if err != nil {
		return t, err
	}
	defer r.Close()
	return t.SetImage(r, conv)
}

// SetImage loads an image from an io.Reader
func (t Texture2D) SetImage(r io.Reader, conv img.ImageConvert) (Texture2D, error) {
	pix, bounds, err := img.Decode(r, conv)
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
func NewTextureCube() TextureCube {
	t := &textureBase{
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

// SetImageFile loads an image from a file.  The index is the number of the image in the cubemap.
func (t TextureCube) SetImageFile(file string, conv img.ImageConvert, index int) (TextureCube, error) {
	r, err := os.Open(file)
	if err != nil {
		return t, err
	}
	defer r.Close()
	return t.SetImage(r, conv, index)
}

// SetImage loads an image, if srgba is set then it is converted to linear RGB space.
// The index is the number of the image in the cubemap.
func (t TextureCube) SetImage(r io.Reader, conv img.ImageConvert, index int) (TextureCube, error) {
	pix, bounds, err := img.Decode(r, conv)
	if err != nil {
		return t, err
	}
	t.dims = []int{bounds.Dx(), bounds.Dy()}
	gl.BindTexture(GL.TEXTURE_CUBE_MAP, t.tex[0])
	target := GL.TEXTURE_CUBE_MAP_POSITIVE_X + glbase.Enum(index)
	gl.TexImage2D(target, 0, GL.RGBA, t.dims[0], t.dims[1], 0, GL.RGBA, GL.UNSIGNED_BYTE, pix)
	gl.BindTexture(GL.TEXTURE_CUBE_MAP, 0)
	CheckError()
	return t, nil
}

type Texture3D struct{ *textureBase }

// NewTexture3D creates a 3D texture mapping.
func NewTexture3D() Texture3D {
	t := &textureBase{
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
func (t Texture3D) SetImageFile(file string, conv img.ImageConvert, dims []int) (Texture3D, error) {
	r, err := os.Open(file)
	if err != nil {
		return t, err
	}
	defer r.Close()
	return t.SetImage(r, conv, dims)
}

// SetImage loads an image. dims is required to set the x,y,z mapping.
func (t Texture3D) SetImage(r io.Reader, conv img.ImageConvert, dims []int) (Texture3D, error) {
	pix, _, err := img.Decode(r, conv)
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
	typ  glbase.Enum
	tex  []glbase.Texture
	dims []int
}

func (t *textureBase) Activate(id int) {
	gl.ActiveTexture(GL.TEXTURE0 + glbase.Enum(id))
	gl.BindTexture(t.typ, t.tex[0])
	if Debug {
		CheckError()
	}
}

func (t *textureBase) Dims() []int {
	return t.dims
}
