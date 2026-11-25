package game

import (
	"testing"

	"github.com/google/uuid"
)

// game_draw_test.go — #6198. CR 104.4c: "An effect may state that the
// game is a draw." CR 104.4a: "If all the players remaining in a game
// lose simultaneously, the game is a draw."

// Divine Intervention: the game ends as a draw for everyone still in it,
// naming its source; nobody wins.
func TestDrawTheGameForEffectEndsTheGameAsADraw(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	src := uuid.New()
	var ended bool
	g.WithWriteLock(func() { ended = g.DrawTheGameForEffect(src) })
	if !ended || g.State != StateEnded {
		t.Fatalf("ended=%v state=%v, want an ended game", ended, g.State)
	}
	o := g.Result()
	if o == nil || o.Kind != OutcomeDraw || o.Cause != OutcomeCauseDrawEffect || o.Source != src || o.Winner != uuid.Nil {
		t.Fatalf("outcome = %+v, want a draw by effect naming the source", o)
	}
	if g.OutcomeKind() != OutcomeDraw {
		t.Errorf("OutcomeKind = %q, want draw", g.OutcomeKind())
	}
	// A second end changes nothing.
	g.WithWriteLock(func() { ended = g.DrawTheGameForEffect(uuid.New()) })
	if ended || g.Result().Source != src {
		t.Error("an ended game was ended again")
	}
}

// Celestial Convergence: a single leader wins; a tie for the highest
// life total is a draw; a player who has left is not compared.
func TestHighestLifeWinsOrDraw(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	for i, p := range g.Seats {
		p.Life = 20 + i
	}
	g.WithWriteLock(func() {
		if _, err := g.HighestLifeWinsOrDrawForEffect(uuid.New()); err != nil {
			t.Fatal(err)
		}
	})
	if o := g.Result(); o == nil || o.Kind != OutcomeWin || o.Winner != g.Seats[3].ID {
		t.Fatalf("outcome = %+v, want seat 3 (23 life) to win", o)
	}

	g = newFourPlayerActiveGame(t)
	g.Seats[0].Life, g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life = 30, 30, 10, 40
	g.Seats[3].Eliminated = true
	g.WithWriteLock(func() { _, _ = g.HighestLifeWinsOrDrawForEffect(uuid.New()) })
	if o := g.Result(); o == nil || o.Kind != OutcomeDraw || o.Cause != OutcomeCauseDrawEffect {
		t.Fatalf("outcome = %+v, want a draw (tie at 30; the 40-life player has left)", o)
	}
}
