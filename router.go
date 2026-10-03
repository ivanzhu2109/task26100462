// Package orthrouter 提供图表正交（直角）连线寻路能力。
// 上层提交节点布局后，可针对该布局请求一条或一批绕开节点的连线。
package orthrouter

// Point 为整数坐标点。
type Point struct {
	X, Y int
}

// Rect 为轴对齐矩形（含边界），宽高为正。
type Rect struct {
	X1, Y1, X2, Y2 int
}

// Endpoint 指定某条边上的一个连接点。
type Endpoint struct {
	Node  int   // 节点在布局中的下标
	Point Point // 连接点，位于该节点某条边上且非角点
}

// Request 描述一条连线请求。
type Request struct {
	Source, Target Endpoint
}

// Path 为求得的路径：折点序列与总代价。
type Path struct {
	Points []Point
	Cost   int
}

// Result 为单条请求的结果；OK 为 false 表示不可达。
type Result struct {
	Path Path
	OK   bool
}

// Router 保存一次提交的节点布局，所有状态仅在内存中。
type Router struct {
	minX, minY, maxX, maxY int
	rects                  []Rect
	turnCost               int
}

// NewRouter 提交节点布局。画布为 [minX,maxX]×[minY,maxY]（含边界）。
// turnCost 为每次转弯的代价，传 0 时使用默认值 20。
func NewRouter(minX, minY, maxX, maxY int, rects []Rect, turnCost int) *Router {
	if turnCost <= 0 {
		turnCost = 20
	}
	rs := make([]Rect, len(rects))
	for i, rc := range rects {
		if rc.X1 > rc.X2 {
			rc.X1, rc.X2 = rc.X2, rc.X1
		}
		if rc.Y1 > rc.Y2 {
			rc.Y1, rc.Y2 = rc.Y2, rc.Y1
		}
		rs[i] = rc
	}
	return &Router{minX: minX, minY: minY, maxX: maxX, maxY: maxY, rects: rs, turnCost: turnCost}
}

// Route 求解单条连线，不可达时 ok 为 false。
func (r *Router) Route(req Request) (Path, bool) {
	if req.Source.Node < 0 || req.Source.Node >= len(r.rects) ||
		req.Target.Node < 0 || req.Target.Node >= len(r.rects) {
		return Path{}, false
	}
	pts, cost, ok := dijkstra(buildGrid(r, req), r, req)
	if !ok {
		return Path{}, false
	}
	return Path{Points: pts, Cost: cost}, true
}

// RouteAll 批量求解，各条连线相互独立。
func (r *Router) RouteAll(reqs []Request) []Result {
	out := make([]Result, len(reqs))
	for i, req := range reqs {
		p, ok := r.Route(req)
		out[i] = Result{Path: p, OK: ok}
	}
	return out
}
