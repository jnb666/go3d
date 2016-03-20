package main

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/glu"
	"github.com/jnb666/go3d/mesh"
	"github.com/jnb666/go3d/scene"
	"gopkg.in/qml.v1"
	"gopkg.in/qml.v1/gl/es2"
	"os"
)

const sceneFile = "gopher.qml"

var (
	cameraPos = glu.Polar{R: 2.2, Theta: 60, Phi: 45}
	camera    = scene.ArcBallCamera(cameraPos, mgl32.Vec3{}, 0, 1, 0, 180)
	light     = scene.DirectionalLight(mgl32.Vec3{0.8, 0.8, 0.8}, 0.2, cameraPos)
)

type GopherCube struct {
	qml.Object
	step  float32
	world scene.Object
	view  *scene.View
}

type ReflectSurface struct {
	mesh.Material
}

func (m ReflectSurface) Enable() *glu.Program {
	gl := glu.GLRef()
	gl.Enable(GL.STENCIL_TEST)
	gl.StencilFunc(GL.ALWAYS, 1, 0xFF)
	gl.StencilOp(GL.KEEP, GL.KEEP, GL.REPLACE)
	gl.StencilMask(0xFF)
	gl.DepthMask(false)
	gl.Clear(GL.STENCIL_BUFFER_BIT)
	return m.Material.Enable()
}

func (m ReflectSurface) Disable() {
	gl := glu.GLRef()
	gl.DepthMask(true)
	gl.Disable(GL.STENCIL_TEST)
}

type ReflectImage struct {
	mesh.Material
}

func (m ReflectImage) Enable() *glu.Program {
	gl := glu.GLRef()
	gl.Enable(GL.STENCIL_TEST)
	gl.StencilFunc(GL.EQUAL, 1, 0xFF)
	gl.StencilMask(0x00)
	return m.Material.Enable()
}

func (m ReflectImage) Disable() {
	gl := glu.GLRef()
	gl.Disable(GL.STENCIL_TEST)
}

func (t *GopherCube) initGL(gl *GL.GL) scene.Object {
	fmt.Println("initialise GL")
	glu.Init(gl)
	t.view = scene.NewView(camera).AddLight(light)

	// load texture
	fmt.Println("load gopher texture")
	tex, err := glu.NewTexture2D(false, true).SetImageFile("gopher.png")
	if err != nil {
		panic(err)
	}

	// materials
	gopher := mesh.DiffuseTex(tex)
	floor := ReflectSurface{mesh.Unshaded()}
	floor.SetColor(glu.Black)
	reflect := ReflectImage{mesh.DiffuseTex(tex)}
	reflect.SetColor(mgl32.Vec4{0.3, 0.3, 0.3, 1})

	// world model
	objs := scene.NewGroup()
	objs.Add(scene.NewItem(mesh.Cube()).SetMaterial(gopher))
	objs.Add(scene.NewItem(mesh.Plane()).SetMaterial(floor).Scale(2, 1, 2).Translate(0, -0.5, 0))
	objs.Add(scene.NewItem(mesh.Cube().Invert()).SetMaterial(reflect).Scale(1, -1, 1).Translate(0, -1, 0))
	objs.Translate(0, 0.2, 0)
	return objs
}

func (t *GopherCube) SetStep(s float32) {
	t.step = s
}

func (t *GopherCube) Rotate() {
	t.world.RotateY(-t.step)
	t.Call("update")
}

func (t *GopherCube) Paint(p *qml.Painter) {
	gl := GL.API(p)
	if t.world == nil {
		t.world = t.initGL(gl)
	}
	t.view.SetProjection(t.Int("width"), t.Int("height"))
	glu.Clear(glu.Blue)
	trans := t.view.Camera.ViewMatrix()
	t.view.UpdateLights(trans, nil)
	t.view.Draw(trans, t.world)
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
