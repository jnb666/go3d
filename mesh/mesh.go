// Package mesh provides functions for managing 3d triangle meshes and material definitions.
package mesh

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/glu"
	"gopkg.in/qml.v1/gl/es2"
	"gopkg.in/qml.v1/gl/glbase"
)

const (
	vertexSize = 11
	epsilon    = 1e-6
)

var vertexLayoutTBN = []glu.Attrib{
	{Name: "position", Size: 3, Offset: 0},
	{Name: "normal", Size: 3, Offset: 3},
	{Name: "texcoord", Size: 2, Offset: 6},
	{Name: "tangent", Size: 3, Offset: 8},
}

var vertexLayout = []glu.Attrib{
	{Name: "position", Size: 3, Offset: 0},
	{Name: "normal", Size: 3, Offset: 3},
	{Name: "texcoord", Size: 2, Offset: 6},
}

var vertexLayoutPoints = []glu.Attrib{
	{Name: "position", Size: 3, Offset: 0},
}

var winding = [2]glbase.Enum{GL.CW, GL.CCW}

type El struct {
	Vert, Tex, Norm int
}

type el2 struct {
	El
	tang int
}

// Mesh type stores a mesh of vertices
type Mesh struct {
	inverted  int
	vdata     []float32
	groups    []*meshGroup
	varray    [2]*glu.VertexArray
	vertices  []mgl32.Vec3
	normals   []mgl32.Vec3
	texcoords []mgl32.Vec2
	tangents  []mgl32.Vec3
	elements  []el2
	ncache    normalCache
	pointSize int
	bumpMap   bool
}

type meshGroup struct {
	mtlName string
	edata   []uint32
	mtl     Material
	earray  *glu.VertexArray
}

type normalCache struct {
	vert2norm map[int]*runningMean
	elem2vert map[int]int
	smooth    bool
}

// NewMesh creates a new empty mesh structure
func New() *Mesh {
	return &Mesh{ncache: newNormalCache(false), groups: []*meshGroup{}, bumpMap: true}
}

func newNormalCache(smooth bool) normalCache {
	return normalCache{vert2norm: map[int]*runningMean{}, elem2vert: map[int]int{}, smooth: smooth}
}

// Clear method wipes the stored vertex data. It does not erase groups which are already built, call this after Build
// if you need to add a new set of vertices separate to the previous ones.
func (m *Mesh) Clear() *Mesh {
	m.vertices = nil
	m.normals = nil
	m.texcoords = nil
	m.tangents = nil
	m.elements = nil
	m.ncache = newNormalCache(true)
	return m
}

// Clone method makes a copy of the mesh with the same vertex data, but a copy of the materials
func (m *Mesh) Clone() *Mesh {
	newMesh := New()
	newMesh.vdata = m.vdata
	newMesh.inverted = m.inverted
	newMesh.varray = m.varray
	newMesh.pointSize = m.pointSize
	for _, grp := range m.groups {
		newMesh.groups = append(newMesh.groups, &meshGroup{mtl: grp.mtl.Clone(), edata: grp.edata, earray: grp.earray})
	}
	return newMesh
}

// Point method returns point size, or zero for non-point
func (m *Mesh) PointSize() int {
	return m.pointSize
}

// Add a new vertex position
func (m *Mesh) AddVertex(x, y, z float32) int {
	m.vertices = append(m.vertices, mgl32.Vec3{x, y, z})
	return len(m.vertices)
}

// Add a new vertex normal
func (m *Mesh) AddNormal(nx, ny, nz float32) int {
	m.normals = append(m.normals, mgl32.Vec3{nx, ny, nz})
	return len(m.normals)
}

// Add tex coordinates
func (m *Mesh) AddTexCoord(tx, ty float32) int {
	m.texcoords = append(m.texcoords, mgl32.Vec2{tx, ty})
	return len(m.texcoords)
}

// Add a triangular or a quad face
func (m *Mesh) AddFace(el ...El) int {
	calcNormal := false
	calcTangent := true
	vtx := make([]mgl32.Vec3, len(el))
	tex := make([]mgl32.Vec2, len(el))
	for i, e := range el {
		if e.Norm == 0 {
			calcNormal = true
		}
		vtx[i] = m.vertex(e.Vert)
		if e.Tex == 0 {
			calcTangent = false
		} else {
			tex[i] = m.texcoord(e.Tex)
		}
	}
	base := len(m.elements)
	ntangent := 0
	if calcTangent {
		if tangent, ok := getTangent(vtx, tex); ok {
			m.tangents = append(m.tangents, tangent)
			ntangent = len(m.tangents)
		}
	}
	switch len(el) {
	case 3:
		m.addElements(ntangent, el)
		if calcNormal {
			normal := vtx[1].Sub(vtx[0]).Cross(vtx[2].Sub(vtx[0]))
			m.ncache.add(m, normal.Normalize(), base, el)
		}
	case 4:
		elquad := []El{el[0], el[1], el[2], el[0], el[2], el[3]}
		m.addElements(ntangent, elquad)
		if calcNormal {
			normal := mgl32.Vec3{}
			for i, v := range vtx {
				v1 := vtx[(i+1)%4]
				normal = normal.Add(mgl32.Vec3{
					(v[1] - v1[1]) * (v[2] + v1[2]),
					(v[2] - v1[2]) * (v[0] + v1[0]),
					(v[0] - v1[0]) * (v[1] + v1[1]),
				})
			}
			m.ncache.add(m, normal.Normalize(), base, elquad)
		}
	default:
		panic("AddFace must have 3 or 4 elements")
	}
	return len(m.elements)
}

func (m *Mesh) addElements(ntangent int, elems []El) {
	for _, elem := range elems {
		m.elements = append(m.elements, el2{El: elem, tang: ntangent})
	}
}

// calc tangent vector for triangle
func getTangent(vtx []mgl32.Vec3, tex []mgl32.Vec2) (tangent mgl32.Vec3, ok bool) {
	dy1 := tex[1][1] - tex[0][1]
	dy2 := tex[2][1] - tex[0][1]
	f := (tex[1][0]-tex[0][0])*dy2 - (tex[2][0]-tex[0][0])*dy1
	// ensure that triangle is not of zero height
	if abs(dy1) < epsilon || abs(dy2) < epsilon || abs(f) < epsilon {
		return
	}
	// calculate tangent vector in texture coordinate space
	edge1, edge2 := vtx[1].Sub(vtx[0]), vtx[2].Sub(vtx[0])
	tangent = edge1.Mul(dy2 / f).Sub(edge2.Mul(dy1 / f)).Normalize()
	return tangent, true
}

// If flag is false then turn off smoothing of vertex normals, else start a new smoothing group
func (m *Mesh) SetNormalSmoothing(on bool) {
	m.ncache.build(m)
	m.ncache = newNormalCache(on)
}

// update the average normal at each vertex
func (n normalCache) add(m *Mesh, norm mgl32.Vec3, base int, elements []El) {
	if n.smooth {
		for i, el := range elements {
			if n.vert2norm[el.Vert] == nil {
				// start accumulating data for this normal
				n.vert2norm[el.Vert] = &runningMean{}
			}
			// add to average so far
			n.vert2norm[el.Vert].push(norm)
			n.elem2vert[base+i] = el.Vert
		}
	} else {
		normal := m.AddNormal(norm[0], norm[1], norm[2])
		for i := range elements {
			m.elements[base+i].Norm = normal
		}
	}
}

// add calculated normals to mesh
func (n normalCache) build(m *Mesh) {
	for i, vid := range n.elem2vert {
		norm := n.vert2norm[vid].mean
		m.elements[i].Norm = m.AddNormal(norm[0], norm[1], norm[2])
	}
}

// Build method processes the data which has been added so far and appends it to the vertex and element buffers.
// It can be called multiple times to add multiple groups of data.
func (m *Mesh) Build(materialName string) {
	grp := &meshGroup{mtlName: materialName}
	if grp.mtlName == "" {
		if m.pointSize != 0 {
			grp.mtlName = "point"
		} else {
			grp.mtlName = "diffuse"
		}
	}
	m.ncache.build(m)
	m.ncache = newNormalCache(true)
	cache := map[el2]uint32{}
	for _, el := range m.elements {
		index, ok := cache[el]
		if !ok {
			index = uint32(len(m.vdata) / vertexSize)
			m.vdata = append(m.vdata, m.getData(el)...)
			cache[el] = index
		}
		grp.edata = append(grp.edata, index)
	}
	//fmt.Printf("mesh group %d: %d vertices, %d elements\n", len(m.groups), len(m.vdata)/vertexSize, len(grp.edata))
	m.groups = append(m.groups, grp)
	m.elements = nil
}

func (m *Mesh) loadMaterials(force bool) (err error) {
	for _, grp := range m.groups {
		if grp.mtl == nil || force {
			if grp.mtl, err = LoadMaterial(grp.mtlName, m.bumpMap); err != nil {
				return err
			}
		}
	}
	return nil
}

// Enable or disable normal mapping
func (m *Mesh) BumpMap(on bool) {
	m.bumpMap = on
	m.loadMaterials(true)
}

// Draw method draws the mesh by calling GL DrawElements, setUniforms callback can be used to set uniforms after
// binding the vertex arrays and enabling the shaders, but prior to drawing.
func (m *Mesh) Draw(setUniforms func(*glu.Program)) error {
	if err := m.loadMaterials(false); err != nil {
		return err
	}
	if m.varray[m.inverted] == nil {
		m.varray[m.inverted] = glu.ArrayBuffer(m.vdata, vertexSize)
	} else {
		m.varray[m.inverted].Enable()
	}
	var lastProg *glu.Program
	for _, grp := range m.groups {
		if grp.earray == nil {
			grp.earray = glu.ElementArrayBuffer(grp.edata)
		} else {
			grp.earray.Enable()
		}
		prog := grp.mtl.Enable()
		if prog != lastProg {
			setUniforms(prog)
			lastProg = prog
		}
		grp.earray.Draw(GL.TRIANGLES, winding[m.inverted])
		grp.mtl.Disable()
	}
	return nil
}

// Invert method reverses the normals and winding order to flip the shape inside out
func (m *Mesh) Invert() *Mesh {
	newMesh := *m
	newMesh.inverted = 1 - m.inverted
	// reverse normal directions
	newMesh.vdata = append([]float32{}, m.vdata...)
	for i := 0; i < len(newMesh.vdata); i += vertexSize {
		newMesh.vdata[i+3] *= -1
		newMesh.vdata[i+4] *= -1
		newMesh.vdata[i+5] *= -1
	}
	return &newMesh
}

// Get the material assocociated with the first mesh group
func (m *Mesh) Material() Material {
	if len(m.groups) == 0 {
		return nil
	}
	if err := m.loadMaterials(false); err != nil {
		return nil
	}
	return m.groups[0].mtl
}

// Update all the materials associated with this mesh.
func (m *Mesh) SetMaterial(mtl Material) *Mesh {
	for _, grp := range m.groups {
		grp.mtl = mtl
	}
	return m
}

// String method for dumping out contents of the mesh
func (m *Mesh) String() (s string) {
	s += fmt.Sprintf("vertices: %f\n", m.vertices)
	s += fmt.Sprintf("normals: %f\n", m.normals)
	s += fmt.Sprintf("tex coords: %f", m.texcoords)
	return s
}

func (m *Mesh) getData(el el2) []float32 {
	data := make([]float32, vertexSize)
	v := m.vertex(el.Vert)
	copy(data, v[:])
	vn := m.normal(el.Norm)
	copy(data[3:], vn[:])
	vt := m.texcoord(el.Tex)
	copy(data[6:], vt[:])
	if el.tang > 0 {
		copy(data[8:], m.tangents[el.tang-1][:])
	}
	return data
}

func (m *Mesh) vertex(n int) mgl32.Vec3 {
	if n > 0 {
		return m.vertices[n-1]
	} else if n < 0 {
		return m.vertices[len(m.vertices)+n]
	} else {
		panic("missing vertex!")
	}
}

func (m *Mesh) texcoord(n int) mgl32.Vec2 {
	if n > 0 {
		return m.texcoords[n-1]
	} else if n < 0 {
		return m.texcoords[len(m.texcoords)+n]
	} else {
		return mgl32.Vec2{}
	}
}

func (m *Mesh) normal(n int) mgl32.Vec3 {
	if n > 0 {
		return m.normals[n-1]
	} else if n < 0 {
		return m.normals[len(m.normals)+n]
	} else {
		panic("missing normals!")
	}
}

type runningMean struct {
	count float32
	mean  mgl32.Vec3
	oldM  mgl32.Vec3
}

func (s *runningMean) push(val mgl32.Vec3) {
	s.count++
	s.mean = s.oldM.Add(val.Sub(s.oldM).Mul(1 / s.count))
	s.oldM = s.mean
}

func abs(x float32) float32 {
	if x >= 0 {
		return x
	}
	return -x
}
