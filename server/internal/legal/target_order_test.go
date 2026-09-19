package legal_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// target_order_test.go — #687, ADR 0033 §1's threat ordering.
//
// The cap (Options.MaxExpansionPerSource) is spent in candidate
// order, so without an ordering the biggest threat on the table can
// simply be absent from the moves a bot is offered, and no policy can
// then pick it. Options.OrderTargets is the hook that fixes it.

// targetIDsOf lists the target instance IDs of every cast move a
// source produced, in the order they were offered.
func targetIDsOf(t *testing.T, moves []legal.Move, src uuid.UUID) []string {
	t.Helper()
	var out []string
	for _, p := range castPayloadsOf(t, moves, src) {
		for _, tg := range p.Targets {
			out = append(out, tg.ID)
		}
	}
	return out
}

// contains reports whether an ID appears anywhere in the list.
func contains(ids []string, want uuid.UUID) bool {
	for _, id := range ids {
		if id == want.String() {
			return true
		}
	}
	return false
}

// boardOfTwenty seeds twenty opposing creatures and returns the one
// planted LAST, which is the one the unordered walk drops off the end
// of the cap.
func boardOfTwenty(g *game.Game, owner *game.Player) (all []uuid.UUID, biggest uuid.UUID) {
	for i := 0; i < 19; i++ {
		all = append(all, battlefieldCard(g, owner, creature(fmt.Sprintf("Squire %d", i), "{1}{W}", 1, 1)))
	}
	biggest = battlefieldCard(g, owner, creature("Blightsteel Colossus", "{12}", 11, 11))
	return append(all, biggest), biggest
}

// Without a hook the enumeration is exactly what it was: candidates
// in engine order, and whatever falls past the cap is dropped. This
// is the control — it states the defect the ordering exists to fix,
// so the test below is measuring something.
func TestWithoutAnOrderingTheBiggestThreatFallsOffTheCap(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 1, "Mountain")
	_, biggest := boardOfTwenty(g, other)

	boltID := handCard(seat, bolt())
	moves := legal.EnumerateFor(g, seat.ID)
	ids := targetIDsOf(t, moves, boltID)
	if len(ids) == 0 {
		t.Fatalf("Lightning Bolt was not offered at all: %v", labels(moves))
	}
	if len(ids) >= 20 {
		t.Fatalf("the cap did not bite (%d offers); this control asserts nothing", len(ids))
	}
	if contains(ids, biggest) {
		t.Errorf("the unordered walk surfaced the 11/11 by luck — the fixture no longer "+
			"demonstrates the defect #687 is about, and the ordering test below is "+
			"measuring nothing (offers: %d)", len(ids))
	}
}

// With a hook, the biggest threat is offered — and it is offered
// FIRST, because the cap is spent in the order the hook asked for.
func TestOrderTargetsPutsTheBiggestThreatInsideTheCap(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 1, "Mountain")
	all, biggest := boardOfTwenty(g, other)

	boltID := handCard(seat, bolt())
	// A stand-in for the policy's scorer: power, read off the
	// authoritative board because this test is about the HOOK rather
	// than about the heuristic (aiseat/heuristic has that one).
	power := map[uuid.UUID]float64{}
	g.ReadSnapshot(func() {
		for _, id := range all {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					power[id] = float64(g.Battlefield.Cards[i].Power)
				}
			}
		}
	})

	moves := legal.EnumerateForWithOptions(g, seat.ID, legal.Options{
		OrderTargets: func(c legal.TargetCandidate) float64 { return power[c.ID] },
	})
	ids := targetIDsOf(t, moves, boltID)
	if len(ids) == 0 {
		t.Fatalf("Lightning Bolt was not offered at all: %v", labels(moves))
	}
	if !contains(ids, biggest) {
		t.Errorf("the 11/11 is not among the %d offered targets — the cap dropped the threat", len(ids))
	}
	if ids[0] != biggest.String() {
		t.Errorf("the first offered target is %s, want the 11/11 %s", ids[0], biggest)
	}
	dispatchAll(t, g, seat.ID, moves)
}

// The ordering may not change WHICH targets are legal, only which
// reach the cap: with a cap big enough for the whole board, the two
// enumerations offer the same set.
func TestOrderTargetsChangesOrderNotLegality(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 1, "Mountain")
	for i := 0; i < 4; i++ {
		battlefieldCard(g, other, creature(fmt.Sprintf("Squire %d", i), "{1}{W}", i+1, 1))
	}

	boltID := handCard(seat, bolt())
	opts := legal.Options{MaxExpansionPerSource: 64}
	plain := targetIDsOf(t, legal.EnumerateForWithOptions(g, seat.ID, opts), boltID)

	opts.OrderTargets = func(c legal.TargetCandidate) float64 {
		// Reverse the engine's order as hard as possible.
		return -float64(c.ID.ID())
	}
	ordered := targetIDsOf(t, legal.EnumerateForWithOptions(g, seat.ID, opts), boltID)

	if len(plain) != len(ordered) {
		t.Fatalf("the ordering changed the number of offers: %d vs %d", len(plain), len(ordered))
	}
	set := map[string]bool{}
	for _, id := range plain {
		set[id] = true
	}
	for _, id := range ordered {
		if !set[id] {
			t.Errorf("the ordering invented a target the unordered enumeration did not have: %s", id)
		}
	}
}

// Determinism: the same board and the same hook give the same move
// list, every time. Equal scores keep the engine's own order, which
// is what makes a decision log replayable.
func TestOrderTargetsIsStableAndDeterministic(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 1, "Mountain")
	for i := 0; i < 6; i++ {
		battlefieldCard(g, other, creature(fmt.Sprintf("Squire %d", i), "{1}{W}", 1, 1))
	}
	handCard(seat, bolt())

	// Every candidate scores the same, so ONLY the stability of the
	// sort decides the answer.
	opts := legal.Options{OrderTargets: func(legal.TargetCandidate) float64 { return 1 }}
	first, err := json.Marshal(legal.EnumerateForWithOptions(g, seat.ID, opts))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for i := 0; i < 5; i++ {
		again, err := json.Marshal(legal.EnumerateForWithOptions(g, seat.ID, opts))
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if string(again) != string(first) {
			t.Fatalf("enumeration %d differs from the first — the ordering is not deterministic", i)
		}
	}

	// And the flat ordering is byte-identical to no ordering at all.
	plain, err := json.Marshal(legal.EnumerateFor(g, seat.ID))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(plain) != string(first) {
		t.Error("an all-equal ordering changed the move list; the sort is not stable")
	}
}
