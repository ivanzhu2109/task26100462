package orthroute

// dir is a unit cardinal direction of travel.
type dir int

const (
	dRight dir = iota
	dUp
	dLeft
	dDown
)

// delta returns the unit step for the direction.
func (d dir) delta() Point {
	switch d {
	case dRight:
		return Point{1, 0}
	case dUp:
		return Point{0, 1}
	case dLeft:
		return Point{-1, 0}
	default:
		return Point{0, -1}
	}
}

// opposite returns the reverse direction.
func (d dir) opposite() dir {
	switch d {
	case dRight:
		return dLeft
	case dUp:
		return dDown
	case dLeft:
		return dRight
	default:
		return dUp
	}
}

// insideStrict reports whether p lies strictly inside r.
func insideStrict(r Rect, p Point) bool {
	return r.Min.X < p.X && p.X < r.Max.X && r.Min.Y < p.Y && p.Y < r.Max.Y
}

// segmentBlocked reports whether the axis-aligned segment a-b contains any
// point strictly inside r (endpoints and mid-segment points included).
func segmentBlocked(r Rect, a, b Point) bool {
	if a.Y == b.Y {
		y := a.Y
		if y <= r.Min.Y || y >= r.Max.Y {
			return false
		}
		lo, hi := a.X, b.X
		if lo > hi {
			lo, hi = hi, lo
		}
		return lo < r.Max.X && hi > r.Min.X
	}
	if a.X == b.X {
		x := a.X
		if x <= r.Min.X || x >= r.Max.X {
			return false
		}
		lo, hi := a.Y, b.Y
		if lo > hi {
			lo, hi = hi, lo
		}
		return lo < r.Max.Y && hi > r.Min.Y
	}
	return false
}

// portSide determines which side of r the boundary point p sits on.
// ok is false when p is not on the boundary or lies on a corner.
func portSide(r Rect, p Point) (side Side, ok bool) {
	if p.X < r.Min.X || p.X > r.Max.X || p.Y < r.Min.Y || p.Y > r.Max.Y {
		return 0, false
	}
	onLeft := p.X == r.Min.X
	onRight := p.X == r.Max.X
	onTop := p.Y == r.Max.Y
	onBottom := p.Y == r.Min.Y
	if (onLeft || onRight) && (onTop || onBottom) {
		return 0, false // corner
	}
	switch {
	case onLeft:
		return SideLeft, true
	case onRight:
		return SideRight, true
	case onTop:
		return SideTop, true
	case onBottom:
		return SideBottom, true
	}
	return 0, false
}

// outwardNormal returns the travel direction leaving a port on side s.
func outwardNormal(s Side) dir {
	switch s {
	case SideLeft:
		return dLeft
	case SideRight:
		return dRight
	case SideTop:
		return dUp
	default:
		return dDown
	}
}
