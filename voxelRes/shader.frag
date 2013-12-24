#version 430 core

#define MAX_COMP(v) max(max(v.x, v.y), v.z)
#define MIN_COMP(v) min(min(v.x, v.y), v.z)

#define PICK_BY(choice, v1, v2) vec3(choice.x == 1 ? v1.x : v2.x, choice.y == 1 ? v1.y : v2.y, choice.z == 1 ? v1.z : v2.z)

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

struct octreeNodeParser {
	bool isGenerated, doneSearching;
	int node;
	float sLMax, sUMin;
	vec3 sMin, sMid, sMax;
	vec3 xMin, xMid, xMax;
	bvec3 childMask;
	bvec3 lastMask;
	bvec3 maskList[3];
	int currentMask;
};

// octreeNodeParser nodeList[MAX_DEPTH];

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
	// Search through the octree.
	// Credit for algorithm: http://diglib.eg.org/EG/DL/WS/EGGH/EGGH89/061-073.pdf

	octreeNodeParser nodeList[MAX_DEPTH];

	dir = normalize(dir);

	vec3 dInv = 1 / dir;

	ivec3 vMask = ivec3(dir.x > 0 ? 0 : 1, dir.y > 0 ? 0 : 1, dir.z > 0 ? 0 : 1);

	nodeList[0].xMin = vec3(0, 0, 0);
	nodeList[0].xMax = worldSize;
	nodeList[0].sMin = (nodeList[0].xMin - start) * dInv;
	nodeList[0].sMax = (nodeList[0].xMax - start) * dInv;
	nodeList[0].node = 0;
	nodeList[0].isGenerated = false;
	nodeList[0].doneSearching = false;

	int i = 0;
	// int maxSteps = 30;
	while (i >= 0) {//} && --maxSteps > 0) {
		if (nodeList[i].doneSearching) i--;
		else if (nodeList[i].isGenerated) {
			/*
			 * Find next intersecting node.
			 */
			ivec3 nextChildDisplacement = ivec3(nodeList[i].childMask) ^ vMask;
			if (nodeList[i].childMask == nodeList[i].lastMask) {
				nodeList[i].doneSearching = true;
			} else {
				int max = 1;	// THIS DOES SOMETHING. I don't know what, but
								// without it the program breaks...
				while(any(nodeList[i].maskList[nodeList[i].currentMask] &&
					nodeList[i].childMask) && bool(max)) // breaks here w/o check.
					nodeList[i].currentMask++;
				nodeList[i].childMask = nodeList[i].childMask ||
										nodeList[i].maskList[nodeList[i].currentMask];
			}

			/*
			 * Prepare to search next intersecting node.
			 */
			int nextNode = nextChildDisplacement.x * 4 +
						   nextChildDisplacement.y * 2 +
						   nextChildDisplacement.z * 1 +
						   int(nodes[nodeList[i].node].childData & CHILD_MASK);
			// Decide if node is a child or solid or neither.
			if ((nodes[nextNode].childData & FINAL_MASK) == uint(0)) {
				ivec3 disp = nextChildDisplacement;
				nodeList[i + 1].node = nextNode;
				nodeList[i + 1].xMin = PICK_BY(disp, nodeList[i].xMid, nodeList[i].xMin);
				nodeList[i + 1].xMax = PICK_BY(disp, nodeList[i].xMax, nodeList[i].xMid);
				nodeList[i + 1].sMin = PICK_BY(disp, nodeList[i].sMid, nodeList[i].sMin);
				nodeList[i + 1].sMax = PICK_BY(disp, nodeList[i].sMax, nodeList[i].sMid);
				nodeList[i + 1].isGenerated = false;
				nodeList[i + 1].doneSearching = false;
				i++;
			} else if ((nodes[nextNode].childData & SOLID_MASK) != 0)
				continue;
			else {	// LEAF NODE
				return vec4(nextChildDisplacement, 1);
				// // return vec4(1, 1, 1, 1);
				// while (nodeList[i].currentMask == 0)
				// 	i--;
				// return vec4(nodeList[i].maskList[nodeList[i].currentMask - 1], 1);
				// return colorAtBrickLoc(nodeBrick(nextNode));
			}
		} else {
			/*
			 * Generate data for a new node.
			 */
			nodeList[i].xMid = (nodeList[i].xMin + nodeList[i].xMax) / 2;
			nodeList[i].sMid = (nodeList[i].xMid - start) * dInv;
			vec3 lowerLimits = PICK_BY(vMask, nodeList[i].sMax, nodeList[i].sMin);
			vec3 upperLimits = PICK_BY(vMask, nodeList[i].sMin, nodeList[i].sMax);
			nodeList[i].sLMax = MAX_COMP(lowerLimits);
			nodeList[i].sUMin = MIN_COMP(upperLimits);
			if (nodeList[i].sLMax >= nodeList[i].sUMin || nodeList[i].sUMin < 0) {
				i--;
				continue;
			}

			// Generate masks.
			bool a = nodeList[i].sMid.x < nodeList[i].sMid.y;
			bool b = nodeList[i].sMid.x < nodeList[i].sMid.z;
			bool c = nodeList[i].sMid.y < nodeList[i].sMid.z;
			nodeList[i].maskList[0].x = a && b;
			nodeList[i].maskList[0].y = !a && c;
			nodeList[i].maskList[0].z = !(b || c);
			nodeList[i].maskList[1].x = a != b;
			nodeList[i].maskList[1].y = a == c;
			nodeList[i].maskList[1].z = b != c;
			nodeList[i].maskList[2].x = !(a || b);
			nodeList[i].maskList[2].y = a && !c;
			nodeList[i].maskList[2].z = b && c;
			nodeList[i].childMask.x   = nodeList[i].sMid.x < nodeList[i].sLMax;
			nodeList[i].childMask.y   = nodeList[i].sMid.y < nodeList[i].sLMax;
			nodeList[i].childMask.z   = nodeList[i].sMid.z < nodeList[i].sLMax;
			nodeList[i].lastMask.x    = nodeList[i].sMid.x < nodeList[i].sUMin;
			nodeList[i].lastMask.y    = nodeList[i].sMid.y < nodeList[i].sUMin;
			nodeList[i].lastMask.z    = nodeList[i].sMid.z < nodeList[i].sUMin;
			nodeList[i].currentMask   = 0;
			nodeList[i].isGenerated   = true;
		}
	}
	return vec4(0, 0, 0, 1);
}

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