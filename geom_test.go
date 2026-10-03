package orthroute

import "testing"

func TestInsideStrict(t *testing.T) {
	r := Rect{Min: Point{2, 2}, Max: Point{6, 8}}
	cases := []struct {
		p    Point
		want bool
	}{
		{Point{4, 5}, true},  // interior
		{Point{2, 5}, false}, // left edge
		{Point{6, 5}, false}, // right edge
		{Point{4, 2}, false}, // bottom edge
		{Point{4, 8}, false}, // top edge
		{Point{2, 2}, false}, // corner
		{Point{6, 8}, false}, // corner
		{Point{1, 5}, false}, // outside
		{Point{4, 9}, false}, // outside
	}
	for _, c := range cases {
		if got := insideStrict(r, c.p); got != c.want {
			t.Errorf("insideStrict(%v)=%v, want %v", c.p, got, c.want)
		}
	}
}

func TestSegmentBlocked(t *testing.T) {
	r := Rect{Min: Point{4, 4}, Max: Point{10, 10}}
	cases := []struct {
		name string
		a, b Point
		want bool
	}{
		{"through interior horizontal", Point{0, 6}, Point{12, 6}, true},
		{"through interior vertical", Point{6, 0}, Point{6, 12}, true},
		{"endpoint inside", Point{0, 6}, Point{6, 6}, true},
		{"along bottom edge", Point{0, 4}, Point{12, 4}, false},
		{"along top edge", Point{0, 10}, Point{12, 10}, false},
		{"along left edge", Point{4, 0}, Point{4, 12}, false},
		{"touching corner", Point{0, 0}, Point{4, 4}, false},
		{"clear above", Point{0, 11}, Point{12, 11}, false},
		{"clear left", Point{3, 0}, Point{3, 12}, false},
		{"reversed through interior", Point{12, 6}, Point{0, 6}, true},
		{"grazes corner only", Point{4, 0}, Point{0, 4}, false},
	}
	for _, c := range cases {
		if got := segmentBlocked(r, c.a, c.b); got != c.want {
			t.Errorf("%s: segmentBlocked(%v,%v)=%v, want %v", c.name, c.a, c.b, got, c.want)
		}
	}
}

// A segment can cross a thin rectangle's strict interior without any
// integer point of the segment being strictly inside: legality must be
// decided on the whole segment, not only on vertices.
func TestSegmentBlockedThinRectMidSegment(t *testing.T) {
	thin := Rect{Min: Point{5, 2}, Max: Point{6, 8}}
	a, b := Point{3, 5}, Point{8, 5}
	// No integer point on the segment is strictly inside.
	for x := a.X; x <= b.X; x++ {
		if insideStrict(thin, Point{x, a.Y}) {
			t.Fatalf("point (%d,5) unexpectedly strictly inside", x)
		}
	}
	if !segmentBlocked(thin, a, b) {
		t.Fatal("mid-segment crossing of a width-1 rect must be blocked")
	}
	// Same for a width-1 rect in the other axis.
	tall := Rect{Min: Point{2, 5}, Max: Point{8, 6}}
	if !segmentBlocked(tall, Point{4, 3}, Point{4, 9}) {
		t.Fatal("mid-segment crossing of a height-1 rect must be blocked")
	}
}

func TestPortSide(t *testing.T) {
	r := Rect{Min: Point{2, 2}, Max: Point{6, 8}}
	cases := []struct {
		p        Point
		wantSide Side
		wantOK   bool
	}{
		{Point{2, 5}, SideLeft, true},
		{Point{6, 5}, SideRight, true},
		{Point{4, 8}, SideTop, true},
		{Point{4, 2}, SideBottom, true},
		{Point{2, 2}, 0, false}, // corner
		{Point{6, 8}, 0, false}, // corner
		{Point{4, 5}, 0, false}, // interior
		{Point{1, 5}, 0, false}, // outside
		{Point{2, 9}, 0, false}, // outside beyond edge
	}
	for _, c := range cases {
		side, ok := portSide(r, c.p)
		if ok != c.wantOK || (ok && side != c.wantSide) {
			t.Errorf("portSide(%v)=(%v,%v), want (%v,%v)", c.p, side, ok, c.wantSide, c.wantOK)
		}
	}
}
