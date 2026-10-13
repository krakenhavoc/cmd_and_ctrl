package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// trigger_independence_3021_test.go — ADR 0018's #3021 amendment: the
// sub-PR 8 skip (one source, one label) no longer ignores what the
// copies read. Two Righteous Valkyrie gains for two creatures entering
// together read a different toughness each, through a layer-7 anthem
// keyed to the life total the first gain moves.

// Two Angels enter together beside Righteous Valkyrie: one source, one
// label, two subjects. The first gain can cross her +7 line and the
// anthem changes the second Angel's toughness, so the order is asked.
func TestTwoValkyrieGainsForTwoAngelsAsk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Righteous Valkyrie", "Creature — Angel Cleric", b18RighteousValkyrieOracle, 2, 4)
	toMain(t, g)
	a := b21Push(g, me.ID, "Small Angel", "Creature — Angel", "", 1, 1)
	b := b21Push(g, me.ID, "Big Angel", "Creature — Angel", "", 4, 4)
	before := promptsByReason(t, game.TriggerOrderReasonCharacteristic)

	enterTogether(g, me.ID, a, b)
	ch := settleUntilOrderPrompt(t, g)
	if ch == nil || len(ch.TriggerOrderIDs) != 2 {
		t.Fatalf("two Valkyrie gains beside her own anthem must ask, got %+v", ch)
	}
	if got := promptsByReason(t, game.TriggerOrderReasonCharacteristic) - before; got != 1 {
		t.Errorf("counted %v under characteristic, want 1", got)
	}
}

// "Never ask" is the seat's choice and wins over the board; the batch
// resolves in collected order.
func TestTwoValkyrieGainsDoNotAskWhenTheSeatNeverAsks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if err := g.SetTriggerOrderPreference(me.ID, game.TriggerOrderNever); err != nil {
		t.Fatal(err)
	}
	b12Push(g, me.ID, "Righteous Valkyrie", "Creature — Angel Cleric", b18RighteousValkyrieOracle, 2, 4)
	toMain(t, g)
	a := b21Push(g, me.ID, "Small Angel", "Creature — Angel", "", 1, 1)
	b := b21Push(g, me.ID, "Big Angel", "Creature — Angel", "", 4, 4)
	life := me.Life
	enterTogether(g, me.ID, a, b)
	if ch := settleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("a seat that never asks was asked: %+v", ch)
	}
	if me.Life != life+5 {
		t.Errorf("life %d → %d, want +5", life, me.Life)
	}
}

// An answered batch is not asked a second time.
func TestAnAnsweredValkyrieBatchIsNotAskedAgain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Righteous Valkyrie", "Creature — Angel Cleric", b18RighteousValkyrieOracle, 2, 4)
	toMain(t, g)
	a := b21Push(g, me.ID, "Small Angel", "Creature — Angel", "", 1, 1)
	b := b21Push(g, me.ID, "Big Angel", "Creature — Angel", "", 4, 4)
	enterTogether(g, me.ID, a, b)
	ch := settleUntilOrderPrompt(t, g)
	if ch == nil {
		t.Fatal("expected the order prompt")
	}
	// A bot seat is offered both orders of the new prompt, and the
	// engine accepts the one it takes (the enumerator is generic over
	// the prompt, not the card: legal/choices.go).
	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) != 2 {
		t.Fatalf("the seat is offered %d moves for the Valkyrie order prompt, want both orders", len(moves))
	}
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Fatalf("move %q offered while the ordering prompt is open", m.Label)
		}
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(moves[0].Type), Player: moves[0].Player, Caller: me.ID, Params: moves[0].Params}); err != nil {
		t.Fatalf("the offered order was rejected: %v", err)
	}
	if again := settleUntilOrderPrompt(t, g); again != nil {
		t.Fatalf("an answered batch was asked again: %+v", again)
	}
}

// Boards that did not ask before still do not: the copies read their
// subjects, but nothing on the board can change what they read.
func TestSameSourceSubjectReadsDoNotAskWithoutAReadingStatic(t *testing.T) {
	t.Run("Trostani alone", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		b12Push(g, me.ID, "Trostani, Selesnya's Voice", "Legendary Creature — Dryad", trostaniOracle, 2, 5)
		toMain(t, g)
		a := b21Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		b := b21Push(g, me.ID, "Ox", "Creature — Ox", "", 2, 4)
		life := me.Life
		enterTogether(g, me.ID, a, b)
		if ch := settleUntilOrderPrompt(t, g); ch != nil {
			t.Fatalf("two Trostani gains asked for an order: %+v", ch)
		}
		if me.Life != life+6 {
			t.Errorf("life %d → %d, want +6", life, me.Life)
		}
	})
	t.Run("Wolverine Riders beside a lord", func(t *testing.T) {
		// A layer-7 static that does not read a life total, and gains
		// that only move life: nothing the first gain writes can change
		// the second read.
		g := newCatalogGame(t)
		me := g.Seats[0]
		b21Push(g, me.ID, "Wolverine Riders", "Creature — Elf Warrior", b27WolverineRidersOracle, 4, 4)
		pushCatalogPermanent(g, me.ID, "Elvish Archdruid", "Creature — Elf Druid", elvishArchdruidOracle, false)
		toMain(t, g)
		a := b21Push(g, me.ID, "Elf One", "Creature — Elf Druid", "", 1, 1)
		b := b21Push(g, me.ID, "Elf Two", "Creature — Elf Druid", "", 1, 2)
		enterTogether(g, me.ID, a, b)
		if ch := settleUntilOrderPrompt(t, g); ch != nil {
			t.Fatalf("two Riders gains beside a life-blind lord asked: %+v", ch)
		}
	})
}
