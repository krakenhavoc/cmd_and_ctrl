package game

import (
	"testing"

	"github.com/google/uuid"
)

// planeswalker_test.go covers the printed-loyalty stamp (CR 306.5b)
// and its interaction with the 0-loyalty state-based action
// (CR 704.5i).
//
// These pin issue #274: a planeswalker cast from hand resolved onto
// the battlefield with no loyalty counters and was immediately moved
// to its owner's graveyard by 704.5i. Loyalty was only ever stamped
// by the card-effect catalog's ETB hook (effects/wire.go), so EVERY
// planeswalker outside the opt-in catalog was unplayable. Starting
// loyalty is printed card data, not card-effect data, so the stamp
// belongs in the game package and runs whether or not a catalog is
// wired in at all.
//
// Most of these tests deliberately register NO effect hooks — the
// game package's hook vars stay nil unless a test sets them, which
// is exactly the "non-catalog card, no catalog compiled in" shape.

// pushPlaneswalkerToHand seeds a planeswalker in hand with the
// printed starting loyalty the deck importer stamps from Scryfall.
func pushPlaneswalkerToHand(p *Player, name string, loyalty int) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = "Legendary Planeswalker — Teferi"
	c.StartingLoyalty = loyalty
	p.Hand.PushTop(c)
	return c.InstanceID
}

// battlefieldCard finds a card on the battlefield, or fails.
func battlefieldCard(t *testing.T, g *Game, id uuid.UUID) *Card {
	t.Helper()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	t.Fatalf("card %s is not on the battlefield", id)
	return nil
}

// TestNonCatalogPlaneswalkerEntersWithPrintedLoyalty is the #274
// regression test. Teferi, Time Raveler is not in the catalog; he
// must still enter the battlefield with his printed 4 loyalty and
// survive the SBA sweep.
func TestNonCatalogPlaneswalkerEntersWithPrintedLoyalty(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	pwID := pushPlaneswalkerToHand(caster, "Teferi, Time Raveler", 4)

	if err := g.CastSpell(caster.ID, pwID, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	// Both seats pass — the spell resolves, then the SBA sweep runs
	// at the next priority-grant boundary.
	_ = g.PassPriority()
	_ = g.PassPriority()

	if caster.Graveyard.Contains(pwID) {
		t.Fatalf("planeswalker went to the graveyard instead of the battlefield (issue #274)")
	}
	pw := battlefieldCard(t, g, pwID)
	if got := pw.Counters[CounterLoyalty]; got != 4 {
		t.Fatalf("loyalty counters = %d, want 4", got)
	}
}

// TestPlaneswalkerWithoutPrintedLoyaltyStillDies keeps the 704.5i
// SBA honest: the fix must not become "planeswalkers never die". A
// card with no printed loyalty and no catalog entry has nothing to
// stamp, so 704.5i still moves it to the graveyard.
func TestPlaneswalkerWithoutPrintedLoyaltyStillDies(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	pwID := pushPlaneswalkerToHand(caster, "Loyalty-less Walker", 0)

	if err := g.CastSpell(caster.ID, pwID, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	_ = g.PassPriority()
	_ = g.PassPriority()

	if !caster.Graveyard.Contains(pwID) {
		t.Errorf("704.5i did not move a 0-loyalty planeswalker to the graveyard")
	}
}

// TestCatalogStartingLoyaltyFallsBackForUnprintedCards proves the
// catalog's Spec.StartingLoyalty still works for a card whose
// game.Card carries no printed loyalty — tokens, test fixtures, and
// any card the deck importer could not resolve type data for.
func TestCatalogStartingLoyaltyFallsBackForUnprintedCards(t *testing.T) {
	prev := CatalogStartingLoyalty
	CatalogStartingLoyalty = func(oracleID string) int {
		if oracleID == "catalog-walker" {
			return 3
		}
		return 0
	}
	t.Cleanup(func() { CatalogStartingLoyalty = prev })

	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	pwID := pushPlaneswalkerToHand(caster, "The Wandering Emperor", 0)
	for i := range caster.Hand.Cards {
		if caster.Hand.Cards[i].InstanceID == pwID {
			caster.Hand.Cards[i].OracleID = "catalog-walker"
			break
		}
	}

	if err := g.CastSpell(caster.ID, pwID, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	_ = g.PassPriority()
	_ = g.PassPriority()

	pw := battlefieldCard(t, g, pwID)
	if got := pw.Counters[CounterLoyalty]; got != 3 {
		t.Fatalf("loyalty counters = %d, want 3 (catalog fallback)", got)
	}
}

// TestPrintedLoyaltyNotDoubleStampedWithCatalog pins the precedence
// rule: when a card has BOTH printed loyalty and a catalog
// StartingLoyalty, exactly one stamp lands. Without it, deck-
// imported The Wandering Emperor would enter with 6.
func TestPrintedLoyaltyNotDoubleStampedWithCatalog(t *testing.T) {
	prev := CatalogStartingLoyalty
	CatalogStartingLoyalty = func(string) int { return 3 }
	t.Cleanup(func() { CatalogStartingLoyalty = prev })

	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	pwID := pushPlaneswalkerToHand(caster, "The Wandering Emperor", 3)
	for i := range caster.Hand.Cards {
		if caster.Hand.Cards[i].InstanceID == pwID {
			caster.Hand.Cards[i].OracleID = "catalog-walker"
			break
		}
	}

	if err := g.CastSpell(caster.ID, pwID, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	_ = g.PassPriority()
	_ = g.PassPriority()

	pw := battlefieldCard(t, g, pwID)
	if got := pw.Counters[CounterLoyalty]; got != 3 {
		t.Fatalf("loyalty counters = %d, want 3 (no double stamp)", got)
	}
}

// TestReanimatedPlaneswalkerGetsPrintedLoyalty covers the other
// battlefield-entry door. docs/decklists/hashaton-the-cheater.md
// recorded that reanimating a planeswalker gave it no loyalty; the
// printed stamp has to ride every entry path, not just the stack.
func TestReanimatedPlaneswalkerGetsPrintedLoyalty(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	owner := g.Seats[0]
	c := NewCard("Teferi, Time Raveler", owner.ID)
	c.TypeLine = "Legendary Planeswalker — Teferi"
	c.StartingLoyalty = 4
	owner.Graveyard.PushTop(c)

	if err := g.ReturnFromGraveyardForEffect(c.InstanceID, ZoneBattlefield); err != nil {
		t.Fatalf("ReturnFromGraveyardForEffect: %v", err)
	}

	pw := battlefieldCard(t, g, c.InstanceID)
	if got := pw.Counters[CounterLoyalty]; got != 4 {
		t.Fatalf("reanimated loyalty counters = %d, want 4", got)
	}
}

// TestManuallyMovedPlaneswalkerGetsPrintedLoyalty covers the admin /
// sandbox door: dragging a planeswalker from hand to the battlefield
// with move_card must stamp loyalty too, otherwise the manual
// escape hatch is broken in exactly the way the cast path was.
func TestManuallyMovedPlaneswalkerGetsPrintedLoyalty(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	owner := g.Seats[0]
	pwID := pushPlaneswalkerToHand(owner, "Teferi, Time Raveler", 4)

	if err := g.MoveCardByID(
		ZoneRef{Kind: ZoneHand, Owner: owner.ID},
		ZoneRef{Kind: ZoneBattlefield},
		pwID,
	); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}

	pw := battlefieldCard(t, g, pwID)
	if got := pw.Counters[CounterLoyalty]; got != 4 {
		t.Fatalf("manually moved loyalty counters = %d, want 4", got)
	}
}
