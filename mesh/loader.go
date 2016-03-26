package mesh

import (
	"bufio"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/glu"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var spaces = regexp.MustCompile("[ \t\r]+")

type elements [][]El

type objData struct {
	*Mesh
	groups  map[string]elements
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
	current, _ := os.Getwd()
	os.Chdir(path.Dir(name))
	defer os.Chdir(current)
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
	obj := &objData{Mesh: New(), groups: map[string]elements{}}
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
			obj.Mesh.SetNormalSmoothing(flds[1] != "off")
		case "mtllib":
			_, err = LoadMtlFile(flds[1])
		case "usemtl":
			obj.mtlName = flds[1]
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

func (o *objData) build(name string) {
	for mat, faces := range o.groups {
		fmt.Printf("group %s with material %s - %d faces\n", o.grpName, mat, len(faces))
		for _, face := range faces {
			o.AddFace(face...)
		}
		o.Build(mat)
	}
	o.grpName = name
	o.groups = map[string]elements{}
}

func (o *objData) parseVertexData(typ string, data mgl32.Vec3) {
	switch typ {
	case "v":
		o.AddVertex(data[0], data[1], data[2])
	case "vt":
		o.AddTexCoord(data[0], data[1])
	case "vn":
		o.AddNormal(data[0], data[1], data[2])
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
	o.groups[o.mtlName] = append(o.groups[o.mtlName], elem)
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
				saveMaterialData(m)
				names = append(names, m.name)
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
			m.diffMap = fullPath(flds[1])
		case "map_Ks":
			m.specMap = fullPath(flds[1])
		case "bump", "map_bump": // TODO
		default:
			fmt.Printf("LoadMtl: skip %s\n", line)
		}
	}
	if err = scanner.Err(); err != nil {
		panic(err)
	}
	if m != nil {
		saveMaterialData(m)
		names = append(names, m.name)
	}
	return names, err
}

func fullPath(name string) string {
	cwd, _ := os.Getwd()
	return path.Join(cwd, name)
}

type mtlData struct {
	name      string
	ambient   mgl32.Vec3
	diffuse   mgl32.Vec3
	specular  mgl32.Vec3
	shininess float32
	alpha     float32
	model     int
	diffMap   string
	specMap   string
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

func (m mtlData) toMaterial() (mtl Material, err error) {
	var textures []glu.Texture
	if m.diffMap != "" {
		if tex, err := glu.NewTexture2D(false, true).SetImageFile(m.diffMap); err == nil {
			textures = append(textures, tex)
		} else {
			return nil, fmt.Errorf("toMaterial: error loading diffuse map for material %s: %s", m.name, err)
		}
		if m.specMap != "" {
			if tex, err := glu.NewTexture2D(false, true).SetImageFile(m.specMap); err == nil {
				textures = append(textures, tex)
			} else {
				return nil, fmt.Errorf("toMaterial: error loading specular map for material %s: %s", m.name, err)
			}
		}
	}
	color := m.diffuse.Vec4(m.alpha)
	ambScale := m.ambient.Vec4(1).Len() / m.diffuse.Vec4(1).Len()
	switch m.model {
	case 0:
		ambScale = 0
		mtl = Diffuse(textures...)
	case 1:
		mtl = Diffuse(textures...)
	default:
		mtl = Reflective(m.specular.Vec4(m.alpha), m.shininess, textures...)
	}
	mtl.SetColor(color).SetAmbient(ambScale)
	return mtl, nil
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
