package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// counter_unless_paid_test.go — #951's enumerator half.
//
// #794's rule is that the engine gate and this enumerator answer "does
// this prompt stop the table" from ONE predicate. #951 adds a reason a
// pay_unless can block — its decline counters a spell still on the
// stack — and the two sides have to pick it up together, or a bot seat
// plays on through a halt the engine is enforcing and every move it
// offers is refused.

// wardTaxOnAStackSpell puts a spell controlled by `payer` on the stack
// and queues the ward tax that guards it.
func wardTaxOnAStackSpell(t *testing.T, g *game.Game, payer *game.Player) uuid.UUID {
	t.Helper()
	spell := uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(game.Card{
			InstanceID: spell, Name: "Doom Blade", TypeLine: "Instant",
			Owner: payer.ID, Controller: payer.ID,
		})
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*game.StackItem)
		}
		g.StackMeta[spell] = &game.StackItem{
			ID: spell, Kind: game.StackItemSpell,
			Controller: payer.ID, Owner: payer.ID, SourceCardID: spell,
		}
		if err := g.QueueCounterUnlessPaidForEffect(game.CounterUnlessPaidPrompt{
			StackItem: spell,
			Source:    uuid.New(),
			Cost:      "{2}",
			Question:  "Ward — pay {2} or the spell is countered",
		}); err != nil {
			t.Fatalf("QueueCounterUnlessPaidForEffect: %v", err)
		}
	})
	return spell
}

// TestAWardTaxStopsEveryOtherSeatAndIsStillOfferedToThePayer. The
// payer is offered their own question and nothing else (#794's second
// half, unchanged); every other seat is offered nothing, which is the
// halt — and the halt has to be answerable or it is a wedge.
func TestAWardTaxStopsEveryOtherSeatAndIsStillOfferedToThePayer(t *testing.T) {
	g := newTable(t)
	active, payer := g.Seats[0], g.Seats[1]
	wardTaxOnAStackSpell(t, g, payer)

	mine := legal.EnumerateFor(g, payer.ID)
	if len(mine) == 0 {
		t.Fatal("the payer was offered nothing — a halt nobody can answer is a wedge")
	}
	for _, m := range mine {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("offered %q (%s) while owing a ward tax", m.Label, m.Type)
		}
	}
	if moves := legal.EnumerateFor(g, active.ID); len(moves) != 0 {
		t.Errorf("the active seat was offered %v while a ward tax is open", labels(moves))
	}

	// Every answer on offer is one the engine accepts (dispatchAll
	// plays each against a clone), and answering for real ends the
	// halt for everybody. Both halves matter: an unanswerable halt and
	// a halt that survives its answer wedge the table the same way.
	dispatchAll(t, g, payer.ID, mine)
	if err := actions.Dispatch(g, actions.Action{
		Type:   actions.Type(mine[0].Type),
		Player: mine[0].Player,
		Caller: payer.ID,
		Params: mine[0].Params,
	}); err != nil {
		t.Fatalf("the payer could not answer the tax: %v", err)
	}
	if moves := legal.EnumerateFor(g, active.ID); len(moves) == 0 {
		t.Error("the table is still stopped after the ward tax was answered")
	}
}

// And the narrowing really is a narrowing: a Rhystic tax raised with
// the same spell on the stack still lets the table play, which is what
// ADR 0018 §6 decided and what #951 must not take away.
func TestARhysticTaxStillLeavesTheOtherSeatsPlaying(t *testing.T) {
	g := newTable(t)
	active, payer := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if err := g.QueuePayUnlessForEffect(payer.ID, uuid.New(), "{1}",
			"Rhystic Study — pay {1}?", func(*game.Game) error { return nil }); err != nil {
			t.Fatalf("QueuePayUnlessForEffect: %v", err)
		}
	})
	if moves := legal.EnumerateFor(g, active.ID); len(moves) == 0 {
		t.Error("an ordinary pay_unless stopped the active seat — ADR 0018 §6 says it must not")
	}
}
