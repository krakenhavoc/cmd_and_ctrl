package actions

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// table_roll_test.go — ADR 0121 §5 at the dispatch layer: roll_table_die
// is player-scoped, takes {die}, passes the opening roll's gate, and
// mints no undo entry.

func lastTableRoll(t *testing.T, g *game.Game) game.Event {
	t.Helper()
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == game.EventTableRoll {
			return g.Events[i]
		}
	}
	t.Fatal("no table roll in the log")
	return game.Event{}
}

func TestRollTableDieThroughDispatch(t *testing.T) {
	g := newGame(t)
	a, b := g.Seats[0].ID, g.Seats[1].ID
	roll := func(player, caller uuid.UUID, params string) error {
		return Dispatch(g, Action{Type: TypeRollTableDie, Player: player, Caller: caller, Params: []byte(params)})
	}

	if err := roll(b, a, `{"die":"d20"}`); !errors.Is(err, ErrPlayerCallerMismatch) {
		t.Fatalf("seat 0 rolled for seat 1: %v", err)
	}
	if err := roll(uuid.Nil, a, `{"die":"d20"}`); !errors.Is(err, ErrInvalidPlayer) {
		t.Fatalf("a roll with no player: %v", err)
	}
	if err := roll(a, a, `{}`); !errors.Is(err, ErrMissingParams) {
		t.Fatalf("a roll with no die: %v", err)
	}
	if err := roll(a, a, `{"die":"d100"}`); !errors.Is(err, game.ErrUnknownTableDie) {
		t.Fatalf("a d100: %v", err)
	}
	for _, tc := range []struct {
		die   string
		sides int
	}{{"d6", 6}, {"d20", 20}, {"coin", 0}} {
		if err := roll(a, a, `{"die":"`+tc.die+`"}`); err != nil {
			t.Fatalf("%s: %v", tc.die, err)
		}
		if ev := lastTableRoll(t, g); ev.Actor != a || ev.Sides != tc.sides {
			t.Errorf("%s rolled %+v", tc.die, ev)
		}
	}
	// The admin rolls for any seat.
	if err := roll(b, uuid.Nil, `{"die":"d6"}`); err != nil {
		t.Fatal(err)
	}
	if ev := lastTableRoll(t, g); ev.Actor != b {
		t.Errorf("the admin's roll for seat 1 is %+v", ev)
	}
}

// Decision 4: legal at any time while the game is active, the opening
// roll included, and it leaves the roll exactly where it was.
func TestRollTableDiePassesTheOpeningRollGate(t *testing.T) {
	g := newOpeningRollGame(t)
	p := g.Seats[2].ID
	if err := Dispatch(g, Action{Type: TypeRollTableDie, Player: p, Caller: p, Params: []byte(`{"die":"coin"}`)}); err != nil {
		t.Fatalf("a table roll during the opening roll: %v", err)
	}
	if !g.OpeningRollOpen() || len(g.OpeningRoll.Rounds[0].Rolls) != 0 {
		t.Fatalf("a table roll moved the opening roll: %+v", g.OpeningRoll)
	}
	if !MintsNoUndo(g, TypeRollTableDie) || !MintsNoUndo(newGame(t), TypeRollTableDie) {
		t.Fatal("roll_table_die mints an undo entry")
	}
}
