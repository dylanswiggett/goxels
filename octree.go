package main

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

type Octree struct {
	Root OctreeNode
	MaxSubdiv int
	Position, Dimension Vec3
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
 * Scales the vector to fit inside a child of an OctreeNode. Basically,
 * the bytes x, y, and z indicate whether or not the vector is going
 * to be used inside the child node that is shifted in the x, y, and z
 * directions (0) or not (1).
 * 
 * Assumes the input vector is in [0,1]^3, and returns a vector in the
 * same range.
 */
func ScaleToHalfDimension(v Vec3, x, y, z byte) Vec3{
	newVec := V3(v.X * 2, v.Y * 2, v.Z * 2)
	if x == 1 {
		newVec.X = newVec.X - 1
	}
	if y == 1 {
		newVec.Y = newVec.Y - 1
	}
	if z == 1 {
		newVec.Z = newVec.Z - 1
	}
	return newVec
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
			node.Contents.SetVoxel(int(pos.X * 8),
								   int(pos.Y * 8),
								   int(pos.Z * 8), v)
			return
		} else {
			node.IsLeaf = false
			node.Subdivide()
		}
	}
	// Pick the right subnode, then calculate the new sub-position vector.
	var x, y, z = byte(0), byte(0), byte(0)
	if pos.X >= .5 {
		x = 1
	}
	if pos.Y >= .5 {
		y = 1
	}
	if pos.Z >= .5 {
		y = 1
	}
	pos = ScaleToHalfDimension(pos, x, y, z)
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