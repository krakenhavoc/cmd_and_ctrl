package game

import (
	"testing"

	"github.com/google/uuid"
)

// loop_progress_test.go — #2450, ADR 0055's amendment of 2026-10-07,
// option B: progress that can happen only finitely often restarts the
// loop runs. A new lowest life total this turn, a new highest poison
// count, a player leaving the game. Nothing else does.
//
// A life low and a poison high are marked (TurnTally.LoopProgressed)
// and applied as the next triggered ability begins to resolve; the
// whole-engine test in internal/aiseat (game_ending_loop_test.go)
// drives that end to end. A player leaving restarts the runs at once.

func progressMarked(g *Game) bool { return g.TurnTally.LoopProgressed }

func clearProgress(g *Game) { g.TurnTally.LoopProgressed = false }

func TestANewLifeLowIsProgressAndAHigherTotalIsNot(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	opp := g.Seats[1]
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -3) }) // 37: the turn's first low
	if !progressMarked(g) {
		t.Fatal("a new life low was not marked as progress")
	}
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 5) // 42
		clearProgress(g)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -2) // 40: above the low of 37
	})
	if progressMarked(g) {
		t.Fatal("a loss that is not a new low was marked as progress")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -4) }) // 36
	if !progressMarked(g) {
		t.Fatal("a new low of 36 was not marked as progress")
	}
}

func TestANewPoisonHighIsProgress(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	opp := g.Seats[1]
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(opp.ID, CounterPoison, 1); err != nil {
			t.Fatalf("poison: %v", err)
		}
	})
	if !progressMarked(g) {
		t.Fatal("a new poison high was not marked as progress")
	}
}

func TestADecisionClearsTheProgressMark(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	g.WithWriteLock(func() {
		g.TurnTally.LoopProgressed = true
		g.notePlayerDecisionLocked()
	})
	if progressMarked(g) {
		t.Fatal("the progress mark outlived a decision")
	}
}

func TestAPlayerLeavingRestartsTheRunsAndClearsTheNotice(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	opp := g.Seats[1]
	const key = "test-progress-key"
	g.WithWriteLock(func() {
		g.TurnTally.LoopRun = map[string]int{key: 10}
		g.LoopNotice = &LoopNotice{Source: uuid.New(), Label: "Somebody's Loop", Count: 25}
		g.leaveGameLocked(opp, LossConcede, uuid.Nil)
	})
	if n := g.TurnTally.LoopRun[key]; n != 0 {
		t.Fatalf("run = %d after a player left, want restarted", n)
	}
	if g.LoopNotice != nil {
		t.Errorf("notice still standing after a player left: %+v", g.LoopNotice)
	}
}
