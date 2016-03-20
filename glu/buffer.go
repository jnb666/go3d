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
	buffer glbase.Buffer
	btype  glbase.Enum
	size   int
}

// ArrayBuffer creates a new empty Vertex array with associated data. Size is the numer of size of each vertex in words.
func ArrayBuffer(data []float32, vertexSize int) *VertexArray {
	buf := gl.GenBuffers(1)
	a := &VertexArray{buffer: buf[0], btype: GL.ARRAY_BUFFER, size: len(data) / vertexSize}
	gl.BindBuffer(a.btype, buf[0])
	gl.BufferData(a.btype, len(data)*4, nil, GL.STATIC_DRAW)
	for start := 0; start < len(data); start += chunkSize {
		end := min(start+chunkSize, len(data))
		gl.BufferSubData(a.btype, start*4, (end-start)*4, data[start:end])
	}
	CheckError()
	runtime.SetFinalizer(a, deleteArray)
	return a
}

// ElementArrayBuffer creates a new empty Vertex array with associated data.
func ElementArrayBuffer(data []uint32) *VertexArray {
	buf := gl.GenBuffers(1)
	a := &VertexArray{buffer: buf[0], btype: GL.ELEMENT_ARRAY_BUFFER, size: len(data)}
	gl.BindBuffer(a.btype, buf[0])
	gl.BufferData(a.btype, len(data)*4, nil, GL.STATIC_DRAW)
	for start := 0; start < len(data); start += chunkSize {
		end := min(start+chunkSize, len(data))
		gl.BufferSubData(a.btype, start*4, (end-start)*4, data[start:end])
	}
	CheckError()
	runtime.SetFinalizer(a, deleteArray)
	return a
}

// Make buffer current
func (a *VertexArray) Enable() {
	gl.BindBuffer(a.btype, a.buffer)
}

// Draw specified elements from the array
func (a *VertexArray) Draw(mode glbase.Enum, winding glbase.Enum) {
	gl.FrontFace(winding)
	if a.btype == GL.ELEMENT_ARRAY_BUFFER {
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
	gl.DeleteBuffers([]glbase.Buffer{a.buffer})
	a.buffer = 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
