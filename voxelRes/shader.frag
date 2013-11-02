#version 330 core
out vec3 color;

uniform int time;

in vec3 normal;
in float lightLevel;
in vec3 lightDir;
in vec3 inColor;
in vec3 worldPos;

void main(){
	float c = clamp(dot(normal, normalize(lightDir)), 0, 1) * .8 + .2;
	c = c * lightLevel;
	color = inColor * c;
}