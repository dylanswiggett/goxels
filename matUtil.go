package main

import "github.com/drakmaniso/glam"

var MatIdentity4 = glam.NewMat4(1, 0, 0, 0,
								0, 1, 0, 0,
								0, 0, 1, 0,
								0, 0, 0, 1)

func MatToSlice(mat *(glam.Mat4)) [16]float32 {
	vals := [4][4]float32(*mat)
	var expanded [16]float32
	for i, _ := range(vals) {
		for j, val := range(vals[i]) {
			expanded[i * 4 + j] = val
		}
	}
	return expanded
}

func ScalarMatrix(s float32) *glam.Mat4 {
	return glam.NewMat4(s, 0, 0, 0,
						0, s, 0, 0,
						0, 0, s, 0,
						0, 0, 0, 1)
}

func VecTimesMat(v glam.Vec3, m glam.Mat4) glam.Vec3 {
	m2 := glam.NewMat4(v.X, v.Y, v.Z, 0,
					   0, 0, 0, 0,
					   0, 0, 0, 0,
					   0, 0, 0, 0)
	rotated := m.Times(m2)
	newVec := MatToSlice(&rotated)
	return glam.Vec3{newVec[0], newVec[4], newVec[8]}
}