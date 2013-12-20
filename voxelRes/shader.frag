#version 430 core

#define LEAF_MASK  uint(0x80000000)
#define CHILD_MASK uint(0x3FFFFFFF)
#define COORD_1 uint(0x3FF00000)
#define COORD_2 uint(0x000FFC00)
#define COORD_3 uint(0x000003FF)

struct octreeNode {
	uint colorData;
	uint childData;
};

layout(std430, binding = 0) buffer octree {
	// int test;
	octreeNode nodes [];
};

// Nodes in the octree reference locations here for actual voxel data.
// Actually a huge number of 8x8x8 voxel blocks, packed into a single texture.
uniform sampler3D voxelBlocks;
uniform vec3 worldSize;

out vec3 color;

uniform int time;
uniform vec3 cameraPos;
uniform vec3 lightPos;

in vec3 cameraDir;

vec4 colorAtBrickLoc(ivec3 pos) {
	return texelFetch(voxelBlocks, pos, 0);
}

ivec3 nodeBrick(uint node) {
	uint loc = nodes[node].colorData;
	return ivec3((loc & COORD_1) >> 20, (loc & COORD_2) >> 10, loc & COORD_3);
}

vec4 colorAtLoc(vec3 pos) {
	uint node = uint(0);
	vec3 scale = worldSize / 2;
	int maxDepth = 100;
	while ((nodes[node].childData & LEAF_MASK) == uint(0) && (maxDepth--) > 0) {
		vec3 childLoc = pos / scale;
		int childOffset = int(childLoc.x) * 4 + int(childLoc.y) * 2 + int(childLoc.z);
		pos = pos - scale * childLoc;
		node = (nodes[node].childData & CHILD_MASK) + uint(childOffset);
		scale = scale / 2;
	}
	ivec3 nodeBrickLoc = nodeBrick(node);
	// return vec4(float(10 - maxDepth) / 10.0, float(node) / 100, 0, 1);
	// return vec4(nodeBrickLoc.xyz, 0);
	return colorAtBrickLoc(nodeBrickLoc);
}

void main(){
	float x = float((time) % 100) / 10.0;
	float y = clamp(gl_FragCoord.x / 120.0, 0, 10);
	float z = clamp(gl_FragCoord.y / 100.0, 0, 10);
	float col = float(nodes[clamp(int(x + y + z), 0, 1000)].colorData) / 100;
	// float col = float(nodes[time].colorData);
	// float col = test;
	color = vec3(col, col, col);
	// color = colorAtLoc(vec3(x, y, z)).rgb;
	// color = vec3(nodeBrick(uint(0))) / 200.0;
}