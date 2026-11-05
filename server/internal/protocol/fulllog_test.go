package protocol

import (
	"reflect"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestFullLogOfHasEveryEntry: the game statistics read the whole log
// (ADR 0165 §2), so FullLogOf keeps what the public ring drops, in the
// same shape and order.
func TestFullLogOfHasEveryEntry(t *testing.T) {
	g := buildActiveGame(t)
	p := g.Seats[0]
	const emitted = PublicLogMax * 3
	g.WithWriteLock(func() {
		for i := range emitted {
			g.EmitEvent(game.Event{
				Kind: game.EventStepBegan, Actor: p.ID, Amount: i + 1, Label: string(game.StepUpkeep),
			})
		}
	})
	view, full := FullLogOf(g)
	var steps []LogEvent
	for _, e := range full {
		if e.Kind == LogStep {
			steps = append(steps, e)
		}
	}
	if len(steps) != emitted {
		t.Fatalf("FullLogOf has %d step entries, want all %d", len(steps), emitted)
	}
	if steps[0].Turn != 1 || steps[len(steps)-1].Turn != emitted {
		t.Errorf("turns run %d..%d, want 1..%d", steps[0].Turn, steps[len(steps)-1].Turn, emitted)
	}
	for i := 1; i < len(full); i++ {
		if full[i].Seq <= full[i-1].Seq {
			t.Fatalf("not in seq order at %d", i)
		}
	}
	// Its tail is the public log: the same fold, the same entries.
	public := ViewOfGame(g).Log
	if !reflect.DeepEqual(full[len(full)-len(public):], public) {
		t.Error("the full log's tail differs from the public log")
	}
	if view.ID != g.ID.String() || len(view.Seats) != len(g.Seats) {
		t.Errorf("the view is not the game's: %s, %d seats", view.ID, len(view.Seats))
	}
}

// TestFullLogOfStatsLog: the stats-only log (ADR 0217 §1) lines up the
// loop breaker, which the public log leaves out, and marks a loop
// replay's line; without the option, and on the wire, nothing changes.
func TestFullLogOfStatsLog(t *testing.T) {
	g := buildActiveGame(t)
	p := g.Seats[0]
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventLoopSuspected, Actor: p.ID, Label: "Untap target permanent", Amount: 40})
		g.EmitEvent(game.Event{Kind: game.EventLoopReplayed, Actor: p.ID, Label: "repeats Basalt Monolith ×5", Amount: 5})
	})
	loops := func(log []LogEvent) []LogEvent {
		var out []LogEvent
		for _, e := range log {
			if e.Kind == LogLoop {
				out = append(out, e)
			}
		}
		return out
	}
	_, plain := FullLogOf(g)
	if got := loops(plain); len(got) != 1 || got[0].Call != "" {
		t.Errorf("the plain full log's loop lines = %+v, want the replay alone, with no call", got)
	}
	if got := loops(ViewOfGame(g).Log); len(got) != 1 || got[0].Call != "" {
		t.Errorf("the public log's loop lines = %+v, want the replay alone, with no call", got)
	}
	_, stats := FullLogOf(g, StatsLog)
	got := loops(stats)
	if len(got) != 2 {
		t.Fatalf("the stats log's loop lines = %+v, want the breaker and the replay", got)
	}
	if got[0].Call != LoopCallSuspected || got[0].Seat != 0 || got[0].Amount != 40 {
		t.Errorf("the breaker's line = %+v", got[0])
	}
	if got[1].Call != LoopCallReplayed || got[1].Amount != 5 {
		t.Errorf("the replay's line = %+v", got[1])
	}
}
