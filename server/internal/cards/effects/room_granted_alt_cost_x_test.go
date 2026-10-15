package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3724: {X} is 0 under a cast that does not pay the mana cost (CR
// 107.3b), so Access Maze's life is the mana value with X = 0.
func TestAccessMazeCastsAnXSpellWithXZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushAccessMaze(g, me.ID, game.DoorRightUnlocked)
	key := GrantedAltCostLifeFromHandOnceYourTurn
	xSpell := handSorcery(me, "X Sorcery", "{X}{X}{1}{B}")
	offer := offerWithKey(grantedCardOffers(g, me.ID, xSpell, game.ZoneHand), key)
	if offer == nil {
		t.Fatal("Access Maze did not offer its life cast to a spell with {X}")
	}
	if offer.Life != 2 || offer.ManaCost != "" {
		t.Fatalf("offer charges %q and %d life, want 2 life (X = 0) and no mana", offer.ManaCost, offer.Life)
	}
	life := me.Life
	if err := g.CastSpell(me.ID, xSpell, game.CastSpellParams{Strict: true, AlternativeCost: key}); err != nil {
		t.Fatalf("life cast of an X spell: %v", err)
	}
	if me.Life != life-2 {
		t.Fatalf("life %d, want %d", me.Life, life-2)
	}
}

// #3724: CR 118.6 lets a spell with no mana cost be cast through an
// alternative cost, and its mana value is 0, so it costs no life.
func TestAccessMazeCastsASpellWithNoManaCostForNoLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushAccessMaze(g, me.ID, game.DoorRightUnlocked)
	key := GrantedAltCostLifeFromHandOnceYourTurn
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Ancestral Vision", Layout: "normal", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	offer := offerWithKey(grantedCardOffers(g, me.ID, id, game.ZoneHand), key)
	if offer == nil {
		t.Fatal("Access Maze did not offer its life cast to a spell with no mana cost")
	}
	if offer.Life != 0 {
		t.Fatalf("offer charges %d life, want 0", offer.Life)
	}
	life := me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: key}); err != nil {
		t.Fatalf("life cast of a spell with no mana cost: %v", err)
	}
	if me.Life != life {
		t.Fatalf("life %d, want %d", me.Life, life)
	}
}

// #3724: Warped Space's free offer has no X gap.
func TestWarpedSpaceCastsAnXSpellFromExile(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushWarpedSpace(g, me.ID, game.DoorRightUnlocked)
	key := GrantedAltCostZeroFromExileOnceEachTurn
	_, xID := impulseExileCost(t, g, "X Sorcery", "{X}{R}")
	if offerWithKey(grantedCardOffers(g, me.ID, xID, game.ZoneExile), key) == nil {
		t.Fatal("Warped Space did not offer its free cast to a spell with {X}")
	}
	if err := g.CastSpell(me.ID, xID, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: key}); err != nil {
		t.Fatalf("free cast of an X spell: %v", err)
	}
}
