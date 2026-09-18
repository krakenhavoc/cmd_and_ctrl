package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// loop_shortcut_test.go — #804, the enumerator's half of the CR 726
// shortcut.
//
// The answer is a number, so the list here is a handful of numbers
// worth offering a policy. It is also where a bot-only table is made
// to terminate: a policy that can rank "100 more" top can rank it top
// every time, so the SECOND ask of a turn offers one answer — stop —
// and there is nothing else for any policy to pick. See docs/bot.md.

// queueLoopShortcut puts a CR 726 prompt on `chooser` by hand. The
// engine only queues one from inside the breaker; the enumerator does
// not care where it came from.
func queueLoopShortcut(g *game.Game, chooser uuid.UUID, repeat bool) uuid.UUID {
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChoiceForEffect(game.PendingChoice{
			Kind:               game.PendingChoiceLoopShortcut,
			Chooser:            chooser,
			Count:              1,
			Source:             uuid.New(),
			Reason:             "Mirror Engine — create a Spark",
			LoopShortcutKey:    "loop-key",
			LoopShortcutCount:  25,
			LoopShortcutRepeat: repeat,
		})
	})
	return id
}

func TestLoopShortcutOffersNumbersToTheLoopsController(t *testing.T) {
	g := newTable(t)
	me, them := g.Seats[0], g.Seats[1]
	queueLoopShortcut(g, me.ID, false)

	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) != 3 {
		t.Fatalf("enumerated %d answers, want 3: %v", len(moves), labels(moves))
	}
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("the seat owing the shortcut was offered %q as well", m.Label)
		}
	}
	// 10 leads, because a bot takes the first offer (docs/bot.md).
	if !strings.HasSuffix(moves[0].Label, ": resolve 10 more times") {
		t.Errorf("first answer = %q, want the default K of 10", moves[0].Label)
	}
	if !strings.HasSuffix(moves[len(moves)-1].Label, ": stop here") {
		t.Errorf("last answer = %q, want stop", moves[len(moves)-1].Label)
	}
	// Stop is the way out the engine can never refuse, and the only
	// one marked so.
	stops := 0
	for _, m := range moves {
		if m.AlwaysLegal {
			stops++
			if !strings.HasSuffix(m.Label, ": stop here") {
				t.Errorf("%q is marked always legal; only stop commits the seat to nothing", m.Label)
			}
		}
	}
	if stops != 1 {
		t.Errorf("%d always-legal answers, want exactly one", stops)
	}
	// A blocking prompt: nobody else gets anything until it is
	// answered.
	if other := legal.EnumerateFor(g, them.ID); len(other) != 0 {
		t.Errorf("another seat was offered %v while the CR 726 prompt was open", labels(other))
	}
	dispatchAll(t, g, me.ID, moves)
}

// TestLoopShortcutSecondAskOffersOnlyStop is the termination
// guarantee, and the reason it lives in the enumerator rather than in
// a policy: with one answer on the list, every policy at the table
// stops, including a random one.
func TestLoopShortcutSecondAskOffersOnlyStop(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	queueLoopShortcut(g, me.ID, true)

	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) != 1 {
		t.Fatalf("enumerated %d answers on the turn's second ask, want just stop: %v", len(moves), labels(moves))
	}
	if !strings.HasSuffix(moves[0].Label, ": stop here") {
		t.Errorf("answer = %q, want stop", moves[0].Label)
	}
	if !moves[0].AlwaysLegal {
		t.Error("stop is the answer the engine can never refuse")
	}
	dispatchAll(t, g, me.ID, moves)
}
