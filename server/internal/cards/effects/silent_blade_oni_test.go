package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

const silentBladeOniOracle = "4f515250-0f1f-4fb7-aa75-b703b81ccd4a"

func TestSilentBladeOniIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(silentBladeOniOracle)
	if !ok || spec.Name != "Silent-Blade Oni" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// oniHits puts the Oni under seat 0, an expensive sorcery, a land and
// (optionally) nothing else in seat 1's hand, and has the Oni deal
// combat damage to seat 1.
func oniHits(t *testing.T, g *game.Game) (me, opp *game.Player, sorcery, land uuid.UUID) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	sorcery = pushHandCardWithManaCost(opp, "Their Big Sorcery", "Sorcery", "{5}{U}{U}")
	land = pushHandCardWithManaCost(opp, "Their Island", "Basic Land — Island", "")
	oni := b43Catalog(g, me.ID, "Silent-Blade Oni", "Creature — Demon Ninja", silentBladeOniOracle, 6, 5)
	attackWith(t, g, opp.ID, oni)
	passPriorityAroundTable(t, g)
	return me, opp, sorcery, land
}

func knows(g *game.Game, p *game.Player, card, viewer uuid.UUID) bool {
	known := false
	g.ReadSnapshot(func() {
		for _, c := range p.Hand.Cards {
			if c.InstanceID == card {
				known = c.KnownBy[viewer]
			}
		}
	})
	return known
}

// The Oni's controller looks at the damaged player's hand, alone, and
// may cast a spell out of it for free while the trigger resolves: a
// sorcery in the combat damage step, with no mana. A land is looked at
// and never offered. The spell is the Oni's controller's, and the card
// goes to its owner's graveyard.
func TestSilentBladeOniCastsASpellFromTheDamagedPlayersHandForFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, sorcery, land := oniHits(t, g)
	third := g.Seats[2]

	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, sorcery)
	if slices.Contains(offer.DiscardOptions, land) {
		t.Error("a land was offered as a spell to cast")
	}
	for _, id := range []uuid.UUID{sorcery, land} {
		if !knows(g, opp, id, me.ID) {
			t.Error("the Oni's controller did not get to look at the whole hand")
		}
		if knows(g, opp, id, third.ID) {
			t.Error("a third player learned the hand: it is a look, not a reveal")
		}
	}
	// The card is still in its owner's hand, and its owner's own cast of
	// it is not this prompt's answer.
	if !opp.Hand.Contains(sorcery) {
		t.Fatal("the card left its owner's hand before it was cast")
	}
	if err := castThroughResolutionCast(t, g, me.ID, sorcery, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast out of their hand: %v", err)
	}
	if !g.Stack.Contains(sorcery) || opp.Hand.Contains(sorcery) {
		t.Fatal("the spell is not on the stack")
	}
	castInsideTheResolution(t, g, sorcery)
	var item *game.StackItem
	g.ReadSnapshot(func() { item = g.StackMeta[sorcery] })
	if item == nil || item.Controller != me.ID {
		t.Fatalf("the spell's controller is not the Oni's controller: %+v", item)
	}
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("a second spell was offered: the text says a spell, one")
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(sorcery) {
		t.Error("the resolved sorcery should be in its owner's graveyard")
	}
	if !opp.Hand.Contains(land) {
		t.Error("the land should still be in its owner's hand")
	}
}

// Declining leaves every card in that hand, and nothing lets the Oni's
// controller cast one of them afterwards.
func TestSilentBladeOniDecliningLeavesTheHandAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, sorcery, _ := oniHits(t, g)

	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, sorcery)
	declineResolutionCast(t, g, me.ID)
	if !opp.Hand.Contains(sorcery) {
		t.Fatal("declining moved the card")
	}
	if g.CastPermissionOnCardByIDForEffect(sorcery) != nil || g.ResolutionPaused() {
		t.Fatal("the decline left the grant, or held the resolution")
	}
	err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: "hand", AlternativeCost: game.AltCostKeyGrantedCost})
	if !errors.Is(err, game.ErrCardNotFound) {
		t.Fatalf("a cast out of their hand after the decline: err = %v, want ErrCardNotFound", err)
	}
}

// While the offer is open, the card in the other player's hand can be
// cast only as the offer's free cast: a cast that pays its printed cost
// instead is not one the permission opens.
func TestSilentBladeOniOnlyTheOfferedCastReachesTheirHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, sorcery, land := oniHits(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, sorcery)

	if err := castThroughResolutionCast(t, g, me.ID, land, game.CastSpellParams{FromZone: "hand"}); err == nil {
		t.Fatal("their land was played out of their hand")
	}
	g.WithWriteLock(func() {
		fillPool(me, 7)
	})
	third := g.Seats[2]
	if err := g.CastSpell(third.ID, sorcery, game.CastSpellParams{FromZone: "hand", AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Fatal("a third player cast the card out of that hand")
	}
	if !opp.Hand.Contains(sorcery) || !opp.Hand.Contains(land) {
		t.Fatal("a refused cast moved a card")
	}
}

// A hand with no spell in it is looked at and nothing is offered.
func TestSilentBladeOniNothingToCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Hand.Cards = nil
	land := pushHandCardWithManaCost(opp, "Their Island", "Basic Land — Island", "")
	oni := b43Catalog(g, me.ID, "Silent-Blade Oni", "Creature — Demon Ninja", silentBladeOniOracle, 6, 5)
	attackWith(t, g, opp.ID, oni)
	passPriorityAroundTable(t, g)
	if c := resolutionCastOfferFor(g, me.ID); c != nil || g.ResolutionPaused() {
		t.Fatalf("an offer was made over a hand with no spell: %+v", c)
	}
	if !knows(g, opp, land, me.ID) {
		t.Error("the hand was not looked at")
	}
}

// The legal-move list, which the client's sheet and a bot both answer
// from, offers the free cast of the card in the other player's hand for
// the Oni's controller, and nothing for that card to anyone else.
func TestSilentBladeOniOfferIsInTheLegalMoves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, sorcery, _ := oniHits(t, g)
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, sorcery)

	castsOf := func(seat uuid.UUID) int {
		n := 0
		for _, m := range legal.EnumerateFor(g, seat) {
			if m.Type == legal.TypeCastSpell && m.Source == sorcery && m.ResolutionCast == offer.ID.String() {
				n++
			}
		}
		return n
	}
	if got := castsOf(me.ID); got != 1 {
		t.Fatalf("legal casts of their sorcery for the Oni's controller = %d, want 1", got)
	}
	if got := castsOf(g.Seats[2].ID); got != 0 {
		t.Fatalf("a third player was offered %d casts of it", got)
	}
	_ = opp
}
