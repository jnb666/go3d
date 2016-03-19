package mesh

import (
	"bytes"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/assets"
	"github.com/jnb666/go3d/glu"
	"gopkg.in/qml.v1/gl/es2"
	"image"
	"image/jpeg"
	"image/png"
	"path"
	"strconv"
)

// Interface type for a material which can be used to render a mesh
type Material interface {
	Enable() *glu.Program
	Disable()
	Color() mgl32.Vec4
	SetColor(c mgl32.Vec4) Material
}

const (
	_ = iota
	mPointShader
	mUnshaded
	mDiffuse
	mBlinnPhong
	mUnshadedTex
	mDiffuseTex
	mBlinnPhongTex
	mUnshadedTexCube
	mDiffuseTexCube
	mBlinnPhongTexCube
	mWoodShader
	mRoughShader
	mEmissiveShader
	mMarbleShader
)

const (
	_ = iota
	tWood
	tTurbulence
	tEarth
	tSkybox
)

var (
	progCache = map[int]*glu.Program{}
	texCache  = map[int]glu.Texture{}
)

// Unshaded colored material with optional texture
type Unshaded struct {
	*baseMaterial
}

func NewUnshaded() *Unshaded {
	m := newMaterial(glu.White)
	m.prog, _ = getProgram(mUnshaded)
	return &Unshaded{baseMaterial: m}
}

func NewUnshadedTex(tex glu.Texture) *Unshaded {
	m := newMaterial(glu.White)
	var cached bool
	switch tex.(type) {
	case glu.Texture2D:
		m.prog, cached = getProgram(mUnshadedTex)
	case glu.TextureCube:
		m.prog, cached = getProgram(mUnshadedTexCube)
	default:
		panic("unsupported texture type")
	}
	m.tex = append(m.tex, tex)
	if !cached {
		m.prog.Uniform("1i", "tex0")
	}
	return &Unshaded{baseMaterial: m}
}

func (m *Unshaded) SetColor(c mgl32.Vec4) Material {
	m.baseMaterial.color = c
	return m
}

type pointMaterial struct {
	*baseMaterial
}

// Material used for drawing points
func PointMaterial() Material {
	m := newMaterial(glu.White)
	var cached bool
	if m.prog, cached = getProgram(mPointShader); !cached {
		m.prog.Uniform("2f", "viewport")
		m.prog.Uniform("v3f", "pointLocation")
		m.prog.Uniform("1f", "pointSize")
	}
	return &Unshaded{baseMaterial: m}
}

func (m *pointMaterial) SetColor(col mgl32.Vec4) Material {
	m.color = col
	return m
}

// Emissive material which looks like it glows
func Emissive() *Unshaded {
	m := newMaterial(glu.White)
	m.prog, _ = getProgram(mEmissiveShader)
	return &Unshaded{baseMaterial: m}
}

// Skybox using a cubemap texture
func Skybox() *Unshaded {
	tex := getTextureCube(tSkybox, "skybox")
	return NewUnshadedTex(tex)
}

// Diffuse colored material with optional texture
type Diffuse struct {
	*baseMaterial
}

func NewDiffuse() *Diffuse {
	m := newMaterial(glu.White)
	m.prog, _ = getProgram(mDiffuse)
	return &Diffuse{baseMaterial: m}
}

func NewDiffuseTex(tex glu.Texture) *Diffuse {
	m := newMaterial(glu.White)
	var cached bool
	switch tex.(type) {
	case glu.Texture2D:
		m.prog, cached = getProgram(mDiffuseTex)
	case glu.TextureCube:
		m.prog, cached = getProgram(mDiffuseTexCube)
	default:
		panic("unsupported texture type")
	}
	m.tex = append(m.tex, tex)
	if !cached {
		m.prog.Uniform("1i", "tex0")
	}
	return &Diffuse{baseMaterial: m}
}

func (m *Diffuse) SetColor(c mgl32.Vec4) Material {
	m.baseMaterial.color = c
	return m
}

// Earth cubemap
func Earth() *Diffuse {
	tex := getTextureCube(tEarth, "earth")
	return NewDiffuseTex(tex)
}

// Reflective material with optional texture
type Reflective struct {
	*baseMaterial
	specular     mgl32.Vec3
	shininess    float32
	ambientScale float32
}

func NewReflective(specular mgl32.Vec4, shininess float32) *Reflective {
	m := newMaterial(glu.White)
	var cached bool
	if m.prog, cached = getProgram(mBlinnPhong); !cached {
		initReflective(m.prog)
	}
	return &Reflective{
		baseMaterial: m,
		specular:     specular.Vec3(),
		shininess:    shininess,
		ambientScale: 1,
	}
}

func NewReflectiveTex(specular mgl32.Vec4, shininess float32, tex glu.Texture) *Reflective {
	m := newMaterial(glu.White)
	var cached bool
	switch tex.(type) {
	case glu.Texture2D:
		m.prog, cached = getProgram(mBlinnPhongTex)
	case glu.TextureCube:
		m.prog, cached = getProgram(mBlinnPhongTexCube)
	default:
		panic("unsupported texture type")
	}
	m.tex = append(m.tex, tex)
	if !cached {
		initReflective(m.prog)
		m.prog.Uniform("1i", "tex0")
	}
	return &Reflective{
		baseMaterial: m,
		specular:     specular.Vec3(),
		shininess:    shininess,
		ambientScale: 1,
	}
}

func initReflective(prog *glu.Program) {
	prog.Uniform("v3f", "specularColor")
	prog.Uniform("1f", "shininess", "ambientScale")
}

func (m *Reflective) SetColor(c mgl32.Vec4) Material {
	m.baseMaterial.color = c
	return m
}

func (m *Reflective) Enable() *glu.Program {
	prog := m.baseMaterial.Enable()
	prog.Set("ambientScale", m.ambientScale)
	prog.Set("specularColor", m.specular)
	prog.Set("shininess", m.shininess)
	return prog
}

// Shiny plastic like material
func Plastic() *Reflective {
	return NewReflective(mgl32.Vec4{0.8, 0.8, 0.8, 1}, 128)
}

// Glass is reflective and has transparency
func Glass() *Reflective {
	mat := NewReflective(mgl32.Vec4{0.7, 0.7, 0.7, 1}, 64)
	mat.SetColor(mgl32.Vec4{1, 1, 1, 0.4})
	return mat
}

// 3d Textured wood material
func Wood() *Reflective {
	m := newMaterial(glu.White)
	var cached bool
	if m.prog, cached = getProgram(mWoodShader); !cached {
		initReflective(m.prog)
		m.prog.Uniform("1i", "tex0", "tex1")
	}
	m.tex = append(m.tex, getTexture(tWood), getTexture(tTurbulence))
	return &Reflective{
		baseMaterial: m,
		specular:     mgl32.Vec3{0.5, 0.5, 0.5},
		shininess:    10,
		ambientScale: 1,
	}
}

// Rough randomly textured material
func Rough() *Reflective {
	m := newMaterial(glu.White)
	var cached bool
	if m.prog, cached = getProgram(mRoughShader); !cached {
		initReflective(m.prog)
		m.prog.Uniform("1i", "tex0")
	}
	m.tex = append(m.tex, getTexture(tTurbulence))
	return &Reflective{
		baseMaterial: m,
		specular:     mgl32.Vec3{0.5, 0.5, 0.5},
		shininess:    32,
		ambientScale: 0.7,
	}
}

// Marble textured material
func Marble() *Reflective {
	m := newMaterial(glu.White)
	var cached bool
	if m.prog, cached = getProgram(mMarbleShader); !cached {
		initReflective(m.prog)
		m.prog.Uniform("1i", "tex0")
	}
	m.tex = append(m.tex, getTexture(tTurbulence))
	return &Reflective{
		baseMaterial: m,
		specular:     mgl32.Vec3{0.8, 0.8, 0.8},
		shininess:    200,
		ambientScale: 1,
	}
}

// base type for all materials
type baseMaterial struct {
	prog  *glu.Program
	tex   []glu.Texture
	color mgl32.Vec4
}

func newMaterial(color mgl32.Vec4) *baseMaterial {
	return &baseMaterial{tex: []glu.Texture{}, color: color}
}

func (m *baseMaterial) Enable() *glu.Program {
	m.prog.Use()
	m.prog.Set("objectColor", m.color)
	for i, tex := range m.tex {
		tex.Activate()
		m.prog.Set("tex"+strconv.Itoa(i), tex.Id())
	}
	return m.prog
}

func (m *baseMaterial) Color() mgl32.Vec4 { return m.color }

func (m *baseMaterial) Disable() {}

// compile program and setup default uniforms
func getProgram(id int) (*glu.Program, bool) {
	if prog, ok := progCache[id]; ok {
		return prog, true
	}
	var prog *glu.Program
	var err error
	if id == mPointShader {
		prog, err = glu.NewProgram(vertexShaderPoints, fragmentShader[id], vertexLayoutPoints, vertexSize)
	} else {
		prog, err = glu.NewProgram(vertexShader, fragmentShader[id], vertexLayout, vertexSize)
	}
	if err != nil {
		panic(err)
	}
	prog.Uniform("m4f", "modelToCamera", "cameraToClip")
	prog.Uniform("v4f", "objectColor")
	if id != mPointShader {
		prog.Uniform("m3f", "normalModelToCamera")
		prog.Uniform("1f", "texScale")
		prog.Uniform("v3f", "modelScale")
		prog.Uniform("1i", "numLights")
		prog.UniformArray(MaxLights, "v4f", "lightPos", "lightCol")
	}
	progCache[id] = prog
	return prog, false
}

// get texture which has been packed using go-bindata
func getTexture(id int) glu.Texture {
	if tex, ok := texCache[id]; ok {
		return tex
	}
	switch id {
	case tWood:
		fmt.Println("load texture2D wood")
		tex := glu.NewTexture2D(GL.MIRRORED_REPEAT)
		img := getImage("wood_rgb.png")
		texCache[id] = tex.SetImage(img, false)
	case tTurbulence:
		fmt.Println("load texture3D turbulence3")
		tex := glu.NewTexture3D()
		img := getImage("turbulence3.png")
		texCache[id] = tex.SetImage(img, false, []int{64, 64, 64})
	default:
		panic("unknown texture")
	}
	return texCache[id]
}

func getTextureCube(id int, baseFile string) glu.Texture {
	if tex, ok := texCache[id]; ok {
		return tex
	}
	tex := glu.NewTextureCube()
	fmt.Printf("load textureCube %s\n", baseFile)
	for i, side := range []string{"posx", "negx", "posy", "negy", "posz", "negz"} {
		img := getImage(baseFile + "_" + side + "_rgb.png")
		tex.SetImage(img, false, i)
	}
	texCache[id] = tex
	return tex
}

func getImage(file string) (img image.Image) {
	data, err := assets.Asset(file)
	if err != nil {
		panic(fmt.Errorf("error loading asset %s: %s", file, err))
	}
	ext := path.Ext(file)
	switch ext {
	case ".png":
		img, err = png.Decode(bytes.NewReader(data))
	case ".jpeg", ".jpg":
		img, err = jpeg.Decode(bytes.NewReader(data))
	default:
		panic("unknown file extension " + ext)
	}
	if err != nil {
		panic(fmt.Errorf("error decoding image: %s", err))
	}
	return img
}
