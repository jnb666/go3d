package scene

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/glu"
	"github.com/jnb666/go3d/mesh"
)

// Default projection settings
var (
	FOV, Near, Far float32    = 45, 0.1, 50
	Up             mgl32.Vec3 = mgl32.Vec3{0, 1, 0}
	ZoomScale      float32    = 1.05
	RotateScale    float32    = 1.0
	StepSize       float32    = 0.05
)

// View settings
type View struct {
	Camera Camera
	Lights []*Light
	Proj   mgl32.Mat4
	ldata  []*Light
	width  float32
	height float32
}

// Setup a new view, makes a copy of the camera which was passed in
func NewView(camera Camera) *View {
	v := new(View)
	v.Camera = camera
	v.Lights = []*Light{}
	return v
}

// Draw the scene with the given view matrix
func (v *View) Draw(worldToCamera mgl32.Mat4, scene Object) {
	scene.Do(NewTransform(worldToCamera), func(o *Item, t Transform) {
		err := o.Mesh.Draw(func(prog *glu.Program) {
			mat := t.Mat4
			if psize := o.Mesh.PointSize(); psize != 0 {
				// points are always facing the camera at a constant size
				pos := mgl32.Vec3{mat[12], mat[13], mat[14]}
				sc := 2 * float32(psize) * pos.Len() / v.height
				mat = mgl32.Mat4{sc, 0, 0, 0, 0, sc, 0, 0, 0, 0, sc, 0, pos[0], pos[1], pos[2], 1}
				prog.Set("pointLocation", pos)
				prog.Set("pointSize", float32(psize))
				prog.Set("viewport", v.width, v.height)
			} else {
				prog.Set("texScale", o.TexScale)
				prog.Set("normalModelToCamera", mat.Mat3().Inv().Transpose())
				prog.Set("modelScale", t.Scale)
				prog.Set("numLights", len(v.ldata))
				for i, light := range v.ldata {
					prog.SetArray("lightPos", i, light.Pos)
					prog.SetArray("lightCol", i, light.Col)
				}
			}
			prog.Set("cameraToClip", v.Proj)
			prog.Set("modelToCamera", mat)
		})
		if err != nil {
			// seems better to panic as caller might otherwise skip checking the error
			panic(err)
		}
	})
}

// Add a new light to the scene
func (v *View) AddLight(l *Light) *View {
	if len(v.Lights) < mesh.MaxLights {
		v.Lights = append(v.Lights, l)
	} else {
		panic("exceeded maximum number of lights!")
	}
	return v
}

// Update the lighting data prior to drawing the scene
func (v *View) UpdateLights(worldToCamera mgl32.Mat4, scene Object) *View {
	v.ldata = []*Light{}
	// static lights
	for _, l := range v.Lights {
		v.addLight(l, worldToCamera)
	}
	// lights attaced to objects in the scene
	if scene != nil {
		trans := NewTransform(worldToCamera)
		scene.Do(trans, func(o *Item, t Transform) {
			v.addLight(o.Light, t.Mat4)
		})
	}
	return v
}

func (v *View) addLight(l *Light, trans mgl32.Mat4) {
	if l != nil && l.On {
		light := *l
		light.Pos = trans.Mul4x1(l.Pos.Vec3().Vec4(l.posw))
		light.Pos[3] = l.Pos[3]
		v.ldata = append(v.ldata, &light)
	}
}

// Set the projection matrix
func (v *View) SetProjection(width, height int) {
	aspect := float32(width) / float32(height)
	// this is inverted as QML puts the y axis upside down
	v.Proj = mgl32.Perspective(FOV, aspect, Near, Far).Mul4(mgl32.Scale3D(1, -1, 1))
	v.width, v.height = float32(width), float32(height)
}

// Get the camera view matrix.
func (v *View) ViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(v.Camera.Eye(), v.Camera.Center(), Up)
}

// view centerd on camera
func (v *View) CenteredView() mgl32.Mat4 {
	pos := v.Camera.Eye()
	return v.ViewMatrix().Mul4(mgl32.Translate3D(pos[0], pos[1], pos[2]))
}

// Camera interface type defines the viewing position
type Camera interface {
	Eye() mgl32.Vec3
	Center() mgl32.Vec3
	Move(dir float32)
	Rotate(dx, dy float32)
	Clone() Camera
}

type povCamera struct {
	pos mgl32.Vec3
	dir mgl32.Vec3
}

// Create a new point of view camera with yaw and pitch controls.
func POVCamera(pos mgl32.Vec3, dir mgl32.Vec3) Camera {
	return &povCamera{pos: pos, dir: dir.Normalize()}
}

func (c *povCamera) Clone() Camera {
	cam := *c
	return &cam
}

func (c *povCamera) Eye() mgl32.Vec3 {
	return c.pos
}

func (c *povCamera) Center() mgl32.Vec3 {
	return c.pos.Add(c.dir)
}

// Step forwards if amount > 0 or backwards if amount <0
func (c *povCamera) Move(amount float32) {
	c.pos = c.pos.Add(c.dir.Mul(StepSize * amount))
}

// Change the direction of the camera: dx controls the yaw, dy controls the pitch
func (c *povCamera) Rotate(dx, dy float32) {
	c.doRotate(dx, mgl32.Vec3{0, 1, 0})
	c.doRotate(dy, c.dir.Cross(Up).Normalize())
}

func (c *povCamera) doRotate(step float32, axis mgl32.Vec3) {
	angle := mgl32.DegToRad(step) * RotateScale
	temp := mgl32.QuatRotate(angle, axis)
	view := mgl32.Quat{V: c.dir}
	result := temp.Mul(view).Mul(temp.Conjugate())
	c.dir = result.V
}

type arcBallCamera struct {
	toEye      glu.Polar
	center     mgl32.Vec3
	minz, maxz float32
	mint, maxt float32
}

// Create a new camera positioned at center + toEye vector and looking at center
func ArcBallCamera(toEye glu.Polar, center mgl32.Vec3, minZ, maxZ, minTheta, maxTheta float32) Camera {
	return &arcBallCamera{toEye: toEye, center: center, minz: minZ, maxz: maxZ, mint: minTheta, maxt: maxTheta}
}

func (c *arcBallCamera) Clone() Camera {
	cam := *c
	return &cam
}

func (c *arcBallCamera) Eye() mgl32.Vec3 {
	return c.center.Add(c.toEye.Vec3())
}

func (c *arcBallCamera) Center() mgl32.Vec3 {
	return c.center
}

// Move towards center if dir > 0 or away if dir < 0
func (c *arcBallCamera) Move(amount float32) {
	if amount < 0 {
		c.toEye.R *= ZoomScale
	} else if amount > 0 {
		c.toEye.R *= 1.0 / ZoomScale

	}
	c.toEye.R = glu.Clamp(c.toEye.R, c.minz, c.maxz)
}

// Rotate the position of the camera around the origin, where dy is relative to up axis and dx
// is around a circle in xz plane.
func (c *arcBallCamera) Rotate(dx, dy float32) {
	c.toEye.Phi -= dx * RotateScale
	c.toEye.Theta -= dy * RotateScale
	c.toEye.Theta = glu.Clamp(c.toEye.Theta, c.mint, c.maxt)
	c.toEye.Clamp()
}

// Light struct represents a light source. Col.W() is the ambient scaling factor.
// Pos.W() is the attenuation or 0 for a directional light.
type Light struct {
	Pos  mgl32.Vec4
	Col  mgl32.Vec4
	On   bool
	posw float32
}

// Directional light source
func DirectionalLight(color mgl32.Vec3, ambient float32, direction glu.Polar) *Light {
	direction.R = 1
	return &Light{
		Pos: direction.Vec4(0),
		Col: color.Vec4(ambient),
		On:  true,
	}
}

// Point light source
func PointLight(color mgl32.Vec3, ambient float32, position mgl32.Vec3, attenuation float32) *Light {
	return &Light{
		Pos:  position.Vec4(attenuation),
		Col:  color.Vec4(ambient),
		On:   true,
		posw: 1,
	}
}

// Rotate position of directional light
func (l *Light) Rotate(dx, dy float32) *Light {
	if l.Pos.W() == 0 {
		polar := new(glu.Polar).Set(l.Pos.Vec3())
		polar.Phi -= float32(dx) * RotateScale
		polar.Theta -= float32(dy) * RotateScale
		l.Pos = polar.Vec4(0)
	} else {
		panic("move not implemented for point lights")
	}
	return l
}
