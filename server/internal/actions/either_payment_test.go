package actions

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestDispatchResolveChoiceRoutesPayOption — ADR 0187 §5 (#3575) at the
// wire's front door. An either payment is answered `{apply: true,
// pay_option: n}` with `card_ids` for a branch that names cards. Each
// malformed answer comes back as game.ErrInvalidParam with the prompt
// kept, and the two well-formed answers pay.
func TestDispatchResolveChoiceRoutesPayOption(t *testing.T) {
	g := newGame(t)
	me := g.Seats[0]
	controller := g.Seats[1]
	source := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: source, Name: "Worms", TypeLine: "Enchantment", Owner: controller.ID, Controller: controller.ID})
	var lands []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		lands = append(lands, id)
	}
	paid := 0
	g.WithWriteLock(func() {
		if err := queueAnyPlayerHooked(g, anyPlayerTestOffer{
			Source: source, Controller: controller.ID, Cost: "sacrifice two lands or take 5 damage", Question: "Worms?",
			Action: &game.PayAction{Kind: game.PayActionEither, Either: []game.PayAction{
				{Kind: game.PayActionSacrifice, Count: 2, Of: game.PermanentQuery{Types: []string{"land"}}},
				{Kind: game.PayActionDamage, Count: 5},
			}},
			OnPaid: func(*game.Game, uuid.UUID) error { paid++; return nil },
		}); err != nil {
			t.Fatalf("queue: %v", err)
		}
	})
	choice := pendingChoiceOfKind(g, game.PendingChoicePayUnless)
	if choice == nil || choice.Chooser != me.ID {
		t.Fatalf("pay-unless %+v, want one for the active seat", choice)
	}

	send := func(p map[string]any) error {
		t.Helper()
		p["choice_id"] = choice.ID.String()
		a, err := Decode(string(TypeResolveChoice), me.ID.String(), params(t, p))
		if err != nil {
			t.Fatalf("Decode: %v", err)
		}
		return Dispatch(g, a)
	}
	ids := func(in ...uuid.UUID) []string {
		out := make([]string, 0, len(in))
		for _, id := range in {
			out = append(out, id.String())
		}
		return out
	}

	for name, p := range map[string]map[string]any{
		"apply without pay_option":    {"apply": true},
		"apply with cards, no option": {"apply": true, "card_ids": ids(lands[0], lands[1])},
		"pay_option out of range":     {"apply": true, "pay_option": 2},
		"a negative pay_option":       {"apply": true, "pay_option": -1},
		"pay_option on a decline":     {"apply": false, "pay_option": 1},
		"cards on the damage branch":  {"apply": true, "pay_option": 1, "card_ids": ids(lands[0])},
		"one land for two":            {"apply": true, "pay_option": 0, "card_ids": ids(lands[0])},
	} {
		if err := send(p); !errors.Is(err, game.ErrInvalidParam) {
			t.Errorf("%s: Dispatch error = %v, want ErrInvalidParam", name, err)
		}
		if pendingChoiceOfKind(g, game.PendingChoicePayUnless) == nil {
			t.Fatalf("%s: the refused answer dropped the prompt", name)
		}
	}
	if paid != 0 || me.Life != 40 {
		t.Fatalf("a refused answer paid (%d) or cost life (%d)", paid, me.Life)
	}

	if err := send(map[string]any{"apply": true, "pay_option": 1}); err != nil {
		t.Fatalf("taking the damage: %v", err)
	}
	if me.Life != 35 || paid != 1 {
		t.Errorf("after the damage branch: life %d, paid %d; want 35 and 1", me.Life, paid)
	}
}

// A pay_option on a prompt that is not an either payment is refused by
// the engine, whatever it pays with.
func TestDispatchPayOptionOnAManaPromptIsRefused(t *testing.T) {
	g := newGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := queueAnyPlayerHooked(g, anyPlayerTestOffer{
			Source: uuid.New(), Controller: g.Seats[1].ID, Cost: "{2}", Question: "Tax?",
		}); err != nil {
			t.Fatalf("queue: %v", err)
		}
	})
	choice := pendingChoiceOfKind(g, game.PendingChoicePayUnless)
	a, err := Decode(string(TypeResolveChoice), me.ID.String(), params(t, map[string]any{
		"choice_id": choice.ID.String(), "apply": true, "pay_option": 0,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err := Dispatch(g, a); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("Dispatch error = %v, want ErrInvalidParam", err)
	}
	if pendingChoiceOfKind(g, game.PendingChoicePayUnless) == nil {
		t.Fatal("the refused answer dropped the prompt")
	}
}
