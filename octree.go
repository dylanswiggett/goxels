package main

type OctreeNode struct {
	Children [2][2][2]*OctreeNode
	ColorData Color
	Emittance, Absorption, Reflection float32
}