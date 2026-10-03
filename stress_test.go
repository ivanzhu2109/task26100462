package orthrouter

import (
	"math/rand"
	"testing"
)

func TestStressSeeds(t *testing.T) {
	for seed := int64(1); seed <= 30; seed++ {
		rng := rand.New(rand.NewSource(seed))
		for trial := 0; trial < 200; trial++ {
			maxC := 5 + rng.Intn(9)
			n := 1 + rng.Intn(5)
			rects := make([]Rect, 0, n)
			for i := 0; i < n; i++ {
				x1 := rng.Intn(maxC)
				y1 := rng.Intn(maxC)
				x2 := x1 + 1 + rng.Intn(maxC-x1)
				y2 := y1 + 1 + rng.Intn(maxC-y1)
				rects = append(rects, Rect{x1, y1, x2, y2})
			}
			r := NewRouter(0, 0, maxC, maxC, rects, 1+rng.Intn(30))
			req, ok := randomRequest(rng, rects)
			if !ok {
				continue
			}
			p, found := r.Route(req)
			best, bFound := bruteOptimal(r, req)
			if found != bFound || (found && p.Cost != best) {
				t.Fatalf("seed=%d trial=%d: 库(%v,%d) 暴力(%v,%d)\nrects=%v req=%+v path=%v",
					seed, trial, found, p.Cost, bFound, best, rects, req, p.Points)
			}
			if found {
				validatePath(t, r, req, p)
			}
		}
	}
}
