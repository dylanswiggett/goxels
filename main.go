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

func InitGL() {
	gl.ClearColor(0, 0, 0.1, 0)
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

	cylinder := LoadModel("voxelRes/models/cylinder.obj")
	suzy := LoadModel("voxelRes/models/suzy.obj")
	torus := LoadModel("voxelRes/models/torus.obj")

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
					fmt.Println("Click!");
				}
			case *sdl.MouseMotionEvent:
				//fmt.Println("Moved the mouse to", e.X, e.Y)
			default:
				break
			}
		}

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		rotVal := float64(float32(ticks) / 10)
		shader.GetUniformLocation("lightPos").Uniform3f(
			float32(8 * math.Cos(rotVal)), 0, float32(8 * math.Sin(rotVal)))

		var camRad float64 = 5

		camera.LookAt(
		    glam.Vec3{float32(camRad * math.Cos(rotVal / 5)),-3 + float32(ticks) * .005 , float32(camRad * math.Sin(rotVal / 5))}, // Eye
		    glam.Vec3{0,0,0}, // Look
		    glam.Vec3{0,1,0}) // Up

		scale := ScaleBy(.3)
		scale = scale.Rotate(.005 * float32(ticks), glam.Vec3{1.0, 1.0, 0}.Normalized())

		num := 5

		for x := -num; x <= num; x++ {
			for y := -num; y <= num; y++ {
				for z := -num; z <= num; z++ {
					if (x * y + y * z + z * x) % 3 == 0 {
						newScale := float32(num) / (float32(math.Abs(float64(x * y * z / num))) + float32(num))
						model := scale.Scale(glam.Vec3{newScale, newScale, newScale})
						model = model.Translate(glam.Vec3{float32(x),float32(y),float32(z)})
						camera.Prepare(shader, model)
						// DrawCube()
						if x % 2 == 0 {
							cylinder.Draw()
						} else if y % 2 == 0{
							suzy.Draw()
						} else {
							torus.Draw()
						}
					}
				}
			}
		}

		// DisableCubeRendering()

		sdl.GL_SwapBuffers()
	}

}