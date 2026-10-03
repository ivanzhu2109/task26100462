package orthrouter

import "container/heap"

// 方向编码：0:+x 1:-x 2:+y 3:-y
var dirDelta = [4]Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

// outwardDir 返回连接点所在边的外法线方向；点不在边上时返回 false。
func outwardDir(rc Rect, p Point) (int, bool) {
	onVert := p.Y > rc.Y1 && p.Y < rc.Y2
	onHoriz := p.X > rc.X1 && p.X < rc.X2
	switch {
	case p.X == rc.X1 && onVert:
		return 1, true
	case p.X == rc.X2 && onVert:
		return 0, true
	case p.Y == rc.Y1 && onHoriz:
		return 3, true
	case p.Y == rc.Y2 && onHoriz:
		return 2, true
	}
	return 0, false
}

func opposite(d int) int { return d ^ 1 }

// stateID 编码 (xi, yi, dir) 状态。
func stateID(nx, xi, yi, dir int) int { return (yi*nx+xi)*4 + dir }

type pqItem struct {
	id, cost int
}
type pq []pqItem

func (h pq) Len() int            { return len(h) }
func (h pq) Less(i, j int) bool  { return h[i].cost < h[j].cost }
func (h pq) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *pq) Push(x any)         { *h = append(*h, x.(pqItem)) }
func (h *pq) Pop() any           { old := *h; it := old[len(old)-1]; *h = old[:len(old)-1]; return it }

// dijkstra 在网格上以 (点, 进入方向) 为状态求最小代价路径。
// 首段强制沿源边外法线，末段强制沿目标边内法线到达。
// 返回折点序列（已压缩共线点）与总代价。
func dijkstra(g grid, r *Router, req Request) ([]Point, int, bool) {
	nx, ny := len(g.xs), len(g.ys)
	index := func(v []int, x int) int {
		for i, e := range v {
			if e == x {
				return i
			}
		}
		return -1
	}
	sx, sy := index(g.xs, req.Source.Point.X), index(g.ys, req.Source.Point.Y)
	tx, ty := index(g.xs, req.Target.Point.X), index(g.ys, req.Target.Point.Y)
	if sx < 0 || sy < 0 || tx < 0 || ty < 0 {
		return nil, 0, false
	}
	srcDir, ok1 := outwardDir(r.rects[req.Source.Node], req.Source.Point)
	tgtOut, ok2 := outwardDir(r.rects[req.Target.Node], req.Target.Point)
	if !ok1 || !ok2 {
		return nil, 0, false
	}
	arrDir := opposite(tgtOut) // 到达目标时须沿内法线方向移动

	const inf = int(^uint(0) >> 1)
	n := nx * ny * 4
	dist := make([]int, n)
	parent := make([]int, n)
	for i := range dist {
		dist[i] = inf
		parent[i] = -1
	}
	start := stateID(nx, sx, sy, srcDir)
	dist[start] = 0
	h := &pq{{start, 0}}
	heap.Init(h)

	final := -1
	for h.Len() > 0 {
		it := heap.Pop(h).(pqItem)
		if it.cost > dist[it.id] {
			continue
		}
		xi := (it.id / 4) % nx
		yi := (it.id / 4) / nx
		dir := it.id % 4
		if xi == tx && yi == ty && dir == arrDir {
			final = it.id
			break
		}
		// 可继续直行或左右转弯（禁止掉头）。
		// 源点处禁止转弯：首段必须沿源边外法线离开。
		dirs := nextDirs(dir)
		if it.id == start {
			dirs = [3]int{dir, -1, -1}
		}
		for _, nd := range dirs {
			if nd < 0 {
				continue
			}
			nxi, nyi := xi+dirDelta[nd].X, yi+dirDelta[nd].Y
			if nxi < 0 || nxi >= nx || nyi < 0 || nyi >= ny {
				continue
			}
			if !g.clear(xi, yi, nxi, nyi) {
				continue
			}
			add := abs(g.xs[nxi]-g.xs[xi]) + abs(g.ys[nyi]-g.ys[yi])
			if nd != dir {
				add += r.turnCost
			}
			nid := stateID(nx, nxi, nyi, nd)
			if it.cost+add < dist[nid] {
				dist[nid] = it.cost + add
				parent[nid] = it.id
				heap.Push(h, pqItem{nid, dist[nid]})
			}
		}
	}
	if final < 0 {
		return nil, 0, false
	}
	// 重建状态序列并压缩共线点。
	var ids []int
	for id := final; id >= 0; id = parent[id] {
		ids = append(ids, id)
	}
	pts := make([]Point, 0, len(ids))
	for i := len(ids) - 1; i >= 0; i-- {
		xi := (ids[i] / 4) % nx
		yi := (ids[i] / 4) / nx
		pts = append(pts, Point{g.xs[xi], g.ys[yi]})
	}
	pts = compress(pts)
	return pts, dist[final], true
}

// nextDirs 返回允许继续的方向：直行与左右转弯（不含掉头）。
func nextDirs(dir int) [3]int {
	// dir 与 dir^1 互为反向；垂直方向为 (dir+1)&3 与 (dir+3)&3 当 dir 为水平时，
	// 统一用查表避免位运算错误。
	switch dir {
	case 0, 1:
		return [3]int{dir, 2, 3}
	default:
		return [3]int{dir, 0, 1}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// compress 删除共线的中间点。
func compress(pts []Point) []Point {
	if len(pts) < 3 {
		return pts
	}
	out := []Point{pts[0]}
	for i := 1; i < len(pts)-1; i++ {
		a, b, c := out[len(out)-1], pts[i], pts[i+1]
		if (a.X == b.X && b.X == c.X) || (a.Y == b.Y && b.Y == c.Y) {
			continue
		}
		out = append(out, b)
	}
	return append(out, pts[len(pts)-1])
}
