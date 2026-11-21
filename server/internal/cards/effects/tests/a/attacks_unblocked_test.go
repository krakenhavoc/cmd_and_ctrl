package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// A defender with no creature has declared none as the step begins, so
// the trigger is on the stack straight away and gives the poison
// counter.
func TestSwampMosquitoPoisonsADefenderWithNothingToBlockWith(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mosquito := effectstest.B12Push(g, me.ID, "Swamp Mosquito", "Creature — Insect", effectstest.SwampMosquitoOracle, 0, 1)
	effectstest.UnblockedAttackIntoBlockers(t, g, mosquito)

	if effectstest.TriggerOnStack(g, mosquito) == nil {
		t.Fatal("the defender declared none as the step began; the trigger should be on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Poison != 1 {
		t.Errorf("defending player poison = %d, want 1", opp.Poison)
	}
}

// While the defender is still deciding, nothing fires; when they finish
// without blocking, it does.
func TestGuiltfeederWaitsForTheDefenderToFinish(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	feeder := effectstest.B12Push(g, me.ID, "Guiltfeeder", "Creature — Horror", effectstest.GuiltfeederOracle, 0, 4)
	// A black creature CAN block through fear, so the defender has a
	// decision to make and stays pending.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Black Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 4, Owner: opp.ID, Controller: opp.ID, Colors: []string{"B"},
	})
	for i := 0; i < 3; i++ {
		effectstest.PushGraveyardCardForTest(opp, "Dead Card")
	}
	effectstest.UnblockedAttackIntoBlockers(t, g, feeder)

	if n := len(g.PendingTriggers) + effectstest.TriggersOnStackFrom(g, feeder); n != 0 {
		t.Fatalf("the defender has not declared yet: %d triggers", n)
	}
	life := opp.Life
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	if effectstest.TriggerOnStack(g, feeder) == nil {
		t.Fatal("declared none: the trigger goes on the stack when the declaration completes")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("defending player lost %d, want 3 (one per graveyard card)", got)
	}
}

// Blocked means no trigger — and the defender's pass is what completes
// the declaration at a manual table.
func TestAbyssalNightstalkerDoesNotTriggerWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stalker := effectstest.B12Push(g, me.ID, "Abyssal Nightstalker", "Creature — Nightstalker", effectstest.AbyssalNightstalkerOracle, 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.UnblockedAttackIntoBlockers(t, g, stalker)

	if err := g.DeclareBlocker(wall, stalker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	if got := g.BlockDeclarationStatusOf(opp.ID); got != game.BlockDeclarationDeclared {
		t.Fatalf("the defender's pass completed the declaration: status %q", got)
	}
	if n := len(g.PendingTriggers) + effectstest.TriggersOnStackFrom(g, stalker); n != 0 {
		t.Errorf("a blocked Nightstalker must not trigger: %d", n)
	}
}

// Unblocked, the defending player is asked to discard — their choice,
// not at random.
func TestAbyssalNightstalkerMakesTheDefenderDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stalker := effectstest.B12Push(g, me.ID, "Abyssal Nightstalker", "Creature — Nightstalker", effectstest.AbyssalNightstalkerOracle, 2, 2)
	effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.PushHandCard(g, opp)
	effectstest.UnblockedAttackIntoBlockers(t, g, stalker)

	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == opp.ID && c.Source == stalker {
			found = true
		}
	}
	if !found {
		t.Errorf("the defending player should be choosing a discard; choices %+v", g.PendingChoices)
	}
}

// Exactly one of the Eternal's two triggers fires, and the defender's
// completed declaration decides which.
func TestEternalOfHarshTruthsAfflictsWhenBlockedAndDrawsWhenNot(t *testing.T) {
	t.Run("blocked", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		eternal := effectstest.B12Push(g, me.ID, "Eternal of Harsh Truths", "Creature — Zombie Cleric", effectstest.EternalOfHarshTruthsOracle, 1, 3)
		wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
		effectstest.UnblockedAttackIntoBlockers(t, g, eternal)
		hand, life := me.Hand.Size(), opp.Life

		if err := g.DeclareBlocker(wall, eternal); err != nil {
			t.Fatal(err)
		}
		if err := g.FinishBlocks(opp.ID); err != nil {
			t.Fatal(err)
		}
		if n := effectstest.TriggersOnStackFrom(g, eternal); n != 1 {
			t.Fatalf("blocked: exactly one trigger (afflict), got %d", n)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := life - opp.Life; got != 2 {
			t.Errorf("afflict 2: defender lost %d", got)
		}
		if me.Hand.Size() != hand {
			t.Error("blocked: the draw must not fire")
		}
	})
	t.Run("unblocked", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		eternal := effectstest.B12Push(g, me.ID, "Eternal of Harsh Truths", "Creature — Zombie Cleric", effectstest.EternalOfHarshTruthsOracle, 1, 3)
		effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
		effectstest.UnblockedAttackIntoBlockers(t, g, eternal)
		hand, life := me.Hand.Size(), opp.Life

		if err := g.FinishBlocks(opp.ID); err != nil {
			t.Fatal(err)
		}
		if n := effectstest.TriggersOnStackFrom(g, eternal); n != 1 {
			t.Fatalf("unblocked: exactly one trigger (the draw), got %d", n)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Size() != hand+1 {
			t.Errorf("unblocked: drew %d, want 1", me.Hand.Size()-hand)
		}
		if opp.Life != life {
			t.Error("unblocked: no afflict")
		}
	})
}
