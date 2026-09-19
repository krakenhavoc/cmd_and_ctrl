package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_wandering_rescuer_test.go — the grant this card shipped with
// in S22 appended "hexproof" to every other tapped creature you
// control and NOTHING READ IT. The card's own file said so. These
// tests are the un-rotting: they assert the grant through the same
// public surfaces a player goes through, so the day something
// changes the Layer 6 static or the keyword table, this fails rather
// than the card quietly going inert again.

// rescuerBoard puts a Rescuer on the battlefield under seat 0 plus
// one other creature under the same controller, and returns their
// IDs. The other creature's tap state is the test's lever.
func rescuerBoard(t *testing.T, g *game.Game, otherTapped bool) (rescuer, other uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	rescuer = uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: rescuer,
		Name:       "The Wandering Rescuer",
		TypeLine:   "Legendary Creature — Human Samurai Noble",
		OracleID:   theWanderingRescuerOracle,
		Power:      3,
		Toughness:  4,
		Owner:      me.ID,
		Controller: me.ID,
	})
	other = uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: other,
		Name:       "Convoked Bear",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Tapped:     otherTapped,
		Owner:      me.ID,
		Controller: me.ID,
	})
	// Both cards were pushed straight onto the zone, so announce the
	// entries the way the real cast path does: the layer engine keys
	// its invalidation and its CR 613 timestamps off EventZoneMove.
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{rescuer, other} {
			g.EmitEvent(game.Event{
				Kind:    game.EventZoneMove,
				CardID:  id,
				OldZone: game.ZoneHand,
				NewZone: game.ZoneBattlefield,
			})
		}
	})
	// Force the layer recompute so Effective() carries the grant.
	g.ReadSnapshot(func() {})
	return rescuer, other
}

// TestWanderingRescuerHexproofGrantIsLive is the headline: the
// creature tapped to convoke the Rescuer is now untargetable by an
// opponent, which is the loop the card is built around.
func TestWanderingRescuerHexproofGrantIsLive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, tapped := rescuerBoard(t, g, true)

	spec := TargetCreature("target creature")
	oppTargets := g.LegalTargetsForEffect(game.SourceChooser(opp.ID), spec)
	for _, id := range oppTargets.Cards {
		if id == tapped {
			t.Fatalf("a tapped creature under the Rescuer is still offered to an opponent — the grant is inert")
		}
	}
	// Its own controller is unaffected: hexproof is "your opponents".
	mine := g.LegalTargetsForEffect(game.SourceChooser(me.ID), spec)
	found := false
	for _, id := range mine.Cards {
		if id == tapped {
			found = true
		}
	}
	if !found {
		t.Errorf("hexproof must not stop the Rescuer's own controller from targeting")
	}
}

// TestWanderingRescuerGrantFollowsTapState: the clause reads "other
// TAPPED creatures you control", so the protection comes and goes
// with the tap — an untapped creature is a legal target again on the
// next recompute.
func TestWanderingRescuerGrantFollowsTapState(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	_, untapped := rescuerBoard(t, g, false)

	spec := TargetCreature("target creature")
	lt := g.LegalTargetsForEffect(game.SourceChooser(opp.ID), spec)
	for _, id := range lt.Cards {
		if id == untapped {
			return
		}
	}
	t.Errorf("an UNTAPPED creature must stay targetable — the grant is gated on Tapped")
}

// TestWanderingRescuerGrantTracksALaterTap is the cache half. The
// creature was untapped when the Rescuer arrived and taps later
// (attacking, or paying another convoke cost). The grant has to
// appear on that tap, which means EventTapCard must invalidate the
// layer engine's cached resolution — it did not before S23, so the
// card looked half-working even once the keyword was honoured.
func TestWanderingRescuerGrantTracksALaterTap(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	_, bear := rescuerBoard(t, g, false)

	spec := TargetCreature("target creature")
	offered := func() bool {
		for _, id := range g.LegalTargetsForEffect(game.SourceChooser(opp.ID), spec).Cards {
			if id == bear {
				return true
			}
		}
		return false
	}
	if !offered() {
		t.Fatalf("untapped creature should start targetable")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
		g.EmitEvent(game.Event{Kind: game.EventTapCard, CardID: bear})
	})
	g.ReadSnapshot(func() {})
	if offered() {
		t.Errorf("creature that tapped under the Rescuer is still targetable — the layer cache did not invalidate on tap")
	}
}

// TestWanderingRescuerDoesNotProtectItself — "OTHER tapped
// creatures". The Rescuer is the one creature its own grant skips.
func TestWanderingRescuerDoesNotProtectItself(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rescuer, _ := rescuerBoard(t, g, true)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rescuer {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
		g.EmitEvent(game.Event{Kind: game.EventTapCard, CardID: rescuer})
	})
	g.ReadSnapshot(func() {})

	lt := g.LegalTargetsForEffect(game.SourceChooser(opp.ID), TargetCreature("target creature"))
	for _, id := range lt.Cards {
		if id == rescuer {
			return
		}
	}
	t.Errorf("the Rescuer must not grant hexproof to itself")
}
