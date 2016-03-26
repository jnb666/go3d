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

const (
	sceneFile = "lighting.qml"
	moveScale = 0.005
)

var (
	lightPos  = glu.Polar{R: 1, Theta: 40, Phi: 90}
	cameraPos = glu.Polar{R: 1.5, Theta: 60, Phi: 60}
	camera    = scene.ArcBallCamera(cameraPos, mgl32.Vec3{}, 0.5, 5.0, 10, 170)
	light     = scene.DirectionalLight(mgl32.Vec3{0.8, 0.8, 0.8}, 0.2, lightPos)
)

type mouseInfo struct {
	x, y, button int
}

type Model struct {
	qml.Object
	mode     string
	texName  string
	scene    *scene.Item
	view     *scene.View
	textures map[string][]glu.Texture
	mouse    mouseInfo
}

func (t *Model) Zoom(amount float32) {
	t.view.Camera.Move(amount)
	t.Call("update")
}

func (t *Model) Mouse(event string, x, y, button int) {
	switch event {
	case "start":
		t.mouse = mouseInfo{x, y, button}
	case "move":
		if t.mouse.button != 0 {
			dx, dy := float32(x-t.mouse.x), float32(y-t.mouse.y)
			if t.mouse.button == 1 {
				axis := t.scene.Mat4.Mul4x1(mgl32.Vec4{1, 0, 0, 0}).Vec3()
				t.scene.Rotate(dy*0.3, axis)
				t.scene.RotateY(dx * 0.3)
			} else {
				t.view.Lights[0].Rotate(dx, dy)
			}
			t.mouse.x, t.mouse.y = x, y
			t.Call("update")
		}
	case "end":
		t.mouse.button = 0
	}
}

func (t *Model) Reset() {
	fmt.Println("reset view")
	t.view = scene.NewView(camera.Clone()).AddLight(light.Clone())
	t.scene.Mat4 = mgl32.Ident4()
	t.Call("update")
}

func (t *Model) Spin() {
	t.scene.RotateY(0.25)
	t.Call("update")
}

func (t *Model) SetTexture(name string) {
	if name != t.texName {
		fmt.Println("set texture", name)
		t.texName = name
		t.Call("update")
	}
}

func (t *Model) SetLighting(mode string) {
	if mode != t.mode {
		fmt.Println("set lighting mode", mode)
		t.mode = mode
		t.Call("update")
	}
}

func getTexture(name string) []glu.Texture {
	fmt.Printf("load %s texture\n", name)
	tex := make([]glu.Texture, 3)
	var err error
	if tex[0], err = glu.NewTexture2D(false, true).SetImageFile(name + ".png"); err != nil {
		panic(err)
	}
	if tex[1], err = glu.NewTexture2D(false, false).SetImageFile(name + "_specular.png"); err != nil {
		panic(err)
	}
	if tex[2], err = glu.NewTexture2D(false, false).SetImageFile(name + "_normal.png"); err != nil {
		panic(err)
	}
	return tex
}

func (t *Model) initialise() *scene.Item {
	fmt.Println("initialise")
	glu.Debug = true
	t.mode = "diffuse"
	t.texName = "brick"
	t.textures = map[string][]glu.Texture{
		"brick":  getTexture("brick"),
		"shield": getTexture("shield"),
	}
	fmt.Println("create scene")
	t.view = scene.NewView(camera.Clone()).AddLight(light.Clone())
	return scene.NewItem(mesh.Cube())
}

func (t *Model) Paint(p *qml.Painter) {
	gl := GL.API(p)
	glu.Init(gl)
	if t.scene == nil {
		t.scene = t.initialise()
	}
	// set material
	var mtl mesh.Material
	switch t.mode {
	case "diffuse":
		mtl = mesh.Diffuse(t.textures[t.texName][0])
	case "specular":
		mtl = mesh.Reflective(glu.White, 64, t.textures[t.texName][:2]...)
	case "normal":
		mtl = mesh.Reflective(glu.White, 64, t.textures[t.texName]...)
	default:
		panic("unknown lighting mode " + t.mode)
	}
	t.scene.SetMaterial(mtl)
	// draw scene
	t.view.SetProjection(t.Int("width"), t.Int("height"))
	glu.Clear(glu.Black)
	view := t.view.ViewMatrix()
	t.view.UpdateLights(view, t.scene)
	t.view.Draw(view, t.scene)
}

func run() error {
	qml.RegisterTypes("GoExtensions", 1, 0, []qml.TypeSpec{{
		Init: func(t *Model, obj qml.Object) { t.Object = obj },
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
