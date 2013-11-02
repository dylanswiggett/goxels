package main

import "math"

/*
 * NOTE: This is in addition to the glam implementation.
 * This is only for use when the features of glam aren't needed,
 * since it allows much less verbose vector manipulation.
 */

/*
 * 3D VECTORS
 */

type Vec3 struct {
	X, Y, Z float32
}

func V3(x, y, z float32) Vec3 {
	return Vec3{x, y, z}
}

func (v *Vec3) Mag2() float32 {
	return v.X * v.X + v.Y * v.Y + v.Z * v.Z
}

func (v *Vec3) Mag() float32 {
	return float32(math.Sqrt(float64(v.Mag2())))
}

func (v *Vec3) Normalize() Vec3 {
	return v.Scale(1 / (v.Mag()))
}

func (v *Vec3) Add(v1 Vec3) Vec3 {
	return V3(v.X + v1.X, v.Y + v1.Y, v.Z + v1.Z)
}

func (v *Vec3) Scale(s float32) Vec3 {
	return V3(v.X * s, v.Y * s, v.Z * s)
}

func (v *Vec3) Dot(v1 Vec3) float32 {
	return v.X * v1.X + v.Y * v1.Y + v.Z * v1.Z
}

func (v *Vec3) Cross(v1 Vec3) Vec3 {
	return V3(v.Y * v1.Z - v.Z * v1.Y, v.Z * v1.X - v.X * v1.Z, v.X * v1.Y - v.Y * v1.X)
}


/*
 * 2D VECTORS
 */

type Vec2 struct {
	X, Y float32
}

func V2(x, y float32) Vec2 {
	return Vec2{x, y}
}

func (v *Vec2) Mag2() float32 {
	return v.X * v.X + v.Y * v.Y
}

func (v *Vec2) Mag() float32 {
	return float32(math.Sqrt(float64(v.Mag2())))
}

func (v *Vec2) Normalize() Vec2 {
	return v.Scale(1 / v.Mag())
}

func (v *Vec2) Add(v1 Vec2) Vec2{
	return V2(v.X + v1.X, v.Y + v1.Y)
}

func (v *Vec2) Scale(s float32) Vec2 {
	return V2(v.X * s, v.Y * s)
}

func (v *Vec2) Dot(v1 Vec2) float32 {
	return v.X * v1.X + v.Y * v1.Y
}

func (v *Vec3) CrossZ(v1 Vec2) float32 {
	return v.X * v1.Y - v.Y * v1.X
}