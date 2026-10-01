package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const defiledCryptOracle = "f97de425-c7ab-4688-8606-159370d179ee"

func defiledCryptCard(g *game.Game, me *game.Player) game.Card {
	return roomsBCard(me.ID, defiledCryptOracle, "Defiled Crypt", "{3}{B}", "Cadaver Lab", "{B}")
}

// Cast Cadaver Lab: its unlock trigger returns a creature card, and the
// still-locked Defiled Crypt does nothing about the card leaving the
// graveyard. Then unlocking Defiled Crypt turns the door on: a card
// leaving the graveyard makes one Horror, and only one each turn.
func TestDefiledCryptCadaverLabDoors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	first := pushGraveyardPermanent(me, "First Corpse", "Creature — Zombie", "{1}{B}")
	second := pushGraveyardPermanent(me, "Second Corpse", "Creature — Zombie", "{1}{B}")
	third := pushGraveyardPermanent(me, "Third Corpse", "Creature — Zombie", "{1}{B}")
	c := defiledCryptCard(g, me)

	roomsBCast(t, g, me, c, 1)
	roomsBSettle(t, g, me, first)
	if !me.Hand.Contains(first) {
		t.Fatal("Cadaver Lab did not return the chosen creature card to hand")
	}
	if n := roomsBCountNamed(g, "Horror"); n != 0 {
		t.Fatalf("a locked Defiled Crypt made %d Horror tokens", n)
	}

	roomsBUnlock(t, g, me, c, game.DoorLeft)
	roomsBSettle(t, g, me)
	if n := roomsBCountNamed(g, "Horror"); n != 0 {
		t.Fatalf("unlocking alone made %d Horror tokens", n)
	}

	// Two cards leave the graveyard in one turn: one Horror.
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: c.InstanceID})
		if err := (ReturnFromGraveyard{Target: second, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			t.Fatal(err)
		}
	})
	roomsBSettle(t, g, me)
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: c.InstanceID})
		if err := (ReturnFromGraveyard{Target: third, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			t.Fatal(err)
		}
	})
	roomsBSettle(t, g, me)
	if n := roomsBCountNamed(g, "Horror"); n != 1 {
		t.Fatalf("Horror tokens = %d, want exactly 1 (once each turn)", n)
	}
	for _, tok := range g.BattlefieldCardsForEffect() {
		if tok.Name != "Horror" {
			continue
		}
		if !tok.IsCreature() || !tok.IsEnchantment() || tok.Power != 2 || tok.Toughness != 2 {
			t.Errorf("Horror token = %+v, want a 2/2 enchantment creature", tok)
		}
	}
}
