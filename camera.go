package main

import (
	"github.com/drakmaniso/glam"
	"github.com/dylanswiggett/gl"
)

var (
	View, Projection *glam.Mat4
	VP *glam.Mat4
	MName, VpName string
)

type Camera struct {
	isChanged bool
	// VP *glam.Mat4
	// View, Projection *glam.Mat4
	position glam.Vec3
	// MName, VpName string
}

func MakeCamera() *Camera {
	VP = MatIdentity4
	View = MatIdentity4
	Projection = MatIdentity4
	MName = ""
	VpName = ""
	return &Camera{false, glam.Vec3{0, 0, 0}}
}

func (c *Camera) ToVPMatrix() *glam.Mat4 {
	if c.isChanged {
		c.isChanged = false
		newVP := View.Times(Projection)
		VP = &newVP
	}
	return VP
}

func (c *Camera) ToVP() *[16]float32 {
	MVPSlice := MatToSlice(c.ToVPMatrix())
	return &MVPSlice
}

func (c *Camera) SetPerspective(FOV float32) {
	c.isChanged = true
	perspecMat := glam.Perspective(FOV,
		float32(WindowW)/float32(WindowH), 0.1, 100.0)
	Projection = &perspecMat
}

func (c *Camera) SetOrthographic(zoom float32) {
	c.isChanged = true
	orthoMat := glam.Orthographic(zoom,
		float32(WindowW)/float32(WindowH), 0.1, 100.0)
	Projection = &orthoMat
}

func (c *Camera) LookAt(eye, look, up glam.Vec3) {
	c.isChanged = true
	lookMat := glam.LookAt(eye, look, up)
	View = &lookMat
}

func (c *Camera) SetView(position, forward, up glam.Vec3) {
	c.LookAt(position, position.Plus(forward), up)
}

// Set the name of the Model matrix used in the vertex shader
func(c *Camera) SetMName(name string) {
	MName = name
}

// Set the name of the View-Projection matrix used in the vertex shader
func(c *Camera) SetVPName(name string) {
	VpName = name
}

// Prepares to draw an object that has undergone the given model transformation
func (c *Camera) Prepare(shader gl.Program, model *Transform) {
	shader.GetUniformLocation(MName).UniformMatrix4fv(
		 					false, MatToSlice(&((*model).TransformMat)))
	shader.GetUniformLocation(VpName).UniformMatrix4fv(
		 					false, *c.ToVP())
}