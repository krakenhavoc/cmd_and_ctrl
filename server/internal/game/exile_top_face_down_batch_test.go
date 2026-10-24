package game

import (
	"testing"

	"github.com/google/uuid"
)

// exile_top_face_down_batch_test.go — #4245 (CR 406.3, CR 616.1):
// ExileTopFaceDownForEffect exiles N cards, and a prompt one of them opens
// (a CR 616.1 ordering of two replacements) must not stop the rest of the
// batch. The cards are planned up front and the rest resume from the answer.

// exileOrderingPair registers two mandatory replacements over `card`'s move
// from a library to exile. Two apply, so CR 616.1 asks the affected player
// which goes first, and the move waits for the answer. Caller must hold g.mu.
func exileOrderingPair(g *Game, card uuid.UUID) {
	for _, label := range []string{"first order", "second order"} {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.CardID == card && ev.NewZone == ZoneExile
			},
			Replace: func(*ReplacementEvent, *Game, *Card) error { return nil },
			Label:   label,
		})
	}
}

// exileTopTable is a table whose first seat has a library of `libraryCards`
// cards, the card at index `promptAt` from the top carrying an ordering prompt.
func exileTopTable(t *testing.T, libraryCards, promptAt int) (*Game, *Player, []uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	me := g.Seats[0]
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		ids = stockLibrary(me, libraryCards)
		exileOrderingPair(g, ids[promptAt])
	})
	return g, me, ids
}

func exileTopFaceDown(t *testing.T, g *Game, p *Player, n int) []uuid.UUID {
	t.Helper()
	var got []uuid.UUID
	g.WithWriteLock(func() {
		var err error
		got, err = g.ExileTopFaceDownForEffect(p.ID, n)
		if err != nil {
			t.Fatalf("ExileTopFaceDownForEffect: %v", err)
		}
	})
	return got
}

func answerExileOrdering(t *testing.T, g *Game) {
	t.Helper()
	c := onlyPrompt(t, g)
	if c.Kind != PendingChoiceReplacementOrder {
		t.Fatalf("prompt kind = %q, want %q", c.Kind, PendingChoiceReplacementOrder)
	}
	if err := g.ResolveReplacementOrder(c.ID, c.Chooser, orderByLabel(t, g, c, "first order", "second order")); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}

// wantExiledFaceDown asserts every id is in exile, face down, and known to
// nobody (CR 406.3a, ADR 0069).
func wantExiledFaceDown(t *testing.T, g *Game, ids []uuid.UUID) {
	t.Helper()
	for _, id := range ids {
		c, zone, ok := cardAnywhere(g, id)
		if !ok || zone != ZoneExile {
			t.Fatalf("card %s is in %q, want exile", id, zone)
		}
		if !c.FaceDown || c.FaceDownKind != FaceDownExiled {
			t.Errorf("card %s exiled face down = %v kind %v, want face down", id, c.FaceDown, c.FaceDownKind)
		}
		for _, s := range g.Seats {
			if c.IsKnownTo(s.ID) {
				t.Errorf("a seat can read the face-down card %s", id)
			}
		}
	}
}

// The headline: three cards, a prompt on the first, and all three end in
// exile face down. A prompt on the second or third card is the same.
func TestExileTopFaceDownFinishesTheBatchAfterAPromptOnAnyCard(t *testing.T) {
	for promptAt := 0; promptAt < 3; promptAt++ {
		g, me, ids := exileTopTable(t, 8, promptAt)
		got := exileTopFaceDown(t, g, me, 3)

		if len(g.PendingChoices) != 1 {
			t.Fatalf("prompt on card %d: %d prompts open, want 1", promptAt, len(g.PendingChoices))
		}
		// The cards before the prompt left; the ones from it on wait.
		if len(got) != promptAt {
			t.Errorf("prompt on card %d: the call reported %d landed, want %d", promptAt, len(got), promptAt)
		}
		if me.Library.Size() != 8-promptAt {
			t.Fatalf("prompt on card %d: library %d before the answer, want %d", promptAt, me.Library.Size(), 8-promptAt)
		}

		answerExileOrdering(t, g)

		wantExiledFaceDown(t, g, ids[:3])
		if me.Library.Size() != 5 {
			t.Errorf("prompt on card %d: library %d after the batch, want 5", promptAt, me.Library.Size())
		}
		if len(g.PendingChoices) != 0 {
			t.Errorf("prompt on card %d: %d prompts left open", promptAt, len(g.PendingChoices))
		}
		for _, id := range ids[3:] {
			if _, zone, _ := cardAnywhere(g, id); zone != ZoneLibrary {
				t.Errorf("prompt on card %d: a card below the batch is in %q", promptAt, zone)
			}
		}
	}
}

// Without a prompt nothing changes: every card, in order, and the list says so.
func TestExileTopFaceDownWithoutAPromptReportsEveryCardInOrder(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var ids []uuid.UUID
	g.WithWriteLock(func() { ids = stockLibrary(me, 6) })

	got := exileTopFaceDown(t, g, me, 3)

	if len(got) != 3 || got[0] != ids[0] || got[1] != ids[1] || got[2] != ids[2] {
		t.Fatalf("exiled %v, want the top three %v in order", got, ids[:3])
	}
	wantExiledFaceDown(t, g, got)
	if me.Library.Size() != 3 {
		t.Errorf("library %d, want 3", me.Library.Size())
	}
}

// A library shorter than N loses what it has and no draw is attempted; N of
// zero does nothing.
func TestExileTopFaceDownOnAShortLibraryTakesWhatIsThere(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { stockLibrary(me, 2) })

	if got := exileTopFaceDown(t, g, me, 0); len(got) != 0 || me.Library.Size() != 2 {
		t.Fatalf("N=0 exiled %d cards and left %d", len(got), me.Library.Size())
	}
	got := exileTopFaceDown(t, g, me, 5)
	if len(got) != 2 || me.Library.Size() != 0 {
		t.Fatalf("exiled %d cards leaving %d, want 2 and an empty library", len(got), me.Library.Size())
	}
	if len(exileTopFaceDown(t, g, me, 3)) != 0 {
		t.Error("an empty library exiled something")
	}
	if me.AttemptedEmptyDraw {
		t.Error("exiling off an empty library set the empty-draw flag")
	}
}

// A card that leaves the library while the prompt is open is skipped rather
// than exiled again or from the wrong zone.
func TestExileTopFaceDownSkipsACardThatLeftWhileTheBatchWaited(t *testing.T) {
	g, me, ids := exileTopTable(t, 5, 0)
	exileTopFaceDown(t, g, me, 3)
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(ids[1]); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
	})

	answerExileOrdering(t, g)

	wantExiledFaceDown(t, g, []uuid.UUID{ids[0], ids[2]})
	if _, zone, _ := cardAnywhere(g, ids[1]); zone != ZoneHand {
		t.Errorf("the card that left is in %q, want the hand it was taken to", zone)
	}
	if me.Library.Size() != 2 {
		t.Errorf("library %d, want 2", me.Library.Size())
	}
}

// An undo into the open prompt replays: answering again exiles the same three
// cards, and the first run's answer does not leak into the replay through the
// shared continuation.
func TestExileTopFaceDownBatchReplaysAfterAnUndoIntoThePrompt(t *testing.T) {
	g, me, ids := exileTopTable(t, 8, 0)
	exileTopFaceDown(t, g, me, 3)
	promptOpen := g.Clone()

	answerExileOrdering(t, g)
	wantExiledFaceDown(t, g, ids[:3])

	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	// An undo restores the game's fields in place, so the seat is read again.
	if n := g.Seats[0].Library.Size(); n != 8 {
		t.Fatalf("the rewind left %d cards on the library, want all 8 back", n)
	}
	answerExileOrdering(t, g)
	wantExiledFaceDown(t, g, ids[:3])
	if n := g.Seats[0].Library.Size(); n != 5 {
		t.Errorf("replayed library %d, want 5", n)
	}
}

// The table is the restore point it was: an ordering prompt holds a
// continuation (the snapshot census counts it), so it is not one while the
// prompt is open and the fix must not change that; once the batch has
// finished, it is, and a restored table has the three cards face down.
func TestExileTopFaceDownBatchLeavesARestorePointOnceItIsDone(t *testing.T) {
	g, me, ids := exileTopTable(t, 8, 0)
	exileTopFaceDown(t, g, me, 3)
	if throughJSON(t, g.CaptureSnapshot()).Restorable() {
		t.Fatal("a table with an ordering prompt open is a restore point: the census no longer counts it")
	}

	answerExileOrdering(t, g)

	snap := throughJSON(t, g.CaptureSnapshot())
	if !snap.Restorable() {
		t.Fatalf("the finished batch is not a restore point: %+v", snap.Continuations)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	wantExiledFaceDown(t, restored, ids[:3])
	if restored.Seats[0].Library.Size() != 5 {
		t.Errorf("restored library %d, want 5", restored.Seats[0].Library.Size())
	}
}
