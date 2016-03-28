package main

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/glu"
	"github.com/jnb666/go3d/img"
	"github.com/jnb666/go3d/mesh"
	"github.com/jnb666/go3d/scene"
	"gopkg.in/qml.v1"
	"gopkg.in/qml.v1/gl/es2"
	"os"
	"strings"
)

const (
	sceneFile = "lighting.qml"
	moveScale = 0.005
)

var (
	lightPos  = glu.Polar{R: 1, Theta: 50, Phi: -110}
	cameraPos = glu.Polar{R: 2, Theta: 65, Phi: 60}
	camera    = scene.ArcBallCamera(cameraPos, mgl32.Vec3{}, 0.5, 5.0, 10, 170)
	light     = scene.DirectionalLight(mgl32.Vec3{0.8, 0.8, 0.8}, 0.2, lightPos)
)

type mouseInfo struct {
	x, y, button int
}

type texInfo struct {
	diffMap glu.Texture
	specMap glu.Texture
	normMap glu.Texture
	specCol mgl32.Vec4
}

type Model struct {
	qml.Object
	mode     string
	texName  string
	scene    scene.Object
	object   *scene.Item
	view     *scene.View
	textures map[string]texInfo
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

func (t *Model) Reset() {
	fmt.Println("reset view")
	t.view = scene.NewView(camera.Clone()).AddLight(light.Clone())
	t.object.Mat4 = mgl32.Ident4()
	t.Call("update")
}

func (t *Model) Spin() {
	t.object.RotateY(0.25)
	t.Call("update")
}

func (t *Model) SetTexture(name string) {
	if name != t.texName {
		t.texName = name
		t.Call("update")
	}
}

func (t *Model) SetLighting(mode string) {
	if mode != t.mode {
		t.mode = mode
		t.Call("update")
	}
}

func getTexture(name string, conv img.ImageConvert) texInfo {
	tex := texInfo{}
	var err error
	fmt.Printf("load diffuse map from %s\n", name)
	if tex.diffMap, err = glu.NewTexture2D(false).SetImageFile(name, img.SRGBToLinear); err != nil {
		panic(err)
	}
	names := strings.Split(name, ".")
	specName := names[0] + "_specular." + names[1]
	if _, err = os.Stat(specName); err == nil {
		fmt.Printf("load specular map from %s\n", specName)
		if tex.specMap, err = glu.NewTexture2D(false).SetImageFile(specName, img.NoConvert); err != nil {
			panic(err)
		}
		tex.specCol = glu.White
	}
	bumpName := names[0] + "_normal." + names[1]
	if conv == img.BumpToNormal {
		bumpName = names[0] + "_bump." + names[1]
	}
	fmt.Printf("load normal map from %s\n", bumpName)
	if tex.normMap, err = glu.NewTexture2D(false).SetImageFile(bumpName, conv); err != nil {
		panic(err)
	}
	return tex
}

func (t *Model) initialise() {
	fmt.Println("initialise")
	glu.Debug = true
	t.mode = "diffuse"
	t.texName = "brick"
	t.textures = map[string]texInfo{
		"brick":  getTexture("brick.png", img.NoConvert),
		"shield": getTexture("shield.png", img.NoConvert),
		"stone":  getTexture("stone.jpg", img.BumpToNormal),
	}
	fmt.Println("create scene")
	t.view = scene.NewView(camera.Clone()).AddLight(light.Clone())
	t.object = scene.NewItem(mesh.Cube())
	t.scene = scene.NewGroup().Add(t.object).Scale(2, 2, 2).Translate(0, -1, 0)
}

func (t *Model) Paint(p *qml.Painter) {
	gl := GL.API(p)
	glu.Init(gl)
	if t.scene == nil {
		t.initialise()
	}
	// set material
	var mtl mesh.Material
	tex := t.textures[t.texName]
	switch t.mode {
	case "diffuse":
		mtl = mesh.Diffuse(tex.diffMap)
	case "specular":
		mtl = mesh.Reflective(tex.specCol, 64, tex.diffMap, tex.specMap)
	case "normal":
		mtl = mesh.Reflective(tex.specCol, 64, tex.diffMap, tex.specMap, tex.normMap)
	default:
		panic("unknown lighting mode " + t.mode)
	}
	t.object.SetMaterial(mtl)
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
