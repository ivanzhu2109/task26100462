package orthrouter

import (
	"math/rand"
	"testing"
)

// 批量请求：各条独立求解，结果与逐条调用一致。
func TestRouteAll(t *testing.T) {
	rects := []Rect{{2, 2, 4, 4}, {6, 6, 8, 8}}
	r := NewRouter(0, 0, 10, 10, rects, 20)
	reqs := []Request{
		{Source: Endpoint{0, Point{2, 3}}, Target: Endpoint{1, Point{6, 7}}},
		{Source: Endpoint{0, Point{3, 2}}, Target: Endpoint{1, Point{7, 6}}},
		{Source: Endpoint{0, Point{4, 3}}, Target: Endpoint{0, Point{3, 4}}},
	}
	results := r.RouteAll(reqs)
	if len(results) != len(reqs) {
		t.Fatalf("结果数量不符: %d", len(results))
	}
	for i, req := range reqs {
		p, ok := r.Route(req)
		if ok != results[i].OK || (ok && p.Cost != results[i].Path.Cost) {
			t.Fatalf("第 %d 条批量结果与单条不一致", i)
		}
		if ok {
			validatePath(t, r, req, p)
		}
	}
}

// 大画布、多节点下的可运行性与合法性（默认转弯代价 20）。
func TestLargeCanvas(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	const n = 100
	rects := make([]Rect, 0, n)
	for i := 0; i < n; i++ {
		x1 := rng.Intn(9900)
		y1 := rng.Intn(9900)
		rects = append(rects, Rect{x1, y1, x1 + 10 + rng.Intn(80), y1 + 10 + rng.Intn(80)})
	}
	r := NewRouter(0, 0, 10000, 10000, rects, 0) // 默认转弯代价
	var reqs []Request
	for i := 0; i+1 < n; i += 2 {
		a, b := rects[i], rects[i+1]
		reqs = append(reqs, Request{
			Source: Endpoint{i, Point{(a.X1 + a.X2) / 2, a.Y2}},
			Target: Endpoint{i + 1, Point{(b.X1 + b.X2) / 2, b.Y1}},
		})
	}
	results := r.RouteAll(reqs)
	okCount := 0
	for i, res := range results {
		if !res.OK {
			continue
		}
		okCount++
		validatePath(t, r, reqs[i], res.Path)
	}
	t.Logf("大画布: %d/%d 可达", okCount, len(reqs))
}
