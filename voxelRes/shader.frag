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
	ivec3 pos = ivec3(start * worldVoxelSize / worldSize);
	ivec3 final  = ivec3((start + dir * tmax) * worldVoxelSize / worldSize);
	
	//Iterate through the region and find a color to render.
	uint node = uint(0);
	vec3 scale = worldSize / 2;

    ivec3 error = ivec3(0, 0, 0);
    ivec3 d = final - pos;
    ivec3 a = ivec3(abs(d)) << 1;
    ivec3 s = ivec3(sign(d));
    ivec3 dom;

    int domDisplacement;
    if (a.x >= max(a.y, a.z)) { /* x dominant */
    	dom = ivec3(1, 0, 0);
    	domDisplacement = a.x >> 1;
    } else if (a.y >= max(a.x, a.z)) { /* y dominant */
    	dom = ivec3(0, 1, 0);
    	domDisplacement = a.y >> 1;
    } else { /* z dominant */
    	dom = ivec3(0, 0, 1);
    	domDisplacement = a.z >> 1;
    }

    ivec3 sub = ivec3(1, 1, 1) - dom;

    ivec3 aD = a * dom;
    ivec3 sD = s * dom;
    ivec3 aS = a * sub;
    ivec3 fD = final * dom;

    ivec3 zeros = ivec3(0, 0, 0);
    ivec3 worldVoxelSizeV = ivec3(worldVoxelSize, worldVoxelSize, worldVoxelSize);
    
    // Uses a 3D version of Bresenham's Algo.
    // From: http://www.luberth.com/plotter/line3d.c.txt.html
    error = sub * (a - domDisplacement);
    for (;all(lessThanEqual(zeros, pos)) && all(greaterThan(worldVoxelSizeV, pos));) {
        // if (pos * dom == fD) return vec4(0, 0, 0, 1);
        vec4 color = colorAtIntegerLoc(pos);
        if (color != vec4(0, 0, 0, 0))
        	return color;
        bvec3 testBools = greaterThanEqual(error * sub, ivec3(0, 0, 0));
        ivec3 testVals = ivec3(testBools) * sub;
        pos += s * testVals;
        error -= domDisplacement * testVals << 1;
        pos += sD;
        error += aS;
    }
    return vec4(0, 0, 0, 1);
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