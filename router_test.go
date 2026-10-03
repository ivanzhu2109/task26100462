package orthrouter

import "testing"

// ---------- 独立暴力参考解：在完整整数网格上逐点搜索 ----------

// stepClear 判断单位步是否合法：开线段中点不在任何矩形严格内部。
// 单位步的开线段无整数点，中点检查与连续检查等价。
func stepClear(rects []Rect, a, b Point) bool {
	mx2 := a.X + b.X // 中点的两倍，避免浮点
	my2 := a.Y + b.Y
	for _, rc := range rects {
		if mx2 > 2*rc.X1 && mx2 < 2*rc.X2 && my2 > 2*rc.Y1 && my2 < 2*rc.Y2 {
			return false
		}
	}
	return true
}

// bruteForce 在画布全部整数点上以 (点,方向) 为状态做 Dijkstra，
// 权函数由 weight(len, turns) 给出，返回最小权。
func bruteForce(r *Router, req Request, weight func(length, turns int) int) (int, bool) {
	srcDir, ok1 := outwardDir(r.rects[req.Source.Node], req.Source.Point)
	tgtOut, ok2 := outwardDir(r.rects[req.Target.Node], req.Target.Point)
	if !ok1 || !ok2 {
		return 0, false
	}
	arrDir := opposite(tgtOut)
	w := r.maxX - r.minX + 1
	h := r.maxY - r.minY + 1
	id := func(x, y, d int) int { return ((y-r.minY)*w+(x-r.minX))*4 + d }
	const inf = int(^uint(0) >> 1)
	dist := make([]int, w*h*4)
	for i := range dist {
		dist[i] = inf
	}
	start := id(req.Source.Point.X, req.Source.Point.Y, srcDir)
	dist[start] = 0
	// 简单数组版 Dijkstra（小画布足够快）。
	done := make([]bool, w*h*4)
	for {
		best, bi := inf, -1
		for i := range dist {
			if !done[i] && dist[i] < best {
				best, bi = dist[i], i
			}
		}
		if bi < 0 {
			return 0, false
		}
		done[bi] = true
		x := r.minX + (bi/4)%w
		y := r.minY + (bi/4)/w
		d := bi % 4
		if x == req.Target.Point.X && y == req.Target.Point.Y && d == arrDir {
			return best, true
		}
		dirs := nextDirs(d)
		if bi == start {
			dirs = [3]int{d, -1, -1}
		}
		for _, nd := range dirs {
			if nd < 0 {
				continue
			}
			nx, ny := x+dirDelta[nd].X, y+dirDelta[nd].Y
			if nx < r.minX || nx > r.maxX || ny < r.minY || ny > r.maxY {
				continue
			}
			if !stepClear(r.rects, Point{x, y}, Point{nx, ny}) {
				continue
			}
			add := weight(1, 0)
			if nd != d {
				add = weight(1, 1)
			}
			if nid := id(nx, ny, nd); dist[bi]+add < dist[nid] {
				dist[nid] = dist[bi] + add
			}
		}
	}
}

// bruteOptimal 返回暴力参考的最小代价（长度 + turnCost*转弯数）。
func bruteOptimal(r *Router, req Request) (int, bool) {
	return bruteForce(r, req, func(length, turns int) int {
		return length + r.turnCost*turns
	})
}

// bruteShortest 返回按 (长度, 转弯数) 字典序最小的长度与转弯数。
func bruteShortest(r *Router, req Request) (length, turns int, ok bool) {
	const big = 1 << 20
	w, found := bruteForce(r, req, func(length, turns int) int {
		return length*big + turns
	})
	if !found {
		return 0, 0, false
	}
	return w / big, w % big, true
}

// ---------- 路径合法性校验 ----------

func validatePath(t *testing.T, r *Router, req Request, p Path) {
	t.Helper()
	pts := p.Points
	if len(pts) < 2 {
		t.Fatalf("路径折点不足: %v", pts)
	}
	srcDir, _ := outwardDir(r.rects[req.Source.Node], req.Source.Point)
	tgtOut, _ := outwardDir(r.rects[req.Target.Node], req.Target.Point)
	arrDir := opposite(tgtOut)
	if pts[0] != req.Source.Point || pts[len(pts)-1] != req.Target.Point {
		t.Fatalf("路径端点与连接点不符: %v", pts)
	}
	length := 0
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		if a == b {
			t.Fatalf("存在零长度段: %v", pts)
		}
		if a.X != b.X && a.Y != b.Y {
			t.Fatalf("存在非轴对齐段: %v -> %v", a, b)
		}
		for _, q := range []Point{a, b} {
			if q.X < r.minX || q.X > r.maxX || q.Y < r.minY || q.Y > r.maxY {
				t.Fatalf("路径越出画布: %v", q)
			}
		}
		if !segClear(r.rects, a, b) {
			t.Fatalf("线段穿过矩形内部: %v -> %v", a, b)
		}
		length += abs(b.X-a.X) + abs(b.Y-a.Y)
		d := segDir(a, b)
		if i == 0 && d != srcDir {
			t.Fatalf("首段未沿源边外法线: %v -> %v", a, b)
		}
		if i == len(pts)-2 && d != arrDir {
			t.Fatalf("末段未沿目标边内法线: %v -> %v", a, b)
		}
	}
	want := length + r.turnCost*(len(pts)-2)
	if want != p.Cost {
		t.Fatalf("代价不符: 重算 %d, 返回 %d", want, p.Cost)
	}
}

func segDir(a, b Point) int {
	switch {
	case b.X > a.X:
		return 0
	case b.X < a.X:
		return 1
	case b.Y > a.Y:
		return 2
	default:
		return 3
	}
}

// ---------- 场景测试 ----------

// 更短但折弯更多的路径不应入选。
func TestShorterButMoreTurnsNotChosen(t *testing.T) {
	rects := []Rect{
		{0, 2, 2, 4},  // 节点 A，源点在其右边
		{8, 7, 10, 9}, // 节点 B，目标点在其上边
		{4, 0, 5, 4},  // 障碍 1：顶到画布上边
		{6, 3, 7, 9},  // 障碍 2：与障碍 1 错位形成狭窄豁口
	}
	r := NewRouter(0, 0, 10, 10, rects, 20)
	req := Request{
		Source: Endpoint{Node: 0, Point: Point{2, 3}},
		Target: Endpoint{Node: 1, Point: Point{9, 7}},
	}
	p, ok := r.Route(req)
	if !ok {
		t.Fatal("应当可达")
	}
	validatePath(t, r, req, p)

	// 与独立暴力解核对最优代价。
	best, ok := bruteOptimal(r, req)
	if !ok || best != p.Cost {
		t.Fatalf("最优代价不符: 库=%d 暴力=%d", p.Cost, best)
	}
	// 存在长度更短但折弯更多的路径，其代价应高于入选路径。
	slen, sturns, ok := bruteShortest(r, req)
	if !ok {
		t.Fatal("暴力解应当可达")
	}
	plen := p.Cost - 20*(len(p.Points)-2)
	if slen >= plen {
		t.Fatalf("应存在更短路径: 最短=%d 入选=%d", slen, plen)
	}
	if sc := slen + 20*sturns; sc <= p.Cost {
		t.Fatalf("更短路径代价 %d 应高于入选代价 %d", sc, p.Cost)
	}
	t.Logf("入选: 长度=%d 代价=%d; 更短路径: 长度=%d 转弯=%d 代价=%d",
		plen, p.Cost, slen, sturns, slen+20*sturns)
}
