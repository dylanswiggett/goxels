package main

import(
	"io/ioutil"
	"os"
	"image"
	"fmt"
)

func readPGM(path string) Octree {
	fImg, _ := os.Open(path)
	defer fImg.Close()
	img, _, err := image.Decode(fImg)

	if err != nil {
		panic(err)
	}

	bounds := img.Bounds()

	fmt.Println(bounds.Size().X)
	panic("exit!")

	tree := NewOctree(V3(0, 0, 0), V3(1, 1, 1), 3)

	return tree
}

func readRAWVoxelFile(path string, d1, d2, d3 int) Octree {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	// for i := 0; i < 100000; i++ {
	// 	if (contents[i] != 0) {
	// 		fmt.Printf("%d: 0x%8x\n", i, contents[i])
	// 	}
	// }
	// panic("exit")

	tree := NewOctree(V3(0, 0, 0), V3(1, 1, 1), 4)

	index := 0

	for x := 0; x < d1; x++ {
		fmt.Println("X:", x)
		for y := 0; y < d2; y++ {
			for z := 0; z < d3; z++ {
				// a := (uint(contents[index + 1]) << 8) + uint(contents[index])
				g := float32(uint(contents[index] >> 4)) / 16
				r := float32(uint(contents[index] & 15)) / 16
				a := float32(uint(contents[index + 1] >> 4)) / 16
				b := float32(uint(contents[index + 1] & 15)) / 16
				// a := contents[index + 0] & 0xF
				// r := contents[index + 1] >> 4
				// g := contents[index + 1] & 0xF

				// r := uint(0)
				// g := uint(0)
				// b := uint(0)
				// a := uint(0)

				// r := float32(10)
				// g := float32(10)
				// b := float32(10)

				// if data > 5 {
				// 	r = 1
				// 	a = 1
				// }

				// if r > 1 {
				// 	a = 100
				// } else {
				// 	a = 0
				// }
				// r = 10
				// g = 10
				// b = 10
				// a = 10
				if a >= .1 {
					
					// r = 100 * float32(x) / float32(d1)
					// g = 100 * float32(y) / float32(d2)
					// b = 100 * float32(z) / float32(d3)
					r *= 30
					g *= 15
					b *= 3
					// r = 40
					// g = 15
					// b = 3
					a *= 100
					// if (r > 20) {
					// 	r *= 2
					// 	a = 50
					// } else {
					// 	a = 10
					// }
					// r = 0
					// b = 0
					testVoxel := NewVoxel(float32(r) / 100,
										  float32(g) / 100,
										  float32(b) / 100,
										  float32(a) / 100, V3(1, 0, 0))
					tree.AddVoxel(&testVoxel, V3(float32(x) / float32(d1),
											     float32(y) / float32(d2),
											     float32(z) / float32(d3)))
				}
				index += 2
			}
		}
	}

	return tree
}