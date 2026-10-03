package orthroute

import "testing"

// Ports on the canvas border: the forced outward segment may immediately
// leave the canvas, in which case the request is unreachable.
func TestCanvasEdgePorts(t *testing.T) {
	l := withWalls(30,
		Rect{Min: Point{0, 0}, Max: Point{4, 4}},  // 0: touches corner
		Rect{Min: Point{8, 0}, Max: Point{12, 4}}, // 1
	)
	// Feasible: right edge of node 0 to left edge of node 1, straight line.
	routeAndCheck(t, l, Request{
		Source: Port{Node: 0, At: Point{4, 2}},
		Target: Port{Node: 1, At: Point{8, 2}},
	}, 30)
	// Infeasible: outward normal points off the canvas.
	routeAndCheck(t, l, Request{
		Source: Port{Node: 0, At: Point{0, 2}},
		Target: Port{Node: 1, At: Point{8, 2}},
	}, 30)
	routeAndCheck(t, l, Request{
		Source: Port{Node: 0, At: Point{4, 2}},
		Target: Port{Node: 1, At: Point{10, 0}},
	}, 30)
}

// Malformed requests are rejected with a validation error.
func TestInvalidRequests(t *testing.T) {
	l := &Layout{Rects: []Rect{{Min: Point{2, 2}, Max: Point{6, 6}}}}
	cases := []Request{
		// port on a corner
		{Source: Port{Node: 0, At: Point{2, 2}}, Target: Port{Node: 0, At: Point{6, 4}}},
		// port not on the boundary
		{Source: Port{Node: 0, At: Point{3, 3}}, Target: Port{Node: 0, At: Point{6, 4}}},
		// port outside the rect
		{Source: Port{Node: 0, At: Point{2, 9}}, Target: Port{Node: 0, At: Point{6, 4}}},
		// node index out of range
		{Source: Port{Node: 1, At: Point{2, 4}}, Target: Port{Node: 0, At: Point{6, 4}}},
		// coincident ports
		{Source: Port{Node: 0, At: Point{2, 4}}, Target: Port{Node: 0, At: Point{2, 4}}},
	}
	for i, req := range cases {
		p, err := l.Route(req)
		if err == nil || err == ErrUnreachable {
			t.Errorf("case %d: err=%v, want a validation error", i, err)
		}
		if p != nil {
			t.Errorf("case %d: path=%v, want nil", i, p)
		}
	}
	// Rect outside the canvas.
	bad := &Layout{Rects: []Rect{{Min: Point{-1, 0}, Max: Point{4, 4}}}}
	if _, err := bad.Route(Request{
		Source: Port{Node: 0, At: Point{0, 2}},
		Target: Port{Node: 0, At: Point{4, 2}},
	}); err == nil {
		t.Error("rect outside canvas: expected validation error")
	}
	// Degenerate rect.
	flat := &Layout{Rects: []Rect{{Min: Point{2, 2}, Max: Point{2, 6}}}}
	if _, err := flat.Route(Request{
		Source: Port{Node: 0, At: Point{2, 3}},
		Target: Port{Node: 0, At: Point{2, 5}},
	}); err == nil {
		t.Error("degenerate rect: expected validation error")
	}
}
