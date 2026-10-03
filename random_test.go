package orthrouter

import (
	"math/rand"
	"testing"
)

// 随机小画布上，用独立暴力解核对可达性与最优代价，
// 并校验返回路径的完整合法性（含整段检查与方向约束）。
func TestRandomCrossCheck(t *testing.T) {
	rng := rand.New(rand.NewSource(20261004))
	for trial := 0; trial < 400; trial++ {
		maxC := 6 + rng.Intn(7) // 画布 [0, 6..12]
		n := 1 + rng.Intn(4)
		rects := make([]Rect, 0, n)
		for i := 0; i < n; i++ {
			x1 := rng.Intn(maxC)
			y1 := rng.Intn(maxC)
			x2 := x1 + 1 + rng.Intn(maxC-x1)
			y2 := y1 + 1 + rng.Intn(maxC-y1)
			rects = append(rects, Rect{x1, y1, x2, y2})
		}
		r := NewRouter(0, 0, maxC, maxC, rects, 20)
		req, ok := randomRequest(rng, rects)
		if !ok {
			continue
		}
		p, found := r.Route(req)
		best, bFound := bruteOptimal(r, req)
		if found != bFound {
			t.Fatalf("trial %d: 可达性不符 库=%v 暴力=%v\nrects=%v req=%+v",
				trial, found, bFound, rects, req)
		}
		if !found {
			continue
		}
		if p.Cost != best {
			t.Fatalf("trial %d: 最优代价不符 库=%d 暴力=%d\nrects=%v req=%+v path=%v",
				trial, p.Cost, best, rects, req, p.Points)
		}
		validatePath(t, r, req, p)
	}
}

// randomRequest 随机选取两个（可能同节点的）边上连接点。
func randomRequest(rng *rand.Rand, rects []Rect) (Request, bool) {
	pick := func() (Endpoint, bool) {
		// 随机尝试若干次找到带整数内部点的边。
		for try := 0; try < 50; try++ {
			ni := rng.Intn(len(rects))
			rc := rects[ni]
			var cands []Point
			if rc.Y2-rc.Y1 >= 2 {
				for y := rc.Y1 + 1; y < rc.Y2; y++ {
					cands = append(cands, Point{rc.X1, y}, Point{rc.X2, y})
				}
			}
			if rc.X2-rc.X1 >= 2 {
				for x := rc.X1 + 1; x < rc.X2; x++ {
					cands = append(cands, Point{x, rc.Y1}, Point{x, rc.Y2})
				}
			}
			if len(cands) > 0 {
				return Endpoint{Node: ni, Point: cands[rng.Intn(len(cands))]}, true
			}
		}
		return Endpoint{}, false
	}
	s, ok1 := pick()
	t, ok2 := pick()
	if !ok1 || !ok2 || s.Point == t.Point {
		return Request{}, false
	}
	return Request{Source: s, Target: t}, true
}

// 高密度重叠矩形下的随机对照。
func TestRandomDenseOverlap(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for trial := 0; trial < 400; trial++ {
		maxC := 8
		n := 3 + rng.Intn(4)
		rects := make([]Rect, 0, n)
		for i := 0; i < n; i++ {
			x1 := rng.Intn(maxC - 1)
			y1 := rng.Intn(maxC - 1)
			x2 := x1 + 1 + rng.Intn(maxC-x1)
			y2 := y1 + 1 + rng.Intn(maxC-y1)
			rects = append(rects, Rect{x1, y1, x2, y2})
		}
		r := NewRouter(0, 0, maxC, maxC, rects, 20)
		req, ok := randomRequest(rng, rects)
		if !ok {
			continue
		}
		p, found := r.Route(req)
		best, bFound := bruteOptimal(r, req)
		if found != bFound {
			t.Fatalf("trial %d: 可达性不符 库=%v 暴力=%v\nrects=%v req=%+v",
				trial, found, bFound, rects, req)
		}
		if !found {
			continue
		}
		if p.Cost != best {
			t.Fatalf("trial %d: 最优代价不符 库=%d 暴力=%d\nrects=%v req=%+v path=%v",
				trial, p.Cost, best, rects, req, p.Points)
		}
		validatePath(t, r, req, p)
	}
}
