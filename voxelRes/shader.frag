#version 430 core

#define LEAF_MASK  uint(0x80000000)
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
uniform vec3 lightPos;

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
	while ((nodes[node].childData & LEAF_MASK) == uint(0) && (maxDepth--) > 0) {
		ivec3 childLoc = ivec3(pos / scale);
		int childOffset = childLoc.x * 4 + childLoc.y * 2 + childLoc.z;
		pos = pos - scale * childLoc;
		node = (nodes[node].childData & CHILD_MASK) + uint(childOffset);
		scale = scale / 2;
	}
	ivec3 nodeBrickLoc = nodeBrick(node);
	nodeBrickLoc += ivec3(vec3(4, 4, 4) * pos / scale);
	// return vec4(float(10 - maxDepth) / 10.0, float(node) / 100, 0, 1);
	// return vec4(0, float(node) / 20000.0, 0, 1);
	// return vec4(float(10 - maxDepth) / 10.0, 0, 0, 1);
	// return vec4(nodeBrickLoc.xyz, 0);
	return colorAtBrickLoc(nodeBrickLoc);
}

void main(){
	float x = float((time) % 100) / 10.0;
	float y = clamp(gl_FragCoord.x / 120.0, 0, 10);
	float z = clamp(gl_FragCoord.y / 100.0, 0, 10);
	// float col = float(nodes[clamp(int((x + y + z) * 5), 0, 1000)].colorData) / 100;
	// float col = float(nodes[time].colorData);
	// float col = nodes[0].colorData;
	// color = vec3(col, col, col);
	color = colorAtLoc(vec3(x, y, z)).rgb;
	// color = vec3(nodeBrick(uint(0))) / 200.0;
	// color = colorAtBrickLoc(ivec3(int(x), int(y), int(z))).rgb;
}