#version 330 core
out vec3 color;

uniform int time;

uniform vec3 cameraPos;

in vec3 normal;
in float lightLevel;
in vec3 lightDir;
in vec3 inColor;
in vec3 worldPos;

smooth in vec3 isEdge;

void main(){
	float c = clamp(dot(normal, normalize(lightDir)), 0, 1) * .8 + .2;
	c = c * lightLevel;

	if (isEdge.x > 0)
		color = 0 * inColor;
	else if (c < .4)
		color = inColor * .2;
	else if (c < .5)
		color = inColor * c * .5;
	else if (c < .7)
		color = inColor * c * .6;
	else
		color = inColor * c;

	/*
	if (normalLevel < .1)
		color = 0 * inColor;
	else
		color = inColor * c * c;
		*/
}
