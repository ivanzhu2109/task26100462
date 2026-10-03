package orthroute

import (
	"container/heap"
	"math/rand"
	"os"
	"strconv"
	"testing"
)

// --- independent reference implementation (full-grid Dijkstra) ---

// sealingWalls confines routes to the [0,max+1] square so the brute-force
// reference can run on a small finite grid. The walls span the whole
// canvas; their boundary lines at max+1 are walkable (and may be part of
// an optimal route), but everything beyond is sealed off by the wall
// interior, and the far strips can never improve an optimal cost.
func sealingWalls(max int) []Rect {
	return []Rect{
		{Min: Point{CanvasMin, max + 1}, Max: Point{CanvasMax, max + 2}},
		{Min: Point{max + 1, CanvasMin}, Max: Point{max + 2, CanvasMax}},
	}
}

type bruteState struct {
	x, y int
	d    dir
}

type bruteItem struct {
	s    bruteState
	cost int
}

type brutePQ []bruteItem

func (q brutePQ) Len() int            { return len(q) }
func (q brutePQ) Less(i, j int) bool  { return q[i].cost < q[j].cost }
func (q brutePQ) Swap(i, j int)       { q[i], q[j] = q[j], q[i] }
func (q *brutePQ) Push(x interface{}) { *q = append(*q, x.(bruteItem)) }
func (q *brutePQ) Pop() interface{} {
	old := *q
	it := old[len(old)-1]
	*q = old[:len(old)-1]
	return it
}

// bruteCost is the reference solver: Dijkstra over every integer point of
// the [0,max] grid with unit steps. Returns (cost, reachable).
func bruteCost(l *Layout, req Request, max int) (int, bool) {
	srcSide, ok1 := portSide(l.Rects[req.Source.Node], req.Source.At)
	tgtSide, ok2 := portSide(l.Rects[req.Target.Node], req.Target.At)
	if !ok1 || !ok2 {
		return 0, false
	}
	startDir := outwardNormal(srcSide)
	arriveDir := outwardNormal(tgtSide).opposite()

	blocked := func(a, b Point) bool {
		for _, r := range l.Rects {
			if segmentBlocked(r, a, b) {
				return true
			}
		}
		return false
	}
	inside := func(p Point) bool {
		return p.X >= 0 && p.X <= max+1 && p.Y >= 0 && p.Y <= max+1
	}

	dist := map[bruteState]int{}
	pq := &brutePQ{}
	// First step must follow the outward normal.
	step := req.Source.At
	step.X += startDir.delta().X
	step.Y += startDir.delta().Y
	if !inside(step) || blocked(req.Source.At, step) {
		return 0, false
	}
	start := bruteState{step.X, step.Y, startDir}
	dist[start] = 1
	heap.Push(pq, bruteItem{start, 1})

	for pq.Len() > 0 {
		it := heap.Pop(pq).(bruteItem)
		if d, seen := dist[it.s]; seen && it.cost > d {
			continue
		}
		if it.s.x == req.Target.At.X && it.s.y == req.Target.At.Y && it.s.d == arriveDir {
			return it.cost, true
		}
		for _, nd := range []dir{dRight, dUp, dLeft, dDown} {
			if nd == it.s.d.opposite() {
				continue
			}
			nx, ny := it.s.x+nd.delta().X, it.s.y+nd.delta().Y
			np := Point{nx, ny}
			if !inside(np) || blocked(Point{it.s.x, it.s.y}, np) {
				continue
			}
			cost := it.cost + 1
			if nd != it.s.d {
				cost += BendCost
			}
			ns := bruteState{nx, ny, nd}
			if d, seen := dist[ns]; !seen || cost < d {
				dist[ns] = cost
				heap.Push(pq, bruteItem{ns, cost})
			}
		}
	}
	return 0, false
}

// bruteLex minimizes (length, turns) lexicographically, ignoring the bend
// cost during optimization. It models "the shortest legal path".
func bruteLex(l *Layout, req Request, max int) (length, turns int, ok bool) {
	srcSide, ok1 := portSide(l.Rects[req.Source.Node], req.Source.At)
	tgtSide, ok2 := portSide(l.Rects[req.Target.Node], req.Target.At)
	if !ok1 || !ok2 {
		return 0, 0, false
	}
	startDir := outwardNormal(srcSide)
	arriveDir := outwardNormal(tgtSide).opposite()

	blocked := func(a, b Point) bool {
		for _, r := range l.Rects {
			if segmentBlocked(r, a, b) {
				return true
			}
		}
		return false
	}
	inside := func(p Point) bool {
		return p.X >= 0 && p.X <= max+1 && p.Y >= 0 && p.Y <= max+1
	}
	type pair struct{ l, t int }
	less := func(a, b pair) bool {
		return a.l < b.l || (a.l == b.l && a.t < b.t)
	}
	dist := map[bruteState]pair{}
	pq := &brutePQ{}
	push := func(s bruteState, p pair) {
		if d, seen := dist[s]; !seen || less(p, d) {
			dist[s] = p
			heap.Push(pq, bruteItem{s, p.l*100000 + p.t})
		}
	}
	step := req.Source.At
	step.X += startDir.delta().X
	step.Y += startDir.delta().Y
	if !inside(step) || blocked(req.Source.At, step) {
		return 0, 0, false
	}
	push(bruteState{step.X, step.Y, startDir}, pair{1, 0})

	for pq.Len() > 0 {
		it := heap.Pop(pq).(bruteItem)
		cur := dist[it.s]
		if it.cost != cur.l*100000+cur.t {
			continue
		}
		if it.s.x == req.Target.At.X && it.s.y == req.Target.At.Y && it.s.d == arriveDir {
			return cur.l, cur.t, true
		}
		for _, nd := range []dir{dRight, dUp, dLeft, dDown} {
			if nd == it.s.d.opposite() {
				continue
			}
			nx, ny := it.s.x+nd.delta().X, it.s.y+nd.delta().Y
			np := Point{nx, ny}
			if !inside(np) || blocked(Point{it.s.x, it.s.y}, np) {
				continue
			}
			nt := cur.t
			if nd != it.s.d {
				nt++
			}
			push(bruteState{nx, ny, nd}, pair{cur.l + 1, nt})
		}
	}
	return 0, 0, false
}

// --- path validation (independent of the solver internals) ---

func validatePath(t *testing.T, l *Layout, req Request, p Path) {
	t.Helper()
	if len(p) < 2 {
		t.Fatalf("path too short: %v", p)
	}
	if p[0] != req.Source.At {
		t.Fatalf("path starts at %v, want %v", p[0], req.Source.At)
	}
	if p[len(p)-1] != req.Target.At {
		t.Fatalf("path ends at %v, want %v", p[len(p)-1], req.Target.At)
	}
	srcSide, _ := portSide(l.Rects[req.Source.Node], req.Source.At)
	tgtSide, _ := portSide(l.Rects[req.Target.Node], req.Target.At)
	if d := heading(p[0], p[1]); d != outwardNormal(srcSide) {
		t.Fatalf("first segment heads %v, want outward normal %v", d, outwardNormal(srcSide))
	}
	if d := heading(p[len(p)-2], p[len(p)-1]); d != outwardNormal(tgtSide).opposite() {
		t.Fatalf("last segment heads %v, want inward normal %v", d, outwardNormal(tgtSide).opposite())
	}
	prev := dRight
	for i := 1; i < len(p); i++ {
		a, b := p[i-1], p[i]
		if a == b {
			t.Fatalf("zero-length segment at %v", a)
		}
		if a.X != b.X && a.Y != b.Y {
			t.Fatalf("segment %v->%v not axis-aligned", a, b)
		}
		for _, q := range []Point{a, b} {
			if q.X < CanvasMin || q.X > CanvasMax || q.Y < CanvasMin || q.Y > CanvasMax {
				t.Fatalf("point %v outside canvas", q)
			}
		}
		d := heading(a, b)
		if i > 1 && d == prev {
			t.Fatalf("collinear consecutive segments at %v", b)
		}
		prev = d
		for _, r := range l.Rects {
			if segmentBlocked(r, a, b) {
				t.Fatalf("segment %v->%v crosses interior of %v", a, b, r)
			}
		}
	}
}

// --- randomized cross-check against the reference solver ---

func randomLayout(rng *rand.Rand, max int) (*Layout, Request) {
	n := 1 + rng.Intn(4)
	rects := make([]Rect, 0, n+2)
	for i := 0; i < n; i++ {
		x1 := 1 + rng.Intn(max-4)
		y1 := 1 + rng.Intn(max-4)
		x2 := x1 + 2 + rng.Intn(max-2-x1)
		y2 := y1 + 2 + rng.Intn(max-2-y1)
		rects = append(rects, Rect{Min: Point{x1, y1}, Max: Point{x2, y2}})
	}
	rects = append(rects, sealingWalls(max)...)
	l := &Layout{Rects: rects}

	randPort := func() Port {
		node := rng.Intn(n)
		r := rects[node]
		// Pick a side long enough to hold a non-corner port.
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
	return l, req
}

func TestRandomAgainstBruteForce(t *testing.T) {
	const max = 30
	iters := 400
	if v := os.Getenv("ORTH_RANDOM_ITERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			iters = n
		}
	}
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < iters; i++ {
		l, req := randomLayout(rng, max)
		wantCost, wantOK := bruteCost(l, req, max)
		path, err := l.Route(req)
		if wantOK != (err == nil) {
			t.Fatalf("case %d: reachability mismatch: lib err=%v, brute ok=%v\nrects=%v\nreq=%+v",
				i, err, wantOK, l.Rects, req)
		}
		if !wantOK {
			continue
		}
		t.Logf("case %d rects=%v req=%+v path=%v", i, l.Rects, req, path)
		validatePath(t, l, req, path)
		if got := path.Cost(); got != wantCost {
			t.Fatalf("case %d: cost=%d, brute optimal=%d\nrects=%v\nreq=%+v\npath=%v",
				i, got, wantCost, l.Rects, req, path)
		}
	}
}
