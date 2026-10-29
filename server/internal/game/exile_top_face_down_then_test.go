package game

import (
	"testing"

	"github.com/google/uuid"
)

// #4259: ExileTopFaceDownThenForEffect runs its continuation once, with every
// card of the batch top first, after a prompt on any card, and again on a
// replay after an undo. The helpers are exile_top_face_down_batch_test.go's.
func TestExileTopFaceDownThenRunsOnceWithTheWholeBatchAfterAPrompt(t *testing.T) {
	for promptAt := 0; promptAt < 3; promptAt++ {
		g, me, ids := exileTopTable(t, 8, promptAt)
		resetProbes()
		g.WithWriteLock(func() {
			err := g.ExileTopFaceDownKeyedForEffect(me.ID, 3, probeTail())
			if err != nil {
				t.Fatalf("ExileTopFaceDownKeyedForEffect: %v", err)
			}
		})
		calls := func() int { return len(testProbeHeard) }
		seenNow := func() []uuid.UUID {
			if len(testProbeHeard) == 0 {
				return nil
			}
			return testProbeHeard[len(testProbeHeard)-1].IDs
		}
		if calls() != 0 {
			t.Fatalf("prompt on card %d: then ran before the answer", promptAt)
		}
		promptOpen := g.Clone()

		answerExileOrdering(t, g)

		if seen := seenNow(); calls() != 1 || len(seen) != 3 || seen[0] != ids[0] || seen[1] != ids[1] || seen[2] != ids[2] {
			t.Fatalf("prompt on card %d: then ran %d times with %v, want once with %v", promptAt, calls(), seen, ids[:3])
		}

		// An undo into the open prompt, answered again: the continuation runs
		// again with the same three cards and no entry left over from the run
		// before.
		g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
		answerExileOrdering(t, g)
		if calls() != 2 || len(seenNow()) != 3 {
			t.Fatalf("prompt on card %d: replay ran then %d times with %d cards, want 2 and 3", promptAt, calls(), len(seenNow()))
		}
		wantExiledFaceDown(t, g, ids[:3])
	}
}

// With nothing to exile (n of zero) `then` still runs, with an empty list; a
// nil `then` is a plain exile; an unknown player is an error and `then` does
// not run.
func TestExileTopFaceDownThenEdgeCases(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { stockLibrary(me, 2) })
	resetProbes()
	g.WithWriteLock(func() {
		if err := g.ExileTopFaceDownKeyedForEffect(me.ID, 0, probeTail()); err != nil {
			t.Fatalf("n=0: %v", err)
		}
		if err := g.ExileTopFaceDownKeyedForEffect(uuid.New(), 1, probeTail()); err == nil {
			t.Error("an unknown player did not error")
		}
		if err := g.ExileTopFaceDownKeyedForEffect(me.ID, 1, KeyedTail{}); err != nil {
			t.Fatalf("zero tail: %v", err)
		}
	})
	if calls := len(testProbeHeard); calls != 1 {
		t.Errorf("then ran %d times, want only the n=0 call", calls)
	} else if len(testProbeHeard[0].IDs) != 0 {
		t.Errorf("then saw %d cards, want none", len(testProbeHeard[0].IDs))
	}
	if me.Library.Size() != 1 {
		t.Errorf("library %d, want 1 after the nil-then exile", me.Library.Size())
	}
}
