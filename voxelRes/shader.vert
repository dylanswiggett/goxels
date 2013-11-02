#version 330 core
layout(location = 0) in vec3 vertexPosition_modelspace;
layout(location = 1) in vec3 vertexColor;
layout(location = 2) in vec3 vertexNormal_modelspace;

uniform int time;
uniform mat4 M;
uniform mat4 VP;

uniform vec3 lightPos;

out vec3 normal;
out float lightLevel;
out vec3 lightDir;
out vec3 inColor;

void main() {
	vec4 v = vec4(vertexPosition_modelspace, 1);
	vec4 worldV = M * v;
	gl_Position = VP * worldV;

	vec4 worldN = M * vec4(vertexNormal_modelspace, 1);
	vec4 world0 = M * vec4(0, 0, 0, 1);
	normal = normalize((worldN - world0).xyz);

	lightDir = lightPos - worldV.xyz;
	lightLevel = clamp(100 / (length(lightDir) * length(lightDir)), 0, 1);

	inColor = vertexColor;
}