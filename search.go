package orthroute

import "container/heap"

// state is a search position: a node plus the heading used to reach it.
type state struct {
	node int
	head dir
}

// shortestPath runs Dijkstra over (node, heading) states. The first move
// out of start must follow startDir; the path is accepted only when it
// arrives at goal heading arriveDir. It returns the bend polyline and cost.
func shortestPath(g *graph, start, goal Point, startDir, arriveDir dir) (Path, int, bool) {
	startID, ok := g.id[start]
	if !ok {
		return nil, 0, false
	}
	goalID, ok := g.id[goal]
	if !ok {
		return nil, 0, false
	}

	dist := make(map[state]int)
	parent := make(map[state]state)
	pq := &prioQueue{}

	// The first segment must leave along startDir: seed only that move.
	seeded := false
	for _, e := range g.adj[startID] {
		if heading(g.nodes[startID], g.nodes[e.to]) != startDir {
			continue
		}
		s := state{node: e.to, head: startDir}
		dist[s] = e.length
		parent[s] = state{node: -1, head: startDir} // sentinel root
		heap.Push(pq, item{s: s, cost: e.length})
		seeded = true
	}
	if !seeded {
		return nil, 0, false
	}

	for pq.Len() > 0 {
		it := heap.Pop(pq).(item)
		if d, seen := dist[it.s]; seen && it.cost > d {
			continue
		}
		if it.s.node == goalID && it.s.head == arriveDir {
			return buildPath(g, parent, startID, it.s), it.cost, true
		}
		for _, e := range g.adj[it.s.node] {
			nd := heading(g.nodes[it.s.node], g.nodes[e.to])
			if nd == it.s.head.opposite() {
				continue // a 180-degree reversal is never a legal segment pair
			}
			cost := it.cost + e.length
			if nd != it.s.head {
				cost += BendCost
			}
			next := state{node: e.to, head: nd}
			if d, seen := dist[next]; !seen || cost < d {
				dist[next] = cost
				parent[next] = it.s
				heap.Push(pq, item{s: next, cost: cost})
			}
		}
	}
	return nil, 0, false
}

// heading returns the travel direction from node a to adjacent node b.
func heading(a, b Point) dir {
	switch {
	case b.X > a.X:
		return dRight
	case b.X < a.X:
		return dLeft
	case b.Y > a.Y:
		return dUp
	default:
		return dDown
	}
}

// buildPath walks parents back from the goal state and merges collinear
// runs into a minimal bend polyline from start to goal. The chain ends at
// the sentinel root, so the path may legally revisit the start node.
func buildPath(g *graph, parent map[state]state, startID int, goal state) Path {
	rev := []int{}
	for cur := goal; cur.node >= 0; cur = parent[cur] {
		rev = append(rev, cur.node)
	}
	rev = append(rev, startID)
	pts := make([]Point, 0, len(rev))
	for i := len(rev) - 1; i >= 0; i-- {
		pts = append(pts, g.nodes[rev[i]])
	}
	out := Path{pts[0]}
	for i := 1; i < len(pts); i++ {
		prev := pts[i-1]
		if i+1 < len(pts) && collinear(prev, pts[i], pts[i+1]) {
			continue
		}
		out = append(out, pts[i])
	}
	return out
}

func collinear(a, b, c Point) bool {
	return (a.X == b.X && b.X == c.X) || (a.Y == b.Y && b.Y == c.Y)
}

type item struct {
	s    state
	cost int
}

type prioQueue []item

func (q prioQueue) Len() int            { return len(q) }
func (q prioQueue) Less(i, j int) bool  { return q[i].cost < q[j].cost }
func (q prioQueue) Swap(i, j int)       { q[i], q[j] = q[j], q[i] }
func (q *prioQueue) Push(x interface{}) { *q = append(*q, x.(item)) }
func (q *prioQueue) Pop() interface{} {
	old := *q
	n := len(old)
	it := old[n-1]
	*q = old[:n-1]
	return it
}
