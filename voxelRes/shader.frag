#version 430 core

#define LEAF_MASK  uint(0x80000000)
#define SOLID_MASK uint(0x40000000)
#define FINAL_MASK uint(0xA0000000)
#define CHILD_MASK uint(0x3FFFFFFF)
#define COORD_MASK uint(0x000003FF)

struct octreeNode {
	uint childData;
	uint colorData;
};

layout(std430, binding = 0) buffer octree {
	octreeNode nodes [];
};

// Nodes in the octree reference locations here for actual voxel data.
// Actually a huge number of 8x8x8 voxel blocks, packed into a single texture.
uniform sampler3D voxelBlocks;
uniform vec3 worldSize;
uniform int worldVoxelSize;

out vec3 color;

uniform int time;
uniform vec3 cameraPos;
uniform vec3 cameraUp;
uniform vec3 cameraRight;
uniform vec3 cameraForwards;
uniform vec3 lightPos;
uniform int widthPix, heightPix;

in vec3 cameraDir;

vec4 colorAtBrickLoc(ivec3 pos) {
	return texelFetch(voxelBlocks, pos.zyx, 0);
}

ivec3 nodeBrick(uint node) {
	uint loc = nodes[node].colorData;
	return ivec3((loc >> 20) & COORD_MASK, (loc >> 10) & COORD_MASK, loc & COORD_MASK);
}

vec4 colorAtLoc(vec3 pos) {
	uint node = uint(0);
	vec3 scale = worldSize / 2;
	int maxDepth = 5;
	while ((nodes[node].childData & FINAL_MASK) == uint(0) && (maxDepth--) > 0) {
		ivec3 childLoc = ivec3(pos / scale);
		int childOffset = childLoc.x * 4 + childLoc.y * 2 + childLoc.z;
		pos = pos - scale * childLoc;
		node = (nodes[node].childData & CHILD_MASK) + uint(childOffset);
		scale = scale / 2;
	}
	if ((nodes[node].childData & SOLID_MASK) != 0) {
		return vec4(0, 0, 0, 0);
	}
	ivec3 nodeBrickLoc = nodeBrick(node);
	nodeBrickLoc += ivec3(vec3(4, 4, 4) * pos / scale);
	// return vec4(, 0, 0, 1);
	// vec4 locColor = colorAtBrickLoc(nodeBrickLoc);
	// return vec4(1.0 - float(10 - maxDepth) / 10.0 * (1.0 - locColor.r),
	// 	locColor.g, locColor.b, locColor.a);
	return colorAtBrickLoc(nodeBrickLoc);
}

vec4 colorAtIntegerLoc(ivec3 pos) {
	uint node = uint(0);
	int scale = worldVoxelSize >> 1;
	while ((nodes[node].childData & FINAL_MASK) == uint(0)) {
		ivec3 childLoc = pos / scale;
		int childOffset = childLoc.x * 4 + childLoc.y * 2 + childLoc.z;
		pos = pos - scale * childLoc;
		node = (nodes[node].childData & CHILD_MASK) + uint(childOffset);
		scale = scale >> 1;
	}
	if ((nodes[node].childData & SOLID_MASK) != 0) {
		return vec4(0, 0, 0, 0);
	}
	ivec3 nodeBrickLoc = nodeBrick(node);
	nodeBrickLoc += ivec3(vec3(4, 4, 4) * pos / scale);
	return colorAtBrickLoc(nodeBrickLoc);
}

vec4 cAlong(vec3 start, vec3 dir) {

	vec4 c = vec4(0, 0, 0, 0);

	dir = normalize(dir);

	ivec3 dirS = ivec3(dir.x < 0 ? 1 : 0, dir.y < 0 ? 1 : 0, dir.z < 0 ? 1 : 0);

	// Find the intersection of the ray and the voxel region
	// Explanation: http://www.scratchapixel.com/lessons/3d-basic-lessons/lesson-7-intersecting-simple-shapes/ray-box-intersection/
	float tmin, tmax, tymin, tymax, tzmin, tzmax;
	if (dir.x == 0){ tmin = -10000; tmax = 100000; }
	else {
		tmin  = (dirS.x * worldSize.x - start.x) / dir.x;
		tmax  = ((1 - dirS.x) * worldSize.x - start.x) / dir.x;
	}
	if (dir.y == 0){ tymin = -10000; tymax = 100000; }
	else {
		tymin = (dirS.y * worldSize.y - start.y) / dir.y;
		tymax = ((1 - dirS.y) * worldSize.y - start.y) / dir.y;
	}
	if ((tmin > tymax) || (tymin > tmax)){ return c;}
	if (tymin > tmin){ tmin = tymin;}
	if (tymax < tmax){ tmax = tymax;}
	if (dir.z == 0){ tzmin = -10000; tzmax = 100000; }
	else {
		tzmin = (dirS.z * worldSize.z - start.z) / dir.z;
		tzmax = ((1 - dirS.z) * worldSize.z - start.z) / dir.z;
	}
	if ((tmin > tzmax) || (tzmin > tmax)){ return c;}
	if (tzmin > tmin){ tmin = tzmin;}
	if (tzmax < tmax){ tmax = tzmax;}
	if (tmax < 0) return c;
	start = start + dir * (tmin + .1);	// The startition at the edge of the voxel region.
	
	//Iterate through the region and find a color to render.
	uint node = uint(0);
	uint LOD = 0;
	int scale = int(log2(worldVoxelSize));
	ivec3 nodeOrigin = ivec3(-1000, -1000, -1000);
	ivec3 brickLoc;
	vec4 newColor;
	bool solid = false;

	// Traverse the tree.
    // From: http://www.cse.chalmers.se/edu/year/2013/course/TDA361/grid.pdf
    vec3 fPos = start * worldVoxelSize / worldSize;
    ivec3 pos = ivec3(fPos);
    ivec3 stepDir = ivec3(sign(dir));
    vec3 deltaT = 1 / dir;
	vec3 stepTmax = (vec3(pos + stepDir) - fPos) * deltaT;
	// return vec4(stepTmax, 1);
	deltaT = abs(deltaT);
    int maxSteps = 500;
    while (maxSteps-- > 0) {
    	if (stepTmax.x < stepTmax.y && stepTmax.x < stepTmax.z) {
    		pos.x += stepDir.x;
    		if (pos.x < 0 || pos.x >= worldVoxelSize)
    			return vec4(0, 0, 0, 0);
    		stepTmax.x += deltaT.x;
    	} else if (stepTmax.y < stepTmax.z) {
    		pos.y += stepDir.y;
    		if (pos.y < 0 || pos.y >= worldVoxelSize)
    			return vec4(0, 0, 0, 0);
    		stepTmax.y += deltaT.y;
    	} else {
    		pos.z += stepDir.z;
    		if (pos.z < 0 || pos.z >= worldVoxelSize)
    			return vec4(0, 0, 0, 0);
    		stepTmax.z += deltaT.z;
    	}
    	ivec3 offset = (pos - nodeOrigin);// >> (scale - 2);
    	if (LOD == 0 ||
    		offset.x <  0 || offset.y < 0  || offset.z <  0 ||
    		offset.x >= 8 || offset.y >= 8 || offset.z >= 8) {
    		node = uint(0);
			LOD = 0;
			nodeOrigin = ivec3(0, 0, 0);
			// brickLoc = nodeBrick(node);
			scale = int(log2(worldVoxelSize));
			ivec3 posTemp = pos;
			int maxLOD = 10;//maxSteps / 200;
			while ((nodes[node].childData & FINAL_MASK) == uint(0) && maxLOD-- >= 0) {
				scale--;
				LOD++;
				ivec3 childLoc = posTemp >> scale;
				int childOffset = childLoc.x * 4 + childLoc.y * 2 + childLoc.z;
				posTemp -= childLoc << scale;
				node = (nodes[node].childData & CHILD_MASK) + uint(childOffset);
				nodeOrigin += childLoc << scale;
			}
			if ((nodes[node].childData & SOLID_MASK) != 0) {
				newColor = vec4(0, 0, 0, 0);
				solid = true;
			} else {
				brickLoc = nodeBrick(node);
				solid = false;
			}
			offset = pos - nodeOrigin;
    	}
    	if (!solid) {
    		newColor = colorAtBrickLoc(brickLoc + offset);
    	}
		if (newColor != vec4(0, 0, 0, 0))
			return newColor;
    }
    return vec4(0, 0, 0, 0);
}

// 	while (all(lessThan(pos, worldSize)) && all(lessThan(vec3(0, 0, 0), pos))) {
// 		pos += normalize(dir) * .1;
// 	}
// }

void main(){
	// Calculate the vector for this fragment into the screen
	float xDisp = (gl_FragCoord.x - float(widthPix) / 2) / widthPix;
	float yDisp = (gl_FragCoord.y - float(heightPix) / 2) / heightPix;
	vec3 cDir = cameraForwards + xDisp * cameraRight - yDisp * cameraUp;
	color = cAlong(cameraPos, cDir).rgb;
	vec3 colorThere = colorAtLoc(vec3(1, 1, 1)).rgb;	// REMOVE
	// float i;
	// for (i = 0; i < 20; i += .1) {
	// 	vec3 dir = cameraPos + cDir * i;
	// 	if (dir.x > 0 && dir.x < 10 && dir.y > 0 && dir.y < 10 && dir.z > 0 && dir.z < 10) {
	// 		vec3 colorThere = colorAtLoc(dir).rgb;
	// 	// 	if (!(colorThere.r == 0 && colorThere.g == 0 && colorThere.b == 0)) {
	// 	// 		// color = colorThere;
	// 	// 		break;
	// 	// 	} 
	// 	}
	// }
}