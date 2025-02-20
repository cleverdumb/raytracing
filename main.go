package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"sync"
	"time"

	_ "image/png"
	"runtime"

	"github.com/deeean/go-vector/vector3"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	scrW = 800
	scrH = 800
	scrY = 600

	maxReflectRecursion = 10
)

type LightType int

const (
	Point LightType = iota
	Ambient
	Directional
)

var viewO = vector3.New(0, 0, 0)
var viewOMutex sync.RWMutex

var yaw = float64(0)   // clockwise, around z-axis
var pitch = float64(0) // clockwise, around x-axis
var roll = float64(0)  // clockwise, around y-axis

var rotMatrix [9]float64
var rotMatrixMutex sync.RWMutex

func makeRotationMatrix() {
	rotMatrixMutex.Lock()
	ca := math.Cos(yaw * math.Pi / 180) // alpha
	sa := math.Sin(yaw * math.Pi / 180)
	cb := math.Cos(roll * math.Pi / 180) // beta
	sb := math.Sin(roll * math.Pi / 180)
	cg := math.Cos(pitch * math.Pi / 180) // gamma
	sg := math.Sin(pitch * math.Pi / 180)
	rotMatrix[0] = ca * cb
	rotMatrix[1] = ca*sb*sg - sa*cg
	rotMatrix[2] = ca*sb*cg + sa*sg
	rotMatrix[3] = sa * cb
	rotMatrix[4] = sa*sb*sg + ca*cg
	rotMatrix[5] = sa*sb*cg - ca*sg
	rotMatrix[6] = -sb
	rotMatrix[7] = cb * sg
	rotMatrix[8] = cb * cg
	rotMatrixMutex.Unlock()
}

func applyMatrix(m [9]float64, v *vector3.Vector3) *vector3.Vector3 {
	return vector3.New(
		v.Dot(vector3.New(m[0], m[1], m[2])),
		v.Dot(vector3.New(m[3], m[4], m[5])),
		v.Dot(vector3.New(m[6], m[7], m[8])),
	)
}

type Object interface {
	normal(pos *vector3.Vector3) *vector3.Vector3
	init()
	p() ObjectProp
}

type Polygon struct {
	v              []*vector3.Vector3
	n              *vector3.Vector3
	D              float64
	specExp        float64
	shiny          bool
	reflect        bool
	reflectiveness float64
	col            *vector3.Vector3
}

type Sphere struct {
	c              *vector3.Vector3
	r              float64
	specExp        float64
	shiny          bool
	reflect        bool
	reflectiveness float64
	col            *vector3.Vector3
}

type ObjectProp struct {
	specExp        float64
	shiny          bool
	reflect        bool
	reflectiveness float64
	col            *vector3.Vector3
}

func (p Polygon) normal(_ *vector3.Vector3) *vector3.Vector3 {
	return p.n
}

func (p Sphere) normal(pos *vector3.Vector3) *vector3.Vector3 {
	return pos.Sub(p.c).Normalize()
}

func (p Polygon) p() ObjectProp {
	return ObjectProp{specExp: p.specExp, shiny: p.shiny, reflect: p.reflect, reflectiveness: p.reflectiveness, col: p.col}
}

func (p Sphere) p() ObjectProp {
	return ObjectProp{specExp: p.specExp, shiny: p.shiny, reflect: p.reflect, reflectiveness: p.reflectiveness, col: p.col}
}

type Ray struct {
	O   *vector3.Vector3
	dir *vector3.Vector3
}

type Hit struct {
	tri Object
	t   float64
	p   *vector3.Vector3
}

type Light struct {
	I   float64
	pos *vector3.Vector3
	t   LightType
	dir *vector3.Vector3
}

func (t *Polygon) init() {
	d1 := t.v[1].Sub(t.v[0])
	d2 := t.v[2].Sub(t.v[0])

	// log.Println(d1.String(), d2.String())

	t.n = d1.Cross(d2).Normalize()
	// log.Println(t.n.String())

	t.D = float64(t.n.Dot(t.v[0]))
	// log.Println(t.D)
}

func (s *Sphere) init() {}

var mesh = []Object{
	// {v0: vector3.New(-300, 700, 300), v1: vector3.New(-300, 700, -300), v2: vector3.New(300, 700, -300), specExp: 5, shiny: true},
	// {v0: vector3.New(-300, 700, 300), v1: vector3.New(300, 700, -300), v2: vector3.New(300, 700, 300), shiny: false},

	// &Polygon{
	// 	v:     []*vector3.Vector3{vector3.New(-300, 400, -250), vector3.New(300, 400, -250), vector3.New(300, 1000, -250), vector3.New(-300, 1000, -250)},
	// 	shiny: false,
	// 	col:   vector3.New(255, 0, 0),
	// },
	// &Polygon{
	// 	v:              []*vector3.Vector3{vector3.New(-200, 600, -250), vector3.New(200, 600, -250), vector3.New(200, 600, 0), vector3.New(-200, 600, 0)},
	// 	col:            vector3.New(255, 255, 255),
	// 	reflect:        true,
	// 	reflectiveness: (0.7),
	// 	shiny:          true,
	// 	specExp:        100,
	// },
	&Sphere{
		c:              vector3.New(-150, 750, 0),
		r:              200,
		col:            vector3.New(255, 0, 0),
		shiny:          false,
		specExp:        50,
		reflect:        false,
		reflectiveness: (0.5),
	},
	&Sphere{
		c:              vector3.New(300, 750, 0),
		r:              100,
		col:            vector3.New(0, 255, 0),
		shiny:          true,
		specExp:        50,
		reflect:        false,
		reflectiveness: (0.5),
	},
	// {
	// 	v:              []*vector3.Vector3{vector3.New(200, 500, -240), vector3.New(-200, 500, -240), vector3.New(0, 500, 0)},
	// 	shiny:          false,
	// 	col:            vector3.New(0, 0, 255),
	// 	reflect:        true,
	// 	reflectiveness: (0.8),
	// },
	// {v: vector3.New(-300, 1000, -250), v1: vector3.New(300, 400, -250), v2: vector3.New(300, 1000, -250), shiny: false, col: vector3.New(20, 230, 0)},

	// {v0: vector3.New(300-20, 700-20, -200), v1: vector3.New(300+20, 700-20, -200), v2: vector3.New(300-20, 700+20, -200), shiny: false},
	// {v0: vector3.New(300-20, 700+20, -200), v1: vector3.New(300+20, 700-20, -200), v2: vector3.New(300+20, 700+20, -200), shiny: false},
}

var lights = []Light{
	// {t: Ambient, I: (0.1)},
	{t: Point, I: (0), pos: vector3.New(-400, 600, 0)},
	{t: Directional, I: (0.8), dir: vector3.New(-1, 0, 0)},
	// {t: Point, I: (0.4), pos: vector3.New(-400, 600, 0)},
	// {t: Point, I: (0.4), pos: vector3.New(-250, 700, -210)},
	// {t: Point, I: (0.4), pos: vector3.New(250, 700, -210)},
	{t: Ambient, I: (0.2)},

	// {t: Point, I: (0.8), pos: vector3.New(400, 400, 0)},,
}

var resultImage = image.NewRGBA(image.Rect(0, 0, scrW, scrH))
var copyImage image.Image

func init() {
	// Lock OS thread to the main thread (necessary for OpenGL context)
	runtime.LockOSThread()
}

func main() {
	makeRotationMatrix()

	for i := range mesh {
		// if p, ok := mesh[i].(Polygon); ok {
		mesh[i].init()
		// }
	}
	raytrace()

	// log.Println(checkShadow(vector3.New(280, 700, -240), lights[1].dir.MulScalar(-1), math.Inf(1)))
	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		panic(fmt.Errorf("failed to initialize glfw: %v", err))
	}
	defer glfw.Terminate()

	// Create a windowed mode window
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(scrW, scrH, "Raytracing", nil, nil)
	if err != nil {
		panic(fmt.Errorf("failed to create window: %v", err))
	}
	window.MakeContextCurrent()

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		panic(fmt.Errorf("failed to initialize OpenGL: %v", err))
	}

	// Load and bind texture
	// texture, err := loadTexture() // Replace with your image file
	// if err != nil {
	// 	panic(fmt.Errorf("failed to load texture: %v", err))
	// }

	// Define quad vertices
	vertices := []float32{
		// Positions   // Texture Coords
		-1.0, -1.0, 0.0, 0.0,
		1.0, -1.0, 1.0, 0.0,
		-1.0, 1.0, 0.0, 1.0,
		1.0, 1.0, 1.0, 1.0,
	}

	// Create VBO and VAO
	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Define vertex attributes
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)

	// Create shader program
	shaderProgram := createShaderProgram()
	gl.UseProgram(shaderProgram)
	gl.Uniform1i(gl.GetUniformLocation(shaderProgram, gl.Str("texture1\x00")), 0)

	// angle := float64(0)

	window.SetKeyCallback(keyCB)

	s := time.Now()
	// Render loop
	for !window.ShouldClose() {
		// angle += 1
		// yaw = angle * math.Pi / 180
		// makeRotationMatrix()
		// radAngle := angle / 180 * math.Pi
		// angle2 := radAngle + math.Pi/2
		// lights[0].pos.X = 300 * math.Cos(radAngle)
		// lights[0].pos.Y = 700 + 300*math.Sin(radAngle)
		// lights[0].pos.Z = -210

		// lights[1].pos.X = 300 * math.Cos(angle2)
		// lights[1].pos.Y = 700 + 300*math.Sin(angle2)
		// lights[1].pos.Z = -210

		// mesh[0].col.X += 10
		// mesh[1].col.X += 10

		// lights[1].pos.Z += 10

		// mesh[1].reflectiveness = math.Sin(radAngle)
		// mesh[2] = Tri{v0: vector3.New(lights[0].pos.X-20, lights[0].pos.Y-20, -200), v1: vector3.New(lights[0].pos.X+20, lights[0].pos.Y-20, -200), v2: vector3.New(lights[0].pos.X-20, lights[0].pos.Y+20, -200), shiny: false}
		// mesh[3] = Tri{v0: vector3.New(lights[0].pos.X-20, lights[0].pos.Y+20, -200), v1: vector3.New(lights[0].pos.X+20, lights[0].pos.Y-20, -200), v2: vector3.New(lights[0].pos.X+20, lights[0].pos.Y+20, -200), shiny: false}

		// mesh[2].init()
		// mesh[3].init()

		doKeyEffects()

		// log.Println(lights[0].pos.String())
		// var copyResult image.Image
		sectorMutex.Lock()
		if sectorsDone >= 8 {
			log.Println(time.Since(s))
			s = time.Now()
			copyImage = resultImage
			resultImage = image.NewRGBA(image.Rect(0, 0, scrW, scrH))
			sectorsDone = 0

			raytrace()
		}
		sectorMutex.Unlock()
		texture, _ := loadTexture()
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// Bind texture and draw quad
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, texture)

		gl.UseProgram(shaderProgram)
		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)

		window.SwapBuffers()

		// raytrace()
		glfw.PollEvents()
	}

	// Cleanup
	gl.DeleteVertexArrays(1, &vao)
	gl.DeleteBuffers(1, &vbo)
	gl.DeleteProgram(shaderProgram)
}

var keyMap = make(map[glfw.Key]bool)

func keyCB(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		keyMap[key] = true
	} else if action == glfw.Release {
		keyMap[key] = false
	}
}

const viewSpd = 5
const rotSpd = 1

func doKeyEffects() {
	viewOMutex.Lock()
	if keyMap[glfw.KeyW] {
		// viewO.Y += viewSpd
		viewO = viewO.Add(applyMatrix(rotMatrix, vector3.New(0, viewSpd, 0)))
	}
	if keyMap[glfw.KeyS] {
		viewO = viewO.Add(applyMatrix(rotMatrix, vector3.New(0, -viewSpd, 0)))
	}
	if keyMap[glfw.KeyA] {
		viewO = viewO.Add(applyMatrix(rotMatrix, vector3.New(-viewSpd, 0, 0)))
	}
	if keyMap[glfw.KeyD] {
		viewO = viewO.Add(applyMatrix(rotMatrix, vector3.New(viewSpd, 0, 0)))
	}
	if keyMap[glfw.KeyQ] {
		viewO = viewO.Add(applyMatrix(rotMatrix, vector3.New(0, 0, -viewSpd)))
	}
	if keyMap[glfw.KeyE] {
		viewO = viewO.Add(applyMatrix(rotMatrix, vector3.New(0, 0, viewSpd)))
	}
	if keyMap[glfw.KeyL] {
		yaw -= rotSpd
		makeRotationMatrix()
	}
	if keyMap[glfw.KeyJ] {
		yaw += rotSpd
		makeRotationMatrix()
	}
	if keyMap[glfw.KeyK] {
		pitch -= rotSpd
		makeRotationMatrix()
	}
	if keyMap[glfw.KeyI] {
		pitch += rotSpd
		makeRotationMatrix()
	}
	if keyMap[glfw.KeyO] {
		roll += rotSpd
		makeRotationMatrix()
	}
	if keyMap[glfw.KeyU] {
		roll -= rotSpd
		makeRotationMatrix()
	}
	viewOMutex.Unlock()
}

func loadTexture() (uint32, error) {
	// file, err := os.Open(filename)
	// if err != nil {
	// 	return 0, err
	// }
	// defer file.Close()

	// img, _, err := image.Decode(file)
	// if err != nil {
	// 	return 0, err
	// }

	img := copyImage

	// Convert to RGBA
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	// Generate OpenGL texture
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(rgba.Bounds().Dx()), int32(rgba.Bounds().Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	return texture, nil
}

func createShaderProgram() uint32 {
	vertexShaderSource := `
		#version 410 core
		layout (location = 0) in vec2 aPos;
		layout (location = 1) in vec2 aTexCoord;
		out vec2 TexCoord;
		void main() {
			gl_Position = vec4(aPos, 0.0, 1.0);
			TexCoord = aTexCoord;
		}
	` + "\x00"

	fragmentShaderSource := `
		#version 410 core
		out vec4 FragColor;
		in vec2 TexCoord;
		uniform sampler2D texture1;
		void main() {
			FragColor = texture(texture1, TexCoord);
		}
	` + "\x00"

	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	cVertexSource, free := gl.Strs(vertexShaderSource)
	gl.ShaderSource(vertexShader, 1, cVertexSource, nil)
	free()
	gl.CompileShader(vertexShader)

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	cFragmentSource, free := gl.Strs(fragmentShaderSource)
	gl.ShaderSource(fragmentShader, 1, cFragmentSource, nil)
	free()
	gl.CompileShader(fragmentShader)

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program
}

func setRes(x, y int, r, g, b int) {
	resultImage.Set(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), 1})
}

var sectorsDone = 0
var sectorMutex sync.Mutex

const (
	sectionX = 4 // 4
	sectionY = 2 // 2
	sectionW = scrW / sectionX
	sectionH = scrH / sectionY
)

func raytrace() {
	// s := time.Now()
	for y := 0; y < sectionY; y++ {
		for x := 0; x < sectionX; x++ {
			go processSection(x, y)
		}
	}
	// log.Println(time.Since(s))
}

func processSection(x, y int) {
	viewOMutex.RLock()
	rotMatrixMutex.RLock()
	for dy := 0; dy < sectionH; dy++ {
		for dx := 0; dx < sectionW; dx++ {
			sx, sy := x*sectionW+dx, y*sectionH+dy
			dir := scrToWorld(sx, sy)
			dir = applyMatrix(rotMatrix, dir)
			oneRay(dir, sx, sy)
		}
	}
	sectorMutex.Lock()
	sectorsDone++
	sectorMutex.Unlock()
	viewOMutex.RUnlock()
	rotMatrixMutex.RUnlock()
	// if sectorsDone >= 8 {
	// 	// render()
	// 	sectorMutex.Lock()
	// 	sectorsDone = 0
	// 	sectorMutex.Unlock()
	// }
}

func scrToWorld(sx, sy int) *vector3.Vector3 {
	return vector3.New(float64(-scrW/2+sx), scrY, float64(-scrH/2+sy)).Normalize()
}

func oneRay(dir *vector3.Vector3, sx, sy int) {
	// log.Println("d", dir)
	ray := Ray{O: viewO, dir: dir}

	// s1 := time.Now()

	hitRecord := getHits(ray, false, 0)

	// hitRecord := make([]Hit, 0)

	if len(hitRecord) <= 0 {
		setRes(sx, sy, 0, 0, 0)
		return
	}

	// log.Println(hitRecord)

	firstHit := getClosestHit(hitRecord)

	// fmt.Printf("TT%+v\n", firstHit.tri)

	// angleIncidence := Angle(firstHit.p.Sub(viewO), firstHit.tri.n)
	// log.Println(angleIncidence)
	// c := int(math.Max(0, angleIncidence-2.1) / (math.Pi - 2.1) * 255)
	i := getIntensity(firstHit.tri, ray.dir.MulScalar(firstHit.t), viewO)
	var col *vector3.Vector3
	if firstHit.tri.p().col != nil {
		col = firstHit.tri.p().col
	} else {
		col = vector3.New(100, 100, 100)
	}
	col = col.MulScalar(i)
	// var r *vector3.Vector3
	if firstHit.tri.p().reflect {
		r := getReflection(firstHit.tri, ray.dir.MulScalar(firstHit.t), firstHit.p, maxReflectRecursion)
		col = col.MulScalar(1 - firstHit.tri.p().reflectiveness).Add(r.MulScalar(firstHit.tri.p().reflectiveness))
	}

	setRes(sx, sy, int(col.X), int(col.Y), int(col.Z))
}

func getHits(ray Ray, breakHit bool, maxT float64) []Hit {
	// log.Printf("M %v", ray.dir.String())
	var hitRecord []Hit

out:
	for _, tri := range mesh {
		if poly, ok := tri.(*Polygon); ok {
			var t float64
			var p *vector3.Vector3
			if backface(ray, *poly) {
				continue
			}

			t, p = intersect(*poly, ray)

			if t <= 0 {
				continue
			} else {
				// log.Println("Marker")
				cont := false
				if _, ok := tri.(*Polygon); ok {
					cont = inTri(*tri.(*Polygon), p)
				}
				if cont {
					// c := int((t - 700) / 30 * 255)
					// angleIncidence := Angle(p.Sub(viewO), tri.n)
					// log.Println(angleIncidence)
					// c := int(math.Max(0, angleIncidence-2.1) / (math.Pi - 2.1) * 255)
					// setRes(sx, sy, c, c, c)

					// log.Println("MARKER")

					if breakHit {
						if t < maxT && t > 0.0000001 {
							hitRecord = append(hitRecord, Hit{tri: tri, t: t, p: p})
							break
						}
					} else {
						if t > 0.000001 {
							hitRecord = append(hitRecord, Hit{tri: tri, t: t, p: p})
						}
					}
					// break
					// setRes(sx, sy, 255, 255, 255)
					// log.Println("GOOD")
				}
			}
		} else if sph, ok := tri.(*Sphere); ok {
			n, t1, t2 := intersectSph(*sph, ray)

			// log.Println(n, t1, t2)

			switch n {
			case 0:
				continue out
			case 1:
				// log.Println(n, t1, t2)
				if breakHit {
					if t1 < maxT && t1 > 0.0000001 {
						hitRecord = append(hitRecord, Hit{tri: tri, t: t1, p: ray.O.Add(ray.dir.MulScalar(t1))})
						break out
					}
				} else {
					if t1 > 0.000001 {
						hitRecord = append(hitRecord, Hit{tri: tri, t: t1, p: ray.O.Add(ray.dir.MulScalar(t1))})
					}
				}
			case 2:
				if breakHit {
					if t1 < maxT && t1 > 0.0000001 {
						hitRecord = append(hitRecord, Hit{tri: tri, t: t1, p: ray.O.Add(ray.dir.MulScalar(t1))})
						break out
					}
				} else {
					if t1 > 0.000001 {
						hitRecord = append(hitRecord, Hit{tri: tri, t: t1, p: ray.O.Add(ray.dir.MulScalar(t1))})
					}
				}

				if breakHit {
					if t2 < maxT && t2 > 0.0000001 {
						hitRecord = append(hitRecord, Hit{tri: tri, t: t2, p: ray.O.Add(ray.dir.MulScalar(t2))})
						break out
					}
				} else {
					if t2 > 0.000001 {
						hitRecord = append(hitRecord, Hit{tri: tri, t: t2, p: ray.O.Add(ray.dir.MulScalar(t2))})
					}
				}
			}
		}
		// log.Printf("%+v\n", tri)
		// log.Println(t)
		// log.Println(p.String())
	}

	return hitRecord
}

func backface(r Ray, t Polygon) bool {
	// log.Println("N", r.dir.String())
	// log.Println("NORM", t.n.String())
	return r.dir.Dot(t.n) > 0
}

func intersect(t Polygon, r Ray) (float64, *vector3.Vector3) {
	// log.Println(t)
	denom := r.dir.Dot(t.n)

	if math.Abs(denom) > 0.000001 {
		t := (t.n.Dot(t.v[0].Sub(r.O))) / denom
		p := r.O.Add(r.dir.MulScalar(t))

		return t, p
	}

	return -1, &vector3.Vector3{}
}

// @ returns (number of intersections, t1, t2)
func intersectSph(s Sphere, r Ray) (int, float64, float64) {
	// A - C
	AC := r.O.Sub(s.c)
	a := r.dir.Dot(r.dir)
	b := 2 * AC.Dot(r.dir)
	c := AC.Dot(AC) - s.r*s.r

	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return 0, -1, -1
	} else if discriminant == 0 {
		return 1, -b / (2 * a), 0
	} else {
		return 2, (-b + math.Sqrt(discriminant)) / (2 * a), (-b - math.Sqrt(discriminant)) / (2 * a)
	}

	// return 0, -1, -1
}

const epsilon = 0.001

func inTri(t Polygon, p *vector3.Vector3) bool {
	max := len(t.v)
	for i, v := range t.v {
		edge := t.v[(i+1)%max].Sub(v)
		toP := p.Sub(v)

		if t.n.Dot(edge.Cross(toP)) < epsilon {
			return false
		}
	}

	return true
}

func Angle(a, b *vector3.Vector3) float64 {
	return math.Acos((a.Dot(b)) / (a.Magnitude() * b.Magnitude()))
}

func getIntensity(tri Object, ray *vector3.Vector3, O *vector3.Vector3) float64 {
	sum := float64(0)

	P := ray.Add(O)
	if !tri.p().shiny {
		for _, l := range lights {
			if l.t == Ambient {
				sum += l.I
			} else {
				var pl *vector3.Vector3
				var maxT float64
				if l.t == Point {
					// L - P <=> L - (R + O)
					pl = l.pos.Sub(P)
					maxT = 1
				} else {
					pl = l.dir.MulScalar(-1)
					maxT = math.Inf(1)
				}

				dot := tri.normal(P).Dot(pl)

				if dot > 0 {
					shadowResult := checkShadow(P.Add(tri.normal(P).MulScalar(0.00001)), pl, maxT)
					// shadowResult := false
					// if shadowResult {
					// 	log.Println("shadow detected")
					// }
					if !shadowResult {
						sum += l.I * dot / pl.Magnitude()
					}
				}
			}
		}
	} else {
		// ray angle to normal
		rayView := ray.MulScalar(-1)

		for _, l := range lights {
			if l.t == Ambient {
				sum += l.I
			} else {
				var pl *vector3.Vector3
				var maxT float64
				if l.t == Point {
					// fmt.Printf("%+v\n", l)
					// log.Println(p.String())
					pl = l.pos.Sub(P)
					maxT = 1
				} else {
					pl = l.dir.MulScalar(-1)
					maxT = math.Inf(1)
				}

				rayReflection := reflect(pl, tri.normal(P))
				dotted := rayView.Dot(rayReflection)
				if dotted > 0 {
					if !checkShadow(P.Add(tri.normal(P).MulScalar(0.00001)), pl, maxT) {
						cosAlpha := dotted / (rayView.Magnitude() * rayReflection.Magnitude())

						sum += l.I * math.Pow(cosAlpha, tri.p().specExp)
					}
				}
			}
		}
	}

	return sum
}

func checkShadow(p *vector3.Vector3, dir *vector3.Vector3, maxT float64) bool {
	hitRecord := getHits(Ray{O: p, dir: dir}, true, maxT)

	// log.Println(hitRecord)
	// if len(hitRecord) > 0 {
	// 	fmt.Printf("%+v\n", hitRecord[0].tri.col.String())
	// }
	return len(hitRecord) > 0
}

func getClosestHit(record []Hit) Hit {
	minT := math.Inf(1)
	var firstHit Hit

	for _, hit := range record {
		if hit.t < minT {
			minT = hit.t
			firstHit = hit
		}
	}

	return firstHit
}

func reflect(in *vector3.Vector3, line *vector3.Vector3) *vector3.Vector3 {
	return line.MulScalar(2 * line.Dot(in)).Sub(in)
}

func getReflection(tri Object, ray *vector3.Vector3, O *vector3.Vector3, layer int) *vector3.Vector3 {
	// P := O
	rayReflection := reflect(ray.MulScalar(-1), tri.normal(O))
	// log.Println(rayReflection.String())
	hits := getHits(Ray{O: O, dir: rayReflection}, false, -1)

	if len(hits) <= 0 {
		// log.Println("No reflection")
		return vector3.New(0, 0, 0)
	}

	// log.Println("REFLECTION")

	firstHit := getClosestHit(hits)

	i := getIntensity(firstHit.tri, rayReflection.MulScalar(firstHit.t), O)

	var col *vector3.Vector3
	if firstHit.tri.p().col != nil {
		col = firstHit.tri.p().col
	} else {
		col = vector3.New(100, 100, 100)
	}
	col = col.MulScalar(i)

	if firstHit.tri.p().reflect && layer > 0 {
		r := getReflection(firstHit.tri, rayReflection.MulScalar(firstHit.t), firstHit.p, layer-1)
		col = col.MulScalar(1 - firstHit.tri.p().reflectiveness).Add(r.MulScalar(firstHit.tri.p().reflectiveness))
	}

	return col

	// return vector3.New(0, 0, 255)
}
