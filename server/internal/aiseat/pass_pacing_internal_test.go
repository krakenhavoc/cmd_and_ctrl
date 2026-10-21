package aiseat

import (
	"context"
	"io"
	"log/slog"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// pass_pacing_internal_test.go pins #2898: a seat whose only legal move
// is to pass skips MinThink, while the ADR 0119 §2 stack hold on another
// seat's item still applies and a window with a real choice is still
// paced.

// firstMove records the runner's first applied move: when, and how many
// moves it was chosen from.
type firstMove struct {
	mu    sync.Mutex
	at    time.Time
	moves []legal.Move
	idx   int
	done  bool
}

func (f *firstMove) Observe(ev DecisionEvent) {
	if !ev.Applied || ev.Index < 0 || ev.Index >= len(ev.Input.Moves) {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.done {
		f.done, f.at, f.moves, f.idx = true, time.Now(), ev.Input.Moves, ev.Index
	}
}

func (f *firstMove) wait(t *testing.T) (time.Time, []legal.Move, int) {
	t.Helper()
	deadline := time.Now().Add(holdBackstop)
	for time.Now().Before(deadline) {
		f.mu.Lock()
		at, moves, idx, ok := f.at, f.moves, f.idx, f.done
		f.mu.Unlock()
		if ok {
			return at, moves, idx
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("the bot never moved")
	return time.Time{}, nil, 0
}

func testLog() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// startRunner is Start for a test in this package, and the only way one
// starts a runner (TestNoTestStartsARunnerItDoesNotStop). It is
// startBot in runner_test.go for the internal tests, for the same
// reason (#4219): a runner a test merely cancels can still be inside a
// step, reading the process-wide catalog hooks, when the next test
// swaps one. The cleanup cancels it and waits for Done.
func startRunner(t *testing.T, ctx context.Context, room *ws.Room, seat uuid.UUID, policy Policy, cfg Config, bc Broadcaster, log *slog.Logger) *Runner {
	t.Helper()
	ctx, cancel := context.WithCancel(ctx)
	r := Start(ctx, room, seat, policy, cfg, bc, log)
	t.Cleanup(func() {
		cancel()
		select {
		case <-r.Done():
		case <-time.After(holdBackstop):
			t.Errorf("the bot runner for seat %s did not exit after its test ended (#4219)", seat)
		}
	})
	return r
}

// waitDone waits for the runner to exit; the caller has cancelled it.
func waitDone(t *testing.T, r *Runner) {
	t.Helper()
	select {
	case <-r.Done():
	case <-time.After(holdBackstop):
		t.Fatal("the runner did not exit after cancel")
	}
}

func TestOnlyPass(t *testing.T) {
	pass := legal.Move{Kind: legal.KindPass}
	land := legal.Move{Kind: legal.KindLand}
	cases := []struct {
		name  string
		moves []legal.Move
		want  bool
	}{
		{"nothing", nil, false},
		{"a lone pass", []legal.Move{pass}, true},
		{"a lone land", []legal.Move{land}, false},
		{"a pass beside a land", []legal.Move{land, pass}, false},
	}
	for _, c := range cases {
		if got := onlyPass(c.moves); got != c.want {
			t.Errorf("%s: onlyPass = %v, want %v", c.name, got, c.want)
		}
	}
}

// A seat at a MinThink of 3 s whose only move is a pass passes at once.
func TestRunnerPassesWithoutMinThinkWhenPassIsTheOnlyMove(t *testing.T) {
	room, active, _ := holdTable(t)
	rec := &firstMove{}
	cfg := Config{MinThink: 3 * time.Second, MaxThink: defaultMaxThink, Observer: rec}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := time.Now()
	r := startRunner(t, ctx, room, active, NewRandomPolicy(rand.NewPCG(2898, 1)), cfg, nil, testLog())
	at, moves, idx := rec.wait(t)
	if !onlyPass(moves) || moves[idx].Kind != legal.KindPass {
		t.Fatalf("fixture drifted: the bot's window is not pass-only: %+v", moves)
	}
	if took := at.Sub(started); took >= time.Second {
		t.Errorf("a pass-only window took %v with MinThink 3s; it should not be paced", took)
	}
	cancel()
	waitDone(t, r)
}

// The same seat with a real choice in front of it is still paced.
func TestRunnerStillPacesAWindowWithAChoice(t *testing.T) {
	room, active, _ := holdTable(t)
	g := room.Game
	// A land in hand at a main phase with an empty stack is a second
	// legal move beside the pass.
	for i := 0; i < 40 && g.Turn.Step != game.StepPrecombatMain; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.PriorityHolder == game.NoPriority {
		t.Skip("no priority at main phase in this fixture")
	}
	if len(legal.EnumerateFor(g, active)) < 2 {
		t.Fatalf("fixture drifted: want a window with a choice, got %+v", legal.EnumerateFor(g, active))
	}
	rec := &firstMove{}
	cfg := Config{MinThink: 400 * time.Millisecond, MaxThink: defaultMaxThink, Observer: rec}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := time.Now()
	r := startRunner(t, ctx, room, active, NewRandomPolicy(rand.NewPCG(2898, 2)), cfg, nil, testLog())
	at, moves, _ := rec.wait(t)
	if onlyPass(moves) {
		t.Fatalf("fixture drifted: the window is pass-only")
	}
	if took := at.Sub(started); took < 350*time.Millisecond {
		t.Errorf("a window with %d moves took %v; MinThink 400ms was not honoured", len(moves), took)
	}
	cancel()
	waitDone(t, r)
}

// A pass-only window over another seat's item still waits out the stack
// hold: skipping MinThink must not skip ADR 0119 §2.
func TestPassOnlyWindowStillHoldsForAnotherSeatsStackItem(t *testing.T) {
	room, active, other := holdTable(t)
	g := room.Game
	announce(t, g, active)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if holder := g.Seats[g.Turn.PriorityHolder].ID; holder != other {
		t.Fatalf("priority is with %s, want %s", holder, other)
	}
	normal := game.BotPaceNormal
	if err := g.UpdateSettings(uuid.Nil, game.SettingsPatch{BotPace: &normal}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	rec := &firstMove{}
	cfg := Config{MinThink: 0, MaxThink: defaultMaxThink, FollowTablePace: true, Observer: rec}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := time.Now()
	r := startRunner(t, ctx, room, other, NewRandomPolicy(rand.NewPCG(2898, 3)), cfg, nil, testLog())
	at, moves, idx := rec.wait(t)
	if !onlyPass(moves) || moves[idx].Kind != legal.KindPass {
		t.Fatalf("fixture drifted: the bot's window is not pass-only: %+v", moves)
	}
	if held := at.Sub(started); held < 2*time.Second {
		t.Errorf("pass-only over an opponent's item landed after %v; the 2 s stack hold was skipped", held)
	}
	cancel()
	waitDone(t, r)
}
