package main

import (
	"io/ioutil"
	"strings"
	"strconv"
	"fmt"
)

func toFloat(s string) float32 {
	val, _ := strconv.ParseFloat(s, 32)
	return float32(val)
}

func toByte(s string) byte {
	val, _ := strconv.Atoi(s)
	return byte(val)
}

// Convert "1/2/3" to [1, 2, 3]
// This format is used in .obj files for vertices.
func toInts(s string) [3]int {
	vals := strings.Split(s, "/")
	var ints [3]int
	if len(vals) == 1 {
		intVal, _ := strconv.Atoi(vals[0])
		ints[0] = intVal
		ints[1] = 1
		ints[2] = intVal
	} else {
		for i, val := range(vals) {
			intVal, _ := strconv.Atoi(val)
			ints[i] = intVal
		}
	}
	return ints
}

func makePolygons(filePath string) []Polygon {
	contents, _ := ioutil.ReadFile(filePath)
	file := string(contents)
	lines := strings.Split(file, "\n")

	positions := make(map[int]Vec3)
	colors := make(map[int]Color)
	normals := make(map[int]Vec3)

	polygons := make([]Polygon, 0)

	for _, line := range(lines) {
		if len(line) == 0 || line[0] == '#' {	// Comment line
			continue
		}

		params := strings.Split(line, " ")
		switch params[0] {
		case "v":	// Vertex
			positions[len(positions) + 1] =
				V3(toFloat(params[1]), toFloat(params[2]), toFloat(params[3]))
		case "c":	// Should be "vt" for texture vertex
			colors[len(colors) + 1] =
				Color{toByte(params[1]), toByte(params[2]), toByte(params[3])}
		case "vn":	// Vertex normal
			normals[len(normals) + 1] =
				V3(toFloat(params[1]), toFloat(params[2]), toFloat(params[3]))
		case "f":
			v1 := toInts(params[1])
			v2 := toInts(params[2])
			v3 := toInts(params[3])
			var vertices [3]Vertex
			vertices[0] = Vertex{positions[v1[0]], normals[v1[2]], colors[v1[1]]}
			vertices[1] = Vertex{positions[v2[0]], normals[v2[2]], colors[v2[1]]}
			vertices[2] = Vertex{positions[v3[0]], normals[v3[2]], colors[v3[1]]}
			polygons = append(polygons, Polygon{vertices})
		default:
			fmt.Println("Unsupported model param: " + line)
		}
	}

	return polygons
}

// Loads a shitty semi-.obj file format.
// See http://www.martinreddy.net/gfx/3d/OBJ.spec for actual spec
// See http://www.opengl-tutorial.org/beginners-tutorials/tutorial-7-model-loading/
// 		for fake spec.
func LoadModel(filePath string) Model {
	return MakeModel(makePolygons(filePath))
}

func LoadScaledModel(filePath string, scale float32) Model {
	return MakeScaledModel(makePolygons(filePath), scale)
}