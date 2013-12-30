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

layout(std430, binding = 1) buffer octree {
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
uniform int widthPix, heightPix;

in vec3 cameraDir;

vec4 colorAtBrickLoc(ivec3 pos) {
	return texelFetch(voxelBlocks, pos.zyx, 0);
}

ivec3 nodeBrick(uint node) {
	uint loc = nodes[node].colorData;
	return ivec3((loc >> 20) & COORD_MASK, (loc >> 10) & COORD_MASK, loc & COORD_MASK);
}

vec4 cAlong(vec3 start, vec3 dir) {
	vec4 c = vec4(0, 0, 0, 0);

	// Search through the octree.
	dir = normalize(dir);
	vec3 dInv = 1 / dir;
	vec3 deltaT = abs(dInv);

	ivec3 vMask = ivec3(INT(dir.x <= 0), INT(dir.y <= 0), INT(dir.z <= 0));
	ivec3 stepDir = ivec3(sign(dir));

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
		if ((nodes[nextNode].childData & FINAL_MASK) != 0) {
			vec3 nextXMax = nextXMin + dim;
			sMin = (nextXMin - start) * dInv;
			sMax = (nextXMax - start) * dInv;

			if ((nodes[nextNode].childData & SOLID_MASK) == 0) {
				/*
				 * Look at a leaf node.
				 */

				sLMax = MAX_COMP(PICK_BY(vMask, sMax, sMin));
				pos = start + dir * (sLMax + .0001) - nextXMin;

				vec3 fLinePos = pos * 8 / dim;
				ivec3 linePos = ivec3(fLinePos);
				vec3 stepTmax = (vec3(linePos + stepDir) - fLinePos) * dInv;

				ivec3 brickLoc = nodeBrick(nextNode);

				while (true) {
					vec4 checkColor = colorAtBrickLoc(brickLoc + linePos);
					if (checkColor.a != 0) {
						c += vec4(checkColor.rgb, 0) * (1.0 - c.a);
						c.a += checkColor.a;
						if (c.a >= 1)
							return c;
					}
					if (stepTmax.x < stepTmax.y && stepTmax.x < stepTmax.z) {
							linePos.x += stepDir.x;
							if (linePos.x < 0 || linePos.x >= 8)
								break;
							stepTmax.x += deltaT.x;
					} else if (stepTmax.y < stepTmax.z) {
							linePos.y += stepDir.y;
							if (linePos.y < 0 || linePos.y >= 8)
								break;
							stepTmax.y += deltaT.y;
					} else {
							linePos.z += stepDir.z;
							if (linePos.z < 0 || linePos.z >= 8)
								break;
							stepTmax.z += deltaT.z;
					}
				}
			}
			/*
			 * Prepare for next node to check.
			 */
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
	vec4 actualColor = cAlong(cameraPos, cDir);
	color = actualColor.rgb * actualColor.a;
}