package main

import (
	"github.com/drakmaniso/glam"
	"github.com/go-gl/gl"
)

type Camera struct {
	isChanged bool
	VP *glam.Mat4
	view, projection *glam.Mat4
	position glam.Vec3
	mName, vpName string
}

func MakeCamera() *Camera {
	return &Camera{false, MatIdentity4,
		MatIdentity4, MatIdentity4, glam.Vec3{0, 0, 0}, "", ""}
}

func (c *Camera) ToVPMatrix() *glam.Mat4 {
	if c.isChanged {
		c.isChanged = false
		newVP := c.view.Times(c.projection)
		c.VP = &newVP
	}
	return c.VP
}

func (c *Camera) ToVP() *[16]float32 {
	MVPSlice := MatToSlice(c.ToVPMatrix())
	return &MVPSlice
}

func (c *Camera) SetPerspective(FOV float32) {
	c.isChanged = true
	perspecMat := glam.Perspective(FOV,
		float32(WindowW)/float32(WindowH), 0.1, 100.0)
	c.projection = &perspecMat
}

func (c *Camera) SetOrthographic(zoom float32) {
	c.isChanged = true
	orthoMat := glam.Orthographic(zoom,
		float32(WindowW)/float32(WindowH), 0.1, 100.0)
	c.projection = &orthoMat
}

func (c *Camera) LookAt(eye, look, up glam.Vec3) {
	c.isChanged = true
	lookMat := glam.LookAt(eye, look, up)
	c.view = &lookMat
}

func (c *Camera) SetView(position, forward, up glam.Vec3) {
	c.LookAt(position, position.Plus(forward), up)
}

// Set the name of the Model matrix used in the vertex shader
func(c *Camera) SetMName(name string) {
	c.mName = name
}

// Set the name of the View-Projection matrix used in the vertex shader
func(c *Camera) SetVPName(name string) {
	c.vpName = name
}

// Prepares to draw an object that has undergone the given model transformation
func (c *Camera) Prepare(shader gl.Program, model *Transform) {
	shader.GetUniformLocation(c.mName).UniformMatrix4fv(
		 					false, MatToSlice(&((*model).TransformMat)))
	shader.GetUniformLocation(c.vpName).UniformMatrix4fv(
		 					false, *c.ToVP())
}