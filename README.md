# orthrouter

图表正交连线寻路库（Go）。上层提交节点布局后，可针对该布局请求一条或一批
绕开节点的直角连线，按「总长度 + 转弯代价 × 转弯次数」选出最优路径。
节点布局与计算结果仅保存在内存中。

## 用法

```go
r := orthrouter.NewRouter(0, 0, 10000, 10000, rects, 20) // 画布含边界；转弯代价传 0 用默认 20
path, ok := r.Route(orthrouter.Request{
    Source: orthrouter.Endpoint{Node: 0, Point: orthrouter.Point{X: 2, Y: 3}},
    Target: orthrouter.Endpoint{Node: 1, Point: orthrouter.Point{X: 8, Y: 7}},
})
// 或批量：results := r.RouteAll(reqs)（各条独立求解，允许相互交叉）
```

- `Rect{X1,Y1,X2,Y2}`：轴对齐矩形（含边界），宽高为正，整数坐标。
- `Endpoint{Node, Point}`：连接点位于该节点某条边上且非角点。
- 返回 `Path{Points, Cost}`：`Points` 为折点序列，`Cost = 总长度 + 20×转弯数`；
  不可达时 `ok == false`。

## 规则

- 路径由水平/竖直非零线段组成，折点为整数坐标，全部留在画布内。
- 首段沿源边外法线离开，末段沿目标边内法线到达（不允许先沿边滑行）。
- 路径任意位置不得进入任一矩形的严格内部；允许沿边界行走、经过角点、
  穿过矩形边界重合形成的通道。

## 方法

在扩展 Hanan 网格（矩形边线及其 ±1 线、连接点及其 ±1 线、画布边界）上
以 (网格点, 进入方向) 为状态做 Dijkstra；±1 线覆盖「禁止滑行」的首末段
与绕角 U 形回转的最优折点。线段合法性按整条开线段与开矩形相交判定，
并预计算相邻网格线通行表加速。

## 测试

```sh
go test ./...
```

- `router_test.go`：独立暴力参考解（完整整数网格逐点 Dijkstra）、路径合法性
  校验，含「距离更短但折弯更多而不应入选」的对比场景。
- `geometry_test.go`：重叠矩形、贴边通道、整段（而非仅折点）合法性、
  不可达、同节点、方向约束。
- `random_test.go` / `stress_test.go`：随机小画布（含高密度重叠）与暴力解
  核对可达性与最优代价。
- `api_test.go`：批量接口一致性、万级画布百节点性能 sanity。
