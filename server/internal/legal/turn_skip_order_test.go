package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// turn_skip_order_test.go — ADR 0188 Decision 12 (#3561): while the turn
// change waits on the which-skip question, the chooser is offered its
// two answers, "use my skip" first and always legal, and nobody is
// offered anything else. Both answers are accepted by the dispatcher.

// pauseOnTheWhichSkipQuestion queues an extra turn for the active seat,
// which owes a skip, under a replacement that skips that seat's extra
// turns, and passes the turn.
func pauseOnTheWhichSkipQuestion(t *testing.T, g *game.Game) *game.Player {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	seat := active.Seat
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(game.ReplacementEffect{
			Watches: []game.EventKind{game.EventExtraTurnBegin},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventExtraTurn && ev.ExtraTurnSeat == seat
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			PureCancel: true,
			Label:      "Trouble in Pairs",
		})
		g.SkipNextTurnsForEffect(active.ID, uuid.Nil, 1, "skip your next turn")
		g.TakeExtraTurnsForEffect(active.ID, uuid.Nil, 1)
	})
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if !g.TurnChangePending {
		t.Fatal("the turn change did not pause")
	}
	return active
}

func TestTheWhichSkipQuestionOffersOnlyItsTwoAnswers(t *testing.T) {
	g := newTable(t)
	chooser := pauseOnTheWhichSkipQuestion(t, g)

	moves := legal.EnumerateFor(g, chooser.ID)
	if len(moves) != 2 {
		t.Fatalf("chooser offered %d moves, want 2: %v", len(moves), labels(moves))
	}
	for i, want := range []bool{true, false} {
		var p struct {
			Apply *bool `json:"apply"`
		}
		if err := json.Unmarshal(moves[i].Params, &p); err != nil || p.Apply == nil || *p.Apply != want {
			t.Errorf("move %d %q: apply %v, want %v", i, moves[i].Label, p.Apply, want)
		}
	}
	if !moves[0].AlwaysLegal || moves[1].AlwaysLegal {
		t.Errorf("always-legal marks %v / %v, want the use-my-skip answer only", moves[0].AlwaysLegal, moves[1].AlwaysLegal)
	}
	dispatchAll(t, g, chooser.ID, moves)

	for _, p := range g.Seats {
		if p.ID == chooser.ID {
			continue
		}
		if got := legal.EnumerateFor(g, p.ID); len(got) != 0 {
			t.Errorf("%s offered %d moves while the turn change waits: %v", p.Name, len(got), labels(got))
		}
	}
}
