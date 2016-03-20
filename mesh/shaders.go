package mesh

const MaxLights = 4

var vertexShader = `
attribute vec3 position;
attribute vec3 normal;
attribute vec2 texcoord;

varying vec3 Normal;
varying vec3 CameraSpacePos;
varying vec2 Texcoord;
varying vec3 ModelPos;

uniform mat4 cameraToClip;
uniform mat4 modelToCamera;
uniform mat3 normalModelToCamera;
uniform vec3 modelScale;
uniform float texScale;

void main() {
	vec4 pos = modelToCamera * vec4(position, 1.0);
	gl_Position = cameraToClip * pos;
	Normal = normalize(normalModelToCamera * normal);
	CameraSpacePos = pos.xyz;
	Texcoord = texcoord * texScale;
	ModelPos = position * modelScale * texScale;
}
`

var vertexShaderPoints = `
attribute vec3 position;

varying vec3 CameraSpacePos;
varying vec2 PointLocation;

uniform mat4 cameraToClip;
uniform mat4 modelToCamera;
uniform vec3 pointLocation;
uniform vec2 viewport;

void main() {
	gl_Position = cameraToClip * modelToCamera * vec4(position, 1.0);
	vec4 pointClip = cameraToClip * vec4(pointLocation, 1.0);
	vec3 pointNdc = pointClip.xyz / pointClip.w;
	PointLocation = viewport * (pointNdc.xy + 1.0) / 2.0;
}
`

var fragShaderHead = `
varying vec3 Normal;
varying vec3 CameraSpacePos;
varying vec2 Texcoord;
varying vec3 ModelPos;

#define MAX_LIGHTS 4
#define GAMMA 2.2

uniform vec4 objectColor;
uniform vec3 specularColor;
uniform float shininess;
uniform float ambientScale;
uniform int numLights;
uniform vec4 lightPos[MAX_LIGHTS];
uniform vec4 lightCol[MAX_LIGHTS];

void gammaCorrect(in vec4 color) {
	gl_FragColor = vec4(pow(color.rgb, vec3(1.0/GAMMA)), color.a);
}

// calculate attenuation and light direction in camera space
float attenuation(in float quadScale, in vec3 pos, in vec3 lgtPos, out vec3 dir) {
	vec3 diff = lgtPos - pos;
	float dist2 = dot(diff, diff);
	dir = diff * inversesqrt(dist2);
	return 1.0 / (1.0 + quadScale*dist2);
}
`

var noise3D = `
// 64 x 64 x 64 cube mapped to 64 x 4096 2D texture
#define TSIZE 64.0
#define SAMPLE(t,p)  	 ( texture2D(t,p).rgb )

// convert XYZ coordinate to mapping on 2D texture
vec2 convXY(in vec3 wpos, in vec3 offpix) {
	vec3 pix = mod(floor(wpos*TSIZE) + offpix, TSIZE);
	return vec2(pix.x/TSIZE, (pix.y+TSIZE*pix.z)/(TSIZE*TSIZE));
}

// interpolate quad values
vec3 interp(in sampler2D tex, in vec3 wpos, in float zoff, in vec2 delta) {
	vec2 p1 = convXY(wpos, vec3(0.5,0.5,zoff));
	vec2 p2 = convXY(wpos, vec3(1.5,0.5,zoff));
	vec2 p3 = convXY(wpos, vec3(0.5,1.5,zoff));
	vec2 p4 = convXY(wpos, vec3(1.5,1.5,zoff));
	vec3 avg1 = mix(SAMPLE(tex,p1), SAMPLE(tex,p2), delta.x);
	vec3 avg2 = mix(SAMPLE(tex,p3), SAMPLE(tex,p4), delta.x);
	return mix(avg1, avg2, delta.y);
}

// unwrap 3d position to position on tiles x tiles 2d texture array
vec3 sample3D(in sampler2D tex, in vec3 wpos) {
	vec3 delta = fract(TSIZE*wpos);
	vec3 avg1 = interp(tex, wpos, 0.0, delta.xy);
	vec3 avg2 =	interp(tex, wpos, 1.0, delta.xy);
	return mix(avg1, avg2, delta.z);
}

// calculate a 3d noise function from a texture packed as a 2d array as es2 does not support 3d samplers
vec3 noise3D(in sampler2D tex, in vec3 matPos, const float secondOctaveScale) {
	const float octave = 8.0;
	return sample3D(tex, matPos) + sample3D(tex, matPos*octave)*(secondOctaveScale/octave);
}
`

var diffuseLighting = `
vec3 diffuseLighting(in vec3 vertexNormal, in vec3 objColor) {
	vec3 color = vec3(0);
	vec3 norm = normalize(vertexNormal);
	for (int i = 0; i < numLights; i++) {
		vec3 lightDir;
		vec3 intensity;
		if (lightPos[i].w == 0.0) {
			// directional light
			lightDir = lightPos[i].xyz;
			intensity = lightCol[i].xyz;
		} else {
			// point light
			float att = attenuation(lightPos[i].w, CameraSpacePos, lightPos[i].xyz, lightDir);
			intensity = att * lightCol[i].xyz;
		}
		float ambient = lightCol[i].w * ambientScale;
		float diffuse = max(dot(norm, lightDir), 0.0);
		color += objColor * intensity * (ambient + diffuse);
	}
	return color;
}
`

var blinnPhongLighting = `
vec3 blinnPhongLighting(in vec3 vertexNormal, in vec3 objColor) {
	vec3 color = vec3(0);
	vec3 norm = normalize(vertexNormal);
	for (int i = 0; i < numLights; i++) {
		vec3 lightDir;
		vec3 intensity;
		if (lightPos[i].w == 0.0) {
			// directional light
			lightDir = lightPos[i].xyz;
			intensity = lightCol[i].xyz;
		} else {
			// point light
			float att = attenuation(lightPos[i].w, CameraSpacePos, lightPos[i].xyz, lightDir);
			intensity = att * lightCol[i].xyz;
		}
		// diffuse component
		float ambient = lightCol[i].w * ambientScale;		
		float diffuse = max(dot(norm, lightDir), 0.0);
		color += objColor * intensity * (ambient + diffuse);
		// specular highlight
		vec3 viewDir = normalize(-CameraSpacePos);
		vec3 halfAngle = normalize(lightDir + viewDir);
		float specular = pow(max(dot(norm, halfAngle), 0.0), shininess);
		color += specularColor * intensity * specular;
	}
	return color;
}
`

var fragmentShader = map[int]string{
	mUnshaded: fragShaderHead + `
void main() {
	gammaCorrect(objectColor);
}
`,
	mUnshadedTex: fragShaderHead + `
uniform sampler2D tex0;

void main() {
	gammaCorrect(objectColor * texture2D(tex0, Texcoord));
}
`,
	mUnshadedTexCube: fragShaderHead + `
uniform samplerCube tex0;

void main() {
	gammaCorrect(objectColor * textureCube(tex0, ModelPos));
}
`,
	mPointShader: `
varying vec2 PointLocation;
uniform vec4 objectColor;
uniform float pointSize;
uniform float ambientScale;

void main() {
	vec2 dist = PointLocation - gl_FragCoord.xy;
	if (pointSize >= 4.0 && dot(dist, dist) > pointSize*pointSize / 4.0) discard;	
	gl_FragColor = objectColor;
}
`,
	mEmissiveShader: fragShaderHead + `
void main() {
	float d = dot(normalize(-CameraSpacePos), normalize(Normal));
	d = max(pow(d*1.5,0.4)*1.1, 1.0);
	gammaCorrect(vec4(objectColor.rgb*d, 1.0));
}
`,
	mDiffuse: fragShaderHead + diffuseLighting + `
void main() {
	vec3 color = diffuseLighting(Normal, objectColor.rgb);
	gammaCorrect(vec4(color, objectColor.a));
}
`,

	mDiffuseTex: fragShaderHead + diffuseLighting + `
uniform sampler2D tex0;
void main() {
	vec4 C = texture2D(tex0, Texcoord);
	vec3 color = diffuseLighting(Normal, objectColor.rgb*C.rgb);
	gammaCorrect(vec4(color, objectColor.a*C.a));
}
`,
	mDiffuseTexCube: fragShaderHead + diffuseLighting + `
uniform samplerCube tex0;
void main() {
	vec4 C = textureCube(tex0, ModelPos);
	vec3 color = diffuseLighting(Normal, objectColor.rgb*C.rgb);
	gammaCorrect(vec4(color, objectColor.a*C.a));
}
`,
	mBlinnPhong: fragShaderHead + blinnPhongLighting + `
void main() {
	vec3 color = blinnPhongLighting(Normal, objectColor.rgb);
	gammaCorrect(vec4(color, objectColor.a));
}
`,
	mBlinnPhongTex: fragShaderHead + blinnPhongLighting + `
uniform sampler2D tex0;

void main() {
	vec4 C = texture2D(tex0, Texcoord);
	vec3 color = blinnPhongLighting(Normal, objectColor.rgb*C.rgb);
	gammaCorrect(vec4(color, objectColor.a*C.a));
}
`,
	mBlinnPhongTexCube: fragShaderHead + blinnPhongLighting + `
uniform samplerCube tex0;

void main() {
	vec4 C = textureCube(tex0, ModelPos);
	vec3 color = blinnPhongLighting(Normal, objectColor.rgb*C.rgb);
	gammaCorrect(vec4(color, objectColor.a*C.a));
}
`,
	mWoodShader: fragShaderHead + blinnPhongLighting + noise3D + `
uniform sampler2D tex0;
uniform sampler2D tex1;

void main() {
	vec2 woodPos = vec2(0.5, 0.5) - 0.85*ModelPos.zy - 0.10*ModelPos.x - 0.05*noise3D(tex1, ModelPos*0.5, 1.0).xy;
	vec3 C = texture2D(tex0, woodPos).rgb;
	vec3 color = blinnPhongLighting(Normal, objectColor.rgb*C);
	gammaCorrect(vec4(color, 1.0));
}
`,
	mRoughShader: fragShaderHead + blinnPhongLighting + noise3D + `
uniform sampler2D tex0;

void main() {
	vec3 pos = ModelPos + vec3(0.5, 0.5, 0.5);
	vec3 N2 = Normal + noise3D(tex0, pos, 1.0) * 0.4;
	vec3 color = blinnPhongLighting(N2, objectColor.rgb);
	gammaCorrect(vec4(color, objectColor.a));
}
`,
	mMarbleShader: fragShaderHead + blinnPhongLighting + noise3D + `
uniform sampler2D tex0;

void main() {
	vec3 pos = ModelPos + vec3(0.5, 0.5, 0.5);
	vec3 noise = noise3D(tex0, pos, 2.0);
	float a = 0.5 + 0.5*sin(ModelPos.y*16.0 + noise.x*10.0);
	vec3 C = mix(vec3(0.4,0.3,0.3), vec3(1.0,1.0,1.0), a);
	vec3 color = blinnPhongLighting(Normal, objectColor.rgb*C);
	gammaCorrect(vec4(color, objectColor.a));
}
`,
}
