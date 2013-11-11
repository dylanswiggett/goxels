package main

import "github.com/go-gl/gl"

/*
 * NO SUPPORT FOR TEXTURED MODELS
 */

var modelUniqueIdCounter = 0
var currentLoadedModel = -1
var vertexAttrib, colorAttrib, normalAttrib gl.AttribLocation

func EnableModelRendering() {
	vertexAttrib = 0
	colorAttrib = 1
	normalAttrib = 2
}

func DisableModelRendering() {
	vertexAttrib.DisableArray()
	colorAttrib.DisableArray()
	normalAttrib.DisableArray()
	currentLoadedModel = -1
}

type Color struct {
	R, G, B byte
}

type Vertex struct {
	P, N Vec3	// Position, Normal,
	C Color		// Color
}

// Note: Only triangles supported
type Polygon struct {
	Vertices [3]Vertex
}

type Model struct {
	uid int
	polygons []Polygon
	pFloats []float32
	cBytes  []byte
	nFloats []float32
	vertexBuffer, normalBuffer, colorBuffer gl.Buffer
}

func MakeModel(polygons []Polygon) Model {
	model := Model{modelUniqueIdCounter, polygons,
		nil, nil, nil, gl.GenBuffer(), gl.GenBuffer(), gl.GenBuffer()}
	modelUniqueIdCounter++
	model.genFloatArrays()
	model.genGLBuffers()
	return model
}

func MakeScaledModel(polygons []Polygon, scale float32) Model {
	for i := 0; i < len(polygons); i++ {
		for j := 0; j < len(polygons[i].Vertices); j++ {
			polygons[i].Vertices[j].P = polygons[i].Vertices[j].P.Scale(scale)
		}
	}

	// for _, poly := range(polygons) {
	// 	for _, vert := range(poly.Vertices) {
	// 		vert.P = vert.P.Scale(scale)
	// 	}
	// }
	return MakeModel(polygons)
}

func (m *Model) Draw() {
	if currentLoadedModel != m.uid {
		vertexAttrib.EnableArray()
		m.vertexBuffer.Bind(gl.ARRAY_BUFFER)
		vertexAttrib.AttribPointer(3, gl.FLOAT, false, 0, nil)

		colorAttrib.EnableArray()
		m.colorBuffer.Bind(gl.ARRAY_BUFFER)
		colorAttrib.AttribPointer(3, gl.UNSIGNED_BYTE, false, 0, nil)

		normalAttrib.EnableArray()
		m.normalBuffer.Bind(gl.ARRAY_BUFFER)
		normalAttrib.AttribPointer(3, gl.FLOAT, false, 0, nil)

		currentLoadedModel = m.uid
	}
	gl.DrawArrays(gl.TRIANGLES, 0, len(m.polygons) * 3)
}

func (m *Model) genGLBuffers() {
	m.vertexBuffer.Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, len(m.pFloats)*4, m.pFloats, gl.STATIC_DRAW)

	m.colorBuffer.Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, len(m.cBytes), m.cBytes, gl.STATIC_DRAW)

	m.normalBuffer.Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, len(m.nFloats)*4, m.nFloats, gl.STATIC_DRAW)
}

// Get the 3 float arrays to use with OpenGL Buffers.
func (m *Model) genFloatArrays() {
	numEntries := 3 * 3 * len(m.polygons)
	vertices := make([]float32, numEntries)
	colors := make([]byte, numEntries)
	normals := make([]float32, numEntries)

	i := 0

	for _, poly := range(m.polygons) {
		for _, vert := range(poly.Vertices) {
			vertices[i] = vert.P.X
			colors[i] = vert.C.R
			normals[i] = vert.N.X
			i++
			vertices[i] = vert.P.Y
			colors[i] = vert.C.G
			normals[i] = vert.N.Y
			i++
			vertices[i] = vert.P.Z
			colors[i] = vert.C.B
			normals[i] = vert.N.Z
			i++
		}
	}

	m.pFloats = vertices
	m.cBytes = colors
	m.nFloats = normals
}

func (m *Model) EnableRendering() {

}