package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// choose_player_test.go — #929's enumerator half. A choose-a-player
// prompt is an option pick whose branches are SEATS, so the existing
// case enumerates it; what this pins is that the seats reach a bot as
// answers it can take, and in the order the policy documented in
// docs/bot.md expects (most life first, which is the always-legal
// first offer a policy with nothing better to say takes).

func TestChoosePlayerEnumeratesOneAnswerPerEligibleSeat(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	g.Seats[1].Life = 11
	g.Seats[2].Life = 44
	g.Seats[3].Eliminated = true
	g.WithWriteLock(func() {
		g.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
			Chooser:  me.ID,
			Among:    []uuid.UUID{g.Seats[1].ID, g.Seats[2].ID, g.Seats[3].ID},
			Question: "Slithermuse",
			Item:     &game.StackItem{ID: uuid.New(), Controller: me.ID},
			Then:     func(*game.Game, uuid.UUID) error { return nil },
		})
	})

	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) != 2 {
		t.Fatalf("enumerated %d answers, want one per eligible seat: %v", len(moves), labels(moves))
	}
	if got := moves[0].Label; got != "Slithermuse: "+g.Seats[2].Name {
		t.Errorf("first offer %q, want the highest-life seat %q", got, g.Seats[2].Name)
	}
	if !hasLabel(moves, "Slithermuse: "+g.Seats[1].Name) {
		t.Errorf("the other eligible seat is offered too: %v", labels(moves))
	}
	if hasLabel(moves, "Slithermuse: "+g.Seats[3].Name) {
		t.Errorf("a seat that has left is not offered: %v", labels(moves))
	}
	for _, m := range moves {
		if m.Cost != nil {
			t.Errorf("naming a seat costs nothing: %q %+v", m.Label, m.Cost)
		}
	}
	dispatchAll(t, g, me.ID, moves)
}
