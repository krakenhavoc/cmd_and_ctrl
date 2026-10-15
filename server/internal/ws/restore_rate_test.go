package ws

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// restore_rate_test.go holds #3823: restore points written at most once
// per interval, flushed at shutdown, and assembled from each event's
// cached encoding.

// rateLimitedManager is restoreTestManager with a restore-point
// interval, as the server sets it.
func rateLimitedManager(t *testing.T, dir string, every time.Duration) *RoomManager {
	t.Helper()
	m := restoreTestManager(t, dir)
	m.SetRestorePointInterval(every)
	return m
}

// restoreFileSeq reads the seq of the restore point on disk, or 0 when
// there is none.
func restoreFileSeq(t *testing.T, dir string, id uuid.UUID) uint64 {
	t.Helper()
	raw, err := os.ReadFile(restorePointPath(dir, id))
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatal(err)
	}
	var f restorePointFile
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	return f.Seq
}

func restoreWrites(r *Room) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.restoreWrites
}

// setLife is a restorable commit.
func setLife(t *testing.T, r *Room, seat uuid.UUID, life int) {
	t.Helper()
	g := r.Game
	if _, _, err := r.Apply(seat, func() error {
		g.WithWriteLock(func() {
			for _, p := range g.Seats {
				if p.ID == seat {
					p.Life = life
				}
			}
		})
		return nil
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
}

// openContinuation is a commit that leaves the game holding a Go
// continuation (an ability whose effect is a closure), so it is not a
// restore point.
func openContinuation(t *testing.T, r *Room, seat uuid.UUID) {
	t.Helper()
	g := r.Game
	if _, _, err := r.Apply(seat, func() error {
		g.WithWriteLock(func() {
			g.Seats[0].Life = 1
			id := uuid.New()
			g.StackMeta = map[uuid.UUID]*game.StackItem{id: {
				ID: id, Kind: game.StackItemActivated, Controller: seat, Label: "some ability",
				Effect: func(*game.Game, *game.StackItem) error { return nil },
			}}
		})
		return nil
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
}

// TestRestorePointWritesAreRateLimited: several commits inside the
// interval produce one write (the first, at once), and the last state
// is on disk by the time the interval runs out, with one more write.
func TestRestorePointWritesAreRateLimited(t *testing.T) {
	const every = 300 * time.Millisecond
	dir := t.TempDir()
	mgr := rateLimitedManager(t, dir, every)
	g := newPersistGame(t)
	room := mgr.Create(g)
	seat := g.Seats[0].ID

	start := time.Now()
	for life := 39; life >= 35; life-- {
		setLife(t, room, seat, life)
	}
	lastSeq := room.Seq()
	if time.Since(start) >= every {
		t.Skip("the host took longer than the interval to make five commits")
	}
	if got := restoreWrites(room); got != 1 {
		t.Fatalf("restore points written inside the interval = %d, want 1 (the first commit's)", got)
	}
	if got := restoreFileSeq(t, dir, g.ID); got != 1 {
		t.Fatalf("restore point seq = %d, want 1 until the interval runs out", got)
	}

	deadline := time.Now().Add(every + 2*time.Second)
	for restoreFileSeq(t, dir, g.ID) != lastSeq {
		if time.Now().After(deadline) {
			t.Fatalf("the last state (seq %d) was not written within the interval; disk has seq %d", lastSeq, restoreFileSeq(t, dir, g.ID))
		}
		time.Sleep(10 * time.Millisecond)
	}
	if took := time.Since(start); took < every {
		t.Errorf("the trailing write came after %v, inside the %v interval", took, every)
	}
	if got := restoreWrites(room); got != 2 {
		t.Errorf("restore points written = %d, want 2 (the first commit's and the trailing one)", got)
	}

	// After a quiet spell longer than the interval, the next commit is
	// written at once.
	time.Sleep(every)
	setLife(t, room, seat, 20)
	if got := restoreFileSeq(t, dir, g.ID); got != room.Seq() {
		t.Errorf("after a quiet interval the commit was not written at once: disk seq %d, live %d", got, room.Seq())
	}
}

// TestShutdownFlushWritesThePendingRestorePoint: with a long interval,
// the pending state is held back until the shutdown flush writes it,
// and a restart from that file is the live game.
func TestShutdownFlushWritesThePendingRestorePoint(t *testing.T) {
	dir := t.TempDir()
	mgr := rateLimitedManager(t, dir, time.Hour)
	g := newPersistGame(t)
	room := mgr.Create(g)
	seat := g.Seats[0].ID

	setLife(t, room, seat, 39)
	setLife(t, room, seat, 31)
	setLife(t, room, seat, 27)
	if got := restoreFileSeq(t, dir, g.ID); got == room.Seq() {
		t.Fatal("setup: the last commit was written inside the interval")
	}

	mgr.FlushRestorePoints()
	if got := restoreFileSeq(t, dir, g.ID); got != room.Seq() {
		t.Fatalf("after the flush disk has seq %d, want the live %d", got, room.Seq())
	}
	if rep := room.ShutdownReport(); !rep.Clean || rep.SeqBehind != 0 {
		t.Errorf("shutdown report after the flush: clean=%v seq_behind=%d, want a clean table at its restore point", rep.Clean, rep.SeqBehind)
	}

	back := restoreTestManager(t, dir)
	if out := back.RestoreRooms(); len(out) != 1 || !out[0].Restored() {
		t.Fatalf("restore: %+v", out)
	}
	if got := back.Get(g.ID).Game.Seats[0].Life; got != 27 {
		t.Errorf("restored life = %d, want 27", got)
	}
}

// TestRateLimitKeepsTheLastCleanState is #3038 / ADR 0195's clean
// restore point under the rate limit: a restorable commit held back by
// the interval is still the one written when the table then walks into
// a continuation, not the older point on disk. A restart from it
// reproduces the game as it was at that commit.
func TestRateLimitKeepsTheLastCleanState(t *testing.T) {
	dir := t.TempDir()
	mgr := rateLimitedManager(t, dir, time.Hour)
	g := newPersistGame(t)
	room := mgr.Create(g)
	seat := g.Seats[0].ID

	setLife(t, room, seat, 38) // written at once
	setLife(t, room, seat, 30) // held back
	cleanSeq := room.Seq()
	want := comparableSnapshot(t, g.CaptureSnapshot())

	openContinuation(t, room, seat)
	if rep := room.ShutdownReport(); rep.Clean {
		t.Fatal("setup: the continuation did not make the table unclean")
	}

	mgr.FlushRestorePoints()
	if got := restoreFileSeq(t, dir, g.ID); got != cleanSeq {
		t.Fatalf("disk seq = %d, want %d (the last clean commit)", got, cleanSeq)
	}
	if rep := room.ShutdownReport(); rep.SeqBehind != 1 {
		t.Errorf("seq_behind = %d, want 1 (only the continuation's commit)", rep.SeqBehind)
	}

	back := restoreTestManager(t, dir)
	if out := back.RestoreRooms(); len(out) != 1 || !out[0].Restored() {
		t.Fatalf("restore: %+v", out)
	}
	restored := back.Get(g.ID)
	if restored.Seq() != cleanSeq {
		t.Errorf("restored seq = %d, want %d", restored.Seq(), cleanSeq)
	}
	if got := comparableSnapshot(t, restored.Game.CaptureSnapshot()); !bytes.Equal(got, want) {
		i := 0
		for i < len(got) && i < len(want) && got[i] == want[i] {
			i++
		}
		t.Errorf("the restored game is not the game at the clean commit, from byte %d\n got: …%.300s\nwant: …%.300s", i, got[max(i-100, 0):], want[max(i-100, 0):])
	}
}

// comparableSnapshot is a snapshot's encoding without its capture time
// and the layer cache's version, which a restore bumps on purpose so no
// cache keyed by it survives.
func comparableSnapshot(t *testing.T, s *game.GameSnapshot) []byte {
	t.Helper()
	s.TakenAt = time.Time{}
	s.LayerVersion, s.LastResolvedVersion = 0, 0
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestEndedTableIsWrittenAtOnceUnderTheRateLimit: the final state is
// what a boot must bring back (#2919), so it does not wait.
func TestEndedTableIsWrittenAtOnceUnderTheRateLimit(t *testing.T) {
	dir := t.TempDir()
	mgr := rateLimitedManager(t, dir, time.Hour)
	g := newPersistGame(t)
	room := mgr.Create(g)
	setLife(t, room, g.Seats[0].ID, 30)
	if _, _, err := room.Apply(g.Seats[0].ID, func() error {
		g.WithWriteLock(func() { g.State = game.StateEnded })
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if got := restoreFileSeq(t, dir, g.ID); got != room.Seq() {
		t.Errorf("ended table's restore point seq = %d, want the final %d at once", got, room.Seq())
	}
}

// TestDeletedRoomWritesNoPendingRestorePoint: neither the shutdown
// flush nor the trailing timer brings a deleted table's file back.
func TestDeletedRoomWritesNoPendingRestorePoint(t *testing.T) {
	for _, tc := range []struct {
		name  string
		every time.Duration
	}{{"flush", time.Hour}, {"timer", 50 * time.Millisecond}} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			mgr := rateLimitedManager(t, dir, tc.every)
			g := newPersistGame(t)
			room := mgr.Create(g)
			setLife(t, room, g.Seats[0].ID, 30)
			setLife(t, room, g.Seats[0].ID, 29) // pending

			mgr.Delete(g.ID)
			room.FlushRestorePoint()
			if tc.every < time.Second {
				time.Sleep(4 * tc.every) // the trailing timer has fired
			}
			setLife(t, room, g.Seats[0].ID, 28) // a straggling commit
			room.FlushRestorePoint()

			if _, err := os.Stat(restorePointPath(dir, g.ID)); !os.IsNotExist(err) {
				t.Errorf("a deleted table's restore point came back (err=%v)", err)
			}
		})
	}
}

// TestEventSeqFollowsEvents guards the splice in
// encodeRestorePointLocked: the cached log goes in just before
// "eventSeq", so GameSnapshot must declare EventSeq right after Events,
// and not omitempty.
func TestEventSeqFollowsEvents(t *testing.T) {
	st := reflect.TypeFor[game.GameSnapshot]()
	f, ok := st.FieldByName("Events")
	if !ok {
		t.Fatal("GameSnapshot has no Events")
	}
	next := st.Field(f.Index[0] + 1)
	if next.Name != "EventSeq" || next.Tag.Get("json") != "eventSeq" {
		t.Fatalf("the field after GameSnapshot.Events is %s `%s`, want EventSeq `json:\"eventSeq\"`; move the splice key in encodeRestorePointLocked", next.Name, next.Tag)
	}
}

// TestRestorePointFromTheCacheIsByteIdentical: a restore point assembled
// from the encoded-event cache is byte for byte the file json.Marshal
// writes for the same capture, as the log grows, after an undo replaces
// it under a new generation, and for every fixture restore point in the
// corpus (the real ones from cmd-dev and every schema's generated set).
func TestRestorePointFromTheCacheIsByteIdentical(t *testing.T) {
	check := func(t *testing.T, r *Room, p pendingRestorePoint) {
		t.Helper()
		parts, err := r.encodeRestorePointLocked(p)
		if err != nil {
			t.Fatal(err)
		}
		got := bytes.Join(parts, nil)
		want, err := json.Marshal(restorePointFile{Seq: p.seq, Generation: r.generation, Snapshot: p.snap})
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			i := 0
			for i < len(got) && i < len(want) && got[i] == want[i] {
				i++
			}
			t.Fatalf("restore point from the cache differs from json.Marshal at byte %d of %d/%d\n got: …%.200s\nwant: …%.200s",
				i, len(got), len(want), got[max(i-40, 0):], want[max(i-40, 0):])
		}
	}

	t.Run("a growing log and an undo", func(t *testing.T) {
		g := midgameRestoreTable(t, 600)
		r := NewRoom(g, nil, t.TempDir())
		capture := func(seq uint64) pendingRestorePoint {
			snap, gen := g.CaptureRestorePoint()
			return pendingRestorePoint{seq: seq, snap: snap, eventLogGen: gen}
		}
		check(t, r, capture(1))
		emitOne(g)
		emitOne(g)
		check(t, r, capture(2))
		older := capture(3)
		before := g.Clone()
		for range 5 {
			emitOne(g)
		}
		check(t, r, capture(4))
		check(t, r, older) // a shorter capture under the same generation
		g.RestoreFrom(before)
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventChangeLife, Target: g.Seats[2].ID, Amount: -3, Label: `<undo & "regrow">`})
		})
		check(t, r, capture(5))
	})

	t.Run("fixtures", func(t *testing.T) {
		paths, err := filepath.Glob(filepath.Join("..", "game", "testdata", "snapshots", "*", "*.json"))
		if err != nil || len(paths) == 0 {
			t.Fatalf("no fixtures: %v", err)
		}
		withEvents := 0
		for _, path := range paths {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var f restorePointFile
			if err := json.Unmarshal(raw, &f); err != nil || f.Snapshot == nil {
				t.Fatalf("%s: %v", path, err)
			}
			if len(f.Snapshot.Events) > 0 {
				withEvents++
			}
			r := NewRoom(game.NewGame(), nil, "")
			r.generation = f.Generation
			t.Run(strings.TrimSuffix(filepath.Base(path), ".json"), func(t *testing.T) {
				check(t, r, pendingRestorePoint{seq: f.Seq, snap: f.Snapshot})
			})
		}
		if withEvents == 0 {
			t.Fatal("no fixture has an event log, so the splice was never exercised")
		}
	})
}
