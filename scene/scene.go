// Package scene provides functions to represend a scene graph of objects.
package scene

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jnb666/go3d/mesh"
)

// Combined transformation matrix and scaling vector
type Transform struct {
	mgl32.Mat4
	Scale mgl32.Vec3
}

func NewTransform(m mgl32.Mat4) Transform {
	return Transform{Mat4: m, Scale: mgl32.Vec3{1, 1, 1}}
}

// Object is the base abstract interface type for something which can be added to the scene
type Object interface {
	Do(trans Transform, f func(*Item, Transform))
	Scale(scaleX, scaleY, scaleZ float32) Object
	Translate(trX, trY, trZ float32) Object
	Rotate(angle float32, axis mgl32.Vec3) Object
	RotateX(angle float32) Object
	RotateY(angle float32) Object
	RotateZ(angle float32) Object
	Clone() Object
	Enabled() bool
	Enable(on bool) Object
	SetMaterial(mtl mesh.Material) Object
}

// Group type represents a set of objects, it implements the Object interface
type Group struct {
	Transform
	objects []Object
	enabled bool
}

// NewGroup function creates a new empty object container
func NewGroup() *Group {
	g := new(Group)
	g.Transform = NewTransform(mgl32.Ident4())
	g.objects = []Object{}
	g.enabled = true
	return g
}

// Add method adds one or more objects to the group
func (g *Group) Add(obj ...Object) *Group {
	g.objects = append(g.objects, obj...)
	return g
}

// Clone method returns a deep copy of the group
func (g *Group) Clone() Object {
	newg := NewGroup()
	newg.Transform = g.Transform
	newg.objects = make([]Object, len(g.objects))
	for i, obj := range g.objects {
		newg.objects[i] = obj.Clone()
	}
	return newg
}

// Do method calls the callback funcion for all items under this root
// matrix transforms are stacked based on the scene tree
func (g *Group) Do(trans Transform, fn func(*Item, Transform)) {
	if !g.enabled {
		return
	}
	newTrans := Transform{
		Mat4:  trans.Mul4(g.Transform.Mat4),
		Scale: vmul(trans.Scale, g.Transform.Scale),
	}
	for _, obj := range g.objects {
		obj.Do(newTrans, fn)
	}
}

// Update the current material associated with all the items in this group.
func (g *Group) SetMaterial(mtl mesh.Material) Object {
	for _, obj := range g.objects {
		obj.SetMaterial(mtl)
	}
	return g
}

// Scale method scales the size of the object
func (g *Group) Scale(scaleX, scaleY, scaleZ float32) Object {
	g.Transform.Mat4 = g.Mul4(mgl32.Scale3D(scaleX, scaleY, scaleZ))
	g.Transform.Scale = vmul(g.Transform.Scale, mgl32.Vec3{scaleX, scaleY, scaleZ})
	return g
}

// Translate method moves the object in world space
func (g *Group) Translate(trX, trY, trZ float32) Object {
	g.Transform.Mat4 = g.Mul4(mgl32.Translate3D(trX/g.Transform.Scale[0], trY/g.Transform.Scale[1], trZ/g.Transform.Scale[2]))
	return g
}

// Rotate method rotates the object around given axis by angle (in degrees)
func (g *Group) Rotate(degrees float32, axis mgl32.Vec3) Object {
	g.Transform.Mat4 = g.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(degrees), axis))
	return g
}

func (g *Group) RotateX(degrees float32) Object {
	g.Transform.Mat4 = g.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(degrees), mgl32.Vec3{1, 0, 0}))
	return g
}

func (g *Group) RotateY(degrees float32) Object {
	g.Transform.Mat4 = g.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(degrees), mgl32.Vec3{0, 1, 0}))
	return g
}

func (g *Group) RotateZ(degrees float32) Object {
	g.Transform.Mat4 = g.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(degrees), mgl32.Vec3{0, 0, 1}))
	return g
}

func (g *Group) Enabled() bool {
	return g.enabled
}

func (g *Group) Enable(on bool) Object {
	g.enabled = on
	return g
}

// Item type represents a single component which is added to the scene
type Item struct {
	Transform
	*mesh.Mesh
	Light    *Light
	TexScale float32
	lightMat map[bool]mesh.Material
	enabled  bool
}

// NewItem function constructs a new object with given material and identity transformation matrix.
// If mtl parameter is nil then use default emissive or point material depending on mesh type.
func NewItem(msh *mesh.Mesh) *Item {
	obj := new(Item)
	obj.Mesh = msh
	obj.Transform = NewTransform(mgl32.Ident4())
	obj.TexScale = 1
	obj.enabled = true
	return obj
}

// Get the current material associated with this item.
func (o *Item) Material() mesh.Material {
	return o.Mesh.Material()
}

// Set the current material associated with this item.
func (o *Item) SetMaterial(mtl mesh.Material) Object {
	o.Mesh.SetMaterial(mtl)
	return o
}

// Associate a point light with this item, position will be calculated dynamically
func (o *Item) Illuminate(intensity, ambient, attenuation float32) *Item {
	color := o.Mesh.Material().Color().Vec3().Mul(intensity)
	o.Light = PointLight(color, ambient, mgl32.Vec3{}, attenuation)
	o.lightMat = map[bool]mesh.Material{
		false: o.Mesh.Material(),
		true:  mesh.Emissive(),
	}
	return o
}

// Do method calls the callback function with the given view
func (o *Item) Do(trans Transform, fn func(*Item, Transform)) {
	if !o.enabled {
		return
	}
	if o.Light != nil {
		o.Mesh.SetMaterial(o.lightMat[o.Light.On])
	}
	fn(o, Transform{Mat4: trans.Mul4(o.Transform.Mat4), Scale: vmul(trans.Scale, o.Transform.Scale)})
}

// Get a copy of the item. Note that the mesh is not copied, it is a reference to the same object
func (o *Item) Clone() Object {
	item := *o
	item.Mesh = o.Mesh.Clone()
	if o.Light != nil {
		lgt := *o.Light
		item.Light = &lgt
		item.lightMat = map[bool]mesh.Material{
			false: o.Mesh.Material().Clone(),
			true:  mesh.Emissive().Clone(),
		}
	}
	return &item
}

// TextureScale method adjusts the relative texture scaling
func (o *Item) TextureScale(scale float32) Object {
	o.TexScale *= scale
	return o
}

// Scale method scales the size of the object
func (o *Item) Scale(scaleX, scaleY, scaleZ float32) Object {
	o.Transform.Mat4 = o.Mul4(mgl32.Scale3D(scaleX, scaleY, scaleZ))
	o.Transform.Scale = vmul(o.Transform.Scale, mgl32.Vec3{scaleX, scaleY, scaleZ})
	return o
}

// Translate method moves the object in world space
func (o *Item) Translate(trX, trY, trZ float32) Object {
	o.Transform.Mat4 = o.Mul4(mgl32.Translate3D(trX/o.Transform.Scale[0], trY/o.Transform.Scale[1], trZ/o.Transform.Scale[2]))
	return o
}

// Rotate method rotates the object around given axis by angle (in degrees)
func (o *Item) Rotate(degrees float32, axis mgl32.Vec3) Object {
	o.Transform.Mat4 = o.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(degrees), axis))
	return o
}

func (o *Item) RotateX(degrees float32) Object {
	o.Transform.Mat4 = o.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(degrees), mgl32.Vec3{1, 0, 0}))
	return o
}

func (o *Item) RotateY(degrees float32) Object {
	o.Transform.Mat4 = o.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(degrees), mgl32.Vec3{0, 1, 0}))
	return o
}

func (o *Item) RotateZ(degrees float32) Object {
	o.Transform.Mat4 = o.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(degrees), mgl32.Vec3{0, 0, 1}))
	return o
}

func (o *Item) Enabled() bool {
	return o.enabled
}

func (o *Item) Enable(on bool) Object {
	o.enabled = on
	return o
}

func vmul(a, b mgl32.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{a[0] * b[0], a[1] * b[1], a[2] * b[2]}
}
