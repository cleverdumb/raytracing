package main

// import (
// 	"fmt"
// 	"image"
// 	"image/color"
// 	"image/draw"
// 	"log"
// 	"math"
// 	"time"

// 	// "github.com/faiface/gui/win"
// 	// "github.com/faiface/mainthread"

// 	"github.com/deeean/go-vector/vector3"
// 	// "github.com/faiface/gui/win"
// 	// "github.com/faiface/mainthread"

// 	"github.com/go-gl/gl/v4.1-core/gl"
// 	"github.com/go-gl/glfw/v3.3/glfw"
// )

// const (
// 	scrW = 800
// 	scrH = 800
// 	scrY = 500
// )

// var vertexShaderSource = `
// #version 410 core
// layout (location = 0) in vec2 aPos;
// layout (location = 1) in vec2 aTexCoord;

// out vec2 TexCoord;

// void main() {
//     gl_Position = vec4(aPos, 0.0, 1.0);
//     TexCoord = aTexCoord;
// }
// ` + "\x00"

// var fragmentShaderSource = `
// #version 410 core
// in vec2 TexCoord;
// out vec4 FragColor;

// uniform sampler2D texture1;

// void main() {
//     FragColor = texture(texture1, TexCoord);
// }
// ` + "\x00"

// type LightType int

// const (
// 	Point LightType = iota
// 	Ambient
// 	Directional
// )

// var viewO = vector3.New(0, 0, 0)

// type Tri struct {
// 	v0 *vector3.Vector3
// 	v1 *vector3.Vector3
// 	v2 *vector3.Vector3

// 	n *vector3.Vector3
// 	D float64
// }

// type Ray struct {
// 	O   *vector3.Vector3
// 	dir *vector3.Vector3
// }

// type Hit struct {
// 	tri Tri
// 	t   float64
// 	p   *vector3.Vector3
// }

// type Light struct {
// 	I   float64
// 	pos *vector3.Vector3
// 	t   LightType
// 	dir *vector3.Vector3
// }

// func (t *Tri) init() {
// 	d1 := t.v1.Sub(t.v0)
// 	d2 := t.v2.Sub(t.v0)

// 	// log.Println(d1.String(), d2.String())

// 	t.n = d1.Cross(d2).Normalize()
// 	// log.Println(t.n.String())

// 	t.D = float64(t.n.Dot(t.v0))
// 	// log.Println(t.D)
// }

// var mesh = []Tri{
// 	{v0: vector3.New(-300, 700, 300), v1: vector3.New(-300, 700, -300), v2: vector3.New(300, 700, -300)},
// }

// var lights = []Light{
// 	// {t: Ambient, I: (0.1)},
// 	{t: Point, I: (0.4), pos: vector3.New(-400, 400, 0)},
// 	// {t: Point, I: (0.8), pos: vector3.New(400, 400, 0)},
// 	// {t: Directional, I: (0.4), dir: vector3.New(10, 1, 0)},
// }

// var resultImage = image.NewRGBA(image.Rect(0, 0, scrW, scrH))

// // func run() {
// // log.Println(vector3.New(2, 3, 4).Sub(vector3.New(1, 0, 5)).String())

// // log.Println(inTri(mesh[0], vector3.New(-200, 700, 0)))
// // render()
// // }

// func init() {
// 	// Initialize GLFW
// 	if err := glfw.Init(); err != nil {
// 		log.Fatalln("failed to initialize glfw:", err)
// 	}
// }

// func main() {
// 	// defer glfw.Terminate()

// 	for i := range mesh {
// 		mesh[i].init()
// 	}
// 	raytrace()

// 	defer glfw.Terminate()

// 	// Create a GLFW window
// 	glfw.WindowHint(glfw.ContextVersionMajor, 4)
// 	glfw.WindowHint(glfw.ContextVersionMinor, 1)
// 	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
// 	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

// 	window, err := glfw.CreateWindow(800, 800, "GLFW Image Renderer", nil, nil)
// 	if err != nil {
// 		log.Fatalln("failed to create window:", err)
// 	}
// 	window.MakeContextCurrent()

// 	// Initialize OpenGL
// 	if err := gl.Init(); err != nil {
// 		log.Fatalln("failed to initialize OpenGL:", err)
// 	}

// 	// Compile shaders and create a shader program
// 	shaderProgram := createShaderProgram()

// 	// Define a full-screen quad with texture coordinates
// 	vertices := []float32{
// 		// Positions    // TexCoords
// 		-1.0, 1.0, 0.0, 1.0, // Top-left
// 		-1.0, -1.0, 0.0, 0.0, // Bottom-left
// 		1.0, -1.0, 1.0, 0.0, // Bottom-right
// 		1.0, 1.0, 1.0, 1.0, // Top-right
// 	}
// 	indices := []uint32{
// 		0, 1, 2, // First triangle
// 		2, 3, 0, // Second triangle
// 	}

// 	// Create and bind a VAO, VBO, and EBO
// 	var VAO, VBO, EBO uint32
// 	gl.GenVertexArrays(1, &VAO)
// 	gl.GenBuffers(1, &VBO)
// 	gl.GenBuffers(1, &EBO)

// 	gl.BindVertexArray(VAO)

// 	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
// 	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

// 	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
// 	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

// 	// Define vertex attributes
// 	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
// 	gl.EnableVertexAttribArray(0)

// 	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
// 	gl.EnableVertexAttribArray(1)

// 	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
// 	gl.BindVertexArray(0)

// 	// Load and bind texture
// 	img := resultImage // Assume image is already an image.Image object
// 	texture, err := createTextureFromImage(img)
// 	if err != nil {
// 		log.Fatalln("failed to create texture:", err)
// 	}

// 	// Main loop
// 	for !window.ShouldClose() {
// 		gl.Clear(gl.COLOR_BUFFER_BIT)

// 		// Render the image as a texture on the full-screen quad
// 		gl.UseProgram(shaderProgram)
// 		gl.BindTexture(gl.TEXTURE_2D, texture)
// 		gl.BindVertexArray(VAO)
// 		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, gl.PtrOffset(0))

// 		window.SwapBuffers()
// 		glfw.PollEvents()
// 	}
// }

// func createShaderProgram() uint32 {
// 	// Compile vertex shader
// 	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
// 	csource, free := gl.Strs(vertexShaderSource)
// 	gl.ShaderSource(vertexShader, 1, csource, nil)
// 	free()
// 	gl.CompileShader(vertexShader)

// 	// Check for errors
// 	var success int32
// 	gl.GetShaderiv(vertexShader, gl.COMPILE_STATUS, &success)
// 	if success == gl.FALSE {
// 		var logLength int32
// 		gl.GetShaderiv(vertexShader, gl.INFO_LOG_LENGTH, &logLength)

// 		log := make([]byte, logLength)
// 		gl.GetShaderInfoLog(vertexShader, logLength, nil, &log[0])

// 		fmt.Printf("Vertex Shader Compilation Error: %s\n", log)
// 	}

// 	// Compile fragment shader
// 	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
// 	csource, free = gl.Strs(fragmentShaderSource)
// 	gl.ShaderSource(fragmentShader, 1, csource, nil)
// 	free()
// 	gl.CompileShader(fragmentShader)

// 	// Check for errors
// 	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &success)
// 	if success == gl.FALSE {
// 		var logLength int32
// 		gl.GetShaderiv(fragmentShader, gl.INFO_LOG_LENGTH, &logLength)

// 		log := make([]byte, logLength)
// 		gl.GetShaderInfoLog(fragmentShader, logLength, nil, &log[0])

// 		fmt.Printf("Fragment Shader Compilation Error: %s\n", log)
// 	}

// 	// Link shaders into a program
// 	shaderProgram := gl.CreateProgram()
// 	gl.AttachShader(shaderProgram, vertexShader)
// 	gl.AttachShader(shaderProgram, fragmentShader)
// 	gl.LinkProgram(shaderProgram)

// 	// Check for errors
// 	gl.GetProgramiv(shaderProgram, gl.LINK_STATUS, &success)
// 	if success == gl.FALSE {
// 		var logLength int32
// 		gl.GetProgramiv(shaderProgram, gl.INFO_LOG_LENGTH, &logLength)

// 		log := make([]byte, logLength)
// 		gl.GetProgramInfoLog(shaderProgram, logLength, nil, &log[0])

// 		fmt.Printf("Shader Program Linking Error: %s\n", log)
// 	}

// 	// Clean up shaders
// 	gl.DeleteShader(vertexShader)
// 	gl.DeleteShader(fragmentShader)

// 	return shaderProgram
// }

// func createTextureFromImage(img image.Image) (uint32, error) {
// 	rgba := image.NewRGBA(img.Bounds())
// 	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

// 	var texture uint32
// 	gl.GenTextures(1, &texture)
// 	gl.BindTexture(gl.TEXTURE_2D, texture)

// 	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(rgba.Bounds().Dx()), int32(rgba.Bounds().Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
// 	gl.GenerateMipmap(gl.TEXTURE_2D)

// 	return texture, nil
// }

// func setRes(x, y int, r, g, b int) {
// 	resultImage.Set(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), 1})
// }

// func raytrace() {
// 	s := time.Now()
// 	for sy := 0; sy < scrH; sy++ {
// 		for sx := 0; sx < scrW; sx++ {
// 			scrWorld := vector3.New(float64(-scrW/2+sx), scrY, float64(-scrH/2+sy))
// 			dir := (scrWorld.Sub(viewO)).Normalize()
// 			// log.Println(dir.String())
// 			// R = t · dir
// 			ray := Ray{O: viewO, dir: dir}

// 			// s1 := time.Now()

// 			var hitRecord []Hit

// 			for _, tri := range mesh {
// 				// log.Printf("%+v\n", tri)
// 				t, p := intersect(tri, ray)
// 				// log.Println(p.String())
// 				if t <= 0 {
// 					continue
// 				} else {
// 					// log.Println("Marker")
// 					if inTri(tri, p) {
// 						// c := int((t - 700) / 30 * 255)
// 						// angleIncidence := Angle(p.Sub(viewO), tri.n)
// 						// log.Println(angleIncidence)
// 						// c := int(math.Max(0, angleIncidence-2.1) / (math.Pi - 2.1) * 255)
// 						// setRes(sx, sy, c, c, c)

// 						// log.Println("MARKER")
// 						hitRecord = append(hitRecord, Hit{tri: tri, t: t, p: p})

// 						// break
// 						// setRes(sx, sy, 255, 255, 255)
// 						// log.Println("GOOD")
// 					}
// 				}
// 			}

// 			if len(hitRecord) <= 0 {
// 				setRes(sx, scrH-sy-1, 0, 0, 0)
// 				continue
// 			}

// 			// log.Println(hitRecord)

// 			minT := math.Inf(1)
// 			var firstHit Hit

// 			for _, hit := range hitRecord {
// 				if hit.t < minT {
// 					minT = hit.t
// 					firstHit = hit
// 				}
// 			}

// 			// fmt.Printf("%+v\n", firstHit)

// 			// angleIncidence := Angle(firstHit.p.Sub(viewO), firstHit.tri.n)
// 			// log.Println(angleIncidence)
// 			// c := int(math.Max(0, angleIncidence-2.1) / (math.Pi - 2.1) * 255)
// 			c := getIntensity(firstHit.tri, firstHit.p)
// 			col := int(c * 255)
// 			setRes(sx, scrH-sy-1, col, col, col)

// 			// log.Println(time.Since(s1))

// 			// break
// 		}
// 		// break
// 	}
// 	log.Println(time.Since(s))
// }

// func intersect(t Tri, r Ray) (float64, *vector3.Vector3) {
// 	// log.Println(t)
// 	denom := r.dir.Dot(t.n)

// 	if denom != 0 {
// 		t := (t.n.Dot(r.O) + t.D) / denom
// 		p := r.O.Add(r.dir.MulScalar(t))

// 		return t, p
// 	}

// 	return -1, &vector3.Vector3{}
// }

// func inTri(t Tri, p *vector3.Vector3) bool {
// 	// log.Println(p.String())
// 	// log.Println(t.n.Dot(t.v1.Sub(t.v0).Cross(p.Sub(t.v0))))
// 	// log.Println("N", t.n.String())
// 	v0v1 := t.v1.Sub(t.v0)
// 	// log.Println(v0v1.String())
// 	v1v2 := t.v2.Sub(t.v1)
// 	// log.Println(v1v2.String())
// 	v2v0 := t.v0.Sub(t.v2)
// 	// log.Println(v2v0.String())

// 	v0p := p.Sub(t.v0)
// 	// log.Println(v0p.String())
// 	v1p := p.Sub(t.v1)
// 	// log.Println(v1p.String())
// 	v2p := p.Sub(t.v2)
// 	// log.Println(v2p.String())

// 	// log.Println(t.n.Dot(v0v1.Cross(v0p)), t.n.Dot(v1v2.Cross(v1p)), t.n.Dot(v2v0.Cross(v2p)))
// 	// log.Println(t.n.Dot(v0v1.Cross(v0p)))
// 	// log.Println(t.n.Dot(v1v2.Cross(v1p)))
// 	// log.Println(t.n.Dot(v2v0.Cross(v2p)))

// 	if t.n.Dot(v0v1.Cross(v0p)) < 0 {
// 		// log.Println("r1")
// 		return false
// 	}

// 	if t.n.Dot(v1v2.Cross(v1p)) < 0 {
// 		// log.Println("r2")
// 		return false
// 	}

// 	if t.n.Dot(v2v0.Cross(v2p)) < 0 {
// 		// log.Println("r3")
// 		return false
// 	}

// 	return true
// }

// func Angle(a, b *vector3.Vector3) float64 {
// 	return math.Acos((a.Dot(b)) / (a.Magnitude() * b.Magnitude()))
// }

// func getIntensity(tri Tri, p *vector3.Vector3) float64 {
// 	sum := float64(0)

// 	for _, l := range lights {
// 		if l.t == Ambient {
// 			sum += l.I
// 		} else {
// 			var pl *vector3.Vector3
// 			if l.t == Point {
// 				// fmt.Printf("%+v\n", l)
// 				// log.Println(p.String())
// 				pl = l.pos.Sub(p)
// 			} else {
// 				pl = l.dir.MulScalar(-1)
// 			}

// 			dot := tri.n.Dot(pl)

// 			if dot > 0 {
// 				sum += l.I * dot / pl.Magnitude()
// 			}
// 		}
// 	}

// 	return sum
// }
