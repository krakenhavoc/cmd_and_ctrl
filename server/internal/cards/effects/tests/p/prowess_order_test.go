package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		var monks []uuid.UUID
		for i := 0; i < n; i++ {
			monks = append(monks, effectstest.PushProwessCreature(g, me.ID, "Monk", 1, 2))
		}

		effectstest.CastNoncreature(t, g)
		if ch := effectstest.TriggerOrderPrompt(g); ch != nil {
			t.Fatalf("%d prowess creatures: a trigger_order prompt opened for %d items", n, len(ch.TriggerOrderIDs))
		}
		if got := prowessItemsInFlight(t, g); got != n {
			t.Fatalf("%d prowess creatures: %d prowess items in flight, want %d", n, got, n)
		}
		effectstest.SettleWithoutPrompts(t, g)
		for _, id := range monks {
			if p, tg := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 2 || tg != 3 {
				t.Errorf("%d prowess creatures: %s is %d/%d, want 2/3", n, id, p, tg)
			}
		}
	}
}

// A batch with one trigger outside the Commutes class (Monastery
// Mentor's token trigger) is no longer this skip's: the #3013
// independence check orders it, from prowess's declared footprint
// (trigger_independence_prowess_test.go).

// TestProwessAutoOrderRoundTripsUndo — undo restores a clone. Taken
// before the cast, the restore rewinds the pumps; taken with the
// triggers in flight, the clone's items still carry the mark and the
// restored batch settles without a prompt and pumps exactly once.
func TestProwessAutoOrderRoundTripsUndo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.PushProwessCreature(g, me.ID, "Monk A", 1, 1)
	b := effectstest.PushProwessCreature(g, me.ID, "Monk B", 1, 1)
	// Reach the main phase first so the clone and the cast see the
	// same step.
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	beforeCast := g.Clone()
	effectstest.CastNoncreature(t, g)
	inFlight := g.Clone()
	if got := prowessItemsInFlight(t, inFlight); got != 2 {
		t.Fatalf("the in-flight clone holds %d prowess items, want 2", got)
	}
	effectstest.SettleWithoutPrompts(t, g)
	if pa, pb := effectstest.EffectivePower(t, g, a), effectstest.EffectivePower(t, g, b); pa != 2 || pb != 2 {
		t.Fatalf("before undo: powers %d/%d, want 2/2", pa, pb)
	}

	// Undo the settle: back to the triggers in flight.
	g.RestoreFrom(inFlight)
	if effectstest.TriggerOrderPrompt(g) != nil {
		t.Fatal("the restored in-flight batch opened an ordering prompt")
	}
	effectstest.SettleWithoutPrompts(t, g)
	if pa, pb := effectstest.EffectivePower(t, g, a), effectstest.EffectivePower(t, g, b); pa != 2 || pb != 2 {
		t.Errorf("after restoring the in-flight batch: powers %d/%d, want 2/2 (once each)", pa, pb)
	}

	// Undo the cast: nothing pumped, and casting again auto-orders.
	g.RestoreFrom(beforeCast)
	if pa, pb := effectstest.EffectivePower(t, g, a), effectstest.EffectivePower(t, g, b); pa != 1 || pb != 1 {
		t.Fatalf("after undoing the cast: powers %d/%d, want 1/1", pa, pb)
	}
	effectstest.CastNoncreature(t, g)
	if effectstest.TriggerOrderPrompt(g) != nil {
		t.Fatal("the re-cast after undo opened an ordering prompt")
	}
	effectstest.SettleWithoutPrompts(t, g)
	if pa, pb := effectstest.EffectivePower(t, g, a), effectstest.EffectivePower(t, g, b); pa != 2 || pb != 2 {
		t.Errorf("after the re-cast: powers %d/%d, want 2/2", pa, pb)
	}
}
