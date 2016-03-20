package main

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/glu"
	"gopkg.in/qml.v1"
	"gopkg.in/qml.v1/gl/es2"
	"math"
	"os"
)

const (
	sceneFile   = "gopher.qml"
	eyeDist     = 1.8
	nearDist    = 1
	farDist     = 10
	fieldOfView = 45
)

var (
	vertices = []float32{
		-0.5, -0.5, -0.5, 0.0, 0.0,
		0.5, -0.5, -0.5, 1.0, 0.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		-0.5, 0.5, -0.5, 0.0, 1.0,

		-0.5, -0.5, 0.5, 0.0, 0.0,
		0.5, -0.5, 0.5, 1.0, 0.0,
		0.5, 0.5, 0.5, 1.0, 1.0,
		-0.5, 0.5, 0.5, 0.0, 1.0,

		-0.5, 0.5, 0.5, 1.0, 0.0,
		-0.5, 0.5, -0.5, 1.0, 1.0,
		-0.5, -0.5, -0.5, 0.0, 1.0,
		-0.5, -0.5, 0.5, 0.0, 0.0,

		0.5, 0.5, 0.5, 1.0, 0.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		0.5, -0.5, -0.5, 0.0, 1.0,
		0.5, -0.5, 0.5, 0.0, 0.0,

		-0.5, -0.5, -0.5, 0.0, 1.0,
		0.5, -0.5, -0.5, 1.0, 1.0,
		0.5, -0.5, 0.5, 1.0, 0.0,
		-0.5, -0.5, 0.5, 0.0, 0.0,

		-0.5, 0.5, -0.5, 0.0, 1.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		0.5, 0.5, 0.5, 1.0, 0.0,
		-0.5, 0.5, 0.5, 0.0, 0.0,

		-1.0, -1.0, -0.5, 0.0, 0.0,
		1.0, -1.0, -0.5, 1.0, 0.0,
		1.0, 1.0, -0.5, 1.0, 1.0,
		-1.0, 1.0, -0.5, 0.0, 1.0,
	}
	elements = []uint32{
		0, 1, 2, 2, 3, 0,
		4, 5, 6, 6, 7, 4,
		8, 9, 10, 10, 11, 8,
		12, 13, 14, 14, 15, 12,
		16, 17, 18, 18, 19, 16,
		20, 21, 22, 22, 23, 20,
		24, 25, 26, 26, 27, 24,
	}
	attribs = []glu.Attrib{
		{Name: "position", Size: 3, Offset: 0},
		{Name: "texcoord", Size: 2, Offset: 3},
	}
	vertexSize   = 5
	vertexShader = `
attribute vec3 position;
attribute vec2 texcoord;
varying vec2 Texcoord;
uniform mat4 model;
uniform mat4 view;
uniform mat4 proj;

void main() {
	Texcoord = texcoord;
    gl_Position = proj * view * model * vec4(position, 1.0);
}
`
	fragmentShader = `
varying vec2 Texcoord;
uniform sampler2D tex;
uniform vec3 color;

void main() {
    gl_FragColor = vec4(color, 1.0) * texture2D(tex, Texcoord);
}
`
)

type GopherCube struct {
	qml.Object
	step     float32
	angle    float32
	program  *glu.Program
	vertices *glu.VertexArray
	cube     *glu.VertexArray
	floor    *glu.VertexArray
	texture  glu.Texture
}

func (t *GopherCube) initGL(gl *GL.GL) {
	fmt.Println("initialise GL")

	// vertex arrays
	t.vertices = glu.ArrayBuffer(vertices, vertexSize)
	t.cube = glu.ElementArrayBuffer(elements[:36])
	t.floor = glu.ElementArrayBuffer(elements[36:])

	// load texture
	img, err := glu.PNGImage("gopher.png")
	if err != nil {
		panic(err)
	}
	t.texture = glu.NewTexture2D(GL.CLAMP_TO_EDGE).SetImage(img, false)

	// compile program
	if t.program, err = glu.NewProgram(vertexShader, fragmentShader, attribs, vertexSize); err != nil {
		panic(err)
	}
	t.program.Uniform("v3f", "color")
	t.program.Uniform("m4f", "model", "view", "proj")
	t.program.Uniform("1i", "tex")
}

func (t *GopherCube) SetStep(s float32) {
	t.step = s
}

func (t *GopherCube) Rotate() {
	t.angle -= t.step
	if t.angle > 360 {
		t.angle -= 360
	}
	t.Call("update")
}

func (t *GopherCube) Paint(p *qml.Painter) {
	gl := GL.API(p)
	glu.Init(gl)
	if t.program == nil {
		t.initGL(gl)
	}
	t.texture.Activate()
	t.vertices.Enable()
	t.cube.Enable()
	t.program.Use()

	// set program uniforms
	aspect := float32(t.Int("width")) / float32(t.Int("height"))
	proj := mgl32.Perspective(mgl32.DegToRad(fieldOfView), aspect, nearDist, farDist)
	t.program.Set("proj", proj.Mul4(mgl32.HomogRotate3DZ(math.Pi)))
	t.program.Set("view", mgl32.LookAt(eyeDist, eyeDist, eyeDist, 0, 0, 0, 0, 0, 1))
	m := mgl32.HomogRotate3DZ(mgl32.DegToRad(t.angle))
	m = m.Mul4(mgl32.Translate3D(0, 0, 0.3))
	t.program.Set("model", m)
	t.program.Set("color", mgl32.Vec3{1.0, 1.0, 1.0})
	t.program.Set("tex", 0)

	// draw the cube
	gl.Enable(GL.DEPTH_TEST)
	gl.ClearColor(0, 0, 1, 1)
	gl.Clear(GL.COLOR_BUFFER_BIT | GL.DEPTH_BUFFER_BIT)
	t.cube.Draw(GL.TRIANGLES, GL.CCW)

	// draw the floor
	t.floor.Enable()
	t.program.Use()
	t.program.Set("color", mgl32.Vec3{0, 0, 0})
	gl.Enable(GL.STENCIL_TEST)
	gl.StencilFunc(GL.ALWAYS, 1, 0xFF)
	gl.StencilOp(GL.KEEP, GL.KEEP, GL.REPLACE)
	gl.StencilMask(0xFF)
	gl.DepthMask(false)
	gl.Clear(GL.STENCIL_BUFFER_BIT)
	t.floor.Draw(GL.TRIANGLES, GL.CCW)

	// draw reflection
	t.cube.Enable()
	t.program.Use()
	t.program.Set("color", mgl32.Vec3{0.3, 0.3, 0.3})
	gl.StencilFunc(GL.EQUAL, 1, 0xFF)
	gl.StencilMask(0x00)
	gl.DepthMask(true)
	m2 := m.Mul4(mgl32.Translate3D(0, 0, -1)).Mul4(mgl32.Scale3D(1, 1, -1))
	t.program.Set("model", m2)
	t.cube.Draw(GL.TRIANGLES, GL.CCW)
	gl.Disable(GL.STENCIL_TEST)
}

func run() error {
	qml.RegisterTypes("GoExtensions", 1, 0, []qml.TypeSpec{{
		Init: func(t *GopherCube, obj qml.Object) { t.Object = obj },
	}})

	engine := qml.NewEngine()
	engine.On("quit", func() { os.Exit(0) })
	component, err := engine.LoadFile(sceneFile)
	if err != nil {
		return err
	}

	window := component.CreateWindow(nil)
	window.Show()
	window.Wait()
	return nil
}

func main() {
	if err := qml.Run(run); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}
