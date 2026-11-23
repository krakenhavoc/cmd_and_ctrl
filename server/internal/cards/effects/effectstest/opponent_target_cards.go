package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opponent_target_cards_test.go — #2055, ADR 0019's 2026-10-16
// amendment: "target … of an opponent's choice". Each card is announced
// for real: the controller's payload names its own targets only, the
// controller names an opponent when more than one could choose, that
// opponent answers a pick_target, and only then is anything paid or put
// on the stack.

const (
	OtNovaPentacle      = "db121503-8a34-498a-829c-72c33798369b"
	OtEvangelize        = "23a60803-6b0f-48ee-aa3b-12a3c5fc62cf"
	OtArena             = "12326d23-1fb5-4084-a23a-d6e2c90ea36c"
	OtMagusOfTheArena   = "44865261-16f8-42d2-a388-a57173142eb0"
	OtCuombajjWitches   = "638eeb16-9e0e-4cc6-b97e-8ff0df81ca58"
	OtVolcanicOffering  = "05e0a10d-532a-451c-a54e-27e64969fe0e"
	OtMausoleumTurnkey  = "d4a61685-7149-4d14-bc17-85b060945687"
	OtOpponentChoiceTag = "of an opponent's choice"
)

// otOptionPick is the open option pick addressed to `chooser`.
func OtOptionPick(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser == chooser {
			return c
		}
	}
	t.Fatalf("no option pick open for %s", chooser)
	return nil
}

// otOfferedSeats is the seats an option pick offers.
func OtOfferedSeats(c *game.PendingChoice) []uuid.UUID {
	var out []uuid.UUID
	for _, o := range c.PickOptions {
		out = append(out, o.Player)
	}
	return out
}

// otChooseOpponent answers the controller's "which opponent" question.
func OtChooseOpponent(t *testing.T, g *game.Game, chooser, opp uuid.UUID) {
	t.Helper()
	c := OtOptionPick(t, g, chooser)
	for i, o := range c.PickOptions {
		if o.Player == opp {
			if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			return
		}
	}
	t.Fatalf("opponent %s is not offered (%v)", opp, OtOfferedSeats(c))
}

// otPickFor is the open pick_target addressed to `chooser`, asserted to
// be an opponent's pick for `controller`.
func OtPickFor(t *testing.T, g *game.Game, chooser, controller uuid.UUID) *game.PendingChoice {
	t.Helper()
	p := LatestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target open for %s", chooser)
	}
	if p.PickTargetFor != controller {
		t.Fatalf("pick_target is for %s, want %s", p.PickTargetFor, controller)
	}
	return p
}

// otAnswer answers chooser's open pick_target with refs.
func OtAnswer(t *testing.T, g *game.Game, chooser uuid.UUID, refs ...game.TargetRef) {
	t.Helper()
	p := LatestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target open for %s", chooser)
	}
	if err := g.ResolvePickTargets(p.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

func OtCard(id uuid.UUID) game.TargetRef   { return game.TargetRef{Kind: game.TargetCard, ID: id} }
func OtPlayer(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetPlayer, ID: id} }

// otItemFrom is the stack item whose source is `source`.
func OtItemFrom(g *game.Game, source uuid.UUID) *game.StackItem {
	for _, it := range g.StackMeta {
		if it != nil && it.SourceCardID == source {
			return it
		}
	}
	return nil
}
