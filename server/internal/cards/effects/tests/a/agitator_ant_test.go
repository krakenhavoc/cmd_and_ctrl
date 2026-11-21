package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const agitatorAntOracle = "abe377f8-35f0-4dea-a681-91d31a36815c"

func TestAgitatorAntIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(agitatorAntOracle)
	if !ok || spec.Name != "Agitator Ant" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

func TestAgitatorAntCountersAndGoadsOnlyTheCreaturesPickedNotTheDecliners(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp, third := g.Seats[seat], g.Seats[(seat+1)%4], g.Seats[(seat+2)%4]
	fourth := g.Seats[(seat+3)%4]
	effectstest.B12Push(g, me.ID, "Agitator Ant", "Creature — Insect", agitatorAntOracle, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	declined := effectstest.B12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)
	_ = fourth // controls no creature: not asked

	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayCounter(t, g, me.ID, mine)
	effectstest.AnswerMayCounter(t, g, opp.ID, theirs)
	effectstest.AnswerMayCounter(t, g, third.ID, uuid.Nil)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mine, theirs} {
		c, _ := g.LookupCardForEffect(id)
		if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 2 {
			t.Errorf("%s has %d counters, want 2", c.Name, n)
		}
		if !c.IsGoadedBy(me.ID) {
			t.Errorf("%s is not goaded", c.Name)
		}
	}
	d, _ := g.LookupCardForEffect(declined)
	if effectstest.CountersOn(g, declined, game.CounterPlusOne) != 0 || d.IsGoaded() {
		t.Error("the player who declined got counters or a goad")
	}
}
