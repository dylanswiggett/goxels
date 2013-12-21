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
	int maxDepth = 10;
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

void main(){
	// Calculate the vector for this fragment into the screen
	float xDisp = (gl_FragCoord.x - float(widthPix) / 2) / widthPix;
	float yDisp = (gl_FragCoord.y - float(heightPix) / 2) / heightPix;
	vec3 cDir = cameraForwards + xDisp * cameraRight - yDisp * cameraUp;
	// float x = float((time) % 100) / 10.0;
	// float y = clamp(gl_FragCoord.x / 100.0, 0, 10);
	// float z = clamp(gl_FragCoord.y / 100.0, 0, 10);
	float i;
	for (i = 0; i < 20; i += .1) {
		vec3 dir = cameraPos + cDir * i;
		if (dir.x > 0 && dir.x < 10 && dir.y > 0 && dir.y < 10 && dir.z > 0 && dir.z < 10) {
			vec3 colorThere = colorAtLoc(dir).rgb;
			if (!(colorThere.r == 0 && colorThere.g == 0 && colorThere.b == 0)) {
				color = colorThere;
				break;
			} 
		}
	}
}