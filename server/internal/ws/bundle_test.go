package ws

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func newBundleRoom(t *testing.T) (*Room, string) {
	t.Helper()
	dir := t.TempDir()
	return NewRoom(seedTestGame(t), slog.New(slog.NewTextHandler(io.Discard, nil)), dir), dir
}

// lifeIn reads a seat's life total under the game's read lock.
//
// PlayerByIDForEffect, not PlayerByID: ReadSnapshot is already holding
// g.mu in read mode, and sync.RWMutex is not reentrant — a second
// RLock from this goroutine blocks as soon as a writer is queued
// between the two, because Go blocks a new reader behind a waiting
// Lock so writers cannot starve. #877; the same recursion was #848's
// aiseat flake. Anything called inside a snapshot body has to be a
// *ForEffect accessor or a plain field read, and
// game/snapshot_lock_guard_test.go now fails on anything else.
func lifeIn(g *game.Game, seat uuid.UUID) int {
	var life int
	g.ReadSnapshot(func() {
		if p := g.PlayerByIDForEffect(seat); p != nil {
			life = p.Life
		}
	})
	return life
}

// TestApplyBundleCommitsAsOneEntry is the contract the S31
// improvisation bundle rests on: several mutations, one seq, one
// replay line, one undo.
func TestApplyBundleCommitsAsOneEntry(t *testing.T) {
	room, dir := newBundleRoom(t)
	g := room.Game
	a, b := g.Seats[0].ID, g.Seats[1].ID
	lifeA, lifeB := lifeIn(g, a), lifeIn(g, b)
	seqBefore := room.Seq()

	_, seq, err := room.ApplyBundle(Bundle{
		Caller: uuid.Nil,
		Steps: []func() error{
			func() error { _, err := g.ChangePlayerLife(a, -3); return err },
			func() error { _, err := g.ChangePlayerLife(b, -2); return err },
		},
		FreeUndo:   true,
		Annotation: &protocol.ReplayAnnotation{Tag: protocol.ReplayTagBotImprovisation, Card: "Grim Tutor"},
	})
	if err != nil {
		t.Fatalf("ApplyBundle: %v", err)
	}
	if seq != seqBefore+1 {
		t.Errorf("bundle allocated seq %d, want one bump from %d", seq, seqBefore)
	}
	if lifeIn(g, a) != lifeA-3 || lifeIn(g, b) != lifeB-2 {
		t.Fatalf("bundle did not apply")
	}

	// Exactly one replay line carries the tag.
	raw, err := os.ReadFile(filepath.Join(dir, "replays", g.ID.String()+".jsonl"))
	if err != nil {
		t.Fatalf("read replay: %v", err)
	}
	if n := strings.Count(string(raw), protocol.ReplayTagBotImprovisation); n != 1 {
		t.Errorf("%d replay lines carry the tag, want 1", n)
	}
	var last protocol.SnapshotPayload
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("replay line: %v", err)
	}
	if last.Annotation == nil || last.Annotation.Card != "Grim Tutor" {
		t.Errorf("annotation missing from the line it belongs to: %+v", last.Annotation)
	}

	// One entry: the single undo puts both mutations back.
	if _, _, err := room.Undo(a); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	if lifeIn(g, a) != lifeA || lifeIn(g, b) != lifeB {
		t.Errorf("undo reverted only part of the bundle")
	}
	if _, _, err := room.Undo(a); !errors.Is(err, ErrNothingToUndo) {
		t.Errorf("second undo = %v, want ErrNothingToUndo", err)
	}
}

// TestApplyBundleRollsBackOnFailure: the mutations before the failing
// step must not survive. Apply cannot promise this — a failing fn
// leaves whatever it already did — which is the whole reason
// ApplyBundle exists.
func TestApplyBundleRollsBackOnFailure(t *testing.T) {
	room, dir := newBundleRoom(t)
	g := room.Game
	a := g.Seats[0].ID
	lifeA := lifeIn(g, a)
	seqBefore := room.Seq()
	boom := errors.New("boom")

	_, _, err := room.ApplyBundle(Bundle{
		Steps: []func() error{
			func() error { _, err := g.ChangePlayerLife(a, -7); return err },
			func() error { return boom },
		},
	})
	if !errors.Is(err, boom) {
		t.Fatalf("ApplyBundle err = %v, want the step's error", err)
	}
	if got := lifeIn(g, a); got != lifeA {
		t.Errorf("life = %d after a failed bundle, want %d — the first step survived", got, lifeA)
	}
	if room.Seq() != seqBefore {
		t.Errorf("a failed bundle advanced seq")
	}
	if _, _, err := room.Undo(uuid.Nil); !errors.Is(err, ErrNothingToUndo) {
		t.Errorf("a failed bundle left an undo entry: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "replays", g.ID.String()+".jsonl")); !os.IsNotExist(err) {
		// Nothing committed, so nothing should have been written.
		t.Errorf("a failed bundle wrote a replay line")
	}

	if _, _, err := room.ApplyBundle(Bundle{}); !errors.Is(err, ErrEmptyBundle) {
		t.Errorf("empty bundle = %v, want ErrEmptyBundle", err)
	}
}

// TestFreeUndoSkipsTheBudget pins the decision in ADR 0033 §8: a
// human who cleans up after a bot does not pay for it out of their
// own per-turn take-backs, while an ordinary undo still does.
func TestFreeUndoSkipsTheBudget(t *testing.T) {
	for name, free := range map[string]bool{"free": true, "charged": false} {
		t.Run(name, func(t *testing.T) {
			room, _ := newBundleRoom(t)
			g := room.Game
			a := g.Seats[0].ID
			before := g.PlayerByID(a).UndosRemaining
			if before <= 0 {
				t.Fatalf("fixture: %d undos", before)
			}

			if _, _, err := room.ApplyBundle(Bundle{
				Caller:   uuid.Nil,
				Steps:    []func() error{func() error { _, err := g.ChangePlayerLife(a, -1); return err }},
				FreeUndo: free,
			}); err != nil {
				t.Fatalf("ApplyBundle: %v", err)
			}
			if _, _, err := room.Undo(a); err != nil {
				t.Fatalf("Undo: %v", err)
			}
			after := g.PlayerByID(a).UndosRemaining
			want := before
			if !free {
				want = before - 1
			}
			if after != want {
				t.Errorf("UndosRemaining = %d, want %d", after, want)
			}
		})
	}
}

// TestFreeUndoWorksAtZeroBudget: a player who has spent every undo
// this turn can still put a bot's improvisation back. Being out of
// take-backs must not mean being stuck with someone else's mistake.
func TestFreeUndoWorksAtZeroBudget(t *testing.T) {
	room, _ := newBundleRoom(t)
	g := room.Game
	a := g.Seats[0].ID
	for g.PlayerByID(a).UndosRemaining > 0 {
		if err := g.SpendUndo(a); err != nil {
			t.Fatalf("SpendUndo: %v", err)
		}
	}

	if _, _, err := room.ApplyBundle(Bundle{
		Caller:   uuid.Nil,
		Steps:    []func() error{func() error { _, err := g.ChangePlayerLife(a, -4); return err }},
		FreeUndo: true,
	}); err != nil {
		t.Fatalf("ApplyBundle: %v", err)
	}
	if _, _, err := room.Undo(a); err != nil {
		t.Fatalf("Undo at zero budget: %v, want success", err)
	}

	// The ordinary rule is unchanged: a normal nil-caller entry at
	// zero budget still refuses.
	if _, _, err := room.Apply(uuid.Nil, func() error { _, err := g.ChangePlayerLife(a, -4); return err }); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if _, _, err := room.Undo(a); !errors.Is(err, game.ErrNoUndosRemaining) {
		t.Errorf("Undo = %v, want ErrNoUndosRemaining", err)
	}
}

// TestAnnotationDoesNotLeakOntoTheNextCommit: the pending annotation
// is consumed by the capture it was set for and by nothing else.
func TestAnnotationDoesNotLeakOntoTheNextCommit(t *testing.T) {
	room, dir := newBundleRoom(t)
	g := room.Game
	a := g.Seats[0].ID

	if _, _, err := room.ApplyBundle(Bundle{
		Steps:      []func() error{func() error { _, err := g.ChangePlayerLife(a, -1); return err }},
		Annotation: &protocol.ReplayAnnotation{Tag: protocol.ReplayTagBotImprovisation},
	}); err != nil {
		t.Fatalf("ApplyBundle: %v", err)
	}
	for range 3 {
		if _, _, err := room.Apply(uuid.Nil, func() error { _, err := g.ChangePlayerLife(a, -1); return err }); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	}
	raw, err := os.ReadFile(filepath.Join(dir, "replays", g.ID.String()+".jsonl"))
	if err != nil {
		t.Fatalf("read replay: %v", err)
	}
	if n := strings.Count(string(raw), protocol.ReplayTagBotImprovisation); n != 1 {
		t.Errorf("%d tagged replay lines, want exactly 1", n)
	}
}
