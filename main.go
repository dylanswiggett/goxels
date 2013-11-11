package main

import(
	"github.com/banthar/Go-SDL/sdl"
	"github.com/go-gl/gl"
	"github.com/drakmaniso/glam"
	"fmt"
	"time"
	"math"
)

var WindowW, WindowH = 1200, 800

var camera *Camera

var cameraPosition, forwardDirection, rightDirection, upDirection glam.Vec3

func InitGL() {
	gl.ClearColor(1, 1, 1, 0)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.LINE_SMOOTH)
}

func main() {
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

	camera = MakeCamera()
	camera.SetPerspective(3.14/4.0)
	camera.SetMName("M")
	camera.SetVPName("VP")

	sdl.WarpMouse(WindowW / 2, WindowH / 2)
	for ev := sdl.PollEvent(); ev != nil; ev = sdl.PollEvent() {
	}

	cameraPosition = glam.Vec3{10, 10, 10}
	forwardDirection = glam.Vec3{-1, 0, 0}
	rightDirection = glam.Vec3{0, 0, 1}
	upDirection = glam.Vec3{0, 1, 0}

	cylinder := LoadModel("voxelRes/models/highres.obj")
	// suzy := LoadModel("voxelRes/models/suzy.obj")
	// torus := LoadModel("voxelRes/models/torus.obj")
	// teapot := LoadScaledModel("voxelRes/models/car.obj", .03)
	// scene := LoadScaledModel("voxelRes/models/car.obj", .2)
	// scene := LoadModel("voxelRes/models/whatever.obj")
	
	/* HANDLE OPENGL RENDERING */

	ticks := 2000
	running := true
	for running {
		time.Sleep(5 * time.Millisecond)
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

		rotVal := float64(float32(ticks) / 50)
		shader.GetUniformLocation("lightPos").Uniform3f(
			float32(3 * math.Cos(rotVal * .9 + 1)), 1, float32(3 * math.Sin(rotVal * .9 + 1)))
		// shader.GetUniformLocation("lightPos").Uniform3f(3, 1, 3)
		shader.GetUniformLocation("cameraPos").Uniform3f(cameraPosition.X, cameraPosition.Y, cameraPosition.Z);
		
		camera.SetView(cameraPosition, forwardDirection, upDirection)

		camera.Prepare(shader, Rotate(float32(rotVal), glam.Vec3{0, 1, 0}))
		// scene.Draw()
		cylinder.Draw()

		camera.Prepare(shader, Rotate(float32(rotVal), glam.Vec3{0, 1, 0}).Translate(glam.Vec3{5, 1.2, -4}))
		// teapot.Draw()

		sdl.GL_SwapBuffers()
	}

}