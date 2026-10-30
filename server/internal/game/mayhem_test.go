package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// mayhem_test.go — #4628, CR 702.187: the engine half of mayhem. The
// effects package pins the real cards; this file pins the record
// (Game.DiscardedThisTurn) and the gate (AlternativeCost.
// RequiresDiscardedThisTurn) against stubs, the way flashback's tests do.

const mayhemTestOracle = "test-mayhem"

func mayhemTestOffer() AlternativeCost {
	return AlternativeCost{
		Key:                       "mayhem",
		Label:                     "Mayhem {1}{B}",
		ManaCost:                  "{1}{B}",
		FromZone:                  ZoneGraveyard,
		RequiresDiscardedThisTurn: true,
	}
}

// withMayhemCard stubs the catalog for one mayhem card: the graveyard is
// opened while the object was discarded this turn, priced by the offer.
func withMayhemCard(t *testing.T, oracle string) {
	t.Helper()
	prevZones, prevCosts := CatalogCastableZonesIf, CatalogAlternativeCosts
	CatalogCastableZonesIf = func(id string) []ConditionalCastZone {
		if id != oracle {
			return nil
		}
		return []ConditionalCastZone{{
			Zone: ZoneGraveyard,
			Cond: func(g *Game, _ uuid.UUID, c Card) bool { return g.DiscardedThisTurn(c.InstanceID) },
		}}
	}
	CatalogAlternativeCosts = altCostFor(oracle, mayhemTestOffer())
	t.Cleanup(func() { CatalogCastableZonesIf, CatalogAlternativeCosts = prevZones, prevCosts })
}

// mayhemHandCard puts one instant with the mayhem oracle in me's hand,
// and nothing else, so a random discard can only take it.
func mayhemHandCard(g *Game, me *Player) uuid.UUID {
	c := NewCard("Test Mayhem", me.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{3}{B}"
	c.OracleID = mayhemTestOracle
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		me.Hand.PushTop(c)
	})
	return c.InstanceID
}

func discardedNow(g *Game, id uuid.UUID) (got bool) {
	g.WithWriteLock(func() { got = g.DiscardedThisTurn(id) })
	return got
}

// The record is written by the discard event for the object that
// landed in the graveyard, and by nothing else.
func TestMayhemRecordDiscardedThisTurnIsWrittenByADiscardIntoTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := mayhemHandCard(g, me)
	if discardedNow(g, id) {
		t.Fatal("a card in hand reads as discarded")
	}
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(me.ID, 1); err != nil {
			t.Errorf("discard: %v", err)
		}
	})
	if !me.Graveyard.Contains(id) {
		t.Fatal("fixture: the card was not discarded")
	}
	if !discardedNow(g, id) {
		t.Fatal("the discard did not record the card")
	}
	if discardedNow(g, uuid.New()) || discardedNow(g, uuid.Nil) {
		t.Error("an unrelated or nil id reads as discarded")
	}
}

// CR 400.7: leaving the graveyard ends the object the record names.
func TestMayhemRecordDiscardedThisTurnEndsWhenTheCardChangesZones(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := mayhemHandCard(g, me)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(id, ZoneHand); err != nil {
			t.Errorf("return: %v", err)
		}
	})
	if discardedNow(g, id) {
		t.Error("a card back in hand still reads as discarded")
	}
	// Back in the graveyard some other way, it is a new object.
	g.WithWriteLock(func() {
		if err := g.TuckToLibraryForEffect(id, false); err != nil {
			t.Errorf("tuck: %v", err)
		}
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Errorf("mill: %v", err)
		}
	})
	if !me.Graveyard.Contains(id) {
		t.Fatal("fixture: the card is not back in the graveyard")
	}
	if discardedNow(g, id) {
		t.Error("the record followed the card through its zone changes")
	}
}

// "This turn" is the turn tally's: it is flushed at the next turn.
func TestMayhemRecordDiscardedThisTurnIsFlushedWithTheTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := mayhemHandCard(g, me)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	if !discardedNow(g, id) {
		t.Fatal("fixture: the discard was not recorded")
	}
	passTurn(t, g)
	if discardedNow(g, id) {
		t.Error("the record outlived the turn")
	}
}

// The gate on the PRICE. A card that opens the graveyard itself while
// it was discarded must not hand its cheap price to a card only another
// permission opens the graveyard for (Yawgmoth's Will on a milled card).
func TestMayhemPriceIsRefusedUnderAnotherPermissionWithoutADiscard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	withMayhemCard(t, mayhemTestOracle)
	withCatalogCastPermissions(t, func(id string) []CastPermission {
		if id != "test-will" {
			return nil
		}
		return []CastPermission{{
			Zone: ZoneGraveyard, Scope: ScopeStanding, Duration: WhileInZoneDuration(),
			Filter: PermissionFilter{NonLandOnly: true},
		}}
	})
	source := permanentFor(g, me, "Test Will", "Enchantment", "{1}{R}")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == source {
				g.Battlefield.Cards[i].OracleID = "test-will"
			}
		}
	})
	milled := NewCard("Test Mayhem", me.ID)
	milled.TypeLine = "Instant"
	milled.ManaCost = "{3}{B}"
	milled.OracleID = mayhemTestOracle
	me.Graveyard.PushTop(milled)

	err := g.CastSpell(me.ID, milled.InstanceID, CastSpellParams{FromZone: "graveyard", AlternativeCost: "mayhem"})
	if !errors.Is(err, ErrCastZoneNotAllowed) {
		t.Fatalf("mayhem claimed for a card that was never discarded: err = %v", err)
	}
	if err := g.CastSpell(me.ID, milled.InstanceID, CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("the permission still opens the graveyard at the printed cost: %v", err)
	}
}
