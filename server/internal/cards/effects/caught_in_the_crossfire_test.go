package effects

import "testing"

// caught_in_the_crossfire_test.go — Spree proof card #4 (CR 702.172a,
// ADR 0065's 2026-09-23 amendment): both bullets are UNTARGETED mass-
// damage sweeps, proving Spree composes with a mode that opens no
// targeting prompt at all.

const caughtInTheCrossfireOracle = "d262f3e1-b5bc-4dab-86c2-cd89c2419e54"

func TestCaughtInTheCrossfireHitsOnlyOutlawsAlone(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rogue := b12Creature(g, opp.ID, "Their Rogue", "Creature — Human Rogue", 2, 2)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castModal(t, g, "Caught in the Crossfire", "Instant", caughtInTheCrossfireOracle,
		[]int{0}, nil)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rogue) {
		t.Error("the outlaw bullet should have killed the 2/2 Rogue")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the outlaw bullet should not touch a non-outlaw creature")
	}
}

func TestCaughtInTheCrossfireHitsOnlyNonOutlawsAlone(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rogue := b12Creature(g, opp.ID, "Their Rogue", "Creature — Human Rogue", 2, 2)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castModal(t, g, "Caught in the Crossfire", "Instant", caughtInTheCrossfireOracle,
		[]int{1}, nil)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(rogue) {
		t.Error("the non-outlaw bullet should not touch an outlaw creature")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the non-outlaw bullet should have killed the 2/2 Bear")
	}
}

// Both bullets chosen together is "deal 2 damage to each creature" —
// outlaw and non-outlaw alike.
func TestCaughtInTheCrossfireBothBulletsHitEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rogue := b12Creature(g, opp.ID, "Their Rogue", "Creature — Human Rogue", 2, 2)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castModal(t, g, "Caught in the Crossfire", "Instant", caughtInTheCrossfireOracle,
		[]int{0, 1}, nil)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rogue) {
		t.Error("both bullets together should have killed the outlaw too")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("both bullets together should have killed the non-outlaw too")
	}
}
