package scene

import (
	"bufio"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/glu"
	"github.com/jnb666/go3d/mesh"
	"gopkg.in/qml.v1/gl/es2"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var spaces = regexp.MustCompile("[ \t\r]+")

type objData struct {
	obj     *Item
	mtl     mesh.Material
	group   string
	mtlName string
}

// Create a new mesh and associated materials from a .obj file
func LoadObjFile(name string) (obj Object, err error) {
	r, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	fmt.Println("load mesh from", name)
	return LoadObj(r)
}

// Create a new mesh from data
func LoadObj(r io.Reader) (o Object, err error) {
	var line string
	defer func() {
		if errPanic := recover(); errPanic != nil {
			err = fmt.Errorf("LoadObj: Error %s parsing line: %s", errPanic, line)
		}
	}()
	obj := newObjData()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line = strings.TrimSpace(scanner.Text())
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		flds := spaces.Split(line, -1)
		switch flds[0] {
		case "v", "vt", "vn":
			obj.parseVertexData(flds[0], parse3fv(flds[1:]))
		case "f":
			obj.parseFaces(flds[1:])
		case "g":
			if len(flds) > 1 {
				obj.build(flds[1])
			}
		case "s":
			// TODO
		case "mtllib":
			err = LoadMtlFile(flds[1])
		case "usemtl":
			if obj.mtl, err = mesh.LoadMaterial(flds[1]); err == nil {
				obj.mtlName = flds[1]
			}
		default:
			fmt.Printf("LoadObj: skip %s\n", line)
		}
		if err != nil {
			return
		}
	}
	if err = scanner.Err(); err != nil {
		return
	}
	obj.build("")
	return obj.obj, err
}

func newObjData() *objData {
	return &objData{
		obj: NewItem(mesh.New()),
	}
}

func (o *objData) build(next string) {
	if o.group != "" {
		fmt.Printf("LoadObj: build group %s with material %s\n", o.group, o.mtlName)
		//fmt.Println(o.obj.Mesh)
		o.obj.Mesh.Build(o.mtl)
	}
	if next != "" {
		o.group = next
		//fmt.Println("LoadObj: start group", o.group)
	}
}

func (o *objData) parseVertexData(typ string, data mgl32.Vec3) {
	m := o.obj.Mesh
	switch typ {
	case "v":
		m.AddVertex(data[0], data[1], data[2])
	case "vt":
		m.AddTexCoord(data[0], data[1])
	case "vn":
		m.AddNormal(data[0], data[1], data[2])
	default:
		panic("unknown vertex type!")
	}
}

func (o *objData) parseFaces(flds []string) {
	m := o.obj.Mesh
	var elem []mesh.El
	for _, str := range flds {
		var el [3]int
		for i, str := range strings.Split(str, "/") {
			el[i] = parseint(str)
		}
		elem = append(elem, mesh.El{el[0], el[1], el[2]})
	}
	if len(elem) == 3 {
		m.AddFace(elem[0], elem[1], elem[2])
	} else if len(elem) == 4 {
		m.AddFaceQuad(elem[0], elem[1], elem[2], elem[3])
	} else {
		panic(fmt.Errorf("face with %d vertices not supported", len(elem)-1))
	}
}

// Load a .mtl file to create one or more new materials
func LoadMtlFile(name string) error {
	r, err := os.Open(name)
	if err != nil {
		return err
	}
	defer r.Close()
	fmt.Println("load materials from", name)
	return LoadMtl(r)
}

// Create a new material from .mtl data
func LoadMtl(r io.Reader) (err error) {
	var line string
	defer func() {
		if errPanic := recover(); errPanic != nil {
			err = fmt.Errorf("LoadMtl: Error %s parsing line: %s", errPanic, line)
		}
	}()
	scanner := bufio.NewScanner(r)
	var m *mtlData
	for scanner.Scan() {
		line = strings.TrimSpace(scanner.Text())
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		flds := spaces.Split(line, -1)
		switch flds[0] {
		case "newmtl":
			if m != nil {
				m.save()
			}
			m = newMtlData(flds[1])
		case "Kd":
			m.diffuse = parse3fv(flds[1:4])
		case "Ks":
			m.specular = parse3fv(flds[1:4])
		case "Ns":
			m.shininess = parsef32(flds[1])
		case "Tr":
			m.alpha = 1 - parsef32(flds[1])
		case "d":
			m.alpha = parsef32(flds[1])
		case "illum":
			m.model = parseint(flds[1])
		case "map_Kd":
			m.diffuseMap = flds[1]
		case "Ka", "Ni", "Ke":
			// noop
		default:
			fmt.Printf("LoadMtl: skip %s\n", line)
		}
	}
	if err = scanner.Err(); err != nil {
		panic(err)
	}
	m.save()
	return
}

type mtlData struct {
	name       string
	diffuse    mgl32.Vec3
	specular   mgl32.Vec3
	shininess  float32
	alpha      float32
	model      int
	diffuseMap string
}

// new material with sensible defaults
func newMtlData(name string) *mtlData {
	return &mtlData{
		name:      name,
		diffuse:   mgl32.Vec3{1, 1, 1},
		specular:  mgl32.Vec3{0.5, 0.5, 0.5},
		shininess: 128,
		alpha:     1,
		model:     2,
	}
}

func (m *mtlData) save() {
	//fmt.Println("LoadMtl: saving", m.name)
	var mtl mesh.Material
	color := m.diffuse.Vec4(m.alpha)
	var tex glu.Texture
	if m.diffuseMap != "" {
		img, err := glu.PNGImage(m.diffuseMap)
		if err != nil {
			panic(err)
		}
		tex = glu.NewTexture2D(GL.CLAMP_TO_EDGE).SetImage(img, true)
	}
	switch {
	case m.model == 0:
		mtl = mesh.Diffuse().SetColor(color).SetAmbient(0)
	case m.model == 0 && tex != nil:
		mtl = mesh.DiffuseTex(tex).SetColor(color).SetAmbient(0)
	case m.model == 1:
		mtl = mesh.Diffuse().SetColor(color)
	case m.model == 1 && tex != nil:
		mtl = mesh.DiffuseTex(tex).SetColor(color)
	case m.model == 2 && tex != nil:
		mtl = mesh.ReflectiveTex(m.specular.Vec4(m.alpha), m.shininess, tex).SetColor(color)
	case m.model == 2:
		mtl = mesh.Reflective(m.specular.Vec4(m.alpha), m.shininess).SetColor(color)
	default:
		panic(fmt.Errorf("LoadMtl: illumination model %d not supported\n", m.model))
	}
	mesh.SaveMaterial(m.name, mtl)
}

func parse3fv(flds []string) (v mgl32.Vec3) {
	for i, fld := range flds {
		v[i] = parsef32(fld)
	}
	return
}

func parsef32(fld string) float32 {
	val, err := strconv.ParseFloat(fld, 32)
	if err != nil {
		panic(err)
	}
	return float32(val)
}

func parseint(fld string) int {
	val, err := strconv.ParseInt(fld, 10, 32)
	if err != nil {
		panic(err)
	}
	return int(val)
}
