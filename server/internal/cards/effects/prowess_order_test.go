package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prowess_order_test.go — #1511. A batch of prowess triggers from
// different creatures commutes, so the CR 603.3b ordering prompt is
// skipped for it (ADR 0018's #1511 amendment). Anything else in the
// batch — even one untargeted trigger of another kind — keeps the
// prompt. The targeted case is pinned at the drain in
// game/trigger_order_commutes_test.go and, since #1529 held the drain
// for a trigger that is still choosing its target, end to end from a
// cast in trigger_batch_order_test.go (prowess plus Caldera Pyremaw).

// settleWithoutPrompts passes priority until the stack settles and
// fails on ANY pending choice — the whole point of the auto-ordered
// shape is that the controller is never asked anything.
func settleWithoutPrompts(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if len(g.PendingChoices) > 0 {
			t.Fatalf("a prompt opened while an all-prowess batch settled: %+v", g.PendingChoices[0])
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				continue
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
}

// prowessItemsInFlight counts every prowess item on or headed for the
// stack, and checks each carries the #1511 mark.
func prowessItemsInFlight(t *testing.T, g *game.Game) int {
	t.Helper()
	n := 0
	check := func(item *game.StackItem) {
		if item == nil || item.Label != "Prowess — +1/+1 until end of turn" {
			return
		}
		n++
		if !item.Commutes {
			t.Errorf("prowess item %s from %s is not marked Commutes", item.ID, item.SourceCardID)
		}
	}
	for _, item := range g.StackMeta {
		check(item)
	}
	for _, item := range g.PendingTriggers {
		check(item)
	}
	return n
}

// TestProwessBoardSkipsTheOrderPrompt — two and then three prowess
// creatures, each a different source, put their triggers on the stack
// with no CR 603.3b prompt, and every one of them resolves.
func TestProwessBoardSkipsTheOrderPrompt(t *testing.T) {
	for _, n := range []int{2, 3} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		var monks []uuid.UUID
		for i := 0; i < n; i++ {
			monks = append(monks, pushProwessCreature(g, me.ID, "Monk", 1, 2))
		}

		castNoncreature(t, g)
		if ch := triggerOrderPrompt(g); ch != nil {
			t.Fatalf("%d prowess creatures: a trigger_order prompt opened for %d items", n, len(ch.TriggerOrderIDs))
		}
		if got := prowessItemsInFlight(t, g); got != n {
			t.Fatalf("%d prowess creatures: %d prowess items in flight, want %d", n, got, n)
		}
		settleWithoutPrompts(t, g)
		for _, id := range monks {
			if p, tg := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 2 || tg != 3 {
				t.Errorf("%d prowess creatures: %s is %d/%d, want 2/3", n, id, p, tg)
			}
		}
	}
}

// TestProwessWithOneUntargetedOtherTriggerStillPrompts — the rule is
// "every item commutes", not "all but one". Monastery Mentor's token
// trigger is untargeted and reads nothing, but a token entering can
// trigger something that reads what the pumps change, so where it sits
// among them is left to its controller.
func TestProwessWithOneUntargetedOtherTriggerStillPrompts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Monastery Mentor", OracleID: oracleMonasteryMentor,
		TypeLine: "Creature — Human Monk", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})

	castNoncreature(t, g)
	if triggerOrderPrompt(g) == nil {
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if triggerOrderPrompt(g) == nil {
		t.Fatal("Mentor's prowess plus its token trigger reached the stack with no ordering prompt")
	}
	settleProwess(t, g)
}

// TestProwessAutoOrderRoundTripsUndo — undo restores a clone. Taken
// before the cast, the restore rewinds the pumps; taken with the
// triggers in flight, the clone's items still carry the mark and the
// restored batch settles without a prompt and pumps exactly once.
func TestProwessAutoOrderRoundTripsUndo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushProwessCreature(g, me.ID, "Monk A", 1, 1)
	b := pushProwessCreature(g, me.ID, "Monk B", 1, 1)
	// Reach the main phase first so the clone and the cast see the
	// same step.
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	beforeCast := g.Clone()
	castNoncreature(t, g)
	inFlight := g.Clone()
	if got := prowessItemsInFlight(t, inFlight); got != 2 {
		t.Fatalf("the in-flight clone holds %d prowess items, want 2", got)
	}
	settleWithoutPrompts(t, g)
	if pa, pb := effectivePower(t, g, a), effectivePower(t, g, b); pa != 2 || pb != 2 {
		t.Fatalf("before undo: powers %d/%d, want 2/2", pa, pb)
	}

	// Undo the settle: back to the triggers in flight.
	g.RestoreFrom(inFlight)
	if triggerOrderPrompt(g) != nil {
		t.Fatal("the restored in-flight batch opened an ordering prompt")
	}
	settleWithoutPrompts(t, g)
	if pa, pb := effectivePower(t, g, a), effectivePower(t, g, b); pa != 2 || pb != 2 {
		t.Errorf("after restoring the in-flight batch: powers %d/%d, want 2/2 (once each)", pa, pb)
	}

	// Undo the cast: nothing pumped, and casting again auto-orders.
	g.RestoreFrom(beforeCast)
	if pa, pb := effectivePower(t, g, a), effectivePower(t, g, b); pa != 1 || pb != 1 {
		t.Fatalf("after undoing the cast: powers %d/%d, want 1/1", pa, pb)
	}
	castNoncreature(t, g)
	if triggerOrderPrompt(g) != nil {
		t.Fatal("the re-cast after undo opened an ordering prompt")
	}
	settleWithoutPrompts(t, g)
	if pa, pb := effectivePower(t, g, a), effectivePower(t, g, b); pa != 2 || pb != 2 {
		t.Errorf("after the re-cast: powers %d/%d, want 2/2", pa, pb)
	}
}
