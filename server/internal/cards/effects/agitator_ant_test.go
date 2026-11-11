package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const agitatorAntOracle = "abe377f8-35f0-4dea-a681-91d31a36815c"

func TestAgitatorAntIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(agitatorAntOracle)
	if !ok || spec.Name != "Agitator Ant" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

// answerMayCounter answers `who`'s "may put two counters" prompt with
// `pick` (uuid.Nil: declines).
func answerMayCounter(t *testing.T, g *game.Game, who, pick uuid.UUID) {
	t.Helper()
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, who)
	if c == nil {
		t.Fatalf("%s was not asked: %+v", who, g.PendingChoices)
	}
	var ids []uuid.UUID
	if pick != uuid.Nil {
		ids = []uuid.UUID{pick}
	}
	if err := g.ResolveOwnPermanents(c.ID, who, ids); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
}

func TestAgitatorAntCountersAndGoadsOnlyTheCreaturesPickedNotTheDecliners(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp, third := g.Seats[seat], g.Seats[(seat+1)%4], g.Seats[(seat+2)%4]
	fourth := g.Seats[(seat+3)%4]
	b12Push(g, me.ID, "Agitator Ant", "Creature — Insect", agitatorAntOracle, 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	declined := b12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)
	_ = fourth // controls no creature: not asked

	advanceToEndStepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	answerMayCounter(t, g, me.ID, mine)
	answerMayCounter(t, g, opp.ID, theirs)
	answerMayCounter(t, g, third.ID, uuid.Nil)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mine, theirs} {
		c, _ := g.LookupCardForEffect(id)
		if n := countersOn(g, id, game.CounterPlusOne); n != 2 {
			t.Errorf("%s has %d counters, want 2", c.Name, n)
		}
		if !c.IsGoadedBy(me.ID) {
			t.Errorf("%s is not goaded", c.Name)
		}
	}
	d, _ := g.LookupCardForEffect(declined)
	if countersOn(g, declined, game.CounterPlusOne) != 0 || d.IsGoaded() {
		t.Error("the player who declined got counters or a goad")
	}
}
