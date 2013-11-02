package main

import(
	"io/ioutil"
	"github.com/go-gl/gl"
	"fmt"
)

func readFileToString(path string) string{
	contents, _ := ioutil.ReadFile(path)
	return string(contents)
}

func createShader(vertex, fragment string) gl.Program{
	// vertex shader
	vshader := gl.CreateShader(gl.VERTEX_SHADER)
	vshader.Source(readFileToString(vertex))
	vshader.Compile()
	fmt.Println("Making vertex shader!")
	if vshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		panic("vertex shader error: " + vshader.GetInfoLog())
	}
	fmt.Println("Done making vertex shader!")

	// fragment shader
	fshader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fshader.Source(readFileToString(fragment))
	fshader.Compile()
	if fshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		panic("fragment shader error: " + fshader.GetInfoLog())
	}

	// program
	prg := gl.CreateProgram()
	prg.AttachShader(vshader)
	prg.AttachShader(fshader)
	prg.Link()
	if prg.Get(gl.LINK_STATUS) != gl.TRUE {
		panic("linker error: " + prg.GetInfoLog())
	}

	return prg
}