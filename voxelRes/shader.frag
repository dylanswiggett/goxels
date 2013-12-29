#version 430 core

#define MAX_COMP(v) max(max(v.x, v.y), v.z)
#define MIN_COMP(v) min(min(v.x, v.y), v.z)

#define INT(b) (b ? 1 : 0)

#define PICK_BY(choice, v1, v2) vec3((choice.x != 0) ? v1.x : v2.x, (choice.y != 0) ? v1.y : v2.y, (choice.z != 0) ? v1.z : v2.z)

#define LEAF_MASK  uint(0x80000000)
#define SOLID_MASK uint(0x40000000)
#define FINAL_MASK uint(0xA0000000)
#define CHILD_MASK uint(0x3FFFFFFF)
#define COORD_MASK uint(0x000003FF)

#define MAX_DEPTH  10

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

uniform int numNodes;

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
	// return c;

	// Search through the octree.

	// octreeNodeParser nodeList[MAX_DEPTH];
	dir = normalize(dir);

	vec3 dInv = 1 / dir;

	// int vMask = TO_BITS(dir.x <= 0, dir.y <= 0, dir.z <= 0);
	ivec3 vMask = ivec3(INT(dir.x <= 0), INT(dir.y <= 0), INT(dir.z <= 0));

	vec3 xMin = vec3(0, 0, 0);
	vec3 xMax = worldSize;
	vec3 sMin = (xMin - start) * dInv;
	vec3 sMax = (xMax - start) * dInv;

	vec3 lowerLimits = PICK_BY(vMask, sMax, sMin);
	vec3 upperLimits = PICK_BY(vMask, sMin, sMax);
	float sLMax = MAX_COMP(lowerLimits);
	float sUMin = MIN_COMP(upperLimits);

	if (sLMax >= sUMin || sUMin < 0)
		return vec4(0, 0, 0, 1);

	if (sLMax < 0)
		sLMax = 0;

	// Keeps track of current position when raymarching.
	vec3 pos = start + dir * (sLMax + .001);
	vec3 dim = xMax;

	ivec3 subNodeLoc;

	vec3 nextXMin = xMin;
	// XMax will always be xMin + dim, so we don't need a separate var.

	uint node = 0;
	uint nextNode = node;
	int depth = 0;
	int maxSteps = 200;
	while (--maxSteps > 0) {
		// if (nextNode < 0 || nextNode >= numNodes)
		// 	return vec4(1, 0, 0, 1)
		if ((nodes[nextNode].childData & FINAL_MASK) != 0) {
			if ((nodes[nextNode].childData & SOLID_MASK) == 0) {
				/*
				 * Look at a leaf node.
				 */
				// return vec4(pos, 1);
				c += vec4(1, 1, 1, 1) * .1;// * float(maxSteps) / 100.0;
				if (c.a >= 1)
					return c;
			}
			/*
			 * Prepare for next node to check.
			 */
			vec3 nextXMax = nextXMin + dim;
			sMin = (nextXMin - start) * dInv;
			sMax = (nextXMax - start) * dInv;
			sUMin = MIN_COMP(PICK_BY(vMask, sMin, sMax));

			nextXMin = xMin;
			nextNode = node;
			pos = start + dir * (sUMin + .0001);
			dim *= 2;
			depth--;
			// return vec4(nextPos / 20, 1);	// For pretty colors!
		} else {
			/*
			 * Parse down to find a leaf node (or empty node)
			 */
			node = nextNode;
			xMin = nextXMin;

			dim /= 2;
			subNodeLoc = ivec3(floor((pos - xMin) / dim));
			if (subNodeLoc.x < 0 || subNodeLoc.x > 1 ||
				subNodeLoc.y < 0 || subNodeLoc.y > 1 ||
				subNodeLoc.z < 0 || subNodeLoc.z > 1) {
				/*
				 * In this case we either need to exit our search,
				 * or restart from the root node.
				 */
				if (depth == 0) {
					return c;
			 	} else {
			 		nextXMin = vec3(0, 0, 0);
			 		dim = worldSize;
			 		nextNode = 0;
			 		depth = 0;
			 	}
			} else {
				nextXMin = xMin + dim * subNodeLoc;
				nextNode = (nodes[node].childData & CHILD_MASK)
					+ (subNodeLoc.x << 2) + (subNodeLoc.y << 1) + subNodeLoc.z;
				depth++;
			}
		}
	}
	return c;
}

void main(){
	// Calculate the vector for this fragment into the screen
	float xDisp = (gl_FragCoord.x - float(widthPix) / 2) / widthPix;
	float yDisp = (gl_FragCoord.y - float(heightPix) / 2) / heightPix;
	vec3 cDir = cameraForwards + xDisp * cameraRight - yDisp * cameraUp;
	color = cAlong(cameraPos, cDir).rgb;
	vec3 colorThere = colorAtLoc(vec3(1, 1, 1)).rgb;	// TO PREVENT STARTUP ERROR (FOR NOW)
}