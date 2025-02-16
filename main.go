package main

import (
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"time"

	"github.com/faiface/gui/win"
	"github.com/faiface/mainthread"

	"github.com/deeean/go-vector/vector3"
)

const (
	scrW = 800
	scrH = 800
	scrY = 500
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
}

type Ray struct {
	O   *vector3.Vector3
	dir *vector3.Vector3
}

type Hit struct {
	tri    Tri
	t      float64
	p      *vector3.Vector3
	sx, sy int
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
	{v0: vector3.New(-300, 700, 300), v1: vector3.New(-300, 700, -300), v2: vector3.New(300, 700, -300)},
}

var lights = []Light{
	// {t: Ambient, I: (0.1)},
	{t: Point, I: (0.4), pos: vector3.New(-400, 400, 0)},
	// {t: Point, I: (0.8), pos: vector3.New(400, 400, 0)},
	// {t: Directional, I: (0.4), dir: vector3.New(10, 1, 0)},
}

var resultImage = image.NewRGBA(image.Rect(0, 0, scrW, scrH))

func run() {
	// log.Println(vector3.New(2, 3, 4).Sub(vector3.New(1, 0, 5)).String())
	for i := range mesh {
		mesh[i].init()
	}
	raytrace()
	// log.Println(inTri(mesh[0], vector3.New(-200, 700, 0)))
	render()
}

func main() {
	mainthread.Run(run)
	// raytrace()
	// render()
}

func render() {
	// resultImage = image.NewRGBA(image.Rect(0, 0, scrW, scrH))
	// raytrace()
	w, err := win.New(win.Title("raytracer"), win.Size(scrW, scrH))
	if err != nil {
		panic(err)
	}

	w.Draw() <- func(drw draw.Image) image.Rectangle {
		draw.Draw(drw, image.Rect(0, 0, scrW, scrH), resultImage, image.ZP, draw.Src)
		return image.Rect(0, 0, scrW, scrH)
	}

	for event := range w.Events() {
		switch event.(type) {
		case win.WiClose:
			close(w.Draw())
		}
	}
}

func setRes(x, y int, r, g, b int) {
	resultImage.Set(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), 1})
}

func raytrace() {
	s := time.Now()
	for sy := 0; sy < scrH; sy++ {
		for sx := 0; sx < scrW; sx++ {
			scrWorld := vector3.New(float64(-scrW/2+sx), scrY, float64(-scrH/2+sy))
			dir := (scrWorld.Sub(viewO)).Normalize()
			// log.Println(dir.String())
			// R = t · dir
			ray := Ray{O: viewO, dir: dir}

			// s1 := time.Now()

			var hitRecord []Hit

			for _, tri := range mesh {
				// log.Printf("%+v\n", tri)
				t, p := intersect(tri, ray)
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
						hitRecord = append(hitRecord, Hit{tri: tri, t: t, p: p, sx: sx, sy: sy})

						// break
						// setRes(sx, sy, 255, 255, 255)
						// log.Println("GOOD")
					}
				}
			}

			if len(hitRecord) <= 0 {
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
			c := getIntensity(firstHit.tri, firstHit.p)
			col := int(c * 255)
			setRes(firstHit.sx, scrH-firstHit.sy-1, col, col, col)

			// log.Println(time.Since(s1))

			// break
		}
		// break
	}
	log.Println(time.Since(s))
}

func intersect(t Tri, r Ray) (float64, *vector3.Vector3) {
	// log.Println(t)
	denom := r.dir.Dot(t.n)

	if denom != 0 {
		t := (t.n.Dot(r.O) + t.D) / denom
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

	if t.n.Dot(v0v1.Cross(v0p)) < 0 {
		// log.Println("r1")
		return false
	}

	if t.n.Dot(v1v2.Cross(v1p)) < 0 {
		// log.Println("r2")
		return false
	}

	if t.n.Dot(v2v0.Cross(v2p)) < 0 {
		// log.Println("r3")
		return false
	}

	return true
}

func Angle(a, b *vector3.Vector3) float64 {
	return math.Acos((a.Dot(b)) / (a.Magnitude() * b.Magnitude()))
}

func getIntensity(tri Tri, p *vector3.Vector3) float64 {
	sum := float64(0)

	for _, l := range lights {
		if l.t == Ambient {
			sum += l.I
		} else {
			var pl *vector3.Vector3
			if l.t == Point {
				pl = l.pos.Sub(p)
			} else {
				pl = l.dir.MulScalar(-1)
			}

			dot := tri.n.Dot(pl)

			if dot > 0 {
				sum += l.I * dot / pl.Magnitude()
			}
		}
	}

	return sum
}
