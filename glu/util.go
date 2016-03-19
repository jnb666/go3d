// Package glu provides OpenGL utility functions.
package glu

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"gopkg.in/qml.v1/gl/es2"
	"gopkg.in/qml.v1/gl/glbase"
	"math"
)

// If debug mode is set then check for GL errors after each call
var Debug = false

// GL error codes
var glErrorText = map[glbase.Enum]string{
	0x500: "invalid enum",
	0x501: "invalid value",
	0x502: "invalid operation",
	0x503: "stack overflow",
	0x504: "stack underflow",
	0x505: "out of memory",
	0x506: "invalid framebuffer operation",
}

// Predefined colors
var (
	Black   = mgl32.Vec4{0, 0, 0, 1}
	White   = mgl32.Vec4{1, 1, 1, 1}
	Red     = mgl32.Vec4{1, 0, 0, 1}
	Green   = mgl32.Vec4{0, 1, 0, 1}
	Blue    = mgl32.Vec4{0, 0, 1, 1}
	Yellow  = mgl32.Vec4{1, 1, 0, 1}
	Cyan    = mgl32.Vec4{0, 1, 1, 1}
	Magenta = mgl32.Vec4{1, 0, 1, 1}
	Grey    = mgl32.Vec4{0.5, 0.5, 0.5, 1}
)

// Globals
var gl *GL.GL

// Init method should be called to set the pointer to GL functions
func Init(glptr *GL.GL) {
	gl = glptr
}

// Clear the screen ready for draing
func Clear(bg mgl32.Vec4) {
	gl.Enable(GL.DEPTH_TEST)
	gl.Enable(GL.CULL_FACE)
	gl.Enable(GL.BLEND)
	gl.BlendFunc(GL.SRC_ALPHA, GL.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(glbase.Clampf(bg[0]), glbase.Clampf(bg[1]), glbase.Clampf(bg[2]), glbase.Clampf(bg[3]))
	gl.Clear(GL.COLOR_BUFFER_BIT | GL.DEPTH_BUFFER_BIT)
}

// Reference to GL function pointers
func GLRef() *GL.GL {
	return gl
}

// Check for GL errors
func CheckError() {
	err := gl.GetError()
	if err != GL.NO_ERROR {
		text, ok := glErrorText[err]
		if !ok {
			text = fmt.Sprintf("unknown error code %x", err)
		}
		panic("GL error: " + text)
	}
}

// Force value into range from min to max
func Clamp(x, min, max float32) float32 {
	if x < min {
		x = min
	} else if x > max {
		x = max
	}
	return x
}

// Polar type represents polar coordinates
type Polar struct {
	R, Theta, Phi float32
}

// bring phi in range 0->360 and theta in range 1->179 degrees
func (p *Polar) Clamp() {
	p.Theta = Clamp(p.Theta, 1, 179)
	phi := float64(p.Phi)
	if p.Phi > 360 {
		p.Phi -= float32(math.Trunc(phi/360) * 360)
	}
	if p.Phi < 0 {
		p.Phi += float32(math.Trunc(1-phi/360) * 360)
	}
}

// Convert polar coords to vector
func (p Polar) Vec3() mgl32.Vec3 {
	sinp, cosp := math.Sincos(float64(p.Phi) * math.Pi / 180)
	sinpf, cospf := float32(sinp), float32(cosp)
	sint, cost := math.Sincos(float64(p.Theta) * math.Pi / 180)
	sintf, costf := float32(sint), float32(cost)
	return mgl32.Vec3{p.R * sintf * cospf, p.R * costf, p.R * sintf * sinpf}
}

func (p Polar) Vec4(w float32) mgl32.Vec4 {
	v := p.Vec3()
	return mgl32.Vec4{v[0], v[1], v[2], w}
}

// Convert vector to polar coords
func (p *Polar) Set(vec mgl32.Vec3) *Polar {
	p.R = vec.Len()
	p.Theta = float32(math.Acos(float64(vec[1]/p.R)) * 180 / math.Pi)
	p.Phi = float32(math.Atan2(float64(vec[2]), float64(vec[0])) * 180 / math.Pi)
	return p
}
