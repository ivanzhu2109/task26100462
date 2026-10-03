package orthroute

import (
	"math/rand"
	"os"
	"testing"
)

// TestFindShortestLosesCase searches for a scenario where the
// minimum-length legal path is not the minimum-cost one.
// Run with: ORTH_SEARCH=1 go test -run TestFindShortestLosesCase
func TestFindShortestLosesCase(t *testing.T) {
	if os.Getenv("ORTH_SEARCH") == "" {
		t.Skip("set ORTH_SEARCH=1 to search for a demo case")
	}
	const max = 30
	for seed := int64(1); seed < 200000; seed++ {
		rng := rand.New(rand.NewSource(seed))
		for i := 0; i < 20; i++ {
			l, req := randomLayout(rng, max)
			cost, ok := bruteCost(l, req, max)
			if !ok {
				continue
			}
			length, turns, ok2 := bruteLex(l, req, max)
			if !ok2 {
				t.Fatalf("lex unreachable but cost reachable: %v %+v", l.Rects, req)
			}
			shortestCost := length + BendCost*turns
			if cost < shortestCost {
				t.Fatalf("FOUND seed=%d case=%d cost=%d shortestLen=%d shortestTurns=%d shortestCost=%d\nrects=%v\nreq=%+v",
					seed, i, cost, length, turns, shortestCost, l.Rects, req)
			}
		}
	}
	t.Log("no case found")
}
