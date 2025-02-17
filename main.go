package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
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
	scrY = 300
)

type LightType int

const (
	Point LightType = iota
	Ambient
	Directional
)

var viewO = vector3.New(0, 0, 0)

type Tri struct {
	v0 *vector3.Vector3
	v1 *vector3.Vector3
	v2 *vector3.Vector3

	n *vector3.Vector3
	D float64

	specExp float64
	shiny   bool

	col *vector3.Vector3
}

type Ray struct {
	O   *vector3.Vector3
	dir *vector3.Vector3
}

type Hit struct {
	tri Tri
	t   float64
	p   *vector3.Vector3
}

type Light struct {
	I   float64
	pos *vector3.Vector3
	t   LightType
	dir *vector3.Vector3
}

func (t *Tri) init() {
	d1 := t.v1.Sub(t.v0)
	d2 := t.v2.Sub(t.v0)

	// log.Println(d1.String(), d2.String())

	t.n = d1.Cross(d2).Normalize()
	// log.Println(t.n.String())

	t.D = float64(t.n.Dot(t.v0))
	// log.Println(t.D)
}

var mesh = []Tri{
	// {v0: vector3.New(-300, 700, 300), v1: vector3.New(-300, 700, -300), v2: vector3.New(300, 700, -300), specExp: 5, shiny: true},
	// {v0: vector3.New(-300, 700, 300), v1: vector3.New(300, 700, -300), v2: vector3.New(300, 700, 300), shiny: false},

	{v0: vector3.New(-300, 400, -250), v1: vector3.New(300, 400, -250), v2: vector3.New(-300, 1000, -250), shiny: false, col: vector3.New(230, 20, 0)},
	{v0: vector3.New(-300, 1000, -250), v1: vector3.New(300, 400, -250), v2: vector3.New(300, 1000, -250), shiny: false, col: vector3.New(20, 230, 0)},

	{v0: vector3.New(-100, 800, -250), v1: vector3.New(-100, 600, -250), v2: vector3.New(-100, 700, 0), shiny: false, col: vector3.New(0, 0, 255)},
	{v2: vector3.New(-100, 800, -250), v1: vector3.New(-100, 600, -250), v0: vector3.New(-100, 700, 0), shiny: false, col: vector3.New(0, 0, 255)},

	{v0: vector3.New(100, 600, -250), v1: vector3.New(100, 800, -250), v2: vector3.New(100, 700, 0), shiny: false, col: vector3.New(0, 0, 255)},
	{v2: vector3.New(100, 600, -250), v1: vector3.New(100, 800, -250), v0: vector3.New(100, 700, 0), shiny: false, col: vector3.New(0, 0, 255)},
	// {v0: vector3.New(300-20, 700-20, -200), v1: vector3.New(300+20, 700-20, -200), v2: vector3.New(300-20, 700+20, -200), shiny: false},
	// {v0: vector3.New(300-20, 700+20, -200), v1: vector3.New(300+20, 700-20, -200), v2: vector3.New(300+20, 700+20, -200), shiny: false},
}

var lights = []Light{
	// {t: Ambient, I: (0.1)},
	// {t: Point, I: (0.4), pos: vector3.New(-400, 600, 0)},
	// {t: Point, I: (0.4), pos: vector3.New(-400, 600, 0)},
	{t: Point, I: (0.7), pos: vector3.New(0, 700, -100)},
	// {t: Point, I: (0.4), pos: vector3.New(0, 700, -100)},
	{t: Ambient, I: (0.3)},

	// {t: Point, I: (0.8), pos: vector3.New(400, 400, 0)},,
}

var resultImage = image.NewRGBA(image.Rect(0, 0, scrW, scrH))

func init() {
	// Lock OS thread to the main thread (necessary for OpenGL context)
	runtime.LockOSThread()
}

func main() {

	for i := range mesh {
		mesh[i].init()
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

	angle := float64(0)

	window.SetKeyCallback(keyCB)
	// Render loop
	for !window.ShouldClose() {
		angle += 3
		radAngle := angle / 180 * math.Pi
		lights[0].pos.X = 300 * math.Cos(radAngle)
		lights[0].pos.Y = 700 + 300*math.Sin(radAngle)
		lights[0].pos.Z = -210

		// lights[1].pos.X = 300 * math.Cos(2*radAngle)
		// lights[1].pos.Y = 700 + 300*math.Sin(2*radAngle)
		// lights[1].pos.Z = -210

		// mesh[0].col.X += 10
		// mesh[1].col.X += 10

		// lights[1].pos.Z += 10

		// mesh[2] = Tri{v0: vector3.New(lights[0].pos.X-20, lights[0].pos.Y-20, -200), v1: vector3.New(lights[0].pos.X+20, lights[0].pos.Y-20, -200), v2: vector3.New(lights[0].pos.X-20, lights[0].pos.Y+20, -200), shiny: false}
		// mesh[3] = Tri{v0: vector3.New(lights[0].pos.X-20, lights[0].pos.Y+20, -200), v1: vector3.New(lights[0].pos.X+20, lights[0].pos.Y-20, -200), v2: vector3.New(lights[0].pos.X+20, lights[0].pos.Y+20, -200), shiny: false}

		// mesh[2].init()
		// mesh[3].init()

		// log.Println(lights[0].pos.String())
		resultImage = image.NewRGBA(image.Rect(0, 0, scrW, scrH))
		raytrace()
		texture, _ := loadTexture()
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// Bind texture and draw quad
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, texture)

		gl.UseProgram(shaderProgram)
		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)

		window.SwapBuffers()
		glfw.PollEvents()
	}

	// Cleanup
	gl.DeleteVertexArrays(1, &vao)
	gl.DeleteBuffers(1, &vbo)
	gl.DeleteProgram(shaderProgram)
}

func keyCB(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		if key == glfw.KeyW {
			lights[1].dir.X -= 1
		} else if key == glfw.KeyS {
			mesh[2].v2.Z -= 10
		}
	}
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

	img := resultImage

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

func raytrace() {
	s := time.Now()
	for sy := 0; sy < scrH; sy++ {
		for sx := 0; sx < scrW; sx++ {
			scrWorld := vector3.New(float64(-scrW/2+sx)+viewO.X, scrY+viewO.Y, float64(-scrH/2+sy)+viewO.Z)
			dir := (scrWorld.Sub(viewO)).Normalize()
			// log.Println(dir.String())
			// R = t · dir
			ray := Ray{O: viewO, dir: dir}

			// s1 := time.Now()

			hitRecord := getHits(ray, false, 0)

			// hitRecord := make([]Hit, 0)

			if len(hitRecord) <= 0 {
				setRes(sx, sy, 0, 0, 0)
				continue
			}

			// log.Println(hitRecord)

			minT := math.Inf(1)
			var firstHit Hit

			for _, hit := range hitRecord {
				if hit.t < minT {
					minT = hit.t
					firstHit = hit
				}
			}

			// fmt.Printf("%+v\n", firstHit)

			// angleIncidence := Angle(firstHit.p.Sub(viewO), firstHit.tri.n)
			// log.Println(angleIncidence)
			// c := int(math.Max(0, angleIncidence-2.1) / (math.Pi - 2.1) * 255)
			i := getIntensity(firstHit.tri, ray.dir.MulScalar(firstHit.t).Add(viewO))
			var col *vector3.Vector3
			if firstHit.tri.col != nil {
				col = firstHit.tri.col
			} else {
				col = vector3.New(100, 100, 100)
			}
			col = col.MulScalar(i)
			setRes(sx, sy, int(col.X), int(col.Y), int(col.Z))

			// log.Println(time.Since(s1))

			// break
		}
		// break
	}
	log.Println(time.Since(s))
}

func getHits(ray Ray, breakHit bool, maxT float64) []Hit {
	// log.Printf("%+v\n", ray.dir.String())
	var hitRecord []Hit

	for _, tri := range mesh {

		if backface(ray, tri) {
			continue
		}
		// log.Printf("%+v\n", tri)
		t, p := intersect(tri, ray)
		// log.Println(t)
		// log.Println(p.String())
		if t <= 0 {
			continue
		} else {
			// log.Println("Marker")
			if inTri(tri, p) {
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
					hitRecord = append(hitRecord, Hit{tri: tri, t: t, p: p})
				}
				// break
				// setRes(sx, sy, 255, 255, 255)
				// log.Println("GOOD")
			}
		}
	}

	return hitRecord
}

func backface(r Ray, t Tri) bool {
	return r.dir.Dot(t.n) > 0
}

func intersect(t Tri, r Ray) (float64, *vector3.Vector3) {
	// log.Println(t)
	denom := r.dir.Dot(t.n)

	if math.Abs(denom) > 0.000001 {
		t := (t.n.Dot(t.v0.Sub(r.O))) / denom
		p := r.O.Add(r.dir.MulScalar(t))

		return t, p
	}

	return -1, &vector3.Vector3{}
}

func inTri(t Tri, p *vector3.Vector3) bool {
	// log.Println(p.String())
	// log.Println(t.n.Dot(t.v1.Sub(t.v0).Cross(p.Sub(t.v0))))
	// log.Println("N", t.n.String())
	v0v1 := t.v1.Sub(t.v0)
	// log.Println(v0v1.String())
	v1v2 := t.v2.Sub(t.v1)
	// log.Println(v1v2.String())
	v2v0 := t.v0.Sub(t.v2)
	// log.Println(v2v0.String())

	v0p := p.Sub(t.v0)
	// log.Println(v0p.String())
	v1p := p.Sub(t.v1)
	// log.Println(v1p.String())
	v2p := p.Sub(t.v2)
	// log.Println(v2p.String())

	// log.Println(t.n.Dot(v0v1.Cross(v0p)), t.n.Dot(v1v2.Cross(v1p)), t.n.Dot(v2v0.Cross(v2p)))
	// log.Println(t.n.Dot(v0v1.Cross(v0p)))
	// log.Println(t.n.Dot(v1v2.Cross(v1p)))
	// log.Println(t.n.Dot(v2v0.Cross(v2p)))

	epsilon := 0.001
	if t.n.Dot(v0v1.Cross(v0p)) < -epsilon {
		// log.Println("r1")
		return false
	}

	if t.n.Dot(v1v2.Cross(v1p)) < -epsilon {
		// log.Println("r2")
		return false
	}

	if t.n.Dot(v2v0.Cross(v2p)) < -epsilon {
		// log.Println("r3")
		return false
	}

	return true
}

func Angle(a, b *vector3.Vector3) float64 {
	return math.Acos((a.Dot(b)) / (a.Magnitude() * b.Magnitude()))
}

func getIntensity(tri Tri, ray *vector3.Vector3) float64 {
	sum := float64(0)

	P := ray.Add(viewO)
	if !tri.shiny {
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

				dot := tri.n.Dot(pl)

				if dot > 0 {
					shadowResult := checkShadow(P, pl, maxT)
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

				rayReflection := tri.n.MulScalar(2 * tri.n.Dot(pl)).Sub(pl)
				dotted := rayView.Dot(rayReflection)
				if dotted > 0 {
					if !checkShadow(P, pl, maxT) {
						cosAlpha := dotted / (rayView.Magnitude() * rayReflection.Magnitude())

						sum += l.I * math.Pow(cosAlpha, tri.specExp)
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
