package game

import (
	"testing"

	"github.com/google/uuid"
)

// next_loss_replacement_test.go — #6244, Stunning Reversal: "The next
// time you would lose the game this turn, instead draw seven cards and
// your life total becomes 1." CR 614.1a (an "instead" is a replacement),
// CR 614.5 (it gets one opportunity), CR 104.3b (a life total of 0 or
// less loses at the next state-based check).

// At 0 life the loss is replaced: seven cards drawn, life 1. The record
// is used up, so the second loss that turn stands. And the replacement
// survives a restore while it waits.
func TestNextLossThisTurnIsReplacedOnce(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	p := g.Seats[1]
	src := uuid.New()
	g.WithWriteLock(func() {
		if !g.ReplaceNextLossThisTurnForEffect(p.ID, src, 7, 1, "Stunning Reversal") {
			t.Fatal("no record written")
		}
	})
	g = restoreStrictForTest(t, g)
	p = g.Seats[1]
	hand := p.Hand.Size()
	g.WithWriteLock(func() { p.Life = 0 })
	g.RunStateChecksForTest()
	if p.Eliminated || p.Life != 1 {
		t.Fatalf("eliminated %v at %d life, want the loss replaced at 1 life", p.Eliminated, p.Life)
	}
	if got := p.Hand.Size() - hand; got != 7 {
		t.Errorf("drew %d cards, want 7", got)
	}
	if ev, ok := lastEvent(g, EventGameLossReplaced); !ok || ev.Actor != p.ID || ev.Label != string(LossLife) {
		t.Errorf("replacement event %+v", ev)
	}
	if g.NextLossReplacementsForLocked(p.ID) != 0 {
		t.Error("the record was not used up")
	}
	g.WithWriteLock(func() { p.Life = 0 })
	g.RunStateChecksForTest()
	if !p.Eliminated {
		t.Error("the second loss that turn was replaced")
	}
}

// An effect loss is replaced too; another player's loss is not; a
// concession never is.
func TestNextLossThisTurnCoversOnlyItsPlayer(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	me, other := g.Seats[1], g.Seats[2]
	g.WithWriteLock(func() {
		g.ReplaceNextLossThisTurnForEffect(me.ID, uuid.Nil, 7, 1, "Stunning Reversal")
		if lost, _ := g.LoseTheGameForEffect(other.ID, uuid.Nil); !lost {
			t.Error("another player's loss was replaced")
		}
		if lost, _ := g.LoseTheGameForEffect(me.ID, uuid.Nil); lost {
			t.Error("my effect loss was not replaced")
		}
	})
	if me.Life != 1 {
		t.Errorf("life = %d, want 1", me.Life)
	}
}

// "This turn": a record not used by the cleanup step ends with it.
func TestNextLossThisTurnEndsAtEndOfTurn(t *testing.T) {
	g := newActiveGameWithSeats(t, 2)
	p := g.Seats[1]
	g.WithWriteLock(func() { g.ReplaceNextLossThisTurnForEffect(p.ID, uuid.Nil, 7, 1, "Stunning Reversal") })
	if err := g.PassTurn(); err != nil {
		t.Fatal(err)
	}
	if g.NextLossReplacementsForLocked(p.ID) != 0 {
		t.Fatal("the record outlived its turn")
	}
	g.WithWriteLock(func() { p.Life = 0 })
	g.RunStateChecksForTest()
	if !p.Eliminated {
		t.Error("a loss next turn was replaced")
	}
}

func TestNextLossInsteadModProblem(t *testing.T) {
	if nextLossInsteadModProblem(Mod{Kind: ModNextLossInstead, Amount: 7, Power: 1}) == "" {
		t.Error("a record naming no player was accepted")
	}
	if nextLossInsteadModProblem(Mod{Kind: ModNextLossInstead, Player: uuid.New(), Amount: -1}) == "" {
		t.Error("a negative draw was accepted")
	}
}
