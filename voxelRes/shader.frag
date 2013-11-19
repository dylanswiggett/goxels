#version 330 core

struct octreeNode {
	int childData;
	int colorData;
};

layout(std140) uniform octree {
	octreeNode nodes[];
};

out vec3 color;

uniform int time;
uniform vec3 cameraPos;
uniform vec3 lightPos;

in vec3 cameraDir;

void main(){
	color = cameraDir;
}
