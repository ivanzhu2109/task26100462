# orthroute

图表正交连线寻路库（Go）。给定一组轴对齐矩形节点的布局，在固定端口之间
生成绕开所有节点内部的直角连线，并按「总长度 + 20 × 折弯次数」的统一定价
选出最小代价路径。布局与计算结果只保存在内存中，无外部依赖。

## 规则

- 画布为 `[0, 10000] × [0, 10000]`（含边界），坐标均为整数。
- 节点为正宽高的轴对齐矩形，可重叠或相接；矩形已含业务留白。
- 每条请求指定源/目标节点及其边上的连接点（非角点、两点不同，允许同一节点）。
- 路径由水平/竖直的非零线段组成，相邻线段垂直，折点为整数坐标。
- 首段必须沿源点所在边的外法线离开，末段必须沿目标点所在边的内法线到达。
- 路径任何位置（含端点与线段中部）不得落在任一矩形的严格内部；
  允许沿矩形边界行走、经过角点以及边界重合形成的通道。
- 代价 = 各段长度之和 + 20 × 转弯次数；返回所有合法路径中的最小代价路径，
  不存在时返回 `ErrUnreachable`。不同连线独立求解，允许互相交叉。

## 用法

```go
layout := &orthroute.Layout{Rects: []orthroute.Rect{
    {Min: orthroute.Point{X: 2, Y: 2}, Max: orthroute.Point{X: 6, Y: 6}},
    {Min: orthroute.Point{X: 10, Y: 2}, Max: orthroute.Point{X: 14, Y: 6}},
}}
path, err := layout.Route(orthroute.Request{
    Source: orthroute.Port{Node: 0, At: orthroute.Point{X: 6, Y: 4}},
    Target: orthroute.Port{Node: 1, At: orthroute.Point{X: 10, Y: 4}},
})
// path 为折点序列，path.Cost() 为代价；err == orthroute.ErrUnreachable 表示不可达。

results := layout.RouteAll([]orthroute.Request{req1, req2}) // 批量，独立求解
```

## 实现概要

- `grid.go`：在画布边界、矩形边线、端口坐标及端口 ±1 直线上构建压缩
  网格图（端口 ±1 线用于强制首末段方向后的最优转弯点）；节点为不在任何
  矩形严格内部的网格交点，边为不穿越任何矩形严格内部的相邻连线。
- `search.go`：在 (节点, 朝向) 状态空间上跑 Dijkstra，首步只沿外法线，
  到达目标时朝向必须为内法线；禁止 180° 掉头（对应非法折线）。
- `geom.go`：严格内部判定与「整条线段」级别的穿越检测（宽度为 1 的矩形
  也能在线段中部拦截）。

## 测试

```sh
go test ./...        # 全部测试
go vet ./...
```

测试内容：

- 随机布局与独立暴力求解器（全网格 Dijkstra）逐一核对可达性与最优代价，
  默认 400 例，`ORTH_RANDOM_ITERS=5000 go test -run TestRandomAgainstBruteForce .`
  可加大规模；`TestDenseStress` 覆盖密集障碍。
- 固定用例：距离更短但折弯更多的路径不应入选（最短 18/6 弯/138，
  最优 20/4 弯/100）。
- 几何测试：重叠矩形、边界重合通道、贴边行走、宽度 1 矩形的线段中部
  拦截（验证整条线段而非仅折点）。
- 边界行为：画布边缘端口、密封不可达、同节点两端口、批量求解、非法请求。
- `ORTH_SEARCH=1 go test -run TestFindShortestLosesCase .` 可重新搜索
  「最短路径落选」的演示用例。
