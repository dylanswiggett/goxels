package main

import "github.com/drakmaniso/glam"

type Transform struct {
	TransformMat glam.Mat4
}

func MakeTransform() Transform {
	return Transform{*MatIdentity4}
}

func Translate(amount glam.Vec3) Transform {
	return Transform{glam.Translation(amount)}
}

func (t *Transform) Translate(amount glam.Vec3) Transform {
	translation := glam.Translation(amount)
	return Transform{t.TransformMat.Times(&translation)}
}

func Rotate(angle float32, axis glam.Vec3) Transform {
	return Transform{glam.Rotation(angle, axis)}
}

func (t *Transform) Rotate(angle float32, axis glam.Vec3) Transform {
	rotation := glam.Rotation(angle, axis)
	return Transform{t.TransformMat.Times(&rotation)}
}

func Scale(amount glam.Vec3) Transform {
	scale := glam.NewMat4(amount.X, 0, 0, 0,
						  0, amount.Y, 0, 0,
						  0, 0, amount.Z, 0,
						  0, 0, 0, 1)
	return Transform{*scale}
}

func (t *Transform) Scale(amount glam.Vec3) Transform {
	return Transform{t.TransformMat.Times(glam.NewMat4(amount.X, 0, 0, 0,
													   0, amount.Y, 0, 0,
													   0, 0, amount.Z, 0,
													   0, 0, 0, 1))}
}

func ScaleBy(amount float32) Transform {
	scale := glam.NewMat4(amount, 0, 0, 0,
						  0, amount, 0, 0,
						  0, 0, amount, 0,
						  0, 0, 0, 1)
	return Transform{*scale}
}

func (t *Transform) ScaleBy(amount float32) Transform {
	return Transform{t.TransformMat.Times(glam.NewMat4(amount, 0, 0, 0,
													   0, amount, 0, 0,
													   0, 0, amount, 0,
													   0, 0, 0, 1))}
}