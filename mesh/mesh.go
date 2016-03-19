// Package mesh provides functions for managing 3d triangle meshes and material definitions.
package mesh

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/glu"
	"gopkg.in/qml.v1/gl/es2"
	"gopkg.in/qml.v1/gl/glbase"
)

const vertexSize = 8

var vertexLayout = []glu.Attrib{
	{Name: "position", Size: 3, Offset: 0},
	{Name: "normal", Size: 3, Offset: 3},
	{Name: "texcoord", Size: 2, Offset: 6},
}

var vertexLayoutPoints = []glu.Attrib{
	{Name: "position", Size: 3, Offset: 0},
	//	{Name: "texcoord", Size: 2, Offset: 6},
}

var winding = [2]glbase.Enum{GL.CW, GL.CCW}

type El struct {
	Vert, Tex, Norm int
}

// Mesh type stores a mesh of vertices
type Mesh struct {
	inverted  int
	vdata     []float32
	edata     []uint32
	array     [2]*glu.VertexArray
	vertices  []mgl32.Vec3
	normals   []mgl32.Vec3
	texcoords []mgl32.Vec2
	elements  []El
	ncache    normalCache
	pointSize int
}

type normalCache struct {
	vert2norm map[int]*runningMean
	elem2vert map[int]int
}

// NewMesh creates a new empty mesh structure
func New() *Mesh {
	return &Mesh{ncache: newNormalCache()}
}

func newNormalCache() normalCache {
	return normalCache{vert2norm: map[int]*runningMean{}, elem2vert: map[int]int{}}
}

// Clear method wipes the stored vertex data
func (m *Mesh) Clear() *Mesh {
	m.vertices = nil
	m.normals = nil
	m.texcoords = nil
	m.elements = nil
	m.ncache = newNormalCache()
	return m
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

// Add a triangular face
func (m *Mesh) AddFace(f1, f2, f3 El) int {
	if f1.Norm == 0 || f2.Norm == 0 || f3.Norm == 0 {
		// calculate face normal
		v1, v2, v3 := m.vertex(f1.Vert), m.vertex(f2.Vert), m.vertex(f3.Vert)
		normal := v2.Sub(v1).Cross(v3.Sub(v1)).Normalize()
		m.ncache.add(normal, len(m.elements), f1, f2, f3)
	}
	m.elements = append(m.elements, f1, f2, f3)
	return len(m.elements)
}

// Add a quad face
func (m *Mesh) AddFaceQuad(f1, f2, f3, f4 El) int {
	if f1.Norm == 0 || f2.Norm == 0 || f3.Norm == 0 || f4.Norm == 0 {
		// calculate face normal using Newells method
		vn := mgl32.Vec3{}
		verts := []mgl32.Vec3{m.vertex(f1.Vert), m.vertex(f2.Vert), m.vertex(f3.Vert), m.vertex(f4.Vert)}
		for i, v := range verts {
			v1 := verts[(i+1)%4]
			vn = vn.Add(mgl32.Vec3{
				(v[1] - v1[1]) * (v[2] + v1[2]),
				(v[2] - v1[2]) * (v[0] + v1[0]),
				(v[0] - v1[0]) * (v[1] + v1[1]),
			})
		}
		normal := vn.Normalize()
		m.ncache.add(normal, len(m.elements), f1, f2, f3, f3, f4, f1)
	}
	m.elements = append(m.elements, f1, f2, f3, f3, f4, f1)
	return len(m.elements)
}

// update the average normal at each vertex
func (n normalCache) add(normal mgl32.Vec3, base int, elements ...El) {
	for i, el := range elements {
		if n.vert2norm[el.Vert] == nil {
			n.vert2norm[el.Vert] = &runningMean{}
		}
		n.vert2norm[el.Vert].push(normal)
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
func (m *Mesh) Build() {
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
		m.edata = append(m.edata, index)
	}
	m.elements = nil
	fmt.Printf("triangle mesh: %d vertices, %d elements\n", len(m.vdata)/vertexSize, len(m.edata))
}

// Enable method loads the vertex and element data on the GPU prior to drawing.
func (m *Mesh) Enable() {
	if m.array[m.inverted] == nil {
		m.array[m.inverted] = glu.NewArray(m.vdata, m.edata, vertexSize)
	} else {
		m.array[m.inverted].Enable()
	}
}

// Draw method draws the mesh by calling GL DrawElements
func (m *Mesh) Draw() {
	m.array[m.inverted].Draw(GL.TRIANGLES, winding[m.inverted])
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
	count float32
	mean  mgl32.Vec3
	oldM  mgl32.Vec3
}

func (s *runningMean) push(val mgl32.Vec3) {
	s.count++
	s.mean = s.oldM.Add(val.Sub(s.oldM).Mul(1 / s.count))
	s.oldM = s.mean
}
