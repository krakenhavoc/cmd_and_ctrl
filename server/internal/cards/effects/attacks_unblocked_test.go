package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attacks_unblocked_test.go — #1279's proof cards. "Whenever this
// creature attacks and isn't blocked" (CR 509.3) fires off the
// defending player's COMPLETED block declaration, so it fires when the
// defender finishes declaring — not when the step begins, and never
// while they are still deciding.

const (
	swampMosquitoOracle        = "4bb4844c-2678-45c7-8f5e-9cf185fd484b"
	abyssalNightstalkerOracle  = "10733767-1c97-4e2e-b02b-54bf346f6583"
	guiltfeederOracle          = "0285bd20-f49e-48f7-8c8c-960f9fbe4d34"
	eternalOfHarshTruthsOracle = "5c09d646-afea-4fd0-9752-04820b81cc5b"
)

// unblockedAttackIntoBlockers declares `attacker` at seat 1 and walks
// into declare_blockers.
func unblockedAttackIntoBlockers(t *testing.T, g *game.Game, attacker uuid.UUID) {
	t.Helper()
	declareAttack(t, g, g.Seats[1].ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
}

// A defender with no creature has declared none as the step begins, so
// the trigger is on the stack straight away and gives the poison
// counter.
func TestSwampMosquitoPoisonsADefenderWithNothingToBlockWith(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mosquito := b12Push(g, me.ID, "Swamp Mosquito", "Creature — Insect", swampMosquitoOracle, 0, 1)
	unblockedAttackIntoBlockers(t, g, mosquito)

	if triggerOnStack(g, mosquito) == nil {
		t.Fatal("the defender declared none as the step began; the trigger should be on the stack")
	}
	passPriorityAroundTable(t, g)
	if opp.Poison != 1 {
		t.Errorf("defending player poison = %d, want 1", opp.Poison)
	}
}

// While the defender is still deciding, nothing fires; when they finish
// without blocking, it does.
func TestGuiltfeederWaitsForTheDefenderToFinish(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	feeder := b12Push(g, me.ID, "Guiltfeeder", "Creature — Horror", guiltfeederOracle, 0, 4)
	// A black creature CAN block through fear, so the defender has a
	// decision to make and stays pending.
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Black Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 4, Owner: opp.ID, Controller: opp.ID, Colors: []string{"B"},
	})
	for i := 0; i < 3; i++ {
		pushGraveyardCardForTest(opp, "Dead Card")
	}
	unblockedAttackIntoBlockers(t, g, feeder)

	if n := len(g.PendingTriggers) + triggersOnStackFrom(g, feeder); n != 0 {
		t.Fatalf("the defender has not declared yet: %d triggers", n)
	}
	life := opp.Life
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	if triggerOnStack(g, feeder) == nil {
		t.Fatal("declared none: the trigger goes on the stack when the declaration completes")
	}
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("defending player lost %d, want 3 (one per graveyard card)", got)
	}
}

// Blocked means no trigger — and the defender's pass is what completes
// the declaration at a manual table.
func TestAbyssalNightstalkerDoesNotTriggerWhenBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stalker := b12Push(g, me.ID, "Abyssal Nightstalker", "Creature — Nightstalker", abyssalNightstalkerOracle, 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	unblockedAttackIntoBlockers(t, g, stalker)

	if err := g.DeclareBlocker(wall, stalker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	if got := g.BlockDeclarationStatusOf(opp.ID); got != game.BlockDeclarationDeclared {
		t.Fatalf("the defender's pass completed the declaration: status %q", got)
	}
	if n := len(g.PendingTriggers) + triggersOnStackFrom(g, stalker); n != 0 {
		t.Errorf("a blocked Nightstalker must not trigger: %d", n)
	}
}

// Unblocked, the defending player is asked to discard — their choice,
// not at random.
func TestAbyssalNightstalkerMakesTheDefenderDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stalker := b12Push(g, me.ID, "Abyssal Nightstalker", "Creature — Nightstalker", abyssalNightstalkerOracle, 2, 2)
	b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	pushHandCard(g, opp)
	unblockedAttackIntoBlockers(t, g, stalker)

	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	passPriorityAroundTable(t, g)
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
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		eternal := b12Push(g, me.ID, "Eternal of Harsh Truths", "Creature — Zombie Cleric", eternalOfHarshTruthsOracle, 1, 3)
		wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
		unblockedAttackIntoBlockers(t, g, eternal)
		hand, life := me.Hand.Size(), opp.Life

		if err := g.DeclareBlocker(wall, eternal); err != nil {
			t.Fatal(err)
		}
		if err := g.FinishBlocks(opp.ID); err != nil {
			t.Fatal(err)
		}
		if n := triggersOnStackFrom(g, eternal); n != 1 {
			t.Fatalf("blocked: exactly one trigger (afflict), got %d", n)
		}
		passPriorityAroundTable(t, g)
		if got := life - opp.Life; got != 2 {
			t.Errorf("afflict 2: defender lost %d", got)
		}
		if me.Hand.Size() != hand {
			t.Error("blocked: the draw must not fire")
		}
	})
	t.Run("unblocked", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		eternal := b12Push(g, me.ID, "Eternal of Harsh Truths", "Creature — Zombie Cleric", eternalOfHarshTruthsOracle, 1, 3)
		b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
		unblockedAttackIntoBlockers(t, g, eternal)
		hand, life := me.Hand.Size(), opp.Life

		if err := g.FinishBlocks(opp.ID); err != nil {
			t.Fatal(err)
		}
		if n := triggersOnStackFrom(g, eternal); n != 1 {
			t.Fatalf("unblocked: exactly one trigger (the draw), got %d", n)
		}
		passPriorityAroundTable(t, g)
		if me.Hand.Size() != hand+1 {
			t.Errorf("unblocked: drew %d, want 1", me.Hand.Size()-hand)
		}
		if opp.Life != life {
			t.Error("unblocked: no afflict")
		}
	})
}
