package main

import(
	"github.com/banthar/Go-SDL/sdl"
	"github.com/hishboy/gocommons/lang"
	"github.com/go-gl/gl"
	"github.com/drakmaniso/glam"
	"fmt"
	"time"
	// "math"
)

const (
	SSBO_BINDING     = 0
)

var WindowW, WindowH = 1200, 800

var camera *Camera

var cameraPosition, forwardDirection, rightDirection, upDirection glam.Vec3

type TestBlock struct {
	test1 int
	test2 int
}

func InitGL() {
	gl.ClearColor(1, 1, 1, 0)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.LINE_SMOOTH)
	gl.Enable(gl.TEXTURE_3D)
}

func main() {
	fmt.Println("Generating simple test octree...")
	tree := NewOctree(V3(0, 0, 0), V3(10, 10, 10), 4)
	data := 0
	for x := float32(0); x <= 10.0; x += .05 {
		for y := float32(0); y <= 10.0; y+= .05 {
			for z := float32(0); z <= 1.0 && x + z <= 10; z+= .05 {
				data++
				testVoxel := NewVoxel(1, x / 10.0, 0, 1, V3(1, 0, 0))
				tree.AddVoxel(&testVoxel, V3(x, y, x + z))
			}
		}
	}
	fmt.Println("Called AddVoxel", data, "times.")
	tree.BuildTree()
	octreeData, brickData, brickDim := tree.BuildGPURepresentation()

	testBools := make([]bool, len(octreeData))
	testQueue := lang.NewQueue()
	testQueue.Push(uint(0))
	for testQueue.Len() != 0 {
		loc := testQueue.Poll().(uint)
		if testBools[loc] == true {
			fmt.Println("Found node loop. Exiting.")
			panic(1)
		}
		testBools[loc] = true
		if octreeData[loc * 2] & 0x80000000 == 0 {
			fmt.Printf("Child data for %3X data:0x%8X\n", loc, octreeData[loc * 2])
			childLoc := octreeData[loc * 2] & 0x3FFFFFFF;
			testQueue.Push(childLoc)
			testQueue.Push(childLoc + 1)
			testQueue.Push(childLoc + 2)
			testQueue.Push(childLoc + 3)
			testQueue.Push(childLoc + 4)
			testQueue.Push(childLoc + 5)
			testQueue.Push(childLoc + 6)
			testQueue.Push(childLoc + 7)
		}
	}

	// for _, test := range(testBools) {
	// 	// fmt.Println("test")
	// 	if test == false {
	// 		fmt.Println("Found disconnected node. Exiting.")
	// 		panic(1)
	// 	}
	// }

	fmt.Println("Initializing rendering...")
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
	// shader.Use()

	cameraPosition = glam.Vec3{-1, 0, 0}
	forwardDirection = glam.Vec3{1, 0, 0}
	rightDirection = glam.Vec3{0, 0, 1}
	upDirection = glam.Vec3{0, 1, 0}

	bricks := gl.GenTexture()
	bricks.Bind(gl.TEXTURE_3D)
	gl.TexImage3D(gl.TEXTURE_3D, 0, gl.RGBA,
		brickDim, brickDim, brickDim,
		0, gl.RGBA, gl.UNSIGNED_BYTE, brickData)
	shader.GetUniformLocation("voxelBlocks").Uniform1i(int(bricks))
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	fmt.Printf("%X\n", gl.GetError())

	shaderOctreeData := make([]int32, len(octreeData))
	for i, data := range(octreeData) {
		shaderOctreeData[i] = int32(data)
	}

	offsetList := make([]int32, 1)
	gl.GetIntegerv(gl.SHADER_STORAGE_BUFFER_OFFSET_ALIGNMENT, offsetList)
	offset:= int(offsetList[0])

	gl.GetIntegerv(gl.SHADER_STORAGE_BUFFER_OFFSET_ALIGNMENT, offsetList)
	fmt.Println("SSBO Offset:",offsetList[0])

	shader.Use()

	octreeBuffer := gl.GenBuffer()
	octreeBuffer.Bind(gl.SHADER_STORAGE_BUFFER)
	octreeBuffer.BindBufferRange(gl.SHADER_STORAGE_BUFFER, SSBO_BINDING, offset, uint(len(octreeData) * 4))
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, len(octreeData) * 4, shaderOctreeData, gl.STATIC_DRAW)
	octreeIndex := shader.GetProgramResourceIndex(gl.SHADER_STORAGE_BLOCK, "octree")
	if octreeIndex == gl.INVALID_INDEX {
		fmt.Println("Failed to allocate GPU buffer for octree data. Exiting.")
		panic(1)
	}
	shader.ShaderStorageBlockBinding(octreeIndex, SSBO_BINDING)
	gl.MemoryBarrier(gl.SHADER_STORAGE_BARRIER_BIT)

	shader.GetUniformLocation("worldSize").Uniform3f(10.0, 10.0, 10.0)

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
		shader.GetUniformLocation("time").Uniform1i(ticks)

		camera.Prepare(shader, Scale(glam.Vec3{0,float32(WindowH) / float32(WindowW), 1}))
		plane.Draw()

		sdl.GL_SwapBuffers()
	}

}