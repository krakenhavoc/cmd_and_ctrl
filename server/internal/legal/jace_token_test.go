package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// jace_token_test.go — the enumerator half of ADR 0139 (#2796). The
// Jace planeswalker token's loyalty abilities live in a TOKEN template,
// reached through the token key rather than an oracle ID, so they are a
// fresh chance for the enumerator and ActivateCatalogAbility to
// disagree (#544). And Empower Jace with two Jaces asks which one, and
// a seat owing that question must be offered an answer the engine
// takes.

// seatJaceToken creates a Jace token for p through the ordinary
// creation path, with `loyalty` counters on it.
func seatJaceToken(t *testing.T, g *game.Game, p *game.Player, loyalty int) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	g.WithWriteLock(func() {
		ids, err := g.CreateTokensForEffect(p.ID, effects.JaceToken(), 1, game.TokenEntryOptions{})
		if err != nil || len(ids) != 1 {
			t.Fatalf("create the Jace token: %v %v", ids, err)
		}
		id = ids[0]
		if err := g.AddCounterForEffect(id, game.CounterLoyalty, loyalty); err != nil {
			t.Fatalf("loyalty: %v", err)
		}
	})
	return id
}

// At 2 loyalty only the −1 is affordable (CR 606.6), it is offered with
// its loyalty cost, and the dispatcher takes it.
func TestJaceTokenLoyaltyAbilitiesAreOfferedAndAccepted(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	advanceTo(t, g, game.StepPrecombatMain)
	jace := seatJaceToken(t, g, active, 2)

	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, jace)
	if len(acts) != 1 {
		t.Fatalf("%d activations offered on a 2-loyalty Jace token, want the −1 alone: %v", len(acts), labels(acts))
	}
	if acts[0].Cost == nil || acts[0].Cost.Loyalty != -1 {
		t.Errorf("offered %q without its −1 loyalty cost", acts[0].Label)
	}
	dispatchAll(t, g, active.ID, acts)

	// At 3 the −3 joins it.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(jace, game.CounterLoyalty, 1) })
	if n := len(activationsOf(legal.EnumerateFor(g, active.ID), jace)); n != 2 {
		t.Fatalf("%d activations offered at 3 loyalty, want 2", n)
	}
}

// Empower Jace with two Jace tokens asks which one gets the counters;
// the seat is offered answers, every one of them is accepted, and the
// counters land on the one it named.
func TestEmpowerJaceChoiceIsOfferedAndAccepted(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	advanceTo(t, g, game.StepPrecombatMain)
	a := seatJaceToken(t, g, active, 1)
	b := seatJaceToken(t, g, active, 1)

	var err error
	g.WithWriteLock(func() {
		err = effects.EmpowerJace{N: 3}.Apply(effects.NewContext(g, &game.StackItem{
			Kind: game.StackItemSpell, Controller: active.ID, Owner: active.ID,
		}))
	})
	if err != nil {
		t.Fatalf("EmpowerJace: %v", err)
	}
	moves := legal.EnumerateFor(g, active.ID)
	choices := 0
	for _, m := range moves {
		if m.Kind == legal.KindChoice {
			choices++
		}
	}
	if choices == 0 {
		t.Fatalf("no answer offered to the Empower Jace question: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)

	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOwnPermanents {
			if err := g.ResolveOwnPermanents(c.ID, active.ID, []uuid.UUID{b}); err != nil {
				t.Fatalf("answer: %v", err)
			}
		}
	}
	ca, _ := g.LookupCardForEffect(a)
	cb, _ := g.LookupCardForEffect(b)
	if ca.Counters[game.CounterLoyalty] != 1 || cb.Counters[game.CounterLoyalty] != 4 {
		t.Fatalf("loyalty after the answer: %d / %d, want 1 / 4", ca.Counters[game.CounterLoyalty], cb.Counters[game.CounterLoyalty])
	}
}
