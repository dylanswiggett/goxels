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

void main(){
	color = cameraDir;
}
