#version 330 core
layout(location = 0) in vec3 vertexPosition_modelspace;
layout(location = 1) in vec3 vertexColor;
layout(location = 2) in vec3 vertexNormal_modelspace;

uniform int time;
uniform mat4 M;
uniform mat4 VP;
uniform vec3 cameraPos;
uniform vec3 lightPos;

out vec3 cameraDir;

void main() {
	vec4 v = vec4(vertexPosition_modelspace, 1);
	vec4 worldV = M * v;
	gl_Position = VP * worldV;

	cameraDir = worldV.xyz - cameraPos;
}