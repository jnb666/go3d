package main

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/glu"
	"gopkg.in/qml.v1"
	"gopkg.in/qml.v1/gl/es2"
	"os"
)

const sceneFile = "triangle.qml"

var (
	attribs = []glu.Attrib{
		{Name: "position", Size: 2, Offset: 0},
		{Name: "color", Size: 3, Offset: 2},
	}
	vertexSize = 5
	vertices   = []float32{
		-.6, -.4, 1, 0, 0,
		0.6, -.4, 0, 1, 0,
		0.0, 0.6, 0, 0, 1,
	}
	vertexShader = `
attribute vec2 position;
attribute vec3 color;
varying vec3 Color;
uniform mat4 trans;

void main() {
	Color = color;
    gl_Position = trans * vec4(position, 0.0, 1.0);
}
`
	fragmentShader = `
varying vec3 Color;

void main() {
    gl_FragColor = vec4(Color, 1.0);
}
`
)

type Triangle struct {
	qml.Object
	angle   float32
	program *glu.Program
	buffer  *glu.VertexArray
}

func (t *Triangle) initGL(gl *GL.GL) {
	fmt.Println("initialise GL")
	glu.Debug = true
	var err error
	if t.program, err = glu.NewProgram(vertexShader, fragmentShader, attribs, vertexSize); err != nil {
		panic(err)
	}
	t.program.Uniform("m4f", "trans")
	t.buffer = glu.NewArray(vertices, nil, 5)
}

func (t *Triangle) Rotate() {
	t.angle += 1
	if t.angle > 360 {
		t.angle -= 360
	}
	t.Call("update")
}

func (t *Triangle) Paint(p *qml.Painter) {
	gl := GL.API(p)
	glu.Init(gl)
	if t.program == nil {
		t.initGL(gl)
	}
	t.buffer.Enable()
	t.program.Use()
	t.program.Set("trans", mgl32.HomogRotate3DZ(mgl32.DegToRad(t.angle)))
	glu.Clear(glu.Black)
	t.buffer.Draw(GL.TRIANGLES, GL.CCW)
}

func run() error {
	qml.RegisterTypes("GoExtensions", 1, 0, []qml.TypeSpec{{
		Init: func(t *Triangle, obj qml.Object) { t.Object = obj },
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
