package ws

import (
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_answer_test.go — ADR 0127 §4 and §6 at the room: a covered
// prompt a commit raises is answered as a commit of its own, stamped
// with the chooser and free to undo, and an undone answer is asked by
// hand.

// newAutoAnswerRoom is a two-seat room past the mulligan, with no hub
// and no socket — the browser-closed case is the only case here.
func newAutoAnswerRoom(t *testing.T) (*Room, *game.Game) {
	t.Helper()
	g := seedTestGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	return NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), t.TempDir()), g
}

// queueTax queues a pay-unless to payer under key, the shape Rhystic
// Study's tax has, and reports when its decline branch runs.
func queueTax(g *game.Game, payer uuid.UUID, key string, declined *int) func() error {
	return func() error {
		g.WithWriteLock(func() {
			_ = g.QueuePayUnlessForEffect(payer, uuid.Nil, "{1}", "Test — pay {1}?", func(*game.Game) error {
				*declined++
				return nil
			})
			g.PendingChoices[len(g.PendingChoices)-1].AutoAnswerKey = key
		})
		return nil
	}
}

func TestRoomAnswersACoveredPromptAsItsOwnCommit(t *testing.T) {
	room, g := newAutoAnswerRoom(t)
	payer, other := g.Seats[0], g.Seats[1]
	if err := g.SetAutoAnswers(payer.ID, map[string]game.AutoAnswer{"tax": game.AutoAnswerNever}); err != nil {
		t.Fatal(err)
	}
	var declined int
	seqBefore := room.Seq()
	// The commit that raises the prompt is the OTHER seat's (a bot's
	// commit takes the same path).
	view, seq, err := room.Apply(other.ID, queueTax(g, payer.ID, "tax", &declined))
	if err != nil {
		t.Fatal(err)
	}
	if declined != 1 || len(g.PendingChoices) != 0 {
		t.Fatalf("prompt not answered: declined %d, open %d", declined, len(g.PendingChoices))
	}
	if len(view.PendingChoices) != 0 {
		t.Error("the returned frame still shows the prompt")
	}
	if seq != seqBefore+2 {
		t.Errorf("seq advanced %d, want 2 (the commit, then the answer)", seq-seqBefore)
	}
	if n := len(room.undoStack); n != 2 {
		t.Fatalf("undo entries = %d, want 2", n)
	}
	top := room.undoStack[1]
	if top.caller != payer.ID || !top.freeUndo || top.autoAnswered == uuid.Nil {
		t.Errorf("top entry = caller %v free %v answered %v; want the chooser's, free, naming the prompt",
			top.caller, top.freeUndo, top.autoAnswered)
	}
	if room.undoStack[0].caller != other.ID {
		t.Error("the raising commit's entry is not the raiser's")
	}
}

func TestRoomUndoOfAnAutomaticAnswerAsksByHand(t *testing.T) {
	room, g := newAutoAnswerRoom(t)
	payer, other := g.Seats[0], g.Seats[1]
	if err := g.SetAutoAnswers(payer.ID, map[string]game.AutoAnswer{"tax": game.AutoAnswerNever}); err != nil {
		t.Fatal(err)
	}
	var declined int
	if _, _, err := room.Apply(other.ID, queueTax(g, payer.ID, "tax", &declined)); err != nil {
		t.Fatal(err)
	}
	// Owner decision 8: free. A seat with no undos left can still take
	// it back.
	payer.UndosRemaining = 0
	if _, _, err := room.Undo(payer.ID); err != nil {
		t.Fatalf("chooser's undo of the automatic answer: %v", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("open prompts after undo = %d, want the tax back", len(g.PendingChoices))
	}
	tax := g.PendingChoices[0]
	if tax.AskedByHand != game.AskedByHandUndone {
		t.Errorf("AskedByHand = %q, want %q", tax.AskedByHand, game.AskedByHandUndone)
	}
	// The next commit does not answer it again.
	if _, _, err := room.Apply(other.ID, func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingChoices) != 1 {
		t.Error("the next commit answered an undone prompt again")
	}
}

func TestRoomAnswersAfterALobbyCommitToo(t *testing.T) {
	room, g := newAutoAnswerRoom(t)
	payer := g.Seats[0]
	if err := g.SetAutoAnswers(payer.ID, map[string]game.AutoAnswer{"tax": game.AutoAnswerNever}); err != nil {
		t.Fatal(err)
	}
	var declined int
	if _, _, err := room.ApplyExternal(queueTax(g, payer.ID, "tax", &declined)); err != nil {
		t.Fatal(err)
	}
	if declined != 1 {
		t.Fatal("a prompt raised by a lobby commit was not answered")
	}
	if len(room.undoStack) != 1 || room.undoStack[0].caller != payer.ID {
		t.Errorf("undo stack = %d entries, want only the answer's, stamped with the chooser", len(room.undoStack))
	}
}

func TestRoomLeavesAPromptWithNoRule(t *testing.T) {
	room, g := newAutoAnswerRoom(t)
	var declined int
	if _, _, err := room.Apply(g.Seats[1].ID, queueTax(g, g.Seats[0].ID, "tax", &declined)); err != nil {
		t.Fatal(err)
	}
	if declined != 0 || len(g.PendingChoices) != 1 || len(room.undoStack) != 1 {
		t.Error("a prompt with no rule was answered")
	}
}
