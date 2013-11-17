package main

type RGBA struct {
	R, G, B, A float32
}

type Voxel struct {
	Properties RGBA
	Normal Vec3
}

func NewVoxel(r, g, b, a float32, normal Vec3) Voxel {
	return Voxel{RGBA{r, g, b, a}, normal}
}

func AverageVoxels(voxels []Voxel) Voxel {
	var avgR, avgG, avgB, avgA float32
	avgR, avgG, avgB, avgA = 0, 0, 0, 0//= avgG = avgB =  avgA = 0
	avgNorm := V3(0, 0, 0)

	for _, vox := range(voxels) {
		avgR += vox.Properties.R
		avgG += vox.Properties.G
		avgB += vox.Properties.B
		avgA += vox.Properties.A
		avgNorm = avgNorm.Add(vox.Normal)
	}

	n := float32(len(voxels))

	return NewVoxel(avgR / n, avgG / n, avgB / n, avgA / n, avgNorm.Scale(1.0 / n))
}

type Brick struct {
	Voxels [8][8][8]*Voxel
}

func NewBrick() Brick{
	var newVoxels [8][8][8]*Voxel
	return Brick{newVoxels}
}

func (b *Brick) SetVoxel(x, y, z int, v *Voxel) {
	b.Voxels[x][y][z] = v
}

func (b *Brick) AverageVoxelsInRange(x1, y1, z1, x2, y2, z2 int) Voxel{
	toAverage := make([]Voxel, (x2 - x1 + 1) * (y2 - y1 + 1) * (z2 - z1 + 1))
	addPos := 0
	for x := x1; x <= x2; x++ {
		for y := y1; y <= y2; y++ {
			for z := z1; z <= z2; z++ {
				if (b.Voxels[x][y][z] == nil) {
					toAverage[addPos] = NewVoxel(0, 0, 0, 0, V3(0, 0, 0))
				} else {
					toAverage[addPos] = *(b.Voxels[x][y][z])
				}
				addPos++
			}
		}
	}
	return AverageVoxels(toAverage)
}

func (b Brick) Dimension() int {
	return len(b.Voxels)
}