package main

import (
	"fmt"
	"gopkg.in/qml.v1"
	"gopkg.in/qml.v1/gl/2.1"
	"os"
)

const sceneFile = "triangle.qml"

type Triangle struct {
	qml.Object
	angle float32
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
	width := t.Int("width")
	height := t.Int("height")
	gl.Viewport(0, 0, width, height)
	ratio := float64(width) / float64(height)
	gl.MatrixMode(GL.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(-ratio, ratio, -1, 1, 1, -1)
	gl.MatrixMode(GL.MODELVIEW)
	gl.LoadIdentity()
	gl.Rotatef(t.angle, 0, 0, 1)
	gl.Begin(GL.TRIANGLES)
	gl.Color3f(1, 0, 0)
	gl.Vertex3f(-0.6, -0.4, 0.)
	gl.Color3f(0, 1, 0)
	gl.Vertex3f(0.6, -0.4, 0)
	gl.Color3f(0, 0, 1)
	gl.Vertex3f(0, 0.6, 0)
	gl.End()
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
