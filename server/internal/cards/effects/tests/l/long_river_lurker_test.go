package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestLongRiverLurkerIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.LongRiverLurkerOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Long River Lurker: %+v %v", spec, ok)
	}
}

func TestLongRiverLurkerMakesTheTargetUnblockableAndMayFlickerItOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	runner := effectstest.B12Creature(g, me.ID, "Runner", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterByKeyedForEffect(uuid.Nil, runner, game.CounterPlusOne, 1, game.KeyedTail{}) })
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.LrlCastLurker(t, g, me, runner)
	r, _ := effectstest.BattlefieldCard(g, runner)
	oldEpoch := r.ObjectEpoch
	effectstest.AssertRestrictions(t, g, runner, game.CantBeBlocked)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.DeclareAttack(t, g, opp.ID, runner)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, runner); err == nil {
		t.Error("the unblockable creature was blocked")
	}
	life := opp.Life
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if life-opp.Life != 3 {
		t.Errorf("the defender lost %d, want 3 (2/2 with a counter)", life-opp.Life)
	}
	if _, still := effectstest.BattlefieldCard(g, runner); still {
		t.Error("the creature is still the same object after being exiled and returned")
	}
	back := effectstest.BattlefieldIDsNamed(g, "Runner")
	if len(back) != 1 {
		t.Fatalf("the battlefield has %d Runners, want the creature back as a new object", len(back))
	}
	r2, _ := effectstest.BattlefieldCard(g, back[0])
	if r2.ObjectEpoch == oldEpoch && back[0] == runner {
		t.Error("the same object")
	}
	if effectstest.CounterOn(g, back[0], game.CounterPlusOne) > 0 {
		t.Error("the counter survived the blink")
	}
}

func TestLongRiverLurkerGivesOtherFrogsWardOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.DskB1Creature(g, me.ID, "Long River Lurker", "Creature — Frog Scout", effectstest.LongRiverLurkerOracle, 2, 3)
	frog := effectstest.B12Creature(g, me.ID, "Other Frog", "Creature — Frog", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	// A Frog is warded; a Bear is not.
	effectstest.CastAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 6; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Error("a non-Frog has ward")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastAtWardedCreature(t, g, opp, frog)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for a spell aimed at the other Frog")
	}
}
