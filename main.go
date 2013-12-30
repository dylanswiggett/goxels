package main

import(
	"github.com/go-gl/glfw"
	"github.com/hishboy/gocommons/lang"
	"github.com/dylanswiggett/gl"
	"github.com/drakmaniso/glam"
	"fmt"
	"time"
	"math"
	"runtime"
	"unsafe"
)

const (
	SSBO_BINDING     = 1 	// Must line up with the # in voxelRes/shader.frag
	MAX_OCTREE_NODES = 1000000
	WindowW, WindowH = 1200, 800
	CAMERA_SLOWDOWN  = .7
)

var (
	Cam *Camera
	CameraPosition, ForwardDirection, RightDirection, UpDirection glam.Vec3
	CameraVelocity glam.Vec3
	CameraRotXVel, CameraRotYVel float32

	Bricks gl.Texture
	OctreeBuffer gl.Buffer
	shader gl.Program
	OctreeData, BrickData []uint32
	BrickDim int
	OctreeDataList *[MAX_OCTREE_NODES]uint32
	OctreeDataPointer unsafe.Pointer
	Running bool
)

type TestBlock struct {
	test1 int
	test2 int
}

func InitGL() {
	gl.Enable(gl.TEXTURE_3D)
}

func main() {
	// fmt.Println("Generating simple test octree...")
	// tree := NewOctree(V3(0, 0, 0), V3(40, 10, 30), 6)
	// data := 0
	// for x := float32(0); x < 10.0; x += .02 {
	// 	for y := float32(0); y < 10.0; y+= .02 {
	// 		// for z := float32(0); float64(z) < math.Sin(float64(x)) + 2.0 && x + z < 10.0; z+= 0.1 {
	// 		// 	data++
	// 		// 	testVoxel := NewVoxel(0, float32(int(x * 100) % 100) / 100.0, 0, .2, V3(1, 0, 0))
	// 		// 	tree.AddVoxel(&testVoxel, V3(x, y, x + z))
	// 		// }
	// 		testVoxel := NewVoxel(x / 10.0, y / 10.0, 0, x / 20.0, V3(1, 0, 0))
	// 		tree.AddVoxel(&testVoxel, V3(x, y, float32(math.Sqrt(math.Pow(float64(x - 5.0), 2) + math.Pow(float64(y - 5.0), 2)))))
	// 		data++
	// 	}
	// }
	// fmt.Println("Called AddVoxel", data, "times.")
	fmt.Println("Loading a data set.")
	tree := readRAWVoxelFile("voxelRes/models/walnut.raw", 352, 296, 400)
	tree.BuildTree()
	OctreeData, BrickData, BrickDim = tree.BuildGPURepresentation()

	/*
	 * A few tests to confirm octree integrity.
	 */

	testBools := make([]bool, len(OctreeData) / 2)
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
		if OctreeData[loc * 2] & 0x80000000 == 0 {
			childLoc := OctreeData[loc * 2] & 0x3FFFFFFF;
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
	glfw.Init()
	defer glfw.Terminate()

	glfw.OpenWindow(WindowW, WindowH, 8, 8, 8, 8, 0, 0, glfw.Windowed)

	glfw.Disable(glfw.MouseCursor)
	defer glfw.Enable(glfw.MouseCursor)
	glfw.SetMousePos(WindowW / 2, WindowH / 2)

	if err := gl.Init(); int(err) != 0 {
		panic(err)
	}

	InitGL()
	EnableModelRendering()

	shader = createShader("voxelRes/shader.vert",
		"voxelRes/shader.frag")

	CameraPosition = glam.Vec3{-1, 0, 0}
	ForwardDirection = glam.Vec3{1, 0, 0}
	RightDirection = glam.Vec3{0, 0, -1}
	UpDirection = glam.Vec3{0, -1, 0}

	Bricks = gl.GenTexture()
	defer Bricks.Delete()
	Bricks.Bind(gl.TEXTURE_3D)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_3D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexImage3D(gl.TEXTURE_3D, 0, gl.RGBA,
		BrickDim, BrickDim, BrickDim,
		0, gl.RGBA, gl.UNSIGNED_BYTE, BrickData)
	Bricks.Unbind(gl.TEXTURE_3D)

	shader.Use()

	// Make the octree buffer
	OctreeBuffer = gl.GenBuffer()
	defer OctreeBuffer.Delete()
	OctreeBuffer.Bind(gl.SHADER_STORAGE_BUFFER)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, len(OctreeData) * 4, nil, gl.STATIC_DRAW)
	
	// Fill the octree buffer
	OctreeDataPointer = gl.MapBuffer(gl.SHADER_STORAGE_BUFFER, gl.WRITE_ONLY)
	OctreeDataList = (*[MAX_OCTREE_NODES]uint32)(OctreeDataPointer)
	OctreeDataPointer = nil
	for i, dataPoint := range(OctreeData) {
		OctreeDataList[i] = dataPoint
	}
	gl.UnmapBuffer(gl.SHADER_STORAGE_BUFFER)

	// Tell the shader where the octree buffer is
	OctreeBuffer.BindBufferBase(gl.SHADER_STORAGE_BUFFER, SSBO_BINDING)

	shader.GetUniformLocation("worldSize").Uniform3f(10.0, 10.0, 10.0)
	shader.GetUniformLocation("worldVoxelSize").Uniform1i(
		int(math.Pow(2, float64(tree.MaxSubdiv))) * BRICK_SIZE);
	shader.GetUniformLocation("numNodes").Uniform1i(len(OctreeData) / 2);

	Cam = MakeCamera()
	Cam.SetOrthographic(1)
	Cam.SetMName("M")
	Cam.SetVPName("VP")
	Cam.SetView(CameraPosition, ForwardDirection, UpDirection)

	plane := LoadModel("voxelRes/models/plane.obj")
	
	/* HANDLE OPENGL RENDERING */

	gl.ActiveTexture(gl.TEXTURE0)
	Bricks.Bind(gl.TEXTURE_3D)
	defer Bricks.Unbind(gl.TEXTURE_3D)

	OctreeBuffer.Bind(gl.SHADER_STORAGE_BUFFER)
	defer OctreeBuffer.Unbind(gl.SHADER_STORAGE_BUFFER)

	// go handleInput()

	ticks := 2000
	startTime := time.Now().UnixNano()
	Running = true
	for Running {
		time.Sleep(10 * time.Millisecond)
		// For some reason, Cgo occasionally segfaults w/o this lock.
		runtime.LockOSThread()
		ticks += 1

		/* HANDLE UI INTERACTIONS */

		mX, mY := glfw.MousePos()
		mX -= WindowW / 2
		mY -= WindowH / 2
		glfw.SetMousePos(WindowW / 2, WindowH / 2)
		CameraRotXVel = (CameraRotXVel + .001 * float32(mX)) * CAMERA_SLOWDOWN
		CameraRotYVel = (CameraRotYVel - .001 * float32(mY)) * CAMERA_SLOWDOWN
		if (CameraRotXVel != 0 || CameraRotYVel != 0) {
			hRot := glam.Rotation(CameraRotXVel, UpDirection)
			ForwardDirection = VecTimesMat(ForwardDirection, hRot)
			RightDirection = VecTimesMat(RightDirection, hRot)

			vRot := glam.Rotation(CameraRotYVel, RightDirection)
			ForwardDirection = VecTimesMat(ForwardDirection, vRot)
			UpDirection = VecTimesMat(UpDirection, vRot)
		}

		if (glfw.Key(glfw.KeyEsc) == glfw.KeyPress) ||
		   (glfw.Key(81) == glfw.KeyPress) {	//q
			Running = false
		}

		if glfw.Key(65) == glfw.KeyPress {	//a
			CameraVelocity = CameraVelocity.Plus(RightDirection.Times(-.5))
		}
		if glfw.Key(68) == glfw.KeyPress {	//d
			CameraVelocity = CameraVelocity.Plus(RightDirection.Times(.5))
		}
		if glfw.Key(87) == glfw.KeyPress {	//w
			CameraVelocity = CameraVelocity.Plus(ForwardDirection.Times(.5))
		}
		if glfw.Key(83) == glfw.KeyPress {	//s
			CameraVelocity = CameraVelocity.Plus(ForwardDirection.Times(-.5))
		}

		CameraPosition = CameraPosition.Plus(CameraVelocity)
		CameraVelocity = CameraVelocity.Times(CAMERA_SLOWDOWN)

		if ticks % 20 == 0 {
			seconds := time.Now().UnixNano() - startTime;
			fmt.Println("20 ticks at", float32(20.0 * 1e9) / float32(seconds), "FPS.");
			startTime = time.Now().UnixNano()
		}

		// go runtime.GC()
		// fmt.Println("ROUTINES: ", runtime.NumGoroutine())
		var mems runtime.MemStats
		runtime.ReadMemStats(&mems)
		fmt.Println("Allocated:", mems.HeapAlloc)

		shader.GetUniformLocation("voxelBlocks").Uniform1i(0)

		shader.GetUniformLocation("cameraPos").Uniform3f(CameraPosition.X, CameraPosition.Y, CameraPosition.Z);
		shader.GetUniformLocation("cameraUp").Uniform3f(UpDirection.X, UpDirection.Y, UpDirection.Z)
		shader.GetUniformLocation("cameraRight").Uniform3f(RightDirection.X, RightDirection.Y, RightDirection.Z)
		shader.GetUniformLocation("cameraForwards").Uniform3f(ForwardDirection.X, ForwardDirection.Y, ForwardDirection.Z)
		shader.GetUniformLocation("time").Uniform1i(ticks)
		shader.GetUniformLocation("widthPix").Uniform1i(WindowW);
		shader.GetUniformLocation("heightPix").Uniform1i(WindowH);

		Cam.Prepare(shader, Scale(glam.Vec3{0,float32(WindowH) / float32(WindowW), 1}))
		plane.Draw()

		glfw.SwapBuffers()
		
		runtime.UnlockOSThread()
	}
}

func handleInput() {
	Running = true
	for Running {
		time.Sleep(10 * time.Millisecond)

		/* HANDLE UI INTERACTIONS */

		mX, mY := glfw.MousePos()
		mX -= WindowW / 2
		mY -= WindowH / 2
		glfw.SetMousePos(WindowW / 2, WindowH / 2)
		if (mX != 0 || mY != 0) {
			hRot := glam.Rotation(.001 * float32(mX), UpDirection)
			ForwardDirection = VecTimesMat(ForwardDirection, hRot)
			RightDirection = VecTimesMat(RightDirection, hRot)

			vRot := glam.Rotation(-.001 * float32(mY), RightDirection)
			ForwardDirection = VecTimesMat(ForwardDirection, vRot)
			UpDirection = VecTimesMat(UpDirection, vRot)
		}

		if (glfw.Key(glfw.KeyEsc) == glfw.KeyPress) ||
		   (glfw.Key(81) == glfw.KeyPress) {	//q
			Running = false
		}

		if glfw.Key(65) == glfw.KeyPress {	//a
			CameraPosition = CameraPosition.Plus(RightDirection.Times(-.5))
		}
		if glfw.Key(68) == glfw.KeyPress {	//d
			CameraPosition = CameraPosition.Plus(RightDirection.Times(.5))
		}
		if glfw.Key(87) == glfw.KeyPress {	//w
			CameraPosition = CameraPosition.Plus(ForwardDirection.Times(.5))
		}
		if glfw.Key(83) == glfw.KeyPress {	//s
			CameraPosition = CameraPosition.Plus(ForwardDirection.Times(-.5))
		}
	}
}