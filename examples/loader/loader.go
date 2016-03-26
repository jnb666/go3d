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
	"sync"
)

const sceneFile = "shapes.qml"

var (
	cameraPos = glu.Polar{R: 2.0, Theta: 70, Phi: 45}
	lightPos  = glu.Polar{R: 1, Theta: 20, Phi: 90}
	rotCamera = scene.ArcBallCamera(cameraPos, mgl32.Vec3{0, 0, 0}, 1, 3, 10, 170)
	povCamera = scene.POVCamera(mgl32.Vec3{2, -0.25, 0}, mgl32.Vec3{-1, 0.25, 0})
	light     = scene.DirectionalLight(mgl32.Vec3{0.8, 0.8, 0.8}, 0.2, lightPos)
	meshes    = map[string]string{
		"cube":    "cube.obj",
		"teapot":  "teapot.obj",
		"shuttle": "shuttle.obj",
		"bunny":   "bunny.obj",
		"dragon":  "dragon.obj",
		"sponza":  "sponza/sponza.obj",
		"sibenik": "sibenik/sibenik.obj",
	}
)

type mouseInfo struct {
	x, y, button int
}

type Model struct {
	qml.Object
	cameraMode int
	setModel   string
	modelName  string
	models     map[string]scene.Object
	background scene.Object
	view       *scene.View
	mouse      mouseInfo
	loading    sync.Mutex
}

func (t *Model) initialise(gl *GL.GL) {
	fmt.Println("initialise")
	glu.Debug = true
	t.view = scene.NewView(rotCamera.Clone()).AddLight(light)
	t.background = scene.NewItem(mesh.Cube().Invert().SetMaterial(mesh.Skybox()))
	t.background.Scale(40, 40, 40)
	t.models = map[string]scene.Object{}
	t.setModel = "cube"
	t.loadMesh(t.setModel)
}

func (t *Model) loadMesh(name string) {
	if _, loaded := t.models[name]; loaded {
		t.modelName = name
		return
	}
	fmt.Println("load mesh", name)
	model, err := mesh.LoadObjFile(meshes[name])
	if err != nil {
		fmt.Printf("error loading %s: %v\n", name, err)
	}
	switch name {
	case "cube":
		t.models[name] = scene.NewItem(model)
	case "teapot":
		t.models[name] = scene.NewItem(model).Scale(0.012, 0.012, 0.012).Translate(0, -0.4, 0)
	case "shuttle":
		t.models[name] = scene.NewGroup().Add(scene.NewItem(model).Scale(0.13, 0.13, 0.13).RotateX(-90))
	case "bunny":
		t.models[name] = scene.NewItem(model).Scale(0.7, 0.7, 0.7)
	case "dragon":
		t.models[name] = scene.NewGroup().Add(scene.NewItem(model).RotateX(-90))
	case "sponza":
		t.models[name] = scene.NewItem(model).Scale(0.5, 0.5, 0.5).Translate(0.5, -1, 0)
	case "sibenik":
		t.models[name] = scene.NewItem(model).Scale(0.5, 0.5, 0.5).RotateY(180).Translate(-0.5, 6, 0)
	}
	t.modelName = name
}

func (t *Model) SetModel(name string) {
	if name == "" {
		return
	}
	fmt.Println("set model", name)
	t.setModel = name
	t.Call("update")
}

func (t *Model) Spin() {
	t.models[t.modelName].RotateY(1)
	t.Call("update")
}

func (t *Model) SetScenery(on bool) {
	t.background.Enable(on)
	t.Call("update")
}

func (t *Model) SetCamera(mode int) {
	if mode == t.cameraMode {
		return
	}
	fmt.Println("set camera mode", mode)
	t.cameraMode = mode
	t.Reset()
}

func (t *Model) Reset() {
	fmt.Println("reset view", rotCamera)
	if t.cameraMode == 0 {
		t.view.Camera = rotCamera.Clone()
	} else {
		t.view.Camera = povCamera.Clone()
	}
	t.Call("update")
}

func (t *Model) Move(amount float32) {
	t.view.Camera.Move(amount)
	t.Call("update")
}

func (t *Model) Rotate(event string, x, y, button int) {
	switch event {
	case "start":
		t.mouse = mouseInfo{x, y, button}
	case "move":
		if t.mouse.button != 0 {
			dx, dy := float32(x-t.mouse.x), float32(y-t.mouse.y)
			if t.mouse.button == 1 {
				t.view.Camera.Rotate(dx, dy)
			} else {
				t.view.Lights[0].Rotate(dx, dy)
			}
			t.mouse.x, t.mouse.y = x, y
			t.Call("update")
		}
	case "end":
		t.mouse.button = 0
	case "keys":
		t.view.Camera.Rotate(float32(x), float32(y))
		t.Call("update")
	}
}

func (t *Model) Paint(p *qml.Painter) {
	gl := GL.API(p)
	glu.Init(gl)
	if t.models == nil {
		t.initialise(gl)
	}
	if t.setModel != t.modelName {
		t.loading.Lock()
		// model loading can take a long time so do it in the background
		go func() {
			t.loadMesh(t.setModel)
			t.loading.Unlock()
			t.Call("update")
		}()
	}
	//fmt.Println("paint", t.modelName)
	t.view.SetProjection(t.Int("width"), t.Int("height"))
	glu.Clear(mgl32.Vec4{0.5, 0.5, 1, 1})
	view := t.view.ViewMatrix()
	t.view.UpdateLights(view, nil)
	// skybox is always centered on the camera
	if t.background != nil && t.background.Enabled() {
		t.view.Draw(t.view.CenteredView(), t.background)
	}
	t.view.Draw(view, t.models[t.modelName])
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
