package game

import (
	"testing"

	"github.com/google/uuid"
)

// "Cards can't enter from exile this turn" (Bad Wolf Bay, ADR 0216 slice
// 4): the stored rule refuses a card leaving exile and nothing else, and it
// is swept with the turn.
func TestCardsCantEnterFromExileRefusesOnlyCardsLeavingExile(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.WithWriteLock(func() {
		card := Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear"}
		q := EntryQuery{Game: g, Card: card, FromZone: ZoneExile}
		if _, refused := g.entryRefusalLocked(q); refused {
			t.Fatal("refused before the rule")
		}
		if !g.CardsCantEnterFromExileThisTurnForEffect(uuid.Nil, "Bad Wolf Bay") {
			t.Fatal("the rule was not registered")
		}
		if !g.anyEntryRestrictionLocked() {
			t.Error("the gate's cheap question does not see the rule")
		}
		if _, refused := g.entryRefusalLocked(q); !refused {
			t.Error("a card leaving exile was not refused")
		}
		for _, other := range []EntryQuery{
			{Game: g, Card: card, FromZone: ZoneGraveyard},
			{Game: g, Card: card, FromZone: ZoneHand},
			{Game: g, Card: card, Token: true},
		} {
			if _, refused := g.entryRefusalLocked(other); refused {
				t.Errorf("refused an entry from %q (token %v)", other.FromZone, other.Token)
			}
		}
	})
}
