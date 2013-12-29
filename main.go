package main

import(
	"github.com/banthar/Go-SDL/sdl"
	"github.com/hishboy/gocommons/lang"
	"github.com/dylanswiggett/gl"
	"github.com/drakmaniso/glam"
	"fmt"
	"time"
	"math"
)

const (
	SSBO_BINDING     = 0 	// Must line up with the # in voxelRes/shader.frag
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
	tree := NewOctree(V3(0, 0, 0), V3(10, 10, 10), 7)
	data := 0
	for x := float32(0); x <= 10.0; x += .05 {
		for y := float32(0); y <= 10.0; y+= .05 {
			// for z := float32(0); float64(z) < math.Sin(float64(x)) + 2.0 && x + z <= 10.0; z+= 0.1 {
			// 	data++
			// 	testVoxel := NewVoxel(0, float32(int(x * 1000) % 100) / 100.0, 0, 1, V3(1, 0, 0))
			// 	tree.AddVoxel(&testVoxel, V3(x, y, x + z))
			// }
			testVoxel := NewVoxel(x / 10.0, y / 10.0, 0, 1, V3(1, 0, 0))
			tree.AddVoxel(&testVoxel, V3(x, y, (y * y / (x + 1)) / 10.0))
			data++
		}
	}
	fmt.Println("Called AddVoxel", data, "times.")
	tree.BuildTree()
	octreeData, brickData, brickDim := tree.BuildGPURepresentation()

	/*
	 * A few tests to confirm octree integrity.
	 */

	testBools := make([]bool, len(octreeData) / 2)
	testQueue := lang.NewQueue()
	testQueue.Push(uint32(0))
	numNonLeaf := 0
	for testQueue.Len() != 0 {
		loc := testQueue.Poll().(uint32)
		if testBools[loc] == true {
			fmt.Println("Found node loop. Exiting.")
			panic(1)
		}
		testBools[loc] = true
		if octreeData[loc * 2] & 0x80000000 == 0 {
			childLoc := octreeData[loc * 2] & 0x3FFFFFFF;
			testQueue.Push(childLoc)
			testQueue.Push(childLoc + 1)
			testQueue.Push(childLoc + 2)
			testQueue.Push(childLoc + 3)
			testQueue.Push(childLoc + 4)
			testQueue.Push(childLoc + 5)
			testQueue.Push(childLoc + 6)
			testQueue.Push(childLoc + 7)
			numNonLeaf++
		}
	}

	disconnectedNodes := 0
	for _, test := range(testBools) {
		if test == false {
			disconnectedNodes++
		}
	}
	if disconnectedNodes != 0 {
		fmt.Println("Found", disconnectedNodes, "disconnected nodes. Exiting.")
		panic(1)
	}

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
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexImage3D(gl.TEXTURE_3D, 0, gl.RGBA,
		brickDim, brickDim, brickDim,
		0, gl.RGBA, gl.UNSIGNED_BYTE, brickData)
	bricks.Unbind(gl.TEXTURE_3D)

	shader.Use()

	shaderOctreeData := make([]int32, len(octreeData))
	for i, data := range(octreeData) {
		shaderOctreeData[i] = int32(data)
	}

	octreeBuffer := gl.GenBuffer()
	octreeBuffer.Bind(gl.SHADER_STORAGE_BUFFER)
	octreeBuffer.BindBufferRange(gl.SHADER_STORAGE_BUFFER, 0, 0, uint(len(octreeData) * 4))
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, len(octreeData) * 4, shaderOctreeData, gl.STATIC_DRAW)
	octreeIndex := shader.GetProgramResourceIndex(gl.SHADER_STORAGE_BLOCK, "octree")
	if octreeIndex == gl.INVALID_INDEX {
		fmt.Println("Failed to allocate GPU buffer for octree data. Exiting.")
		panic(1)
	}
	shader.ShaderStorageBlockBinding(octreeIndex, SSBO_BINDING)
	gl.MemoryBarrier(gl.SHADER_STORAGE_BARRIER_BIT)
	octreeBuffer.Unbind(gl.SHADER_STORAGE_BUFFER)

	shader.GetUniformLocation("worldSize").Uniform3f(10.0, 10.0, 10.0)
	shader.GetUniformLocation("worldVoxelSize").Uniform1i(
		int(math.Pow(2, float64(tree.MaxSubdiv))) * BRICK_SIZE);
	shader.GetUniformLocation("numNodes").Uniform1i(len(octreeData) / 2);

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
	startTime := time.Now().UnixNano()
	running := true
	for running {
		time.Sleep(10 * time.Millisecond)
		ticks += 1

		if ticks % 20 == 0 {
			seconds := time.Now().UnixNano() - startTime;
			fmt.Println("20 ticks at", float32(20.0 * 1e9) / float32(seconds), "FPS.");
			startTime = time.Now().UnixNano()
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
			cameraPosition = cameraPosition.Plus(rightDirection.Times(-.5))
		}
		if sdl.GetKeyState()[sdl.K_d] == 1 {
			cameraPosition = cameraPosition.Plus(rightDirection.Times(.5))
		}
		if sdl.GetKeyState()[sdl.K_w] == 1 {
			cameraPosition = cameraPosition.Plus(forwardDirection.Times(.5))
		}
		if sdl.GetKeyState()[sdl.K_s] == 1 {
			cameraPosition = cameraPosition.Plus(forwardDirection.Times(-.5))
		}


		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		shader.GetUniformLocation("voxelBlocks").Uniform1i(0)
		gl.ActiveTexture(gl.TEXTURE0)
		bricks.Bind(gl.TEXTURE_3D)

		shader.GetUniformLocation("lightPos").Uniform3f(3, 1, 3)
		shader.GetUniformLocation("cameraPos").Uniform3f(cameraPosition.X, cameraPosition.Y, cameraPosition.Z);
		shader.GetUniformLocation("cameraUp").Uniform3f(upDirection.X, upDirection.Y, upDirection.Z)
		shader.GetUniformLocation("cameraRight").Uniform3f(rightDirection.X, rightDirection.Y, rightDirection.Z)
		shader.GetUniformLocation("cameraForwards").Uniform3f(forwardDirection.X, forwardDirection.Y, forwardDirection.Z)
		shader.GetUniformLocation("time").Uniform1i(ticks)

		shader.GetUniformLocation("widthPix").Uniform1i(WindowW);
		shader.GetUniformLocation("heightPix").Uniform1i(WindowH);

		camera.Prepare(shader, Scale(glam.Vec3{0,float32(WindowH) / float32(WindowW), 1}))
		plane.Draw()

		bricks.Unbind(gl.TEXTURE_3D)

		sdl.GL_SwapBuffers()
	}

}