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

func NewOctree(position, dimension Vec3, maxSubdiv int) Octree {
	return Octree{NewOctreeNode(), maxSubdiv, position, dimension}
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
						Contents.AverageVoxelsInRange(
						xMod, yMod, zMod, xMod + 1, yMod + 1, zMod + 1)
					
					node.Contents.SetVoxel(x, y, z, &avgVox)
				}
			}
		}
	}
}

func (tree *Octree) BuildTree() {
	tree.Root.BuildTree()
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