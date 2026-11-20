package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_cards_cost_test.go — #1297: "Exile N cards from your graveyard"
// and "Exile a card from your hand" as a CR 602 activated ability's
// cost (AbilityCost.ExileCards), on the five proof cards.
//
// What is pinned, card by card:
//
//   - Grim Lavamancer: the count, the pile, the distinctness, and that
//     the payment happens at ANNOUNCE — the cards are in exile while
//     the ability is still on the stack — through the exit primitive,
//     never the discard door.
//   - Moorland Haunt, Tome Shredder: the clause's predicate.
//   - Mines of Moria: the caveat is gone and the ability is real.
//   - Holistic Wisdom: the hand form, the paid-cost record the effect
//     reads, and the auto-tapper keeping its hands off the named card.

const (
	GrimLavamancerOracle = "37445e06-88a1-4e2e-a432-383736c9b977"
	MoorlandHauntOracle  = "5324192b-6687-41e4-8e56-326b21a5dbf3"
	TomeShredderOracle   = "b145952b-52e3-4a66-b47e-f08a489f9443"
	MinesOfMoriaOracle   = "583cdebe-0195-45be-bd2e-5765f07cb902"
	HolisticWisdomOracle = "7e108285-52da-473c-accd-d48e646a49c0"
	SimianSpiritGuideOr  = "44e0ffa3-8915-4c1f-8f1a-4aeea1365f07"
)

// exileCostTable is a catalog game in the active seat's main phase with
// an empty hand and graveyard.
func ExileCostTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		me.Graveyard.Cards = nil
	})
	return g, me, opp
}

func FloatMana(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect %s: %v", mana, err)
	}
}

func NoDiscardEvents(t *testing.T, g *game.Game) {
	t.Helper()
	for _, ev := range g.Events {
		if ev.Kind == game.EventDiscardCard {
			t.Fatalf("an EventDiscardCard fired (%+v) — the cost exiles, it does not discard", ev)
		}
	}
}
