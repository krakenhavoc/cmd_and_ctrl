package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const orzhovAdvokistOracle = "d5e9942e-096f-433f-8668-4b594f3c1b02"

func TestOrzhovAdvokistIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(orzhovAdvokistOracle)
	if !ok || spec.Name != "Orzhov Advokist" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

func TestOrzhovAdvokistBarsOnlyThePlayersWhoTookTheCountersFromAttackingItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp, third := g.Seats[seat], g.Seats[(seat+1)%4], g.Seats[(seat+2)%4]
	effectstest.B12Push(g, me.ID, "Orzhov Advokist", "Creature — Human Advisor", orzhovAdvokistOracle, 1, 4)
	var walker uuid.UUID
	g.WithWriteLock(func() {
		walker = effectstest.PushPermanent(g, me.ID, game.Card{Name: "My Walker", TypeLine: "Legendary Planeswalker — Test",
			Counters: map[string]int{game.CounterLoyalty: 3}})
	})
	oppBear := effectstest.B12Creature(g, opp.ID, "Opp Bear", "Creature — Bear", 2, 2)
	oppOther := effectstest.B12Creature(g, opp.ID, "Opp Other", "Creature — Bear", 2, 2)
	thirdBear := effectstest.B12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)

	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayCounter(t, g, me.ID, uuid.Nil)
	effectstest.AnswerMayCounter(t, g, opp.ID, oppBear)
	effectstest.AnswerMayCounter(t, g, third.ID, uuid.Nil)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CountersOn(g, oppBear, game.CounterPlusOne); n != 2 {
		t.Fatalf("counters = %d, want 2", n)
	}
	// It reaches the player's other creatures too: "creatures that player
	// controls", not only the one that got the counters.
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	effectstest.EndTurn(t, g)
	// The restriction holds on the other players' turns, until mine.
	effectstest.AdvanceToDeclareAttackersOf(t, g, (seat+1)%4)
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
	effectstest.EndTurn(t, g)
	effectstest.AdvanceToDeclareAttackersOf(t, g, (seat+2)%4)
	if err := g.Clone().DeclareAttacker(thirdBear, me.ID); err != nil {
		t.Errorf("the declining player's creature was barred: %v", err)
	}
}
