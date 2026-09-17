package actions

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choice_gate_test.go is #730 at the wire seam: the engine's refusal
// reaches the dispatcher unchanged, so a hand-built frame gets the
// same answer the client's own gating gives, and the verbs that exist
// to ANSWER a prompt or to unstick a table are not caught by it.

// queueSacrificePrompt asks seat 0 to sacrifice a creature it
// controls and returns the prompt.
func queueSacrificePrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	me := g.Seats[0]
	c := game.NewCard("Test Creature", me.ID)
	c.TypeLine = "Creature — Test"
	c.Power, c.Toughness = 2, 2
	c.Controller = me.ID
	g.Battlefield.PushTop(c)
	g.WithWriteLock(func() {
		if n := g.PlayerSacrificesForEffect(uuid.New(), me.ID, nil, "sacrifice a creature"); n != 1 {
			t.Fatalf("queued %d prompts, want 1", n)
		}
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1", len(g.PendingChoices))
	}
	return g.PendingChoices[0]
}

// TestDispatchRefusesStepVerbsWhileAChoiceIsPending — advance_step,
// pass_priority and pass_turn all come back as ErrChoicePending, with
// the message naming what is owed.
func TestDispatchRefusesStepVerbsWhileAChoiceIsPending(t *testing.T) {
	for _, verb := range []Type{TypeAdvanceStep, TypePassPriority, TypePassTurn} {
		g := newGame(t)
		prompt := queueSacrificePrompt(t, g)
		before := g.Snapshot().Turn

		a, err := Decode(string(verb), g.Seats[0].ID.String(), nil)
		if err != nil {
			t.Fatalf("Decode %s: %v", verb, err)
		}
		a.Caller = g.Seats[0].ID
		err = Dispatch(g, a)
		if !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("Dispatch %s = %v, want ErrChoicePending", verb, err)
		}
		var cpe *game.ChoicePendingError
		if !errors.As(err, &cpe) || cpe.ChoiceID != prompt.ID {
			t.Errorf("%s: the error does not name the outstanding prompt: %v", verb, err)
		}
		if got := g.Snapshot().Turn; got != before {
			t.Errorf("%s: a refused verb moved the cursor to %v", verb, got)
		}
	}
}

// TestDispatchStillTakesTheAnswerAndTheUnstickVerbs — the gate must
// never be able to wedge a table. The answer itself, concede, and the
// admin context menu's raw sandbox moves stay dispatchable.
func TestDispatchStillTakesTheAnswerAndTheUnstickVerbs(t *testing.T) {
	// The answer.
	g := newGame(t)
	prompt := queueSacrificePrompt(t, g)
	victim := prompt.SacrificeOptions[0]
	a, err := Decode(string(TypeResolveChoice), g.Seats[0].ID.String(), params(t, map[string]any{
		"choice_id": prompt.ID.String(),
		"card_ids":  []string{victim.String()},
	}))
	if err != nil {
		t.Fatalf("Decode resolve_choice: %v", err)
	}
	a.Caller = g.Seats[0].ID
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("resolve_choice while the prompt is open: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("the prompt survived its own answer: %+v", g.PendingChoices)
	}

	// The admin / sandbox unstick moves, each with a prompt open.
	for _, tc := range []struct {
		name   string
		verb   Type
		params any
	}{
		{"move_card", TypeMoveCard, nil},
		{"change_life", TypeChangeLife, map[string]any{"delta": -1}},
		{"concede", TypeConcede, nil},
	} {
		g := newGame(t)
		prompt := queueSacrificePrompt(t, g)
		seat := g.Seats[0]
		raw := []byte(nil)
		if tc.params != nil {
			raw = params(t, tc.params)
		}
		if tc.verb == TypeMoveCard {
			raw = params(t, map[string]any{
				"src":         map[string]any{"kind": "battlefield"},
				"dst":         map[string]any{"kind": "graveyard", "player": seat.ID.String()},
				"instance_id": prompt.SacrificeOptions[0].String(),
			})
		}
		a, err := Decode(string(tc.verb), seat.ID.String(), raw)
		if err != nil {
			t.Fatalf("Decode %s: %v", tc.name, err)
		}
		a.Caller = seat.ID
		if err := Dispatch(g, a); errors.Is(err, game.ErrChoicePending) {
			t.Errorf("%s was gated by the pending choice; it is an unstick verb", tc.name)
		}
	}
}

// TestDispatchDoesNotGateOnAPayUnlessPrompt — ADR 0018 §6's accepted
// looseness, asserted at the seam a client actually uses.
func TestDispatchDoesNotGateOnAPayUnlessPrompt(t *testing.T) {
	g := newGame(t)
	g.WithWriteLock(func() {
		if err := g.QueuePayUnlessForEffect(g.Seats[1].ID, uuid.New(), "{1}", "tax", nil); err != nil {
			t.Fatalf("QueuePayUnlessForEffect: %v", err)
		}
	})
	a, err := Decode(string(TypeAdvanceStep), "", nil)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("advance_step with a pay_unless open: %v", err)
	}
}
