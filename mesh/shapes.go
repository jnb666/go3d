package mesh

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/glu"
	"math"
)

var cache = map[key]*Mesh{}

type key struct {
	typ   int
	level int
}

const (
	mPoint = iota
	mPlane
	mCube
	mPrism
	mPyramid
	mCircle
	mCylinder
	mCone
	mIcosohedron
	mSphere
)

// Generate a zero dimensional point
func Point(pointSize int) *Mesh {
	m, ok := cache[key{mPoint, pointSize}]
	if ok {
		return m.Clone()
	}
	if pointSize <= 0 {
		panic("point size must be > 0!")
	}
	// emulate points using a front facing quad
	m = New()
	m.pointSize = pointSize
	m.AddVertex(-0.5, -0.5, 0)
	m.AddVertex(-0.5, 0.5, 0)
	m.AddVertex(0.5, 0.5, 0)
	m.AddVertex(0.5, -0.5, 0)
	m.AddTexCoord(0, 0)
	m.AddTexCoord(0, 1)
	m.AddTexCoord(1, 1)
	m.AddTexCoord(1, 0)
	m.AddNormal(0, 0, 1)
	m.AddFace(El{4, 1, 1}, El{3, 2, 1}, El{2, 3, 1}, El{1, 4, 1})
	m.Build("")
	cache[key{mPoint, 0}] = m
	return m
}

// Plane object is a flat two sided unit square in the xz plane centered at the origin facing in +ve z direction
func Plane() *Mesh {
	m, ok := cache[key{mPlane, 0}]
	if ok {
		return m.Clone()
	}
	m = New()
	m.AddVertex(-0.5, 0, -0.5)
	m.AddVertex(-0.5, 0, 0.5)
	m.AddVertex(0.5, 0, 0.5)
	m.AddVertex(0.5, 0, -0.5)
	m.AddTexCoord(0, 0)
	m.AddTexCoord(0, 1)
	m.AddTexCoord(1, 1)
	m.AddTexCoord(1, 0)
	m.AddNormal(0, 1, 0)
	m.AddFace(El{1, 1, 1}, El{2, 2, 1}, El{3, 3, 1}, El{4, 4, 1})
	m.Build("")
	cache[key{mPlane, 0}] = m
	return m
}

// Cube object is a cube centered at the origin with sides of unit length aligned with the axes
func Cube() *Mesh {
	m, ok := cache[key{mCube, 0}]
	if ok {
		return m.Clone()
	}
	m = New()
	m.AddVertex(-0.5, 0.5, -0.5)
	m.AddVertex(-0.5, 0.5, 0.5)
	m.AddVertex(0.5, 0.5, 0.5)
	m.AddVertex(0.5, 0.5, -0.5)
	m.AddVertex(-0.5, -0.5, -0.5)
	m.AddVertex(-0.5, -0.5, 0.5)
	m.AddVertex(0.5, -0.5, 0.5)
	m.AddVertex(0.5, -0.5, -0.5)
	m.AddTexCoord(0, 0)
	m.AddTexCoord(0, 1)
	m.AddTexCoord(1, 1)
	m.AddTexCoord(1, 0)
	m.AddNormal(0, 1, 0)
	m.AddNormal(-1, 0, 0)
	m.AddNormal(1, 0, 0)
	m.AddNormal(0, 0, -1)
	m.AddNormal(0, 0, 1)
	m.AddNormal(0, -1, 0)
	m.AddFace(El{1, 1, 1}, El{2, 2, 1}, El{3, 3, 1}, El{4, 4, 1})
	m.AddFace(El{1, 1, 2}, El{5, 2, 2}, El{6, 3, 2}, El{2, 4, 2})
	m.AddFace(El{3, 1, 3}, El{7, 2, 3}, El{8, 3, 3}, El{4, 4, 3})
	m.AddFace(El{4, 1, 4}, El{8, 2, 4}, El{5, 3, 4}, El{1, 4, 4})
	m.AddFace(El{2, 1, 5}, El{6, 2, 5}, El{7, 3, 5}, El{3, 4, 5})
	m.AddFace(El{6, 1, 6}, El{5, 2, 6}, El{8, 3, 6}, El{7, 4, 6})
	m.Build("")
	cache[key{mCube, 0}] = m
	return m
}

// Prism object centered on the origin
// base is in xz plane and is square of unit lenght, height is sqrt(3)/2
func Prism() *Mesh {
	m, ok := cache[key{mPrism, 0}]
	if ok {
		return m.Clone()
	}
	m = New()
	h := float32(math.Sqrt(3) / 2)
	m.AddVertex(-0.5, -h/2, -0.5)
	m.AddVertex(-0.5, -h/2, 0.5)
	m.AddVertex(0.5, -h/2, 0.5)
	m.AddVertex(0.5, -h/2, -0.5)
	m.AddVertex(-0.5, h/2, 0)
	m.AddVertex(0.5, h/2, 0)
	m.AddTexCoord(0, 0)
	m.AddTexCoord(0, 1)
	m.AddTexCoord(1, 1)
	m.AddTexCoord(1, 0)
	m.AddTexCoord(0.5, 0)
	m.AddNormal(0, -1, 0)
	m.AddNormal(0, 0.5, -h)
	m.AddNormal(0, 0.5, h)
	m.AddNormal(-1, 0, 0)
	m.AddNormal(1, 0, 0)
	// base
	m.AddFace(El{2, 1, 1}, El{1, 2, 1}, El{4, 3, 1}, El{3, 4, 1})
	// sides
	m.AddFace(El{1, 3, 2}, El{5, 4, 2}, El{6, 1, 2}, El{4, 2, 2})
	m.AddFace(El{3, 3, 3}, El{6, 4, 3}, El{5, 1, 3}, El{2, 2, 3})
	// ends
	m.AddFace(El{2, 3, 4}, El{5, 5, 4}, El{1, 2, 4})
	m.AddFace(El{4, 3, 5}, El{6, 5, 5}, El{3, 2, 5})
	m.Build("")
	cache[key{mPrism, 0}] = m
	return m
}

// Circle is a flat circlular triangle fan with given number of segments
func Circle(segments int) *Mesh {
	m, ok := cache[key{mCircle, segments}]
	if ok {
		return m.Clone()
	}
	m = New()
	pts := getCircle(segments)
	doCircle(m, pts, 0, 1)
	m.Build("")
	cache[key{mCircle, segments}] = m
	return m
}

// draw a circle around the last vertex in the xz plane
func doCircle(m *Mesh, pts []mgl32.Vec2, y, yNormal float32) int {
	norm := m.AddNormal(0, yNormal, 0)
	base := m.AddVertex(0, y, 0)
	texc := m.AddTexCoord(0.5, 0.5)
	centre := El{base, texc, norm}
	for _, pt := range pts {
		x, z := 0.5*pt[0], 0.5*pt[1]
		m.AddVertex(x, y, z)
		m.AddTexCoord(0.5+x, 0.5+yNormal*z)
	}
	for ix := range pts {
		var p1, p2 El
		if ix == 0 {
			p1, p2 = El{base + len(pts), texc + len(pts), norm}, El{base + ix + 1, texc + ix + 1, norm}
		} else {
			p1, p2 = El{base + ix, texc + ix, norm}, El{base + ix + 1, texc + ix + 1, norm}
		}
		if yNormal < 0 {
			m.AddFace(centre, p1, p2)
		} else {
			m.AddFace(centre, p2, p1)
		}
	}
	return base
}

func getCircle(segments int) []mgl32.Vec2 {
	res := make([]mgl32.Vec2, segments)
	for i := range res {
		angle := 2 * math.Pi * float64(i) / float64(segments)
		sina, cosa := math.Sincos(angle)
		res[i] = mgl32.Vec2{float32(cosa), float32(sina)}
	}
	return res
}

// Cylinder object has a circle with unit diameter aligned with the y axis and is of unit height
func Cylinder(segments int) *Mesh {
	m, ok := cache[key{mCylinder, segments}]
	if ok {
		return m.Clone()
	}
	m = New()
	// ends
	pts := getCircle(segments)
	top := doCircle(m, pts, 0.5, 1)
	base := doCircle(m, pts, -0.5, -1)
	// sides
	for i, pt := range pts {
		tx := 3 * (1 - float32(i)/float32(segments))
		norm := m.AddNormal(pt[0], 0, pt[1])
		t1 := m.AddTexCoord(tx, 1)
		t2 := m.AddTexCoord(tx, 0)
		if i > 0 {
			m.AddFace(El{base + i, t1 - 2, norm - 1}, El{top + i, t2 - 2, norm - 1},
				El{top + i + 1, t2, norm}, El{base + i + 1, t1, norm})
		}
	}
	// close the cylinder
	m.AddTexCoord(0, 1)
	m.AddTexCoord(0, 0)
	m.AddFace(El{base + segments, -3, -1}, El{top + segments, -4, -1},
		El{top + 1, -2, -segments}, El{base + 1, -1, -segments})
	m.Build("")
	cache[key{mCylinder, segments}] = m
	return m
}

// Cone object has a circular base with unit diameter and unit height aligned with the y axis
// if 8 sides or less then map 2d texture to each side, else wrap it around
func Cone(segments int) *Mesh {
	m, ok := cache[key{mCone, segments}]
	if ok {
		return m.Clone()
	}
	m = New()
	// top
	m.AddVertex(0, 0.5, 0)
	m.AddTexCoord(0.5, 0)
	// base
	m.AddTexCoord(1, 1)
	m.AddTexCoord(0, 1)
	base := doCircle(m, getCircle(segments), -0.5, -1)
	// sides
	n := float32(1 / math.Sqrt2)
	pts := getCircle(2 * segments)
	for i := 0; i < segments; i++ {
		fnorm := m.AddNormal(n*pts[2*i+1][0], n, n*pts[2*i+1][1])
		if segments <= 8 {
			// normal for each face
			if i > 0 {
				m.AddFace(El{base + i, 2, fnorm - 1}, El{1, 1, fnorm - 1}, El{base + i + 1, 3, fnorm - 1})
			}
		} else {
			tx := 2 * (1 - float32(i)/float32(segments))
			t1 := m.AddTexCoord(tx, 1)
			t2 := m.AddTexCoord(tx, 0)
			// normal for each vertex to make this smooth
			vnorm := m.AddNormal(n*pts[2*i][0], n, n*pts[2*i][1])
			if i > 0 {
				m.AddFace(El{base + i, t1 - 2, vnorm - 2}, El{1, t2 - 2, fnorm}, El{base + i + 1, t1, vnorm})
			}
		}
	}
	// close the surface
	if segments <= 8 {
		m.AddFace(El{base + segments, 2, -1}, El{1, 1, -1}, El{base + 1, 3, -1})
	} else {
		m.AddFace(El{base + segments, -2, -1}, El{1, -1, -1}, El{base + 1, 3, -2 * segments})
	}
	m.Build("")
	cache[key{mCone, segments}] = m
	return m
}

// Create an icosohedron on a sphere of unit diameter centered on the origin
func Icosohedron() *Mesh {
	m, ok := cache[key{mIcosohedron, 0}]
	if ok {
		return m.Clone()
	}
	m = New()
	m.SetNormalSmoothing(false)
	faces := doIcosohedron(m)
	m.addElementTriangles(faces, false)
	m.Build("")
	cache[key{mIcosohedron, 0}] = m
	return m
}

type point struct {
	x, y float32
}

func (m *Mesh) addElementTriangles(faces [][3]int, smooth bool) {
	elem := make([]El, 3)
	tex := make([]point, 3)
	for _, face := range faces {
		for i, ix := range face {
			polar := new(glu.Polar).Set(m.vertex(ix).Normalize())
			tex[i] = point{x: 1 + polar.Phi/180, y: polar.Theta / 180}
		}
		for i, ix := range face {
			elem[i].Vert = ix
			if smooth {
				elem[i].Norm = ix
			}
			// wrap around the texture to eliminate longditudinal seam
			if tex[i].x > 1.9 && (tex[(i+1)%3].x < 0.1 || tex[(i+2)%3].x < 0.1) {
				tex[i].x -= 2
			}
			elem[i].Tex = m.AddTexCoord(tex[i].x, tex[i].y)
		}
		m.AddFace(elem...)
	}
}

func doIcosohedron(m *Mesh) [][3]int {
	t := float32(1+math.Sqrt(5)) / 2

	// 12 points of the icosohedron
	m.addSphereVertex(mgl32.Vec3{-1, t, 0})
	m.addSphereVertex(mgl32.Vec3{1, t, 0})
	m.addSphereVertex(mgl32.Vec3{-1, -t, 0})
	m.addSphereVertex(mgl32.Vec3{1, -t, 0})

	m.addSphereVertex(mgl32.Vec3{0, -1, t})
	m.addSphereVertex(mgl32.Vec3{0, 1, t})
	m.addSphereVertex(mgl32.Vec3{0, -1, -t})
	m.addSphereVertex(mgl32.Vec3{0, 1, -t})

	m.addSphereVertex(mgl32.Vec3{t, 0, -1})
	m.addSphereVertex(mgl32.Vec3{t, 0, 1})
	m.addSphereVertex(mgl32.Vec3{-t, 0, -1})
	m.addSphereVertex(mgl32.Vec3{-t, 0, 1})

	// 20 faces in total
	return [][3]int{
		{1, 12, 6}, {1, 6, 2}, {1, 2, 8}, {1, 8, 11}, {1, 11, 12},
		{2, 6, 10}, {6, 12, 5}, {12, 11, 3}, {11, 8, 7}, {8, 2, 9},
		{4, 10, 5}, {4, 5, 3}, {4, 3, 7}, {4, 7, 9}, {4, 9, 10},
		{5, 10, 6}, {3, 5, 12}, {7, 3, 11}, {9, 7, 8}, {10, 9, 2},
	}
}

// Create a sphere by recursively dividing the faces on the icosohedron into smaller triangles
func Sphere(recursionLevel int) *Mesh {
	m, ok := cache[key{mSphere, recursionLevel}]
	if ok {
		return m.Clone()
	}
	m = New()
	faces := doIcosohedron(m)
	scache := make(map[[2]int]int)
	for i := 0; i < recursionLevel; i++ {
		faces2 := [][3]int{}
		for _, tri := range faces {
			a := m.getMiddlePoint(scache, [2]int{tri[0], tri[1]})
			b := m.getMiddlePoint(scache, [2]int{tri[1], tri[2]})
			c := m.getMiddlePoint(scache, [2]int{tri[2], tri[0]})
			faces2 = append(faces2, [3]int{tri[0], a, c})
			faces2 = append(faces2, [3]int{tri[1], b, a})
			faces2 = append(faces2, [3]int{tri[2], c, b})
			faces2 = append(faces2, [3]int{a, b, c})
		}
		faces = faces2
	}
	m.addElementTriangles(faces, true)
	m.Build("")
	cache[key{mSphere, recursionLevel}] = m
	return m
}

// create a new vertex midway between two given points
func (m *Mesh) getMiddlePoint(scache map[[2]int]int, p [2]int) int {
	// order such that p[0] < p[1]
	if p[1] < p[0] {
		p[0], p[1] = p[1], p[0]
	}
	// check the cache to see if already done
	if val, ok := scache[p]; ok {
		return val
	}
	// calc the point
	index := m.addSphereVertex(m.vertex(p[0]).Add(m.vertex(p[1])))
	scache[p] = index
	return index
}

func (m *Mesh) addSphereVertex(v mgl32.Vec3) int {
	v = v.Normalize()
	m.AddNormal(v[0], v[1], v[2])
	return m.AddVertex(v[0]/2, v[1]/2, v[2]/2)
}

// Return a range of indices from start to end, inclusive with step of step
func irange(start, end, step int) []int {
	r := make([]int, 0, (end-start)/step)
	for i := start; (step > 0 && i <= end) || (step < 0 && i >= end); i += step {
		r = append(r, i)
	}
	return r
}
