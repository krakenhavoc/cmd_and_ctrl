package effects

import "testing"

// rhox_faithmender_test.go covers the card through the paths a game
// actually takes, which is what #482 was about. Batch 14 shipped the
// Faithmender CompletenessFull on a test that only ever called the
// public ChangePlayerLife sandbox verb — the life change a player makes
// by dragging their own counter — and every other way a life total
// moves skipped the CR 614 window entirely. The card was wrong on the
// table and right in its test.
//
// TestB14RhoxFaithmenderDoublesYourLifeGainOnly stays where it is: it
// pins the sandbox verb and the two AppliesTo boundaries (a loss is not
// a gain; an opponent's gain is not yours). What follows is the rest of
// the card.

// TestRhoxFaithmenderDoublesACatalogLifeGain is the repro from the
// issue, verbatim: Lightning Helix at an opponent with a Faithmender
// out gains SIX, not three. The gain goes through
// ChangePlayerLifeForEffect, which is the entry point every catalog
// GainLife and every drain uses.
func TestRhoxFaithmenderDoublesACatalogLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Rhox Faithmender", "Creature — Rhino Monk", b14RhoxFaithmenderOracle, 1, 5)
	mine, theirs := me.Life, opp.Life

	b11Helix(t, g, opp.ID)

	if me.Life != mine+6 {
		t.Errorf("Lightning Helix's gain of 3: %d → %d, want %d", mine, me.Life, mine+6)
	}
	if opp.Life != theirs-3 {
		t.Errorf("the Helix's DAMAGE is not a life-change event: %d → %d, want %d",
			theirs, opp.Life, theirs-3)
	}
}

// TestRhoxFaithmenderDoublesItsOwnLifelink — the half the 2026-09-16
// note on #482 found. CR 702.15b makes lifelink life gain, so the
// Faithmender's own printed lifelink is doubled; the card comment has
// claimed this since batch 14 and it only became true with #482.
func TestRhoxFaithmenderDoublesItsOwnLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rhox := b12Push(g, me.ID, "Rhox Faithmender", "Creature — Rhino Monk", b14RhoxFaithmenderOracle, 1, 5)
	if !hasEffectiveKeyword(t, g, rhox, "lifelink") {
		t.Fatal("the Faithmender has lifelink")
	}
	mine, theirs := me.Life, opp.Life

	attackWith(t, g, opp.ID, rhox)

	if me.Life != mine+2 {
		t.Errorf("1 power of lifelink doubled: %d → %d, want %d", mine, me.Life, mine+2)
	}
	if theirs-opp.Life != 1 {
		t.Errorf("the combat damage itself is not doubled: %d → %d, want %d",
			theirs, opp.Life, theirs-1)
	}
}

// TestTwoRhoxFaithmendersQuadruple — two applicable replacements on one
// event, and both of them are the same printed effect, so CR 616 has
// nothing to ask: x2 then x2 is x2 then x2 whichever Faithmender is
// named first. #792 made that window skip the prompt, so the gain
// lands inline and the player pays no clicks for it. Before #482 a
// paused life change landed through a second copy of the mutation and
// lost its source; this is the one path that used to exercise it.
func TestTwoRhoxFaithmendersQuadruple(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Rhox Faithmender", "Creature — Rhino Monk", b14RhoxFaithmenderOracle, 1, 5)
	b12Push(g, me.ID, "Rhox Faithmender", "Creature — Rhino Monk", b14RhoxFaithmenderOracle, 1, 5)
	mine := me.Life

	b11Helix(t, g, opp.ID)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("two copies of ONE replacement queued %d prompts, want 0 (#792)", len(g.PendingChoices))
	}
	if me.Life != mine+12 {
		t.Errorf("two Faithmenders on a gain of 3: %d → %d, want %d (3 × 2 × 2)", mine, me.Life, mine+12)
	}
}

// TestOneRhoxFaithmenderNeverPrompts — the other side of the CR 616
// rule. One applicable replacement applies inline, so the common case
// costs the player no clicks. Guards against a fix that routed the life
// change through the pipeline and started prompting on every Soul
// Warden trigger.
func TestOneRhoxFaithmenderNeverPrompts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Rhox Faithmender", "Creature — Rhino Monk", b14RhoxFaithmenderOracle, 1, 5)

	b11Helix(t, g, opp.ID)

	if len(g.PendingChoices) != 0 {
		t.Errorf("one life replacement queued %d prompts, want 0", len(g.PendingChoices))
	}
}
