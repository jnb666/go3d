package mesh

import (
	"bufio"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/glu"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var spaces = regexp.MustCompile("[ \t\r]+")

type objData struct {
	*Mesh
	mtl     Material
	grpName string
	mtlName string
}

// Create a new mesh and associated materials from a .obj file
func LoadObjFile(name string) (m *Mesh, err error) {
	r, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	fmt.Println("load mesh from", name)
	return LoadObj(r)
}

// Create a new mesh from data
func LoadObj(r io.Reader) (m *Mesh, err error) {
	var line string
	defer func() {
		if errPanic := recover(); errPanic != nil {
			err = fmt.Errorf("LoadObj: Error %s parsing line: %s", errPanic, line)
		}
	}()
	obj := &objData{Mesh: New()}
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
			obj.Mesh.SetNormalSmoothing(flds[1])
		case "mtllib":
			_, err = LoadMtlFile(flds[1])
		case "usemtl":
			if obj.mtl, err = LoadMaterial(flds[1]); err == nil {
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
	obj.build("__END")
	return obj.Mesh, err
}

func (o *objData) build(next string) {
	if o.grpName != "" || next == "__END" {
		o.Mesh.Build(o.mtl)
		size := len(o.Mesh.groups[len(o.Mesh.groups)-1].edata)
		fmt.Printf("group %s with material %s - %d elements\n", o.grpName, o.mtlName, size)
	}
	if next != "" && next != "__END" {
		o.grpName = next
	}
}

func (o *objData) parseVertexData(typ string, data mgl32.Vec3) {
	switch typ {
	case "v":
		o.Mesh.AddVertex(data[0], data[1], data[2])
	case "vt":
		o.Mesh.AddTexCoord(data[0], data[1])
	case "vn":
		o.Mesh.AddNormal(data[0], data[1], data[2])
	default:
		panic("unknown vertex type!")
	}
}

func (o *objData) parseFaces(flds []string) {
	var elem []El
	for _, str := range flds {
		var el [3]int
		for i, str := range strings.Split(str, "/") {
			el[i] = parseint(str)
		}
		elem = append(elem, El{el[0], el[1], el[2]})
	}
	o.Mesh.AddFace(elem...)
}

// Load a .mtl file to create one or more new materials, returns list of material names
func LoadMtlFile(name string) ([]string, error) {
	r, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	fmt.Println("load materials from", name)
	return LoadMtl(r)
}

// Create a new material from .mtl data, returns list of material names
func LoadMtl(r io.Reader) (names []string, err error) {
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
				names = append(names, m.save())
			}
			m = newMtlData(flds[1])
		case "Ka":
			m.ambient = parse3fv(flds[1:4])
		case "Kd":
			m.diffuse = parse3fv(flds[1:4])
		case "Ke": // emissive?
		case "Ks":
			m.specular = parse3fv(flds[1:4])
		case "Ni": // optical density
		case "Ns":
			// increase this since we using blinn phong
			m.shininess = parsef32(flds[1]) * 2
		case "Tr":
			m.alpha = 1 - parsef32(flds[1])
		case "Tf": // transmission filter
		case "d":
			m.alpha = parsef32(flds[1])
		case "illum":
			m.model = parseint(flds[1])
		case "map_Ka": // assume this matches diffuse map
		case "map_Kd":
			m.diffuseMap = flds[1]
		case "map_Ks": // specular map: TODO
		case "bump", "map_bump": // TODO
		default:
			fmt.Printf("LoadMtl: skip %s\n", line)
		}
	}
	if err = scanner.Err(); err != nil {
		panic(err)
	}
	names = append(names, m.save())
	return names, err
}

type mtlData struct {
	name       string
	ambient    mgl32.Vec3
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
		ambient:   mgl32.Vec3{1, 1, 1},
		diffuse:   mgl32.Vec3{1, 1, 1},
		specular:  mgl32.Vec3{0.5, 0.5, 0.5},
		shininess: 128,
		alpha:     1,
		model:     2,
	}
}

func (m *mtlData) save() string {
	var mtl Material
	color := m.diffuse.Vec4(m.alpha)
	var tex glu.Texture
	var err error
	if m.diffuseMap != "" {
		if tex, err = glu.NewTexture2D(false, true).SetImageFile(m.diffuseMap); err != nil {
			panic(err)
		}
	}
	ambScale := m.ambient.Vec4(1).Len() / m.diffuse.Vec4(1).Len()
	switch m.model {
	case 0:
		ambScale = 0
		mtl = DiffuseTex(tex)
	case 1:
		mtl = DiffuseTex(tex)
	case 2:
		mtl = ReflectiveTex(m.specular.Vec4(m.alpha), m.shininess, tex)
	default:
		panic(fmt.Errorf("LoadMtl: illumination model %d not supported\n", m.model))
	}
	mtl.SetColor(color).SetAmbient(ambScale)
	SaveMaterial(m.name, mtl)
	return m.name
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
