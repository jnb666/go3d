// Package mesh provides functions for managing 3d triangle meshes and material definitions.
package mesh

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/glu"
	"gopkg.in/qml.v1/gl/es2"
	"gopkg.in/qml.v1/gl/glbase"
	"strings"
)

const vertexSize = 8

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

// Mesh type stores a mesh of vertices
type Mesh struct {
	inverted  int
	vdata     []float32
	groups    []meshGroup
	varray    [2]*glu.VertexArray
	vertices  []mgl32.Vec3
	normals   []mgl32.Vec3
	texcoords []mgl32.Vec2
	elements  []El
	faces     int
	ncache    normalCache
	pointSize int
}

type meshGroup struct {
	mtl    Material
	edata  []uint32
	earray *glu.VertexArray
}

type normalCache struct {
	vert2norm map[int]*runningMean
	elem2vert map[int]int
	smooth    int
}

// NewMesh creates a new empty mesh structure
func New() *Mesh {
	return &Mesh{ncache: newNormalCache(), groups: []meshGroup{}}
}

func newNormalCache() normalCache {
	return normalCache{vert2norm: map[int]*runningMean{}, elem2vert: map[int]int{}, smooth: 9999}
}

// Clear method wipes the stored vertex data. It does not erase groups which are already built, call this after Build
// if you need to add a new set of vertices separate to the previous ones.
func (m *Mesh) Clear() *Mesh {
	m.vertices = nil
	m.normals = nil
	m.texcoords = nil
	m.elements = nil
	m.faces = 0
	m.ncache = newNormalCache()
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
		newMesh.groups = append(newMesh.groups, meshGroup{mtl: grp.mtl.Clone(), edata: grp.edata, earray: grp.earray})
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
	if len(el) != 3 && len(el) != 4 {
		panic("AddFace must have 3 or 4 elements")
	}
	calcNormal := false
	vtx := make([]mgl32.Vec3, len(el))
	for i, e := range el {
		if e.Norm == 0 {
			calcNormal = true
		}
		vtx[i] = m.vertex(e.Vert)
	}
	var elements []El
	if len(el) == 3 {
		elements = []El{el[0], el[1], el[2]}
	} else {
		elements = []El{el[0], el[1], el[2], el[2], el[3], el[0]}
	}
	if calcNormal {
		// calculate face normal
		var normal mgl32.Vec3
		if len(el) == 3 {
			normal = vtx[1].Sub(vtx[0]).Cross(vtx[2].Sub(vtx[0]))
		} else {
			for i, v := range vtx {
				v1 := vtx[(i+1)%4]
				normal = normal.Add(mgl32.Vec3{
					(v[1] - v1[1]) * (v[2] + v1[2]),
					(v[2] - v1[2]) * (v[0] + v1[0]),
					(v[0] - v1[0]) * (v[1] + v1[1]),
				})
			}
		}
		m.ncache.add(normal.Normalize(), m.faces, len(m.elements), elements)
	}
	m.elements = append(m.elements, elements...)
	m.faces++
	return len(m.elements)
}

func (m *Mesh) SetNormalSmoothing(opt string) {
	if strings.ToLower(opt) == "off" {
		m.ncache.smooth = 0
	} else {
		m.ncache.smooth = parseint(opt)
	}
}

// update the average normal at each vertex
func (n normalCache) add(normal mgl32.Vec3, face, base int, elements []El) {
	for i, el := range elements {
		counter := n.vert2norm[el.Vert]
		if counter == nil || face-counter.start > n.smooth {
			// start accumulating data for this normal
			n.vert2norm[el.Vert] = &runningMean{start: face}
			counter = n.vert2norm[el.Vert]
		}
		// add to average so far
		counter.push(normal)
		n.elem2vert[base+i] = el.Vert
	}
}

// add calculated normals to mesh
func (n normalCache) build(m *Mesh) {
	for i, vid := range n.elem2vert {
		norm := n.vert2norm[vid].mean
		m.AddNormal(norm[0], norm[1], norm[2])
		m.elements[i].Norm = len(m.normals)
	}
}

// Build method processes the data which has been added so far and appends it to the vertex and element buffers.
// It can be called multiple times to add multiple groups of data.
func (m *Mesh) Build(mtl Material) {
	grp := meshGroup{mtl: mtl}
	if grp.mtl == nil {
		if m.pointSize != 0 {
			grp.mtl = PointMaterial()
		} else {
			grp.mtl = Diffuse()
		}
	}
	m.ncache.build(m)
	m.ncache = newNormalCache()
	cache := map[El]uint32{}
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
	grp.earray = glu.ElementArrayBuffer(grp.edata)
	m.groups = append(m.groups, grp)
	m.elements = nil
	m.faces = 0
}

// Draw method draws the mesh by calling GL DrawElements, setUniforms callback can be used to set uniforms after
// binding the vertex arrays and enabling the shaders, but prior to drawing.
func (m *Mesh) Draw(setUniforms func(*glu.Program)) {
	if m.varray[m.inverted] == nil {
		m.varray[m.inverted] = glu.ArrayBuffer(m.vdata, vertexSize)
	}
	var lastProg *glu.Program
	m.varray[m.inverted].Enable()
	for _, grp := range m.groups {
		grp.earray.Enable()
		prog := grp.mtl.Enable()
		if prog != lastProg {
			setUniforms(prog)
			lastProg = prog
		}
		grp.earray.Draw(GL.TRIANGLES, winding[m.inverted])
		grp.mtl.Disable()
	}
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
	if len(m.groups) > 0 {
		return m.groups[0].mtl
	}
	return nil
}

// Update all the materials associated with this mesh.
func (m *Mesh) SetMaterial(mtl Material) *Mesh {
	for i := range m.groups {
		m.groups[i].mtl = mtl
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

func (m *Mesh) getData(el El) []float32 {
	data := make([]float32, vertexSize)
	v := m.vertex(el.Vert)
	copy(data, v[:])
	vn := m.normal(el.Norm)
	copy(data[3:], vn[:])
	vt := m.texcoord(el.Tex)
	copy(data[6:], vt[:])
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
	start int
	count float32
	mean  mgl32.Vec3
	oldM  mgl32.Vec3
}

func (s *runningMean) push(val mgl32.Vec3) {
	s.count++
	s.mean = s.oldM.Add(val.Sub(s.oldM).Mul(1 / s.count))
	s.oldM = s.mean
}
