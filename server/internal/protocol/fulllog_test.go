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
