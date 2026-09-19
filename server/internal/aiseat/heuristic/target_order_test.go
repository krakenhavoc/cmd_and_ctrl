package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// target_order_test.go — #687. The heuristic ranks the enumerator's
// candidate targets with the SAME scoring it ranks everything else
// with, so the part of the board that survives the expansion cap is
// the part the policy would have wanted.

// mustUUID parses a fixture's wire instance ID. The hook takes UUIDs
// because `legal` speaks in them; the view speaks in strings.
func mustUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return id
}

// orderOf prices a set of candidates through the policy's hook.
func orderOf(t *testing.T, p *heuristic.Policy, v protocol.GameView, me int) legal.TargetOrder {
	t.Helper()
	f := p.TargetOrder(aiseat.Input{View: v, Seat: seatID(me)})
	if f == nil {
		t.Fatal("the heuristic returned no ordering")
	}
	return f
}

// A bigger creature outranks a smaller one, whoever controls it. The
// ordering is by IMPORTANCE rather than by desirability for a
// particular spell: its only power is to stop a target being dropped
// off the cap, and dropping the board's biggest permanent is wrong
// for every spell.
func TestTargetOrderRanksPermanentsByTheirBoardValue(t *testing.T) {
	small := creature(cardID(1), 1, "Squire", 1, 1)
	big := creature(cardID(2), 1, "Colossus", 11, 11)
	mine := creature(cardID(3), 0, "Bear", 2, 2)
	v := newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(small, big, mine),
	)

	order := orderOf(t, heuristic.New(), v, 0)
	sv := order(legal.TargetCandidate{ID: mustUUID(t, small.InstanceID)})
	bv := order(legal.TargetCandidate{ID: mustUUID(t, big.InstanceID)})
	mv := order(legal.TargetCandidate{ID: mustUUID(t, mine.InstanceID)})
	if bv <= sv {
		t.Errorf("the 11/11 scored %v and the 1/1 scored %v", bv, sv)
	}
	if mv <= sv {
		t.Errorf("a 2/2 I control scored %v, below the opponent's 1/1 at %v — "+
			"the ordering is by value, not by ownership", mv, sv)
	}
}

// A seat is priced by Weights.Threat, the same function the attack
// rotation ranks opponents with: the seat closest to dying with the
// most board is the first target offered.
func TestTargetOrderRanksSeatsByThreat(t *testing.T) {
	v := newView(
		[]protocol.PlayerView{
			newSeat(0),
			newSeat(1, withLife(12), withHandCount(7)),
			newSeat(2, withLife(40)),
		},
		withBattlefield(creature(cardID(1), 1, "Colossus", 11, 11)),
	)

	order := orderOf(t, heuristic.New(), v, 0)
	leader := order(legal.TargetCandidate{ID: seatID(1), Player: true})
	quiet := order(legal.TargetCandidate{ID: seatID(2), Player: true})
	if leader <= quiet {
		t.Errorf("the seat at 12 life with a board and a full hand scored %v, "+
			"the untouched empty seat %v", leader, quiet)
	}
}

// Anything that is not a seat and not a battlefield permanent scores
// zero, which keeps the engine's own candidate order for it. Stated
// as a test because "zero" is a decision (see target_order.go) rather
// than a gap.
func TestTargetOrderScoresOffBoardCandidatesZero(t *testing.T) {
	onStack := spell(cardID(9), 1, "Lightning Bolt", "{R}")
	v := newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withStack(onStack),
	)
	order := orderOf(t, heuristic.New(), v, 0)
	if got := order(legal.TargetCandidate{ID: mustUUID(t, onStack.InstanceID)}); got != 0 {
		t.Errorf("a spell on the stack scored %v, want 0", got)
	}
}

// The heuristic is the TargetOrderer the runner looks for. A
// compile-time assertion lives in target_order.go; this is the
// runtime half, because a type assertion that quietly answers false
// restores exactly the defect #687 describes.
func TestHeuristicIsATargetOrderer(t *testing.T) {
	var p aiseat.Policy = heuristic.New()
	if _, ok := p.(aiseat.TargetOrderer); !ok {
		t.Fatal("the heuristic is no longer an aiseat.TargetOrderer; " +
			"the runner's assertion will silently stop ordering")
	}
}
