package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// hand_mana_ability_test.go — the enumerator half of #1228. The
// invariant is #544's, one ability kind over from hand_ability_test.go:
// offer exactly the hand mana activations the engine accepts, offer
// none it refuses, and hand a bot a payload the dispatcher takes.

// spiritGuideCard is the printed Spirit Guide: a creature card whose
// only text is a mana ability that functions from a hand.
func spiritGuideCard(name, produced string) game.Card {
	return game.Card{
		Name:     name,
		TypeLine: "Creature — Ape Spirit",
		ManaCost: "{2}{R}",
		ManaAbilities: []game.ManaAbilityShape{{
			Zones:     []game.ZoneKind{game.ZoneHand},
			ExileSelf: true,
			Produced:  produced,
			Label:     "Exile this card from your hand: Add " + produced,
		}},
	}
}

// A Spirit Guide in the seat's own hand is an enumerated mana move,
// and the move the enumerator offers is one the dispatcher accepts.
func TestSpiritGuideIsEnumeratedFromHand(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	guide := handCard(active, spiritGuideCard("Simian Spirit Guide", "{R}"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	got := movesFrom(moves, guide, legal.KindMana)
	if len(got) != 1 {
		t.Fatalf("want exactly one mana move from the hand card, got %v", labels(got))
	}
	dispatchAll(t, g, active.ID, moves)
}

// CR 113.6, both directions, on the mana entry point: a Spirit Guide
// on the battlefield offers nothing and a Forest in hand offers
// nothing. Neither is a source, and the enumerator must not say so —
// a move the engine refuses is #544.
func TestManaAbilityZoneFiltersTheEnumeration(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	advanceTo(t, g, game.StepPrecombatMain)

	onBoard := battlefieldCard(g, active, spiritGuideCard("Simian Spirit Guide", "{R}"))
	inHand := handCard(active, basic("Forest", "Forest"))

	moves := legal.EnumerateFor(g, active.ID)
	if got := movesFrom(moves, onBoard, legal.KindMana); len(got) != 0 {
		t.Errorf("a Spirit Guide on the battlefield offered %v", labels(got))
	}
	if got := movesFrom(moves, inHand, legal.KindMana); len(got) != 0 {
		t.Errorf("a Forest in hand offered %v", labels(got))
	}
	dispatchAll(t, g, active.ID, moves)
}

// CR 108.4: an opponent's Spirit Guide is nobody else's mana move.
func TestSpiritGuideIsNotEnumeratedForAnotherSeat(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	clearHand(other)
	guide := handCard(other, spiritGuideCard("Simian Spirit Guide", "{R}"))
	advanceTo(t, g, game.StepPrecombatMain)

	if got := movesFrom(legal.EnumerateFor(g, active.ID), guide, legal.KindMana); len(got) != 0 {
		t.Errorf("another seat's hand card was offered to me: %v", labels(got))
	}
}
