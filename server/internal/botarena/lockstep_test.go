package botarena_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
)

// lockstep_test.go pins #1503: `boteval arena --lockstep` makes a seed
// a game rather than a deal, and the concurrent schedule stays the
// default.

// moveLog records an arena game move by move, in a form two runs of
// one seed can be compared in: the seat's order of first appearance,
// the room sequence the move committed at, whether it applied, the
// label, the move's parameters and the policy's reason. Seat and card
// IDs are fresh UUIDs on every run, so each is replaced by the order it
// first appeared in — which is the same exactly when the games are.
//
// It is the arena's twin of aiseat's moveLog (heuristic_game_test.go),
// fed through Config.Runner.Observer, the arena's public observer slot.
type moveLog struct {
	mu  sync.Mutex
	ids map[string]string
	out []string
}

var uuidPattern = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

func (l *moveLog) canon(id string) string {
	if c, ok := l.ids[id]; ok {
		return c
	}
	c := fmt.Sprintf("#%d", len(l.ids))
	l.ids[id] = c
	return c
}

func (l *moveLog) Observe(ev aiseat.DecisionEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.ids == nil {
		l.ids = map[string]string{}
	}
	params := ""
	if ev.Index >= 0 && ev.Index < len(ev.Input.Moves) {
		params = uuidPattern.ReplaceAllStringFunc(string(ev.Input.Moves[ev.Index].Params), l.canon)
	}
	l.out = append(l.out, fmt.Sprintf("%s seq=%d applied=%v %q %s {%s}",
		l.canon(ev.Seat.String()), ev.Seq, ev.Applied, ev.Label, params, ev.Reason))
}

func (l *moveLog) lines() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.out...)
}

// playLogged runs one arena game with a move log attached.
func playLogged(t *testing.T, cfg botarena.Config) (botarena.GameResult, []string) {
	t.Helper()
	log := &moveLog{}
	cfg.Runner.Observer = log
	var got []botarena.GameResult
	if _, err := botarena.Run(context.Background(), cfg, func(r botarena.GameResult) { got = append(got, r) }); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("sink saw %d games, want 1", len(got))
	}
	return got[0], log.lines()
}

// The same seed played twice under --lockstep is the same game: the
// move logs match line for line, and so do the turn count and the
// winner. Four seats, because a four-seat table is where the
// concurrent schedule forks most — three defenders racing the active
// player's pass in every combat.
//
// The ungated half plays two rounds, a few hundred windows and a few
// seconds under -race: this is the one property the flag exists for,
// and a seed that stops replaying is exactly the regression nobody
// notices until a red night cannot be reproduced. The gated half plays
// four heuristic seats to a winner.
func TestArenaLockstepSeedReplaysMoveForMove(t *testing.T) {
	t.Run("opening", func(t *testing.T) {
		assertReplays(t, botarena.Config{
			Seats: []botarena.SeatSpec{
				{Tier: tiers.Heuristic}, {Tier: tiers.Random},
				{Tier: tiers.Heuristic}, {Tier: tiers.Random},
			},
			Games: 1, Seed: 1503, TurnBudget: 2, Wall: 2 * time.Minute, Lockstep: true,
		})
	})
	t.Run("whole game", func(t *testing.T) {
		requireGameTests(t)
		assertReplays(t, botarena.Config{
			Seats: []botarena.SeatSpec{
				{Tier: tiers.Heuristic}, {Tier: tiers.Heuristic},
				{Tier: tiers.Heuristic}, {Tier: tiers.Heuristic},
			},
			Games: 1, Seed: 107, TurnBudget: 50, Wall: 5 * time.Minute, Lockstep: true,
		})
	})
}

func assertReplays(t *testing.T, cfg botarena.Config) {
	t.Helper()
	a, logA := playLogged(t, cfg)
	b, logB := playLogged(t, cfg)
	if a.Stalled || b.Stalled {
		t.Fatalf("a lockstep run stalled:\n%s\n%s", a.StallDump, b.StallDump)
	}
	if len(logA) < 50 {
		t.Fatalf("the first run logged %d moves; too few for the comparison to mean anything", len(logA))
	}
	for i := 0; i < len(logA) && i < len(logB); i++ {
		if logA[i] != logB[i] {
			t.Fatalf("seed %d diverged at move %d of %d/%d:\n  run 1: %s\n  run 2: %s",
				cfg.Seed, i, len(logA), len(logB), logA[i], logB[i])
		}
	}
	if len(logA) != len(logB) {
		t.Fatalf("seed %d: run 1 made %d moves, run 2 made %d", cfg.Seed, len(logA), len(logB))
	}
	if a.Turns != b.Turns || a.Winner != b.Winner || a.State != b.State {
		t.Fatalf("same moves, different outcome: turns %d/%d, winner %d/%d, state %s/%s",
			a.Turns, b.Turns, a.Winner, b.Winner, a.State, b.State)
	}
	t.Logf("seed %d replayed exactly: %d moves, turn %d, state %s, winner %d, in %s and %s",
		cfg.Seed, len(logA), a.Turns, a.State, a.Winner, a.Elapsed.Round(time.Millisecond), b.Elapsed.Round(time.Millisecond))
}

// The concurrent schedule is the default and says so: the zero Config
// is not lockstep, and a run's report names its schedule, so two
// reports can be told apart on the one point that decides whether
// their seeds are comparable.
func TestConcurrentIsTheDefaultSchedule(t *testing.T) {
	if (botarena.Config{}).Lockstep {
		t.Fatal("the zero Config is lockstep; the arena's default schedule must stay one goroutine per seat")
	}
	for _, lockstep := range []bool{false, true} {
		cfg := botarena.Config{
			Seats:    []botarena.SeatSpec{{Tier: tiers.Heuristic}, {Tier: tiers.Heuristic}},
			Games:    1,
			Seed:     1,
			Wall:     time.Nanosecond, // a deal, not a game
			Lockstep: lockstep,
		}
		sum, err := botarena.Run(context.Background(), cfg, nil)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if sum.Config.Lockstep != lockstep {
			t.Errorf("Lockstep=%v reached the report as %v", lockstep, sum.Config.Lockstep)
		}
		want := "schedule " + botarena.Schedule(lockstep)
		if md := sum.Markdown(); !strings.Contains(md, want) {
			t.Errorf("Lockstep=%v: the report does not say %q:\n%s", lockstep, want, md)
		}
	}
}

// A lockstep game ends the way a concurrent one does: its own wall
// clock is a reported stall, a cancelled RUN is an abort that is
// dropped. stepTable and watchTable share the classification; this
// pins that the lockstep side reaches it.
func TestLockstepWallClockAndCancellation(t *testing.T) {
	seats := []botarena.SeatSpec{{Tier: tiers.Heuristic}, {Tier: tiers.Heuristic}}
	t.Run("wall clock", func(t *testing.T) {
		var got []botarena.GameResult
		_, err := botarena.Run(context.Background(), botarena.Config{
			Seats: seats, Games: 1, Seed: 303, Wall: time.Nanosecond, Lockstep: true,
		}, func(r botarena.GameResult) { got = append(got, r) })
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if len(got) != 1 || !got[0].Stalled || got[0].Aborted || !strings.Contains(got[0].StallDump, "wall clock") {
			t.Fatalf("want one stalled, not aborted, wall-clock game; got %+v", got)
		}
	})
	t.Run("cancelled run", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			time.Sleep(150 * time.Millisecond)
			cancel()
		}()
		sum, err := botarena.Run(ctx, botarena.Config{
			Seats: seats, Games: 50, Seed: 404, TurnBudget: 60, Wall: 5 * time.Minute, Lockstep: true,
		}, nil)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run returned %v, want a cancellation", err)
		}
		for _, g := range sum.Games {
			if g.Stalled {
				t.Errorf("game seed %d is reported as stalled; a cancelled game is not a stalled one", g.Seed)
			}
		}
	})
}
