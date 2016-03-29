package main

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/glu"
	"github.com/jnb666/go3d/mesh"
	"github.com/jnb666/go3d/scene"
	"gopkg.in/qml.v1"
	"gopkg.in/qml.v1/gl/es2"
	"math"
	"os"
)

const sceneFile = "table.qml"

var (
	cameraPos = glu.Polar{R: 4, Theta: 80, Phi: 130}
	centerPos = mgl32.Vec3{0, 1, 0}
	camera    = scene.ArcBallCamera(cameraPos, centerPos, 1, 6, 40, 95)
)

type Scene struct {
	qml.Object
	scene  scene.Object
	view   *scene.View
	mouseX int
	mouseY int
	table  scene.Object
	lamp   [2]*scene.Item
	balls  [5]*scene.Group
	time   float64
}

func (t *Scene) initialise(gl *GL.GL) scene.Object {
	fmt.Println("initialise")
	glu.Init(gl)
	glu.Debug = true
	t.mouseX, t.mouseY = -1, -1
	t.view = scene.NewView(camera)
	world := scene.NewGroup()

	// room with marble floor
	blue := mesh.Diffuse().SetColor(mgl32.Vec4{0.3, 0.3, 1, 1})
	marble := mesh.Marble().SetColor(mgl32.Vec4{0.2, 0.3, 0.3, 1})
	room := scene.NewGroup()
	walls := scene.NewItem(mesh.Cube().Invert().SetMaterial(blue)).Translate(0, 0.5, 0)
	floor := scene.NewItem(mesh.Plane().SetMaterial(marble)).Translate(0, 0.01, 0)
	room.Add(walls, floor)
	room.Scale(10, 4, 10)
	world.Add(room)

	// ceiling light
	t.lamp[0] = scene.NewItem(mesh.Cylinder(60).SetMaterial(mesh.Rough())).Illuminate(2, 0.2, 0.5)
	t.lamp[0].Scale(0.5, 0.2, 0.5).Translate(0, 3.8, 0)
	world.Add(t.lamp[0])

	// wooden table
	wood := mesh.Wood().SetColor(mgl32.Vec4{0.90, 0.43, 0.14, 1})
	rough := mesh.Rough().SetColor(glu.Grey)
	top := scene.NewItem(mesh.Cube().SetMaterial(wood)).Scale(2, 0.05, 1).Translate(0, 1, 0)
	leg := scene.NewItem(mesh.Cylinder(36).SetMaterial(rough)).Scale(0.1, 1, 0.1).Translate(0, 0.5, 0)
	table := scene.NewGroup().Add(top,
		leg.Clone().Translate(-0.9, 0, -0.4),
		leg.Clone().Translate(-0.9, 0, 0.4),
		leg.Clone().Translate(0.9, 0, -0.4),
		leg.Clone().Translate(0.9, 0, 0.4),
	)
	world.Add(table.Translate(0, 0, -0.1))
	t.table = table

	// desk lamp
	redPlastic := mesh.Plastic().SetColor(glu.Red)
	t.lamp[1] = scene.NewItem(mesh.Sphere(2).SetMaterial(mesh.Rough())).Illuminate(1, 0.1, 5)
	t.lamp[1].Light.Col[2] = 0.4
	lamp := scene.NewGroup()
	lamp.Add(scene.NewItem(mesh.Cone(60).SetMaterial(redPlastic)).Scale(2, 2, 2).Translate(0, -1, 0))
	lamp.Add(t.lamp[1])
	lamp.Scale(0.1, 0.1, 0.1).Translate(-0.5, 1.2, 0)
	table.Add(lamp)

	// Newton's cradle
	rod := mesh.Cylinder(36).SetMaterial(mesh.Plastic().SetColor(glu.Black))
	tubev := scene.NewItem(rod).Scale(0.06, 1, 0.06)
	tubet := scene.NewGroup().Add(scene.NewItem(rod).RotateZ(90).Scale(0.06, 1, 0.06))
	tubeb := scene.NewGroup().Add(scene.NewItem(rod).RotateX(90).Scale(0.06, 1, 0.06))
	cradle := scene.NewGroup().Add(
		tubev.Clone().Translate(-0.5, -0.5, -0.5),
		tubev.Clone().Translate(-0.5, -0.5, 0.5),
		tubev.Clone().Translate(0.5, -0.5, -0.5),
		tubev.Clone().Translate(0.5, -0.5, 0.5),
		tubet.Clone().Translate(0, 0, -0.5),
		tubet.Clone().Translate(0, 0, 0.5),
		tubeb.Clone().Translate(-0.5, -1, 0),
		tubeb.Clone().Translate(0.5, -1, 0),
	)
	ball := mesh.Sphere(3).SetMaterial(mesh.Metallic())
	wire := mesh.Cube().SetMaterial(mesh.Unshaded().SetColor(glu.Black))
	cradleItem := scene.NewGroup().Add(
		scene.NewItem(ball).Scale(0.2, 0.2, 0.2).Translate(0, -0.5, 0),
		scene.NewItem(wire).RotateX(45).Scale(0.005, 1/math.Sqrt2, 0.005).Translate(0, 0, 0.25*math.Sqrt2),
		scene.NewItem(wire).RotateX(-45).Scale(0.005, 1/math.Sqrt2, 0.005).Translate(0, 0, -0.25*math.Sqrt2),
	)
	for i := range t.balls {
		t.balls[i] = cradleItem.Clone().(*scene.Group)
		cradle.Add(scene.NewGroup().Add(t.balls[i]).Translate(0.2*float32(i-2), 0, 0))
	}
	table.Add(cradle.Scale(0.25, 0.25, 0.25).Translate(0, 1.285, 0))

	// glass sphere
	glass := mesh.Glass()
	sphere := scene.NewItem(mesh.Sphere(3).SetMaterial(glass)).Scale(0.3, 0.3, 0.3).Translate(0.5, 1.175, 0)
	table.Add(sphere)

	return world
}

func (t *Scene) Animate(step float64) {
	t.time += step
	angle := float32(0.33 * math.Pi * math.Cos(10*t.time))
	if angle > 0 {
		t.balls[3].Mat4 = mgl32.HomogRotate3DZ(angle)
		t.balls[4].Mat4 = mgl32.HomogRotate3DZ(angle)
		t.balls[0].Mat4 = mgl32.Ident4()
		t.balls[1].Mat4 = mgl32.Ident4()
	} else {
		t.balls[0].Mat4 = mgl32.HomogRotate3DZ(angle)
		t.balls[1].Mat4 = mgl32.HomogRotate3DZ(angle)
		t.balls[3].Mat4 = mgl32.Ident4()
		t.balls[4].Mat4 = mgl32.Ident4()
	}
	t.Call("update")
}

func (t *Scene) Zoom(amount float32) {
	t.view.Camera.Move(amount)
	t.Call("update")
}

func (t *Scene) Mouse(event string, x, y int) {
	switch event {
	case "start":
		t.mouseX, t.mouseY = x, y
	case "move":
		if t.mouseX >= 0 && t.mouseY >= 0 {
			dx, dy := float32(x-t.mouseX), float32(y-t.mouseY)
			t.view.Camera.Rotate(dx, dy)
			t.mouseX, t.mouseY = x, y
			t.Call("update")
		}
	case "end":
		t.mouseX, t.mouseY = -1, -1
	}
}

type key struct {
	id int
	on bool
}

func (t *Scene) ShowLight(id int, on bool) {
	t.lamp[id].Light.On = on
	t.Call("update")
}

func (t *Scene) ShowTable(on bool) {
	t.table.Enable(on)
	t.ShowLight(1, on)
	t.Call("update")
}

func (t *Scene) Paint(p *qml.Painter) {
	gl := GL.API(p)
	if t.scene == nil {
		t.scene = t.initialise(gl)
	}
	t.view.SetProjection(t.Int("width"), t.Int("height"))
	glu.Clear(glu.Black)
	trans := t.view.ViewMatrix()
	t.view.UpdateLights(trans, t.scene)
	t.view.Draw(trans, t.scene)
}

func run() error {
	qml.RegisterTypes("GoExtensions", 1, 0, []qml.TypeSpec{{
		Init: func(t *Scene, obj qml.Object) { t.Object = obj },
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
