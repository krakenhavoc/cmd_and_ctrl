package game

import (
	"testing"

	"github.com/google/uuid"
)

// monarch_test.go pins #375: the monarch designation is two triggered
// abilities (CR 724.2), not a sticker. Combat damage to the monarch
// hands the crown to the attacker's controller, and the monarch draws
// at the beginning of their own end step. Both go through the stack,
// and CR 724.4 keeps the crown on the table when its holder leaves.

// settleMonarchStack passes priority until nothing is on, or headed
// for, the stack — answering any CR 603.3b trigger-ordering prompt in
// the order the prompt offered, since two monarch triggers from
// different creatures under different controllers are not
// interchangeable.
//
// A sibling of delayed_test.go's settleStack; it cannot reuse that one
// because PassPriority refuses to move while a choice is outstanding.
func settleMonarchStack(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if answered := answerTriggerOrderPrompts(t, g); answered {
			continue
		}
		if g.Stack.Size() == 0 && len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0 {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatalf("stack did not empty after 64 priority passes")
}

// answerTriggerOrderPrompts clears every outstanding trigger-ordering
// prompt, keeping the offered order. Reports whether it answered any.
func answerTriggerOrderPrompts(t *testing.T, g *Game) bool {
	t.Helper()
	answered := false
	for {
		var choiceID, chooser uuid.UUID
		var ids []uuid.UUID
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == PendingChoiceTriggerOrder {
				choiceID, chooser, ids = c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)
				break
			}
		}
		if choiceID == uuid.Nil {
			return answered
		}
		if err := g.ResolveTriggerOrder(choiceID, chooser, ids); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
		answered = true
	}
}

// attackWith drops a vanilla creature on the battlefield for `owner`,
// declares it as an attacker against `defender`, and walks the cursor
// into the combat damage step. Returns nothing — the tests read the
// board state afterwards.
func attackUnblocked(t *testing.T, g *Game, owner *Player, defender uuid.UUID, power int, extra int) {
	t.Helper()
	advanceIntoStep(t, g, StepDeclareAttackers)
	for i := 0; i < 1+extra; i++ {
		id := pushKeywordCreature(t, g, owner, power, power)
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceIntoStep(t, g, StepCombatDamage)
}

func crown(t *testing.T, g *Game, playerID uuid.UUID) {
	t.Helper()
	if err := g.SetMonarch(playerID); err != nil {
		t.Fatalf("SetMonarch: %v", err)
	}
}

// TestMonarchPassesOnCombatDamage is the headline of #375: the
// reporter had the monarchy, the opponent connected, and nothing
// happened.
func TestMonarchPassesOnCombatDamage(t *testing.T) {
	g := newActiveGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	crown(t, g, defender.ID)

	attackUnblocked(t, g, attacker, defender.ID, 3, 0)
	settleMonarchStack(t, g)

	if g.Monarch != attacker.ID {
		t.Errorf("monarch = %v, want the attacker %v", g.Monarch, attacker.ID)
	}
}

// TestMonarchTransferUsesTheStack: the crown moves on RESOLUTION, not
// on the damage. Everyone gets a window to respond to the trigger.
func TestMonarchTransferUsesTheStack(t *testing.T) {
	g := newActiveGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	crown(t, g, defender.ID)

	advanceIntoStep(t, g, StepDeclareAttackers)
	id := pushKeywordCreature(t, g, attacker, 3, 3)
	if err := g.DeclareAttacker(id, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepCombatDamage)

	if g.Monarch != defender.ID {
		t.Errorf("crown moved before the trigger resolved: monarch = %v", g.Monarch)
	}
	if len(g.PendingTriggers)+len(g.StackMeta) == 0 {
		t.Fatal("combat damage to the monarch produced no trigger")
	}

	settleMonarchStack(t, g)
	if g.Monarch != attacker.ID {
		t.Errorf("monarch = %v, want %v after the trigger resolved", g.Monarch, attacker.ID)
	}
}

// TestMonarchUnmovedByDamageToAnotherPlayer: a hit on somebody who is
// not the monarch is just a hit. Four seats so the defender and the
// monarch can be different people.
func TestMonarchUnmovedByDamageToAnotherPlayer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	attacker, victim, monarch := g.Seats[0], g.Seats[1], g.Seats[2]
	crown(t, g, monarch.ID)

	attackUnblocked(t, g, attacker, victim.ID, 3, 0)
	settleMonarchStack(t, g)

	if g.Monarch != monarch.ID {
		t.Errorf("monarch = %v, want it unchanged at %v", g.Monarch, monarch.ID)
	}
	if victim.Life != 40-3 {
		t.Errorf("victim life = %d, want 37 — the damage should still land", victim.Life)
	}
}

// TestMonarchUnmovedByNoncombatDamage: CR 724.2 says COMBAT damage. A
// Lightning Bolt to the face does not crown the caster.
func TestMonarchUnmovedByNoncombatDamage(t *testing.T) {
	g := newActiveGame(t)
	caster, monarch := g.Seats[0], g.Seats[1]
	crown(t, g, monarch.ID)
	bolt := pushKeywordCreature(t, g, caster, 1, 1)

	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(bolt, monarch.ID, 3); err != nil {
			t.Errorf("DealDamageToPlayerForEffect: %v", err)
		}
	})
	settleMonarchStack(t, g)

	if g.Monarch != monarch.ID {
		t.Errorf("noncombat damage moved the crown: monarch = %v, want %v", g.Monarch, monarch.ID)
	}
}

// TestMonarchNoTransferWithoutAMonarch: no designation, no triggers —
// a hit on a plain player crowns nobody.
func TestMonarchNoTransferWithoutAMonarch(t *testing.T) {
	g := newActiveGame(t)
	attackUnblocked(t, g, g.Seats[0], g.Seats[1].ID, 3, 0)
	settleMonarchStack(t, g)
	if g.Monarch != uuid.Nil {
		t.Errorf("monarch = %v, want none", g.Monarch)
	}
}

// TestMonarchMultipleAttackersInOneStep: two creatures connecting is
// two triggers (CR 724.3 — the crown changes hands once per creature),
// and they share a controller so the table is not asked to order a
// choice with one outcome.
func TestMonarchMultipleAttackersInOneStep(t *testing.T) {
	g := newActiveGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	crown(t, g, defender.ID)

	attackUnblocked(t, g, attacker, defender.ID, 2, 1)

	if got := len(g.PendingTriggers) + len(g.StackMeta); got != 2 {
		t.Errorf("two connecting creatures produced %d triggers, want 2", got)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceTriggerOrder {
			t.Error("asked to order two interchangeable monarch triggers")
		}
	}

	settleMonarchStack(t, g)
	if g.Monarch != attacker.ID {
		t.Errorf("monarch = %v, want %v", g.Monarch, attacker.ID)
	}
	if defender.Life != 40-4 {
		t.Errorf("defender life = %d, want 36", defender.Life)
	}
}

// TestMonarchDrawsAtTheirOwnEndStep is the second half of #375: the
// reporter got no card for holding the crown.
func TestMonarchDrawsAtTheirOwnEndStep(t *testing.T) {
	g := newActiveGame(t)
	monarch := g.Seats[0] // seat 0 takes the first turn
	crown(t, g, monarch.ID)

	before := monarch.Hand.Size()
	advanceTo(t, g, StepEnd)
	if monarch.Hand.Size() != before {
		t.Errorf("the draw skipped the stack: hand %d → %d", before, monarch.Hand.Size())
	}
	settleMonarchStack(t, g)

	if got := monarch.Hand.Size() - before; got != 1 {
		t.Errorf("monarch drew %d cards at their end step, want 1", got)
	}
}

// TestMonarchDoesNotDrawOnSomebodyElsesEndStep: "the MONARCH's end
// step" (CR 724.2), so the crown pays out once per turn cycle, on the
// holder's own turn.
func TestMonarchDoesNotDrawOnSomebodyElsesEndStep(t *testing.T) {
	g := newActiveGame(t)
	monarch := g.Seats[1] // seat 0 is the active player
	crown(t, g, monarch.ID)

	before := monarch.Hand.Size()
	advanceTo(t, g, StepEnd)
	settleMonarchStack(t, g)

	if monarch.Hand.Size() != before {
		t.Errorf("monarch drew on an opponent's end step: hand %d → %d", before, monarch.Hand.Size())
	}
}

// TestMonarchEndStepDrawFollowsTheCrownWithinTheTurn: the active
// player takes the monarchy in combat, so the end step that arrives a
// few steps later is now THEIR end step as monarch — they draw, and
// the player who started the turn wearing the crown does not.
func TestMonarchEndStepDrawFollowsTheCrownWithinTheTurn(t *testing.T) {
	g := newActiveGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	crown(t, g, defender.ID)

	attackUnblocked(t, g, attacker, defender.ID, 3, 0)
	settleMonarchStack(t, g)
	if g.Monarch != attacker.ID {
		t.Fatalf("setup: monarch = %v, want the attacker %v", g.Monarch, attacker.ID)
	}

	attackerHand, defenderHand := attacker.Hand.Size(), defender.Hand.Size()
	advanceTo(t, g, StepEnd)
	settleMonarchStack(t, g)

	if got := attacker.Hand.Size() - attackerHand; got != 1 {
		t.Errorf("new monarch drew %d at their end step, want 1", got)
	}
	if got := defender.Hand.Size() - defenderHand; got != 0 {
		t.Errorf("deposed player drew %d, want 0", got)
	}
}

// TestMonarchCrownPassesToActivePlayerWhenItsHolderIsEliminated is
// CR 724.4. Lethal combat damage is the case that matters: the
// transfer trigger is controlled by the dying monarch, so CR 800.4a
// takes it off the stack with them — without 724.4 the crown would be
// stuck on an empty seat forever.
func TestMonarchCrownPassesToActivePlayerWhenItsHolderIsEliminated(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	attacker, monarch := g.Seats[0], g.Seats[1]
	crown(t, g, monarch.ID)
	monarch.Life = 3

	attackUnblocked(t, g, attacker, monarch.ID, 10, 0)
	settleMonarchStack(t, g)

	if !monarch.Eliminated {
		t.Fatalf("setup: monarch survived 10 damage at 3 life")
	}
	if g.Monarch != attacker.ID {
		t.Errorf("monarch = %v, want the active player %v (CR 724.4)", g.Monarch, attacker.ID)
	}
}

// TestMonarchEliminatedMonarchDrawsNothing: a crown nobody holds pays
// out nothing. Belt-and-braces on becomeMonarchLocked's refusal to
// crown an eliminated player, via the concede path rather than combat.
func TestMonarchEliminatedMonarchDrawsNothing(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	monarch := g.Seats[2]
	crown(t, g, monarch.ID)

	// Guard the probe itself: both assertions below are about a hand
	// that should stay empty and a draw that should never fire, so a
	// seat that started with nothing to lose would pass either of them
	// without testing anything.
	if monarch.Hand.Size() == 0 {
		t.Fatalf("setup: monarch started with an empty hand, so a draw would not show")
	}
	if err := g.Concede(monarch.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if g.Monarch == monarch.ID {
		t.Fatalf("crown stayed with the conceded player")
	}
	// A hand-size DELTA cannot be the probe. CR 800.4a (#769) sweeps
	// every zone the departed player owns, hand included, so comparing
	// against the cards they were holding reads the sweep itself as a
	// draw of -7. The two facts worth asserting are independent:
	//
	//   - the hand is still empty after a full cycle. The sweep runs
	//     ONCE, at elimination, so anything drawn afterwards would
	//     still be sitting there. This catches a draw that LANDED.
	//   - no EventDrawCard names the seat. It fires per card and
	//     carries the drawer in Actor. This catches a draw that was
	//     ATTEMPTED, including one whose card never reached the hand.
	firstAfterConcede := len(g.Events)

	// Walk a full cycle; the eliminated seat's end step never comes,
	// so its monarch draw never triggers.
	for i := 0; i < len(g.Seats); i++ {
		advanceTo(t, g, StepEnd)
		settleMonarchStack(t, g)
		advanceTo(t, g, StepUpkeep)
	}

	if got := monarch.Hand.Size(); got != 0 {
		t.Errorf("eliminated monarch's hand holds %d cards; CR 800.4a emptied it and nothing should have refilled it", got)
	}

	drew := 0
	for _, ev := range g.Events[firstAfterConcede:] {
		if ev.Kind == EventDrawCard && ev.Actor == monarch.ID {
			drew++
		}
	}
	if drew != 0 {
		t.Errorf("eliminated monarch drew %d cards", drew)
	}
}

// TestMonarchSurvivesCloneAndRestore: the designation is snapshot
// state and the listener is a process-lifetime singleton, so an undo
// must neither lose the crown nor lose the ability to move it.
func TestMonarchSurvivesCloneAndRestore(t *testing.T) {
	g := newActiveGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	crown(t, g, defender.ID)

	before := g.Clone()
	attackUnblocked(t, g, attacker, defender.ID, 3, 0)
	settleMonarchStack(t, g)
	if g.Monarch != attacker.ID {
		t.Fatalf("setup: monarch = %v, want %v", g.Monarch, attacker.ID)
	}

	g.RestoreFrom(before)
	if g.Monarch != defender.ID {
		t.Errorf("after restore monarch = %v, want %v", g.Monarch, defender.ID)
	}
	if len(g.Listeners) != len(before.Listeners) {
		t.Errorf("restore dropped listeners: %d, want %d", len(g.Listeners), len(before.Listeners))
	}
}
