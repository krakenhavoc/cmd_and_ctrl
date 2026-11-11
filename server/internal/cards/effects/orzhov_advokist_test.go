package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const orzhovAdvokistOracle = "d5e9942e-096f-433f-8668-4b594f3c1b02"

func TestOrzhovAdvokistIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(orzhovAdvokistOracle)
	if !ok || spec.Name != "Orzhov Advokist" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

func TestOrzhovAdvokistBarsOnlyThePlayersWhoTookTheCountersFromAttackingItsController(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp, third := g.Seats[seat], g.Seats[(seat+1)%4], g.Seats[(seat+2)%4]
	b12Push(g, me.ID, "Orzhov Advokist", "Creature — Human Advisor", orzhovAdvokistOracle, 1, 4)
	var walker uuid.UUID
	g.WithWriteLock(func() {
		walker = pushPermanent(g, me.ID, game.Card{Name: "My Walker", TypeLine: "Legendary Planeswalker — Test",
			Counters: map[string]int{game.CounterLoyalty: 3}})
	})
	oppBear := b12Creature(g, opp.ID, "Opp Bear", "Creature — Bear", 2, 2)
	oppOther := b12Creature(g, opp.ID, "Opp Other", "Creature — Bear", 2, 2)
	thirdBear := b12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)

	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	answerMayCounter(t, g, me.ID, uuid.Nil)
	answerMayCounter(t, g, opp.ID, oppBear)
	answerMayCounter(t, g, third.ID, uuid.Nil)
	passPriorityAroundTable(t, g)

	if n := countersOn(g, oppBear, game.CounterPlusOne); n != 2 {
		t.Fatalf("counters = %d, want 2", n)
	}
	// It reaches the player's other creatures too: "creatures that player
	// controls", not only the one that got the counters.
	advanceToDeclareAttackersOf(t, g, seat)
	endTurn(t, g)
	// The restriction holds on the other players' turns, until mine.
	advanceToDeclareAttackersOf(t, g, (seat+1)%4)
	for _, id := range []uuid.UUID{oppBear, oppOther} {
		for _, target := range []uuid.UUID{me.ID, walker} {
			if err := g.Clone().DeclareAttacker(id, target); err == nil {
				t.Errorf("a creature of the player who took counters attacked %s", target)
			}
		}
		if err := g.Clone().DeclareAttacker(id, third.ID); err != nil {
			t.Errorf("attacking another player: %v", err)
		}
	}
	// The player who declined is free to attack.
	endTurn(t, g)
	advanceToDeclareAttackersOf(t, g, (seat+2)%4)
	if err := g.Clone().DeclareAttacker(thirdBear, me.ID); err != nil {
		t.Errorf("the declining player's creature was barred: %v", err)
	}
}
