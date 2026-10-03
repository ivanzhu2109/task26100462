package orthroute_test

import (
	"fmt"

	"orthroute"
)

func ExampleLayout_Route() {
	layout := &orthroute.Layout{Rects: []orthroute.Rect{
		{Min: orthroute.Point{X: 2, Y: 2}, Max: orthroute.Point{X: 6, Y: 6}},
		{Min: orthroute.Point{X: 10, Y: 2}, Max: orthroute.Point{X: 14, Y: 6}},
	}}
	path, err := layout.Route(orthroute.Request{
		Source: orthroute.Port{Node: 0, At: orthroute.Point{X: 6, Y: 4}},
		Target: orthroute.Port{Node: 1, At: orthroute.Point{X: 10, Y: 4}},
	})
	fmt.Println(path, err, path.Cost())
	// Output: [{6 4} {10 4}] <nil> 4
}
