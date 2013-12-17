package main

import(
	"github.com/banthar/Go-SDL/sdl"
	"github.com/go-gl/gl"
	"github.com/drakmaniso/glam"
	"fmt"
	"time"
	// "math"
)

var WindowW, WindowH = 1200, 800

var camera *Camera

var cameraPosition, forwardDirection, rightDirection, upDirection glam.Vec3

func InitGL() {
	gl.ClearColor(1, 1, 1, 0)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.LINE_SMOOTH)
	gl.Enable(gl.TEXTURE_3D)
}

func main() {
	fmt.Println("Generating simple test octree")
	tree := NewOctree(V3(0, 0, 0), V3(10, 10, 10), 5)
	data := 0
	for x := float32(0); x <= 10.0; x += .01 {
		for y := float32(0); y <= 10.0; y+= .01 {
			data++
			testVoxel := NewVoxel(1, 1, 1, 1, V3(1, 0, 0))
			tree.AddVoxel(&testVoxel, V3(x, y, x))
		}
	}
	fmt.Println("Called AddVoxel", data, "times.")
	tree.BuildTree()
	octreeData, brickData := tree.BuildGPURepresentation()

	fmt.Println("Initializing...")
	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		panic(sdl.GetError())
	}
	defer sdl.Quit()

	if sdl.SetVideoMode(WindowW, WindowH, 32, sdl.OPENGL) == nil {
		panic("sdl error")
	}

	if err := gl.Init(); int(err) != 0 {
		panic(err)
	}

	InitGL()
	EnableModelRendering()

	shader := createShader("voxelRes/shader.vert",
		"voxelRes/shader.frag")
	shader.Use()

	cameraPosition = glam.Vec3{-1, 0, 0}
	forwardDirection = glam.Vec3{1, 0, 0}
	rightDirection = glam.Vec3{0, 0, 1}
	upDirection = glam.Vec3{0, 1, 0}

	shader.GetUniformLocation("octree").Uniform2iv(len(octreeData), octreeData)

	bricks := gl.GenTexture()
	bricks.Bind(gl.TEXTURE_3D)
	gl.TexImage3D(gl.TEXTURE_3D, 0, gl.RGBA,
		len(brickData), len(brickData[0]), len(brickData[0][0]), 
		0, gl.RGBA, gl.BYTE, brickData)
	shader.GetUniformLocation("voxelBlocks").Uniform1i(int(bricks))

	camera = MakeCamera()
	camera.SetOrthographic(1)
	camera.SetMName("M")
	camera.SetVPName("VP")
	camera.SetView(cameraPosition, forwardDirection, upDirection)

	sdl.WarpMouse(WindowW / 2, WindowH / 2)
	for ev := sdl.PollEvent(); ev != nil; ev = sdl.PollEvent() {
	}

	plane := LoadModel("voxelRes/models/plane.obj")
	
	/* HANDLE OPENGL RENDERING */

	ticks := 2000
	running := true
	for running {
		time.Sleep(20 * time.Millisecond)
		ticks += 1

		if ticks % 100 == 0 {
			fmt.Println("100 ticks!")
		}

		/* HANDLE UI INTERACTIONS */

		for ev := sdl.PollEvent(); ev != nil; ev = sdl.PollEvent() {

			switch e := ev.(type){
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseButtonEvent:
				if e.Type == sdl.MOUSEBUTTONDOWN {
					fmt.Println("Click!")
				}
			case *sdl.MouseMotionEvent:
				if e.State == 1 {
					hRot := glam.Rotation(.001 * float32(e.Xrel), upDirection)
					forwardDirection = VecTimesMat(forwardDirection, hRot)
					rightDirection = VecTimesMat(rightDirection, hRot)

					vRot := glam.Rotation(-.001 * float32(e.Yrel), rightDirection)
					forwardDirection = VecTimesMat(forwardDirection, vRot)
					upDirection = VecTimesMat(upDirection, vRot)
					// forwardDirection = forwardDirection.Plus(upDirection.Times(float32(e.Yrel) * .01))
				}
			case *sdl.KeyboardEvent:
				if e.Keysym.Sym == sdl.K_ESCAPE || e.Keysym.Sym == sdl.K_q{
					running = false
				}
			default:
				break
			}
		}

		if sdl.GetKeyState()[sdl.K_a] == 1 {
			cameraPosition = cameraPosition.Plus(rightDirection.Times(.5))
		}
		if sdl.GetKeyState()[sdl.K_d] == 1 {
			cameraPosition = cameraPosition.Plus(rightDirection.Times(-.5))
		}
		if sdl.GetKeyState()[sdl.K_w] == 1 {
			cameraPosition = cameraPosition.Plus(forwardDirection.Times(.5))
		}
		if sdl.GetKeyState()[sdl.K_s] == 1 {
			cameraPosition = cameraPosition.Plus(forwardDirection.Times(-.5))
		}


		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// rotVal := float64(float32(ticks) / 50)
		// shader.GetUniformLocation("lightPos").Uniform3f(
		// 	float32(3 * math.Cos(rotVal * .9 + 1)), 1, float32(3 * math.Sin(rotVal * .9 + 1)))
		shader.GetUniformLocation("lightPos").Uniform3f(3, 1, 3)
		shader.GetUniformLocation("cameraPos").Uniform3f(cameraPosition.X, cameraPosition.Y, cameraPosition.Z);
		
		camera.Prepare(shader, Scale(glam.Vec3{0,float32(WindowH) / float32(WindowW), 1}))
		plane.Draw()

		sdl.GL_SwapBuffers()
	}

}