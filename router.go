// Package orthroute computes minimum-cost orthogonal (Manhattan) routes
// between fixed ports on rectangular nodes, avoiding rectangle interiors.
package orthroute

import (
	"errors"
	"fmt"
)

// Canvas bounds: both axes range in [CanvasMin, CanvasMax], inclusive.
const (
	CanvasMin = 0
	CanvasMax = 10000
)

// BendCost is added to the path length for every 90-degree turn.
const BendCost = 20

// ErrUnreachable is returned when no legal path exists for a request.
var ErrUnreachable = errors.New("orthroute: no legal path")

// Point is an integer coordinate on the canvas.
type Point struct {
	X, Y int
}

// Rect is an axis-aligned rectangle with positive width and height.
// Min and Max are inclusive bounds; the boundary itself is walkable.
type Rect struct {
	Min, Max Point
}

// Side identifies one edge of a node rectangle.
type Side int

const (
	SideLeft Side = iota
	SideRight
	SideTop
	SideBottom
)

// Port is a connection point on the boundary of a node rectangle.
type Port struct {
	Node int   // index into Layout.Rects
	At   Point // must lie on the node's boundary, not on a corner
}

// Request asks for a route between two fixed ports.
type Request struct {
	Source, Target Port
}

// Path is the polyline of bend points from source port to target port.
// Consecutive segments are axis-aligned, non-zero and perpendicular.
type Path []Point

// Cost returns total segment length plus BendCost per turn.
func (p Path) Cost() int {
	if len(p) < 2 {
		return 0
	}
	total := 0
	prev := dRight
	for i := 1; i < len(p); i++ {
		a, b := p[i-1], p[i]
		dx, dy := b.X-a.X, b.Y-a.Y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		total += dx + dy
		d := heading(a, b)
		if i > 1 && d != prev {
			total += BendCost
		}
		prev = d
	}
	return total
}

// Result pairs a routed Path with its error (ErrUnreachable or nil).
type Result struct {
	Path Path
	Err  error
}

// Layout is an immutable node placement to route against.
type Layout struct {
	Rects []Rect
}

// Route computes a minimum-cost legal path for one request.
func (l *Layout) Route(req Request) (Path, error) {
	if err := l.validate(req); err != nil {
		return nil, err
	}
	srcSide, _ := portSide(l.Rects[req.Source.Node], req.Source.At)
	tgtSide, _ := portSide(l.Rects[req.Target.Node], req.Target.At)

	// Besides the ports themselves, add grid lines one step around each
	// port: the forced first/last segments may turn optimally only there.
	extra := []Point{req.Source.At, req.Target.At}
	for _, p := range []Point{req.Source.At, req.Target.At} {
		extra = append(extra,
			Point{p.X + 1, p.Y}, Point{p.X - 1, p.Y},
			Point{p.X, p.Y + 1}, Point{p.X, p.Y - 1})
	}
	g := buildGraph(l.Rects, extra...)
	startDir := outwardNormal(srcSide)
	arriveDir := outwardNormal(tgtSide).opposite()

	path, _, ok := shortestPath(g, req.Source.At, req.Target.At, startDir, arriveDir)
	if !ok {
		return nil, ErrUnreachable
	}
	return path, nil
}

// RouteAll solves each request independently; paths may cross each other.
func (l *Layout) RouteAll(reqs []Request) []Result {
	out := make([]Result, len(reqs))
	for i, req := range reqs {
		p, err := l.Route(req)
		out[i] = Result{Path: p, Err: err}
	}
	return out
}

func (l *Layout) validate(req Request) error {
	for i, r := range l.Rects {
		if r.Min.X >= r.Max.X || r.Min.Y >= r.Max.Y {
			return fmt.Errorf("orthroute: rect %d has non-positive size", i)
		}
		if r.Min.X < CanvasMin || r.Max.X > CanvasMax ||
			r.Min.Y < CanvasMin || r.Max.Y > CanvasMax {
			return fmt.Errorf("orthroute: rect %d outside canvas", i)
		}
	}
	for _, port := range []Port{req.Source, req.Target} {
		if port.Node < 0 || port.Node >= len(l.Rects) {
			return fmt.Errorf("orthroute: node index %d out of range", port.Node)
		}
		if _, ok := portSide(l.Rects[port.Node], port.At); !ok {
			return fmt.Errorf("orthroute: port %+v not on node %d edge", port.At, port.Node)
		}
	}
	if req.Source.At == req.Target.At {
		return fmt.Errorf("orthroute: source and target ports coincide")
	}
	return nil
}
