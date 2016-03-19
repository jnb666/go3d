package glu

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"gopkg.in/qml.v1/gl/es2"
	"gopkg.in/qml.v1/gl/glbase"
)

// Type to encapsulate opengl shader program.
type Program struct {
	prog    glbase.Program
	uniform map[string]uniform
	attr    []Attrib
	stride  int
}

type Attrib struct {
	Name   string
	Size   int
	Offset int
}

// NewProgram compiles and links a shader. Args are vertex shader, fragment shader source
func NewProgram(vertexShader, fragmentShader string, attr []Attrib, stride int) (p *Program, err error) {
	p = new(Program)
	p.prog = gl.CreateProgram()
	p.uniform = make(map[string]uniform)
	vs, err := compileShader(GL.VERTEX_SHADER, vertexShader)
	if err != nil {
		return p, err
	}
	gl.AttachShader(p.prog, vs)
	fs, err := compileShader(GL.FRAGMENT_SHADER, fragmentShader)
	if err != nil {
		return p, err
	}
	gl.AttachShader(p.prog, fs)
	gl.LinkProgram(p.prog)
	gl.DeleteShader(vs)
	gl.DeleteShader(fs)
	var status [1]int32
	gl.GetProgramiv(p.prog, GL.LINK_STATUS, status[:])
	if status[0] == 0 {
		log := gl.GetProgramInfoLog(p.prog)
		return p, fmt.Errorf("error linking program: %s", log)
	}
	CheckError()
	p.attr = attr
	p.stride = stride
	return p, nil
}

func compileShader(typ glbase.Enum, src string) (glbase.Shader, error) {
	shader := gl.CreateShader(typ)
	gl.ShaderSource(shader, src)
	gl.CompileShader(shader)
	var status [1]int32
	gl.GetShaderiv(shader, GL.COMPILE_STATUS, status[:])
	if status[0] == 0 {
		log := gl.GetShaderInfoLog(shader)
		return 0, fmt.Errorf("error compiling %s %s\n", src, log)
	}
	return shader, nil
}

// Use sets this as the current program.
func (p *Program) Use() {
	gl.UseProgram(p.prog)
	for _, att := range p.attr {
		loc := gl.GetAttribLocation(p.prog, att.Name)
		gl.VertexAttribPointer(loc, att.Size, GL.FLOAT, false, p.stride*4, uintptr(att.Offset*4))
		gl.EnableVertexAttribArray(loc)
	}
	if Debug {
		CheckError()
	}
}

// Uniform adds one or more uniforms of the given type
func (p *Program) Uniform(typ string, names ...string) {
	for _, name := range names {
		u := gl.GetUniformLocation(p.prog, name)
		CheckError()
		switch typ {
		case "1i":
			p.uniform[name] = uniform1i(u)
		case "1f":
			p.uniform[name] = uniform1f(u)
		case "2i":
			p.uniform[name] = uniform2i(u)
		case "2f":
			p.uniform[name] = uniform2f(u)
		case "v3f":
			p.uniform[name] = uniformVector3f(u)
		case "v4f":
			p.uniform[name] = uniformVector4f(u)
		case "m3f":
			p.uniform[name] = uniformMatrix3f(u)
		case "m4f":
			p.uniform[name] = uniformMatrix4f(u)
		default:
			panic("unsupported uniform type " + typ)
		}
	}
}

// UniformArray adds one or more uniforms arrays of the given type
func (p *Program) UniformArray(size int, typ string, names ...string) {
	for _, name := range names {
		for i := 0; i < size; i++ {
			p.Uniform(typ, fmt.Sprintf("%s[%d]", name, i))
		}
	}
}

// Set sets the specified uniform value for the program.
func (p *Program) Set(name string, v ...interface{}) {
	if u, ok := p.uniform[name]; ok {
		u.set(v...)
	} else {
		panic(fmt.Errorf("uniform %s not defined", name))
	}
}

// Set an element in a uniform array arr[index].field
func (p *Program) SetArray(arr string, index int, v ...interface{}) {
	p.Set(fmt.Sprintf("%s[%d]", arr, index), v...)
}

// Interface type to represent an opengl uniform
type uniform interface {
	set(v ...interface{})
}

type uniform1i glbase.Uniform

func (u uniform1i) set(v ...interface{}) {
	gl.Uniform1i(glbase.Uniform(u), toi(v[0]))
}

type uniform1f int32

func (u uniform1f) set(v ...interface{}) {
	gl.Uniform1f(glbase.Uniform(u), tof(v[0]))
}

type uniform2i int32

func (u uniform2i) set(v ...interface{}) {
	gl.Uniform2i(glbase.Uniform(u), toi(v[0]), toi(v[1]))
}

type uniform2f int32

func (u uniform2f) set(v ...interface{}) {
	gl.Uniform2f(glbase.Uniform(u), tof(v[0]), tof(v[1]))
}

type uniformVector3f int32

func (u uniformVector3f) set(v ...interface{}) {
	if vec, ok := v[0].(mgl32.Vec3); ok {
		gl.Uniform3fv(glbase.Uniform(u), vec[:])
	} else {
		panic("invalid type for v3f uniform - should be mgl32.Vec3")
	}
}

type uniformVector4f int32

func (u uniformVector4f) set(v ...interface{}) {
	if vec, ok := v[0].(mgl32.Vec4); ok {
		gl.Uniform4fv(glbase.Uniform(u), vec[:])
	} else {
		panic("invalid type for v4f uniform - should be mgl32.Vec4")
	}
}

type uniformMatrix3f int32

func (u uniformMatrix3f) set(v ...interface{}) {
	if mat, ok := v[0].(mgl32.Mat3); ok {
		gl.UniformMatrix3fv(glbase.Uniform(u), false, mat[:])
	} else {
		panic("invalid type for m3f uniform - should be mgl32.Mat3")
	}
}

type uniformMatrix4f int32

func (u uniformMatrix4f) set(v ...interface{}) {
	if mat, ok := v[0].(mgl32.Mat4); ok {
		gl.UniformMatrix4fv(glbase.Uniform(u), false, mat[:])
	} else {
		panic("invalid type for m4f uniform - should be mgl32.Mat4")
	}
}

// conversion utils
func toi(v interface{}) int32 {
	switch t := v.(type) {
	case int:
		return int32(t)
	case int32:
		return t
	case uint32:
		return int32(t)
	default:
		panic(fmt.Sprintf("incompatible type %T for integer uniform", t))
	}
}

func tof(v interface{}) float32 {
	switch t := v.(type) {
	case float32:
		return t
	case float64:
		return float32(t)
	default:
		panic(fmt.Sprintf("incompatible type %T for float uniform", t))
	}
}
