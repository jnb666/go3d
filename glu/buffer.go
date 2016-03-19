package glu

import (
	"fmt"
	"gopkg.in/qml.v1/gl/es2"
	"gopkg.in/qml.v1/gl/glbase"
	"runtime"
)

const chunkSize = 4096

// an array of buffer data
type VertexArray struct {
	buffer []glbase.Buffer
	btype  []glbase.Enum
	size   int
}

// NewArray creates a new Vertex array with associated data.
func NewArray(vertices []float32, elements []uint32, vertexSize int) *VertexArray {
	a := new(VertexArray)
	if vertices == nil || len(vertices) == 0 {
		panic("NewArray: vertex data is required!")
	}
	a.add(GL.ARRAY_BUFFER, len(vertices), vertices)
	a.size = len(vertices) / vertexSize
	if elements != nil && len(elements) > 0 {
		a.add(GL.ELEMENT_ARRAY_BUFFER, len(elements), elements)
		a.size = len(elements)
	}
	CheckError()
	runtime.SetFinalizer(a, deleteArray)
	return a
}

// add a data buffer to the array
func (a *VertexArray) add(typ glbase.Enum, size int, data interface{}) {
	buf := gl.GenBuffers(1)
	a.buffer = append(a.buffer, buf[0])
	a.btype = append(a.btype, typ)
	gl.BindBuffer(typ, buf[0])
	gl.BufferData(typ, size*4, nil, GL.STATIC_DRAW)
	for start := 0; start < size; start += chunkSize {
		end := start + chunkSize
		if end > size {
			end = size
		}
		switch dat := data.(type) {
		case []float32:
			gl.BufferSubData(typ, start*4, (end-start)*4, dat[start:end])
		case []uint32:
			gl.BufferSubData(typ, start*4, (end-start)*4, dat[start:end])
		default:
			panic("invalid type: expecting []float32 or []uint32")
		}
	}
}

// Make buffers current
func (a *VertexArray) Enable() {
	for i, buf := range a.buffer {
		gl.BindBuffer(a.btype[i], buf)
	}
}

// Draw specified elements from the array
func (a *VertexArray) Draw(mode glbase.Enum, winding glbase.Enum) {
	gl.FrontFace(winding)
	if len(a.buffer) > 1 {
		gl.DrawElements(mode, a.size, GL.UNSIGNED_INT, nil)
	} else {
		gl.DrawArrays(mode, 0, a.size)
	}
	if Debug {
		CheckError()
	}
}

func deleteArray(a *VertexArray) {
	fmt.Println("finalizer called for VertexArray")
	gl.DeleteBuffers(a.buffer)
	a.buffer = nil
}
