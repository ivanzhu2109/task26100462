package orthroute

import "testing"

// withWalls appends sealing walls so brute-force comparison is exact.
func withWalls(max int, rects ...Rect) *Layout {
	return &Layout{Rects: append(rects, sealingWalls(max)...)}
}

// routeAndCheck runs the library solver and cross-checks against the
// brute-force reference on the sealed [0,max] grid.
func routeAndCheck(t *testing.T, l *Layout, req Request, max int) Path {
	t.Helper()
	wantCost, wantOK := bruteCost(l, req, max)
	path, err := l.Route(req)
	if wantOK != (err == nil) {
		t.Fatalf("reachability mismatch: lib err=%v, brute ok=%v", err, wantOK)
	}
	if !wantOK {
		return nil
	}
	validatePath(t, l, req, path)
	if got := path.Cost(); got != wantCost {
		t.Fatalf("cost=%d, brute optimal=%d, path=%v", got, wantCost, path)
	}
	return path
}

// A shorter path with more bends must lose to a longer one with fewer.
// Shortest legal path: length 18, 6 turns, cost 138.
// Optimal: length 20, 4 turns, cost 100.
func TestBendCostBeatsShorterPath(t *testing.T) {
	l := withWalls(30,
		Rect{Min: Point{7, 18}, Max: Point{19, 26}},
		Rect{Min: Point{19, 22}, Max: Point{24, 26}},
		Rect{Min: Point{15, 26}, Max: Point{22, 29}},
	)
	req := Request{
		Source: Port{Node: 2, At: Point{19, 29}},
		Target: Port{Node: 1, At: Point{20, 22}},
	}
	path := routeAndCheck(t, l, req, 30)
	if path == nil {
		t.Fatal("expected a legal path")
	}
	if got := path.Cost(); got != 100 {
		t.Fatalf("cost=%d, want 100 (path=%v)", got, path)
	}
	length, turns, ok := bruteLex(l, req, 30)
	if !ok {
		t.Fatal("bruteLex unreachable")
	}
	if length+turns*BendCost <= path.Cost() {
		t.Fatalf("test premise broken: shortest path len=%d turns=%d cost=%d <= %d",
			length, turns, length+turns*BendCost, path.Cost())
	}
	if length >= 20 {
		t.Fatalf("test premise broken: shortest path length %d not shorter than chosen", length)
	}
}

// The only way across the barrier is the walkable line where the two
// rectangles share a coincident boundary (x=6).
func TestCoincidentBoundaryChannel(t *testing.T) {
	l := withWalls(30,
		Rect{Min: Point{2, 0}, Max: Point{5, 3}},    // 0: source node
		Rect{Min: Point{7, 22}, Max: Point{11, 25}}, // 1: target node
		Rect{Min: Point{0, 4}, Max: Point{6, 10}},   // 2: shares edge x=6 with 3
		Rect{Min: Point{6, 4}, Max: Point{12, 10}},  // 3
		Rect{Min: Point{0, 11}, Max: Point{6, 20}},  // 4: shares edge x=6 with 5
		Rect{Min: Point{6, 11}, Max: Point{12, 20}}, // 5
	)
	req := Request{
		Source: Port{Node: 0, At: Point{3, 3}},
		Target: Port{Node: 1, At: Point{9, 22}},
	}
	path := routeAndCheck(t, l, req, 30)
	if path == nil {
		t.Fatal("expected a legal path through the shared boundary")
	}
	// The optimal route funnels through the coincident boundary x=6.
	usesChannel := false
	for i := 1; i < len(path); i++ {
		a, b := path[i-1], path[i]
		if a.X == 6 && b.X == 6 && a.Y <= 10 && b.Y >= 11 {
			usesChannel = true
		}
		if a.X == 6 && b.X == 6 && b.Y <= 10 && a.Y >= 11 {
			usesChannel = true
		}
	}
	if !usesChannel {
		t.Fatalf("path %v does not use the coincident boundary channel", path)
	}
}

// Overlapping rectangles: a boundary of one rect lying strictly inside
// another is not walkable there.
func TestOverlappingRects(t *testing.T) {
	l := withWalls(30,
		Rect{Min: Point{1, 1}, Max: Point{3, 3}},     // 0: source node
		Rect{Min: Point{20, 20}, Max: Point{23, 23}}, // 1: target node
		Rect{Min: Point{4, 4}, Max: Point{12, 12}},   // 2: overlaps with 3
		Rect{Min: Point{8, 8}, Max: Point{16, 16}},   // 3
	)
	req := Request{
		Source: Port{Node: 0, At: Point{3, 2}},
		Target: Port{Node: 1, At: Point{20, 21}},
	}
	routeAndCheck(t, l, req, 30)
	// Sanity: the shared-boundary segment x=8, y in (8,12) lies strictly
	// inside rect 2, so walking there is illegal.
	if !segmentBlocked(l.Rects[2], Point{8, 9}, Point{8, 11}) {
		t.Fatal("segment along rect 3 boundary inside rect 2 must be blocked")
	}
}

// A width-1 obstacle blocks a crossing segment even though no integer
// point of the segment is strictly inside it.
func TestThinObstacleBlocksMidSegment(t *testing.T) {
	thin := Rect{Min: Point{5, 0}, Max: Point{6, 20}}
	if insideStrict(thin, Point{5, 10}) || insideStrict(thin, Point{6, 10}) {
		t.Fatal("boundary points of a width-1 rect are not strictly inside")
	}
	if !segmentBlocked(thin, Point{4, 10}, Point{7, 10}) {
		t.Fatal("segment crossing a width-1 rect mid-segment must be blocked")
	}
	l := withWalls(30,
		Rect{Min: Point{1, 8}, Max: Point{3, 12}},  // 0: source node
		Rect{Min: Point{9, 8}, Max: Point{11, 12}}, // 1: target node
		thin, // 2: vertical barrier, forces a detour around y=20
	)
	req := Request{
		Source: Port{Node: 0, At: Point{3, 10}},
		Target: Port{Node: 1, At: Point{9, 10}},
	}
	routeAndCheck(t, l, req, 30)
}

// A target sealed inside a rectangle ring is unreachable.
func TestUnreachableSealedTarget(t *testing.T) {
	l := withWalls(30,
		Rect{Min: Point{1, 1}, Max: Point{3, 3}},     // 0: source node
		Rect{Min: Point{10, 10}, Max: Point{14, 14}}, // 1: target node, sealed
		Rect{Min: Point{8, 14}, Max: Point{16, 16}},  // ring: top
		Rect{Min: Point{8, 8}, Max: Point{16, 10}},   // ring: bottom
		Rect{Min: Point{8, 8}, Max: Point{10, 16}},   // ring: left
		Rect{Min: Point{14, 8}, Max: Point{16, 16}},  // ring: right
	)
	req := Request{
		Source: Port{Node: 0, At: Point{3, 2}},
		Target: Port{Node: 1, At: Point{12, 14}},
	}
	path, err := l.Route(req)
	if err != ErrUnreachable {
		t.Fatalf("err=%v, want ErrUnreachable", err)
	}
	if path != nil {
		t.Fatalf("path=%v, want nil", path)
	}
	if _, ok := bruteCost(l, req, 30); ok {
		t.Fatal("brute reference disagrees: reachable")
	}
}

// Two ports on the same node: the path must leave one side and arrive
// at the other without crossing the node.
func TestSameNodeTwoPorts(t *testing.T) {
	l := withWalls(30,
		Rect{Min: Point{6, 6}, Max: Point{14, 14}},
	)
	req := Request{
		Source: Port{Node: 0, At: Point{6, 10}},
		Target: Port{Node: 0, At: Point{14, 10}},
	}
	routeAndCheck(t, l, req, 30)
}

// RouteAll solves requests independently and matches individual Route.
func TestRouteAllIndependent(t *testing.T) {
	l := withWalls(30,
		Rect{Min: Point{2, 2}, Max: Point{5, 5}},
		Rect{Min: Point{10, 2}, Max: Point{13, 5}},
		Rect{Min: Point{6, 8}, Max: Point{9, 20}},
	)
	reqs := []Request{
		{Source: Port{Node: 0, At: Point{5, 3}}, Target: Port{Node: 1, At: Point{10, 3}}},
		{Source: Port{Node: 1, At: Point{11, 5}}, Target: Port{Node: 0, At: Point{3, 5}}},
	}
	results := l.RouteAll(reqs)
	if len(results) != len(reqs) {
		t.Fatalf("got %d results, want %d", len(results), len(reqs))
	}
	for i, req := range reqs {
		p, err := l.Route(req)
		if (err == nil) != (results[i].Err == nil) {
			t.Fatalf("req %d: batch err=%v, single err=%v", i, results[i].Err, err)
		}
		if err == nil && results[i].Path.Cost() != p.Cost() {
			t.Fatalf("req %d: batch cost=%d, single cost=%d",
				i, results[i].Path.Cost(), p.Cost())
		}
		if err == nil {
			validatePath(t, l, req, results[i].Path)
		}
	}
}
