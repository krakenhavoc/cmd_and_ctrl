package game

import (
	"testing"

	"github.com/google/uuid"
)

// flashback_test.go — S29, the mechanic half. The cast path itself is
// pinned in cast_zones_test.go; what this file is about is CR
// 702.34a's second sentence, which is the one that keeps flashback
// from being a one-card infinite:
//
//	"If the flashback cost was paid, exile this card instead of
//	 putting it anywhere else ANY TIME it would leave the stack."
//
// "Any time" is the whole test plan. An implementation that exiles
// the card at the end of the resolution passes the happy-path test
// and is wrong on both of the other two exits — a fizzle and a
// counter — and wrong in the player's favour, because the card comes
// back to the graveyard to be flashed back again.

func resolveTop(t *testing.T, g *Game) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.resolveTopOfStackLocked(); err != nil {
			t.Errorf("resolveTopOfStackLocked: %v", err)
		}
	})
}

// flashbackOffer is the constructor's shape, spelled out here so the
// game package can test the mechanic without importing the effects
// package.
func flashbackOffer(cost string) AlternativeCost {
	return AlternativeCost{
		Key:                 "flashback",
		Label:               "Flashback " + cost,
		ManaCost:            cost,
		FromZone:            ZoneGraveyard,
		ExileOnLeavingStack: true,
	}
}

// seedFlashbackCard wires the catalog for a flashback sorcery and
// puts one in the active seat's graveyard at a main phase.
func seedFlashbackCard(t *testing.T, g *Game, me *Player, oracle string) uuid.UUID {
	t.Helper()
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	withCatalogAlternativeCosts(t, altCostFor(oracle, flashbackOffer("{2}{R}")))
	return looterInGraveyard(t, g, me, oracle)
}

// The happy path: cast, resolve, and the card is in exile rather
// than back in the graveyard where it started.
func TestFlashbackSpellIsExiledOnResolution(t *testing.T) {
	const oracle = "test-looting"
	g := newActiveGame(t)
	me := g.Seats[0]
	id := seedFlashbackCard(t, g, me, oracle)

	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	resolveTop(t, g)
	if me.Graveyard.Contains(id) {
		t.Errorf("a flashed-back spell returned to the graveyard — it can be flashed back again")
	}
	if !g.Exile.Contains(id) {
		t.Errorf("a flashed-back spell did not reach exile")
	}
}

// The same card cast for its printed cost out of hand goes to the
// graveyard as usual. The replacement is a property of the COST that
// was paid, not of the card.
func TestPrintedCostCastOfAFlashbackCardStillGoesToTheGraveyard(t *testing.T) {
	const oracle = "test-looting"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	withCatalogAlternativeCosts(t, altCostFor(oracle, flashbackOffer("{2}{R}")))
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Looting", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{R}"
	c.OracleID = oracle
	me.Hand.PushTop(c)

	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("printed-cost cast: %v", err)
	}
	resolveTop(t, g)
	if !me.Graveyard.Contains(c.InstanceID) {
		t.Errorf("a hand cast of a flashback card did not reach the graveyard")
	}
	if g.Exile.Contains(c.InstanceID) {
		t.Errorf("a hand cast of a flashback card was exiled")
	}
}

// Exit two: the spell fizzles. Every target became illegal, so it is
// "countered by game rules" and leaves the stack without resolving —
// and CR 702.34a still applies.
func TestFlashbackSpellIsExiledWhenItFizzles(t *testing.T) {
	const oracle = "test-bolt"
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	withCatalogAlternativeCosts(t, altCostFor(oracle, flashbackOffer("{2}{R}")))
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Bolt", me.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{R}"
	c.OracleID = oracle
	me.Graveyard.PushTop(c)

	// A player target that leaves the game before resolution is the
	// cheapest way to make every target illegal.
	targets := []TargetRef{{Kind: TargetPlayer, ID: them.ID}}
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Targets: targets,
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	if err := g.Concede(them.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	resolveTop(t, g)
	if me.Graveyard.Contains(c.InstanceID) {
		t.Errorf("a fizzled flashback spell returned to the graveyard")
	}
	if !g.Exile.Contains(c.InstanceID) {
		t.Errorf("a fizzled flashback spell did not reach exile")
	}
}

// Exit three, and the one that separates a replacement effect from
// an exile appended to the resolution: a counter that names a
// destination. Hinder puts the spell on top of its owner's library;
// CR 614 replaces that destination with exile, so the spell does not
// come back.
func TestFlashbackSpellIsExiledOverACountersDestination(t *testing.T) {
	const oracle = "test-looting"
	g := newActiveGame(t)
	me := g.Seats[0]
	id := seedFlashbackCard(t, g, me, oracle)

	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	if err := g.CounterSpell(id, &ZoneRef{Kind: ZoneLibrary, Owner: me.ID}); err != nil {
		t.Fatalf("CounterSpell: %v", err)
	}
	if me.Library.Contains(id) {
		t.Errorf("Hinder beat the flashback replacement — the spell went to the library")
	}
	if me.Graveyard.Contains(id) {
		t.Errorf("a countered flashback spell returned to the graveyard")
	}
	if !g.Exile.Contains(id) {
		t.Errorf("a countered flashback spell did not reach exile")
	}
}

// Once exiled, the card is inert. Exile carries no grant and the
// card declares only the graveyard, so there is no second cast — the
// property the whole replacement exists to guarantee.
func TestFlashbackSpellInExileCannotBeCastAgain(t *testing.T) {
	const oracle = "test-looting"
	g := newActiveGame(t)
	me := g.Seats[0]
	id := seedFlashbackCard(t, g, me, oracle)

	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	resolveTop(t, g)
	err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile", AlternativeCost: "flashback"})
	if err != ErrNoPlayPermission {
		t.Fatalf("recast from exile: got %v, want ErrNoPlayPermission", err)
	}
}
