package orthroute

import "sort"

// edge links two adjacent walkable grid nodes.
type edge struct {
	to     int
	length int
}

// graph is the compressed routing graph over Hanan-grid lines:
// canvas borders, rectangle edges and port coordinates (plus the lines
// one step around each port, which the forced first/last segments may
// require for an optimal turn).
type graph struct {
	nodes []Point
	id    map[Point]int
	adj   [][]edge
}

// buildGraph constructs the routing graph for rects, guaranteeing that
// every point in extra exists as a node (unless strictly inside a rect).
// Coordinates of extra points also become grid lines; out-of-canvas
// coordinates are ignored.
func buildGraph(rects []Rect, extra ...Point) *graph {
	xs := map[int]bool{CanvasMin: true, CanvasMax: true}
	ys := map[int]bool{CanvasMin: true, CanvasMax: true}
	for _, r := range rects {
		xs[r.Min.X], xs[r.Max.X] = true, true
		ys[r.Min.Y], ys[r.Max.Y] = true, true
	}
	for _, p := range extra {
		if p.X >= CanvasMin && p.X <= CanvasMax {
			xs[p.X] = true
		}
		if p.Y >= CanvasMin && p.Y <= CanvasMax {
			ys[p.Y] = true
		}
	}
	xlist := sortedKeys(xs)
	ylist := sortedKeys(ys)

	g := &graph{id: make(map[Point]int)}
	for _, y := range ylist {
		for _, x := range xlist {
			p := Point{x, y}
			if blockedPoint(rects, p) {
				continue
			}
			g.id[p] = len(g.nodes)
			g.nodes = append(g.nodes, p)
		}
	}
	g.adj = make([][]edge, len(g.nodes))

	// Horizontal sweeps: connect consecutive walkable nodes on each row.
	for _, y := range ylist {
		prev := -1
		for _, x := range xlist {
			cur, ok := g.id[Point{x, y}]
			if !ok {
				continue
			}
			if prev >= 0 {
				g.link(rects, prev, cur)
			}
			prev = cur
		}
	}
	// Vertical sweeps: connect consecutive walkable nodes on each column.
	for _, x := range xlist {
		prev := -1
		for _, y := range ylist {
			cur, ok := g.id[Point{x, y}]
			if !ok {
				continue
			}
			if prev >= 0 {
				g.link(rects, prev, cur)
			}
			prev = cur
		}
	}
	return g
}

// link adds an undirected edge between nodes a and b when the segment
// between them stays clear of every rectangle's strict interior.
func (g *graph) link(rects []Rect, a, b int) {
	pa, pb := g.nodes[a], g.nodes[b]
	for _, r := range rects {
		if segmentBlocked(r, pa, pb) {
			return
		}
	}
	length := pa.X - pb.X
	if length < 0 {
		length = -length
	}
	dy := pa.Y - pb.Y
	if dy < 0 {
		dy = -dy
	}
	length += dy
	g.adj[a] = append(g.adj[a], edge{to: b, length: length})
	g.adj[b] = append(g.adj[b], edge{to: a, length: length})
}

func blockedPoint(rects []Rect, p Point) bool {
	for _, r := range rects {
		if insideStrict(r, p) {
			return true
		}
	}
	return false
}

func sortedKeys(set map[int]bool) []int {
	out := make([]int, 0, len(set))
	for v := range set {
		out = append(out, v)
	}
	sort.Ints(out)
	return out
}
