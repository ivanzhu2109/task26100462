package orthrouter

import "sort"

// grid 为一次求解使用的候选坐标网格。
// 最优路径可限制在以下候选线上（整数坐标、允许贴边、禁止滑行的前提下）：
//   - 矩形边线及其 ±1 线（绕角 U 形回转的最优深度为边外一格）；
//   - 连接点坐标及其 ±1 线（首末段长度非零）；
//   - 画布边界。
type grid struct {
	xs, ys []int
	// hClear[yi][xi]: (xs[xi],ys[yi]) -> (xs[xi+1],ys[yi]) 是否畅通
	// vClear[yi][xi]: (xs[xi],ys[yi]) -> (xs[xi],ys[yi+1]) 是否畅通
	hClear [][]bool
	vClear [][]bool
}

// buildGrid 收集候选坐标并预计算相邻线间通行表。
func buildGrid(r *Router, req Request) grid {
	s, t := req.Source.Point, req.Target.Point
	xs := []int{r.minX, r.maxX, s.X, s.X - 1, s.X + 1, t.X, t.X - 1, t.X + 1}
	ys := []int{r.minY, r.maxY, s.Y, s.Y - 1, s.Y + 1, t.Y, t.Y - 1, t.Y + 1}
	for _, rc := range r.rects {
		xs = append(xs, rc.X1-1, rc.X1, rc.X2, rc.X2+1)
		ys = append(ys, rc.Y1-1, rc.Y1, rc.Y2, rc.Y2+1)
	}
	g := grid{
		xs: uniqSorted(xs, r.minX, r.maxX),
		ys: uniqSorted(ys, r.minY, r.maxY),
	}
	nx, ny := len(g.xs), len(g.ys)
	g.hClear = make([][]bool, ny)
	g.vClear = make([][]bool, ny)
	for yi, y := range g.ys {
		g.hClear[yi] = make([]bool, nx-1)
		for xi := 0; xi+1 < nx; xi++ {
			g.hClear[yi][xi] = segClear(r.rects, Point{g.xs[xi], y}, Point{g.xs[xi+1], y})
		}
	}
	for yi := 0; yi+1 < ny; yi++ {
		g.vClear[yi] = make([]bool, nx)
		for xi, x := range g.xs {
			g.vClear[yi][xi] = segClear(r.rects, Point{x, g.ys[yi]}, Point{x, g.ys[yi+1]})
		}
	}
	return g
}

// clear 查询相邻网格点间是否畅通。
func (g grid) clear(xi, yi, nxi, nyi int) bool {
	if yi == nyi {
		return g.hClear[yi][min(xi, nxi)]
	}
	return g.vClear[min(yi, nyi)][xi]
}

func uniqSorted(v []int, lo, hi int) []int {
	sort.Ints(v)
	out := v[:0]
	for _, x := range v {
		if x < lo || x > hi {
			continue
		}
		if len(out) == 0 || out[len(out)-1] != x {
			out = append(out, x)
		}
	}
	return out
}

// segClear 判断水平或竖直线段（含端点）是否不经过任何矩形的严格内部。
// 矩形边界允许行走；开线段与开矩形相交则禁止。
func segClear(rects []Rect, a, b Point) bool {
	if a.X == b.X {
		y1, y2 := min(a.Y, b.Y), max(a.Y, b.Y)
		for _, rc := range rects {
			if a.X > rc.X1 && a.X < rc.X2 && max(y1, rc.Y1) < min(y2, rc.Y2) {
				return false
			}
		}
		return true
	}
	if a.Y == b.Y {
		x1, x2 := min(a.X, b.X), max(a.X, b.X)
		for _, rc := range rects {
			if a.Y > rc.Y1 && a.Y < rc.Y2 && max(x1, rc.X1) < min(x2, rc.X2) {
				return false
			}
		}
		return true
	}
	return false // 非轴对齐线段不合法
}
