package main

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/glu"
	"github.com/jnb666/go3d/mesh"
	"github.com/jnb666/go3d/scene"
	"gopkg.in/qml.v1"
	"gopkg.in/qml.v1/gl/es2"
	"image/color"
	"os"
)

const sceneFile = "shapes.qml"

var (
	cameraPos = glu.Polar{R: 2.0, Theta: 70, Phi: 45}
	lightPos  = glu.Polar{R: 1, Theta: 20, Phi: 90}
	camera    = scene.ArcBallCamera(cameraPos, mgl32.Vec3{}, 0.5, 5.0, 10, 170)
	light     = scene.DirectionalLight(mgl32.Vec3{0.8, 0.8, 0.8}, 0.2, lightPos)
)

type mouseInfo struct {
	x, y, button int
}

type Shapes struct {
	qml.Object
	shapeName  string
	matName    string
	material   map[string]mesh.Material
	shapes     map[string]scene.Object
	background scene.Object
	view       *scene.View
	mouse      mouseInfo
}

func (t *Shapes) initialise() {
	fmt.Println("initialise")
	glu.Debug = true
	t.view = scene.NewView(camera).AddLight(light)

	fmt.Println("load gopher texture")
	tex2d, err := glu.NewTexture2D(false, false).SetImageFile("gopher_rgb.png")
	if err != nil {
		panic(err)
	}
	texCube := glu.NewTextureCube(false)
	for i := 0; i < 6; i++ {
		texCube.SetImageFile("gopher_rgb.png", i)
	}
	t.background = scene.NewItem(mesh.Cube().Invert()).SetMaterial(mesh.Skybox())
	t.background.Enable(false).Scale(10, 10, 10)
	spec := mgl32.Vec4{0.5, 0.5, 0.5, 1}
	t.matName = "plastic"
	t.material = map[string]mesh.Material{
		"plastic":     mesh.Plastic().SetColor(glu.Red),
		"wood":        mesh.Wood(),
		"rough":       mesh.Rough().SetColor(glu.Grey),
		"marble":      mesh.Marble(),
		"glass":       mesh.Glass(),
		"earth":       mesh.Earth(),
		"emissive":    mesh.Emissive().SetColor(mgl32.Vec4{0.95, 0.82, 0.25, 1}),
		"diffuse":     mesh.Diffuse().SetColor(glu.Red),
		"unshaded":    mesh.Unshaded().SetColor(glu.Red),
		"texture2d":   mesh.ReflectiveTex(spec, 32, tex2d),
		"texturecube": mesh.ReflectiveTex(spec, 32, texCube),
		"point":       mesh.PointMaterial().SetColor(glu.Red),
	}
	t.shapeName = "cube"
	t.shapes = map[string]scene.Object{
		"cube":        scene.NewItem(mesh.Cube()),
		"prism":       scene.NewItem(mesh.Prism()).Scale(1.1, 1.1, 1.1),
		"pyramid":     scene.NewItem(mesh.Cone(4)).Scale(1.4, 1.1, 1.4),
		"point":       scene.NewGroup().Add(scene.NewItem(mesh.Point(10)).Translate(0.5, 0, 0.5)),
		"plane":       scene.NewItem(mesh.Plane()).Scale(2, 1, 2),
		"circle":      scene.NewItem(mesh.Circle(60)).Scale(2, 1, 2),
		"cylinder":    scene.NewGroup().Add(scene.NewItem(mesh.Cylinder(60)).RotateX(90)),
		"cone":        scene.NewGroup().Add(scene.NewItem(mesh.Cone(120)).Scale(1.2, 1.2, 1.2).RotateX(90)),
		"icosohedron": scene.NewItem(mesh.Icosohedron()).Scale(1.5, 1.5, 1.5),
		"sphere":      scene.NewItem(mesh.Sphere(3)).Scale(1.4, 1.4, 1.4),
	}
}

func (t *Shapes) SetShape(name string) {
	if _, ok := t.shapes[name]; ok {
		fmt.Println("set shape to", name)
		t.shapeName = name
		t.Call("update")
	}
}

func (t *Shapes) SetMaterial(name string) {
	if _, ok := t.material[name]; ok {
		fmt.Println("set material to", name)
		t.matName = name
		t.Call("update")
	}
}

func (t *Shapes) GetColor() color.RGBA {
	var c mgl32.Vec4
	if t.shapeName == "point" {
		c = t.material["point"].Color()
	} else {
		c = t.material[t.matName].Color()
	}
	return color.RGBA{uint8(255 * c[0]), uint8(255 * c[1]), uint8(255 * c[2]), uint8(255 * c[3])}
}

func (t *Shapes) SetColor(c color.RGBA) {
	col := mgl32.Vec4{float32(c.R) / 255, float32(c.G) / 255, float32(c.B) / 255, float32(c.A) / 255}
	fmt.Printf("set color to %.3f\n", col)
	if t.shapeName == "point" {
		t.material["point"].SetColor(col)
	} else {
		t.material[t.matName].SetColor(col)
	}
	t.Call("update")
}

func (t *Shapes) SetScenery(on bool) {
	t.background.Enable(on)
	t.Call("update")
}

func (t *Shapes) Spin() {
	t.shapes[t.shapeName].RotateY(1)
	t.Call("update")
}

func (t *Shapes) Zoom(amount float32) {
	t.view.Camera.Move(amount)
	t.Call("update")
}

func (t *Shapes) Mouse(event string, x, y, button int) {
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
	}
}

func (t *Shapes) Paint(p *qml.Painter) {
	gl := GL.API(p)
	glu.Init(gl)
	if t.shapes == nil {
		t.initialise()
	}
	t.view.SetProjection(t.Int("width"), t.Int("height"))
	glu.Clear(mgl32.Vec4{0.5, 0.5, 1, 1})
	// set current material
	t.shapes[t.shapeName].Do(scene.NewTransform(mgl32.Ident4()),
		func(obj *scene.Item, trans scene.Transform) {
			if t.shapeName == "point" {
				obj.SetMaterial(t.material["point"])
			} else {
				obj.SetMaterial(t.material[t.matName])
			}
		},
	)
	// skybox is always centered on the camera
	if t.background != nil && t.background.Enabled() {
		t.view.Draw(t.view.CenteredView(), t.background)
	}
	view := t.view.ViewMatrix()
	t.view.UpdateLights(view, nil)
	t.view.Draw(view, t.shapes[t.shapeName])
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
