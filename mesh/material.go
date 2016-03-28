package mesh

import (
	"bytes"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/assets"
	"github.com/jnb666/go3d/glu"
	"github.com/jnb666/go3d/img"
	"io"
	"strconv"
	"strings"
)

// Interface type for a material which can be used to render a mesh
type Material interface {
	Enable() *glu.Program
	Disable()
	Color() mgl32.Vec4
	SetColor(c mgl32.Vec4) Material
	Ambient() float32
	SetAmbient(s float32) Material
	Clone() Material
}

const (
	mFirstShader = iota
	mPointShader
	mUnshaded
	mDiffuse
	mBlinnPhong
	mUnshadedTex
	mDiffuseTex
	mBlinnPhongTex
	mBlinnPhongTexNorm
	mUnshadedTexCube
	mDiffuseTexCube
	mBlinnPhongTexCube
	mBlinnPhongCubeNorm
	mWoodShader
	mRoughShader
	mEmissiveShader
	mMarbleShader
	mLastShader
)

const (
	tFirstTexture = iota
	tWood
	tTurbulence
	tEarth
	tEarthSpec
	tSkybox
	tMetallic
	tMetallicSpec
	tLastTexture
)

var (
	progCache    = map[int]*glu.Program{}
	texCache     = map[int]glu.Texture{}
	mtlCache     = map[string]Material{}
	mtlDataCache = map[string]mtlData{}
)

// Get material by name
func LoadMaterial(name string, bumpMap bool) (mtl Material, err error) {
	cname := fmt.Sprintf("%s:%v", strings.ToLower(name), bumpMap)
	var ok bool
	if mtl, ok = mtlCache[cname]; ok {
		return mtl, nil
	}
	if data, ok := mtlDataCache[name]; ok {
		mtl, err = data.toMaterial(bumpMap)
		if err != nil {
			return nil, err
		}
		mtlCache[cname] = mtl
		return mtl, nil
	}
	switch name {
	case "point":
		mtl = PointMaterial()
	case "diffuse":
		mtl = Diffuse()
	case "earth":
		mtl = Earth()
	case "emissive":
		mtl = Emissive()
	case "glass":
		mtl = Glass()
	case "marble":
		mtl = Marble()
	case "plastic":
		mtl = Plastic()
	case "rough":
		mtl = Rough()
	case "skybox":
		mtl = Skybox()
	case "unshaded":
		mtl = Unshaded()
	case "wood":
		mtl = Wood()
	default:
		return nil, fmt.Errorf("LoadMaterial: no material called %s", name)
	}
	return mtl, nil
}

// Save material data to cache - called from mtl loader
func saveMaterialData(m *mtlData) {
	mtlDataCache[strings.ToLower(m.name)] = *m
}

func ntex(tex []glu.Texture) (n int) {
	for _, t := range tex {
		if t != nil {
			n++
		}
	}
	return n
}

// Unshaded colored material with optional texture
func Unshaded(tex ...glu.Texture) Material {
	m := newMaterial(glu.White)
	if ntex(tex) == 0 {
		m.prog = getProgram(mUnshaded)
	} else {
		switch tex[0].(type) {
		case glu.Texture2D:
			m.prog = getProgram(mUnshadedTex)
		case glu.TextureCube:
			m.prog = getProgram(mUnshadedTexCube)
		default:
			panic("unsupported texture type")
		}
		m.tex = append(m.tex, tex[0])
	}
	return m
}

// Material used for drawing points
func PointMaterial() Material {
	m := newMaterial(glu.White)
	m.prog = getProgram(mPointShader)
	return m
}

// Emissive material which looks like it glows
func Emissive() Material {
	m := newMaterial(mgl32.Vec4{0.9, 0.9, 0.9, 1})
	m.prog = getProgram(mEmissiveShader)
	return m
}

// Skybox using a cubemap texture
func Skybox() Material {
	return Unshaded(getTexture(tSkybox))
}

// Diffuse colored material with optional texture
func Diffuse(tex ...glu.Texture) Material {
	m := newMaterial(glu.White)
	if ntex(tex) == 0 {
		m.prog = getProgram(mDiffuse)
	} else {
		switch tex[0].(type) {
		case glu.Texture2D:
			m.prog = getProgram(mDiffuseTex)
		case glu.TextureCube:
			m.prog = getProgram(mDiffuseTexCube)
		default:
			panic("unsupported texture type")
		}
		m.tex = append(m.tex, tex[0])
	}
	return m
}

type reflective struct {
	*baseMaterial
	specular  mgl32.Vec3
	shininess float32
}

// Coloured material with specular highlights using Blinn-Phong model.
// tex paramater is optional list of associated textures for diffuse, specular and normal map.
func Reflective(specular mgl32.Vec4, shininess float32, tex ...glu.Texture) Material {
	m := newMaterial(glu.White)
	if ntex(tex) == 0 {
		m.prog = getProgram(mBlinnPhong)
	} else {
		switch tex[0].(type) {
		case glu.Texture2D:
			if len(tex) > 2 {
				m.prog = getProgram(mBlinnPhongTexNorm)
			} else {
				m.prog = getProgram(mBlinnPhongTex)
			}
		case glu.TextureCube:
			if len(tex) > 2 {
				m.prog = getProgram(mBlinnPhongCubeNorm)
			} else {
				m.prog = getProgram(mBlinnPhongTexCube)
			}
		default:
			panic("unsupported texture type")
		}
		m.tex = append(m.tex, tex...)
	}
	return &reflective{
		baseMaterial: m,
		specular:     specular.Vec3(),
		shininess:    shininess,
	}
}

func (m *reflective) Clone() Material {
	newMat := *m
	newMat.baseMaterial = m.baseMaterial.Clone().(*baseMaterial)
	return &newMat
}

func (m *reflective) SetColor(c mgl32.Vec4) Material {
	m.baseMaterial.color = c
	return m
}

func (m *reflective) SetAmbient(scale float32) Material {
	m.baseMaterial.ambient = scale
	return m
}

func (m *reflective) Enable() *glu.Program {
	prog := m.baseMaterial.Enable()
	prog.Set("specularColor", m.specular)
	prog.Set("shininess", m.shininess)
	return prog
}

// Shiny plastic like material
func Plastic() Material {
	return Reflective(mgl32.Vec4{0.8, 0.8, 0.8, 1}, 128)
}

// Glass is reflective and has transparency
func Glass() Material {
	mat := Reflective(mgl32.Vec4{0.7, 0.7, 0.7, 1}, 64)
	mat.SetColor(mgl32.Vec4{1, 1, 1, 0.4})
	return mat
}

// Earth cubemap
func Earth() Material {
	return Reflective(mgl32.Vec4{0.5, 0.5, 0.5, 1}, 32, getTexture(tEarth), getTexture(tEarthSpec))
}

type metallic struct {
	Material
}

// Textured metallic material
func Metallic() Material {
	return &metallic{
		Reflective(glu.White, 16, getTexture(tMetallic), getTexture(tMetallicSpec)).SetAmbient(0.3),
	}
}

func (m *metallic) Clone() Material {
	return &metallic{m.Material.Clone()}
}

func (m *metallic) SetColor(c mgl32.Vec4) Material {
	m.Material.SetColor(c)
	m.Material.(*reflective).specular = c.Vec3()
	return m
}

func (m *metallic) SetAmbient(amb float32) Material {
	m.Material.SetAmbient(amb)
	return m
}

// 3d Textured wood material
func Wood() Material {
	m := newMaterial(glu.White)
	m.prog = getProgram(mWoodShader)
	m.tex = append(m.tex, getTexture(tWood), getTexture(tTurbulence))
	return &reflective{
		baseMaterial: m,
		specular:     mgl32.Vec3{0.5, 0.5, 0.5},
		shininess:    10,
	}
}

// Rough randomly textured material
func Rough() Material {
	m := newMaterial(glu.White)
	m.ambient = 0.3
	m.prog = getProgram(mRoughShader)
	m.tex = append(m.tex, getTexture(tTurbulence))
	return &reflective{
		baseMaterial: m,
		specular:     mgl32.Vec3{0.5, 0.5, 0.5},
		shininess:    32,
	}
}

// Marble textured material
func Marble() Material {
	m := newMaterial(glu.White)
	m.prog = getProgram(mMarbleShader)
	m.tex = append(m.tex, getTexture(tTurbulence))
	return &reflective{
		baseMaterial: m,
		specular:     mgl32.Vec3{0.8, 0.8, 0.8},
		shininess:    200,
	}
}

// base type for all materials
type baseMaterial struct {
	prog    *glu.Program
	tex     []glu.Texture
	color   mgl32.Vec4
	ambient float32
}

func newMaterial(color mgl32.Vec4) *baseMaterial {
	return &baseMaterial{tex: []glu.Texture{}, color: color, ambient: 1}
}

func (m *baseMaterial) Enable() *glu.Program {
	m.prog.Use()
	m.prog.Set("objectColor", m.color)
	m.prog.Set("ambientScale", m.ambient)
	m.prog.Set("numTex", ntex(m.tex))
	for i, tex := range m.tex {
		if tex != nil {
			tex.Activate(i)
			m.prog.Set("tex"+strconv.Itoa(i), i)
		}
	}
	return m.prog
}

func (m *baseMaterial) Clone() Material {
	return &baseMaterial{
		prog:    m.prog,
		tex:     append([]glu.Texture{}, m.tex...),
		color:   m.color,
		ambient: m.ambient,
	}
}

func (m *baseMaterial) Color() mgl32.Vec4 { return m.color }

func (m *baseMaterial) SetColor(c mgl32.Vec4) Material {
	m.color = c
	return m
}

func (m *baseMaterial) Ambient() float32 { return m.ambient }

func (m *baseMaterial) SetAmbient(amb float32) Material {
	m.ambient = amb
	return m
}

func (m *baseMaterial) Disable() {}

// compile program and setup default uniforms
func getProgram(id int) *glu.Program {
	if prog, ok := progCache[id]; ok {
		return prog
	}
	var prog *glu.Program
	var err error
	switch id {
	case mBlinnPhongTexNorm, mBlinnPhongCubeNorm:
		prog, err = glu.NewProgram(vertexShaderTBN, fragmentShader[id], vertexLayoutTBN, vertexSize)
	case mPointShader:
		prog, err = glu.NewProgram(vertexShaderPoints, fragmentShader[id], vertexLayoutPoints, vertexSize)
	default:
		prog, err = glu.NewProgram(vertexShader, fragmentShader[id], vertexLayout, vertexSize)
	}
	if err != nil {
		panic(err)
	}
	prog.Uniform("m4f", "modelToCamera", "cameraToClip")
	prog.Uniform("v4f", "objectColor")
	prog.Uniform("1f", "ambientScale")
	prog.Uniform("v3f", "specularColor")
	prog.Uniform("1f", "shininess")
	if id == mPointShader {
		prog.Uniform("2f", "viewport")
		prog.Uniform("v3f", "pointLocation")
		prog.Uniform("1f", "pointSize")
	} else {
		prog.Uniform("m3f", "normalModelToCamera")
		prog.Uniform("v3f", "modelScale")
		prog.Uniform("1i", "numLights")
		prog.UniformArray(MaxLights, "v4f", "lightPos", "lightCol")
	}
	prog.Uniform("1i", "numTex")
	for i := 0; i < numSamplers[id]; i++ {
		prog.Uniform("1i", fmt.Sprintf("tex%d", i))
	}
	progCache[id] = prog
	return prog
}

// get texture which has been packed using go-bindata
func getTexture(id int) glu.Texture {
	tex, ok := texCache[id]
	if ok {
		return tex
	}
	var err error
	switch id {
	case tWood:
		tex, err = glu.NewTexture2D(false).SetImage(getImage("wood.png"), img.NoConvert)
	case tTurbulence:
		tex, err = glu.NewTexture3D().SetImage(getImage("turbulence3.png"), img.NoConvert, []int{64, 64, 64})
	case tMetallic:
		tex, err = glu.NewTexture2D(false).SetImage(getImage("metallic.png"), img.NoConvert)
	case tMetallicSpec:
		tex, err = glu.NewTexture2D(false).SetImage(getImage("metallic_spec.png"), img.NoConvert)
	case tEarth:
		tex, err = textureCube("earth")
	case tEarthSpec:
		tex, err = textureCube("earth_spec")
	case tSkybox:
		tex, err = textureCube("skybox")
	default:
		err = fmt.Errorf("unknown texture %d", id)
	}
	if err != nil {
		panic(err)
	}
	texCache[id] = tex
	return tex
}

func textureCube(baseFile string) (glu.Texture, error) {
	tex := glu.NewTextureCube()
	for i, side := range []string{"posx", "negx", "posy", "negy", "posz", "negz"} {
		image := getImage(baseFile + "_" + side + ".png")
		if _, err := tex.SetImage(image, img.NoConvert, i); err != nil {
			return tex, err
		}
	}
	return tex, nil
}

func getImage(file string) io.Reader {
	data, err := assets.Asset(file)
	if err != nil {
		panic(fmt.Errorf("error loading asset %s: %s", file, err))
	}
	return bytes.NewReader(data)
}
