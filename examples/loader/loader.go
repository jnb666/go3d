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
	camera    = scene.ArcBallCamera(cameraPos, mgl32.Vec3{0, 0, 0}, 0.5, 3, 10, 170)
	light     = scene.DirectionalLight(mgl32.Vec3{0.8, 0.8, 0.8}, 0.2, lightPos)
	meshes    = []string{"cube", "teapot", "shuttle", "bunny", "dragon", "sponza"}
)

type mouseInfo struct {
	x, y, button int
}

type Shapes struct {
	qml.Object
	setModel   string
	modelName  string
	models     map[string]scene.Object
	background scene.Object
	view       *scene.View
	mouse      mouseInfo
	loading    sync.Mutex
}

func (t *Shapes) initialise(gl *GL.GL) {
	fmt.Println("initialise")
	glu.Debug = true
	t.view = scene.NewView(camera).AddLight(light)
	t.background = scene.NewItem(mesh.Cube().Invert().SetMaterial(mesh.Skybox())).Enable(false)
	t.background.Scale(40, 40, 40)
	t.models = map[string]scene.Object{}
	t.setModel = "cube"
	t.loadMesh(t.setModel)
}

func (t *Shapes) loadMesh(name string) {
	if _, loaded := t.models[name]; loaded {
		t.modelName = name
		return
	}
	fmt.Println("load mesh", name)
	model, err := mesh.LoadObjFile(name + ".obj")
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
	}
	t.modelName = name
}

func (t *Shapes) SetModel(name string) {
	if name == "" {
		return
	}
	fmt.Println("set model", name)
	t.setModel = name
	t.Call("update")
}

func (t *Shapes) Spin() {
	t.models[t.modelName].RotateY(1)
	t.Call("update")
}

func (t *Shapes) SetScenery(on bool) {
	t.background.Enable(on)
	t.Call("update")
}

func (t *Shapes) Zoom(amount int) {
	t.view.Camera.Zoom(-amount)
	t.Call("update")
}

func (t *Shapes) Mouse(event string, x, y, button int) {
	switch event {
	case "start":
		t.mouse = mouseInfo{x, y, button}
	case "move":
		if t.mouse.button != 0 {
			dx, dy := x-t.mouse.x, y-t.mouse.y
			if t.mouse.button == 1 {
				t.view.Camera.Move(dx, dy)
			} else {
				t.view.Lights[0].Move(dx, dy)
			}
			t.mouse.x, t.mouse.y = x, y
			t.Call("update")
		}
	case "end":
		t.mouse.button = 0
	}
}

func (t *Shapes) Paint(p *qml.Painter) {
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
	view := t.view.Camera.ViewMatrix()
	t.view.UpdateLights(view, nil)
	// skybox is always centered on the camera
	if t.background != nil && t.background.Enabled() {
		t.view.Draw(t.view.CenteredView(), t.background)
	}
	t.view.Draw(view, t.models[t.modelName])
}

func run() error {
	qml.RegisterTypes("GoExtensions", 1, 0, []qml.TypeSpec{{
		Init: func(t *Shapes, obj qml.Object) { t.Object = obj },
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
