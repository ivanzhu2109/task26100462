package orthroute

import (
	"math/rand"
	"testing"
)

// Denser obstacle stress: many small rects on a small grid.
func TestDenseStress(t *testing.T) {
	const max = 20
	rng := rand.New(rand.NewSource(7))
	for i := 0; i < 3000; i++ {
		n := 3 + rng.Intn(5)
		rects := make([]Rect, 0, n+2)
		for j := 0; j < n; j++ {
			x1 := rng.Intn(max - 3)
			y1 := rng.Intn(max - 3)
			x2 := x1 + 2 + rng.Intn(max-1-x1)
			y2 := y1 + 2 + rng.Intn(max-1-y1)
			rects = append(rects, Rect{Min: Point{x1, y1}, Max: Point{x2, y2}})
		}
		rects = append(rects, sealingWalls(max)...)
		l := &Layout{Rects: rects}
		randPort := func() Port {
			node := rng.Intn(n)
			r := rects[node]
			var sides []int
			if r.Max.Y-r.Min.Y >= 2 {
				sides = append(sides, 0, 1)
			}
			if r.Max.X-r.Min.X >= 2 {
				sides = append(sides, 2, 3)
			}
			var p Point
			switch sides[rng.Intn(len(sides))] {
			case 0:
				p = Point{r.Min.X, r.Min.Y + 1 + rng.Intn(r.Max.Y-r.Min.Y-1)}
			case 1:
				p = Point{r.Max.X, r.Min.Y + 1 + rng.Intn(r.Max.Y-r.Min.Y-1)}
			case 2:
				p = Point{r.Min.X + 1 + rng.Intn(r.Max.X-r.Min.X-1), r.Max.Y}
			default:
				p = Point{r.Min.X + 1 + rng.Intn(r.Max.X-r.Min.X-1), r.Min.Y}
			}
			return Port{Node: node, At: p}
		}
		req := Request{Source: randPort(), Target: randPort()}
		for req.Source.At == req.Target.At {
			req.Target = randPort()
		}
		wantCost, wantOK := bruteCost(l, req, max)
		path, err := l.Route(req)
		if wantOK != (err == nil) {
			t.Fatalf("case %d: reachability mismatch: lib err=%v, brute ok=%v\nrects=%v\nreq=%+v",
				i, err, wantOK, l.Rects, req)
		}
		if !wantOK {
			continue
		}
		validatePath(t, l, req, path)
		if got := path.Cost(); got != wantCost {
			t.Fatalf("case %d: cost=%d, brute optimal=%d\nrects=%v\nreq=%+v\npath=%v",
				i, got, wantCost, l.Rects, req, path)
		}
	}
}
