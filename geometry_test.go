package orthrouter

import "testing"

// 整段合法性：折点合法但线段中部穿过矩形内部，必须判非法。
func TestSegClearWholeSegment(t *testing.T) {
	rects := []Rect{{2, 2, 4, 4}}
	cases := []struct {
		a, b Point
		want bool
	}{
		{Point{1, 3}, Point{5, 3}, false}, // 两端在界外，中部穿过内部
		{Point{1, 2}, Point{5, 2}, true},  // 沿顶边行走
		{Point{2, 1}, Point{2, 5}, true},  // 沿左边行走
		{Point{1, 4}, Point{5, 4}, true},  // 沿底边行走
		{Point{3, 1}, Point{3, 2}, true},  // 端点在边上，线段不进入内部
		{Point{3, 2}, Point{3, 5}, false}, // 从边进入内部
		{Point{1, 1}, Point{1, 5}, true},  // 完全在外
	}
	for _, c := range cases {
		if got := segClear(rects, c.a, c.b); got != c.want {
			t.Errorf("segClear(%v,%v)=%v, 期望 %v", c.a, c.b, got, c.want)
		}
	}
}

// 重叠矩形：一个矩形的边界落在另一矩形内部的部分禁止通过。
func TestSegClearOverlappingRects(t *testing.T) {
	rects := []Rect{{0, 0, 4, 4}, {2, 2, 6, 6}}
	// 沿矩形1右边 x=4 行走：y∈(2,4) 段位于矩形2内部，禁止。
	if segClear(rects, Point{4, 0}, Point{4, 8}) {
		t.Error("穿过重叠区的边界段应判非法")
	}
	// 仅走不重叠的部分则允许。
	if !segClear(rects, Point{4, 0}, Point{4, 2}) {
		t.Error("未进入矩形2内部的边界段应合法")
	}
	// x=4 是矩形1的边但穿过矩形2内部，整段禁止。
	if segClear(rects, Point{4, 4}, Point{4, 8}) {
		t.Error("矩形1边界落在矩形2内部的段应判非法")
	}
	// x=6 是矩形2的边，允许行走。
	if !segClear(rects, Point{6, 4}, Point{6, 8}) {
		t.Error("矩形2边界上的段应合法")
	}
}

// 贴边通道：两个矩形边界重合形成的通道允许通过。
func TestTouchingRectsChannel(t *testing.T) {
	rects := []Rect{{0, 0, 4, 2}, {0, 2, 4, 4}} // 共享边界 y=2
	if !segClear(rects, Point{0, 2}, Point{4, 2}) {
		t.Error("沿共享边界行走应合法")
	}
}

// 路由级贴边通道：唯一通路是两墙之间的缝隙，路径贴着墙边与障碍边穿过。
func TestRouteThroughEdgeChannel(t *testing.T) {
	rects := []Rect{
		{0, 0, 10, 5},  // 上墙
		{0, 7, 10, 10}, // 下墙
		{0, 5, 2, 7},   // 节点 A（堵住通道左端）
		{8, 5, 10, 7},  // 节点 B（堵住通道右端）
		{4, 5, 6, 7},   // 障碍：堵住通道中部，只留 y=5、y=7 贴边缝
	}
	r := NewRouter(0, 0, 10, 10, rects, 20)
	req := Request{
		Source: Endpoint{Node: 2, Point: Point{2, 6}},
		Target: Endpoint{Node: 3, Point: Point{8, 6}},
	}
	p, ok := r.Route(req)
	if !ok {
		t.Fatal("贴边通道应当可达")
	}
	validatePath(t, r, req, p)
	best, ok := bruteOptimal(r, req)
	if !ok || best != p.Cost {
		t.Fatalf("最优代价不符: 库=%d 暴力=%d", p.Cost, best)
	}
	// 路径必须贴边：存在 y=5 或 y=7 的段。
	hug := false
	for i := 0; i+1 < len(p.Points); i++ {
		a, b := p.Points[i], p.Points[i+1]
		if a.Y == b.Y && (a.Y == 5 || a.Y == 7) && abs(b.X-a.X) >= 2 {
			hug = true
		}
	}
	if !hug {
		t.Fatalf("路径未贴边穿过通道: %v", p.Points)
	}
}

// 重叠矩形导致连接点落在另一矩形内部：不可达。
func TestEndpointInsideOtherRect(t *testing.T) {
	rects := []Rect{
		{0, 0, 6, 4}, // 节点 A
		{2, 1, 8, 3}, // 节点 B，覆盖 A 的右边中段
		{8, 0, 10, 1},
	}
	r := NewRouter(0, 0, 10, 10, rects, 20)
	// A 右边上的点 (6,2) 严格位于 B 内部。
	req := Request{
		Source: Endpoint{Node: 0, Point: Point{6, 2}},
		Target: Endpoint{Node: 2, Point: Point{8, 0}},
	}
	if p, ok := r.Route(req); ok {
		t.Fatalf("端点在另一矩形内部应不可达, 得到 %v", p.Points)
	}
	if _, ok := bruteOptimal(r, req); ok {
		t.Fatal("暴力解也应不可达")
	}
}

// 目标节点被完整封闭：不可达。
func TestUnreachableEnclosed(t *testing.T) {
	rects := []Rect{
		{0, 0, 2, 2},   // 节点 A
		{8, 8, 12, 12}, // 节点 B，被大矩形完整包围
		{4, 4, 16, 16}, // 大矩形
	}
	r := NewRouter(0, 0, 20, 20, rects, 20)
	req := Request{
		Source: Endpoint{Node: 0, Point: Point{1, 2}},
		Target: Endpoint{Node: 1, Point: Point{10, 8}},
	}
	if _, ok := r.Route(req); ok {
		t.Fatal("被封闭的节点应不可达")
	}
	if _, ok := bruteOptimal(r, req); ok {
		t.Fatal("暴力解也应不可达")
	}
}

// 同一节点上的两个连接点。
func TestSameNode(t *testing.T) {
	rects := []Rect{{4, 4, 6, 6}}
	r := NewRouter(0, 0, 10, 10, rects, 20)
	req := Request{
		Source: Endpoint{Node: 0, Point: Point{4, 5}},
		Target: Endpoint{Node: 0, Point: Point{6, 5}},
	}
	p, ok := r.Route(req)
	if !ok {
		t.Fatal("同一节点两连接点应可达")
	}
	validatePath(t, r, req, p)
	best, _ := bruteOptimal(r, req)
	if best != p.Cost {
		t.Fatalf("最优代价不符: 库=%d 暴力=%d", p.Cost, best)
	}
}

// 方向约束：首段不得先沿节点边滑行。
func TestNoSlidingAlongEdge(t *testing.T) {
	// 源点在 A 左边，目标在正上方；若允许先沿边滑行到角再转弯会更短，
	// 但首段必须沿外法线（-x）离开。
	rects := []Rect{{4, 4, 6, 8}}
	r := NewRouter(0, 0, 10, 10, rects, 20)
	req := Request{
		Source: Endpoint{Node: 0, Point: Point{4, 6}},
		Target: Endpoint{Node: 0, Point: Point{5, 4}},
	}
	p, ok := r.Route(req)
	if !ok {
		t.Fatal("应当可达")
	}
	validatePath(t, r, req, p)
	if p.Points[1].X != 4-1 && p.Points[1].Y != 6 {
		t.Fatalf("首段应水平向左离开: %v", p.Points)
	}
	if d := segDir(p.Points[0], p.Points[1]); d != 1 {
		t.Fatalf("首段方向应为 -x, 得到 %v", p.Points)
	}
}
