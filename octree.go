package main

import(
	"github.com/hishboy/gocommons/lang"
	"math"
	"fmt"
)

type OctreeNode struct {
	Children [2][2][2]*OctreeNode
	IsLeaf, IsSolid bool
	Contents Brick
	SolidProperties RGBA
}

func NewOctreeNode() OctreeNode{
	var children [2][2][2]*OctreeNode
	return OctreeNode{children, true, false, NewBrick(), RGBA{0, 0, 0, 0}}
}

type Octree struct {
	Root OctreeNode
	MaxSubdiv int
	Position, Dimension Vec3
}

func NewOctree(position, dimension Vec3, maxSubdiv int) Octree {
	return Octree{NewOctreeNode(), maxSubdiv, position, dimension}
}

func (o *OctreeNode) Subdivide() {
	for x := 0; x < len(o.Children); x++ {
		for y := 0; y < len(o.Children[x]); y++ {
			for z := 0; z < len(o.Children[x][y]); z++ {
				newNode := NewOctreeNode()
				o.Children[x][y][z] = &newNode
			}
		}
	}
}

/*
 * Scales the vector so that each of its elements is in the range [0,1],
 * as long as those values were non-negative and less than or equal to
 * the corresponding value in dimension.
 */
func ScaleToUnitDimension(v, dimension Vec3) Vec3 {
	return V3(v.X / dimension.X, v.Y / dimension.Y, v.Z / dimension.Z)
}

/*
 * Recursively adds the given voxel at the most precise point possible
 * in the octree. Subdivides the tree as necessary to put the voxel in
 * the correct location.
 * 
 * Assumes that the given voxel falls within the bounds of the octree.
 */
func (node *OctreeNode) AddVoxel(v *Voxel, pos Vec3, maxSubdiv int) {
	if node.IsLeaf {
		if maxSubdiv == 0 {
			// TODO: If there's already a voxel here, average!
			node.Contents.SetVoxel(int(pos.X * float32(BRICK_SIZE)),
								   int(pos.Y * float32(BRICK_SIZE)),
								   int(pos.Z * float32(BRICK_SIZE)), v)
			return
		} else {
			node.IsLeaf = false
			node.Subdivide()
		}
	}
	// Pick the right subnode, then calculate the new sub-position vector.
	pos = pos.Scale(2)
	var x, y, z = int(pos.X), int(pos.Y), int(pos.Z)
	if pos.X >= 1 {
		pos.X -= 1
	}
	if pos.Y >= 1 {
		pos.Y -= 1
	}
	if pos.Z >= 1 {
		pos.Z -= 1
	}
	node.Children[x][y][z].AddVoxel(v, pos, maxSubdiv - 1)
}

/*
 * Adds the given voxel at the most precise point possible in the octree.
 * Assumes that the given voxel falls within the bounds of the octree.
 */
func (tree *Octree) AddVoxel(v *Voxel, pos Vec3) {
	pos = pos.Subtract(tree.Position)	// Produce a relative vector
	pos = ScaleToUnitDimension(pos, tree.Dimension)
		// It's easier to index when the position is in [0,1]^3
	tree.Root.AddVoxel(v, pos, tree.MaxSubdiv)
}

func (tree *Octree) BuildTree() {
	tree.Root.BuildTree()
}

func (node *OctreeNode) BuildTree() {
	if !node.IsLeaf {
		for x := 0; x < 2; x++ {
			for y := 0; y < 2; y++ {
				for z := 0; z < 2; z++ {
					node.Children[x][y][z].BuildTree()
				}
			}
		}

		dim := node.Contents.Dimension()
		for x := 0; x < dim; x++ {
			for y := 0; y < dim; y++ {
				for z := 0; z < dim; z++ {
					xMod := (x * 2) % dim
					yMod := (y * 2) % dim
					zMod := (z * 2) % dim
					avgVox := node.Children[(x * 2) / dim][(y * 2) / dim][(z * 2) / dim].
						Contents.AverageVoxelsInRange(xMod, yMod, zMod, xMod + 1, yMod + 1, zMod + 1)
					node.Contents.SetVoxel(x, y, z, &avgVox)
				}
			}
		}
	}
}

func (tree *Octree) BuildGPURepresentation() ([]uint32, []uint32, int) {
	nodes := make([]OctreeNode, 0, 1000)
	childOffsets := make([]int, 0)
	nodeCount := 0
	nodeQueue := lang.NewQueue()
	nodeQueue.Push(tree.Root)

	// Find every node. Store in a list as a prefix traversal
	for nodeQueue.Len() != 0 {
		n := nodeQueue.Poll().(OctreeNode)
		nodeCount++
		nodes = append(nodes, n)
		childOffsets = append(childOffsets, nodeQueue.Len() + 1)
		if !n.IsLeaf {
			for x := 0; x < len(n.Children); x++ {
				for y := 0; y < len(n.Children[x]); y++ {
					for z := 0; z < len(n.Children[x][y]); z++ {
						nodeQueue.Push(*n.Children[x][y][z])
					}
				}
			}
		}
	}

	fmt.Println("Processing", nodeCount, "octree nodes.")

	// Build a small block for voxel data
	brickDimension := int(math.Ceil(math.Pow(float64(nodeCount), 1.0/3.0)))
	brickBlockDimension := BRICK_SIZE * brickDimension
	bricks := make([][][]uint32, brickBlockDimension)
	for i, _ := range(bricks) {
		bricks[i] = make([][]uint32, brickBlockDimension)
		for j, _ := range(bricks[i]) {
			bricks[i][j] = make([]uint32, brickBlockDimension)
		}
	}

	fmt.Println("Using a", brickBlockDimension, "cubed block of voxel data.")

	// Populate both of the return lists
	nodeData := make([]uint32, nodeCount * 2)
	totalVoxels := 0
	for pos := 0; pos < len(nodes); pos++ {
		brickX := BRICK_SIZE * (pos % brickDimension)
		brickY := BRICK_SIZE * ((pos / brickDimension) % brickDimension)
		brickZ := BRICK_SIZE * ((pos / (brickDimension * brickDimension)) % brickDimension)
		for x := 0; x < BRICK_SIZE; x++ {
			for y := 0; y < BRICK_SIZE; y++ {
				for z := 0; z < BRICK_SIZE; z++ {
					vox := nodes[pos].Contents.Voxels[x][y][z]
					colorInt := uint32(0)
					if vox != nil {
						colorInt = vox.ColorInt()
						totalVoxels++
					}
					bricks[brickX + x][brickY + y][brickZ + z] = colorInt
				}
			}
		}
		nodeVal := uint32(0)
		if nodes[pos].IsLeaf {
			nodeVal = 1
		}
		nodeVal = (nodeVal << 1)
		// TODO: set data type (not yet implemented)
		nodeVal = nodeVal << 30
		if !nodes[pos].IsLeaf {
			nodeVal += uint32(pos + childOffsets[pos])
		}

		nodeData[pos * 2] = nodeVal
		nodeData[pos * 2 + 1] = uint32((brickX << 10 + brickY) << 10 + brickZ)
	}
	fmt.Println("Found", totalVoxels, "individual voxels.")

	// Produce the final, 1D slice of brick data.
	// TODO: Make everything write to this 1D array to start with, rather than copying.
	flattenedBricks := make([]uint32, brickBlockDimension * brickBlockDimension * brickBlockDimension)
	for x := 0; x < brickBlockDimension; x++ {
		for y := 0; y < brickBlockDimension; y++ {
			for z := 0; z < brickBlockDimension; z++ {
				flattenedBricks[(x * brickBlockDimension + y) * brickBlockDimension + z] = bricks[x][y][z]
			}
		}
	}
	return nodeData, flattenedBricks, brickBlockDimension
}