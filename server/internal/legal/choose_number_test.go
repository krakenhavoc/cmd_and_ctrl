package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// choose_number_test.go — ADR 0129's amendment of 2026-10-09 (#1941):
// the enumerator's answers to a number chosen, or life paid, as an
// ability resolves.

func moveAmounts(t *testing.T, moves []legal.Move) []int {
	t.Helper()
	var out []int
	for _, m := range moves {
		var p struct {
			Amount *int `json:"amount"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil || p.Amount == nil {
			t.Fatalf("move %q carries no amount: %s", m.Label, m.Params)
		}
		out = append(out, *p.Amount)
	}
	return out
}

// A number with no ceiling is a handful of moves: the floor, the card's
// goal and its marks, never the overflow guard. Nothing is paid, and
// every move is accepted.
func TestChooseNumberWithNoCeilingMoves(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		_ = g.QueueChooseNumberForEffect(game.ChooseNumber{
			Chooser: active.ID, Resource: game.PayResourceNone, NoMax: true,
			Goal: 3, Marks: []int{3, active.Life}, Unit: game.PayAmountDamage, SelfDamage: true, Question: "how much",
		})
	})
	moves := legal.EnumerateFor(g, active.ID)
	got := moveAmounts(t, moves)
	if len(got) != 3 || got[0] != 0 || got[1] != 3 || got[2] != active.Life {
		t.Fatalf("amounts = %v, want [0 3 %d]", got, active.Life)
	}
	for i, m := range moves {
		if m.Cost != nil {
			t.Errorf("%q: cost %+v, want none (nothing is paid)", m.Label, m.Cost)
		}
		if m.AlwaysLegal != (i == 0) {
			t.Errorf("%q: AlwaysLegal = %v", m.Label, m.AlwaysLegal)
		}
	}
	if moves[1].Label != "how much: choose 3" {
		t.Errorf("label %q", moves[1].Label)
	}
	dispatchAll(t, g, active.ID, moves)
}

// A life payment offers nothing, the goal and the whole life total, each
// priced in life.
func TestPayLifeAmountMoves(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	active.Life = 15
	g.WithWriteLock(func() {
		_ = g.QueueChooseNumberForEffect(game.ChooseNumber{
			Chooser: active.ID, Resource: game.PayResourceLife, Goal: 4, Unit: game.PayAmountCards, Question: "pay life",
		})
	})
	moves := legal.EnumerateFor(g, active.ID)
	got := moveAmounts(t, moves)
	if len(got) != 3 || got[0] != 0 || got[1] != 4 || got[2] != 15 {
		t.Fatalf("amounts = %v, want [0 4 15]", got)
	}
	for _, m := range moves[1:] {
		if m.Cost == nil || m.Cost.Life == 0 || m.Cost.Energy != 0 {
			t.Errorf("%q: cost %+v, want life", m.Label, m.Cost)
		}
	}
	if moves[0].Label != "pay life: pay no life" || moves[1].Label != "pay life: pay 4 life" {
		t.Errorf("labels %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}
