#version 330 core

struct octreeNode {
	int childData;
	int colorData;
};

layout(std140) uniform octree {
	octreeNode nodes[];
};

// Nodes in the octree reference locations here for actual voxel data.
// Actually a huge number of 8x8x8 voxel blocks, packed into a single texture.
uniform sampler3D voxelBlocks;

out vec3 color;

uniform int time;
uniform vec3 cameraPos;
uniform vec3 lightPos;

in vec3 cameraDir;

vec4 colorAtBrickLoc(ivec3 pos) {
	return texelFetch(voxelBlocks, pos, 0);
}

void main(){
	color = colorAtBrickLoc(ivec3((time / 5) % 160,gl_FragCoord.x / 4,gl_FragCoord.y / 3)).rgb;
//	if (color.x == 0 && color.y == 0 && color.z == 0) {
//		color = vec3(1, 0, 0);
//	}
}