package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

var gameID = uuid.MustParse("6f1c1a52-1b44-4c55-9a1c-3c5e4f8d2a10")

// frame is one replay line: the view carrying a log window.
func frame(seq uint64, log ...protocol.LogEvent) protocol.SnapshotPayload {
	return protocol.SnapshotPayload{
		Seq: seq,
		Game: protocol.GameView{
			ID: gameID.String(),
			Seats: []protocol.PlayerView{
				{ID: "p-ana", Seat: 0, Life: 31, CommanderDamage: map[string]int{}},
				{ID: "p-bo", Seat: 1, Life: 0, Eliminated: true, CommanderDamage: map[string]int{}},
			},
			Log: log,
		},
	}
}

func jsonl(t *testing.T, frames ...protocol.SnapshotPayload) string {
	t.Helper()
	var b strings.Builder
	for _, f := range frames {
		raw, err := json.Marshal(f)
		if err != nil {
			t.Fatal(err)
		}
		b.Write(raw)
		b.WriteByte('\n')
	}
	return b.String()
}

func ev(seq uint64, kind protocol.LogKind, turn, seat int) protocol.LogEvent {
	return protocol.LogEvent{Seq: seq, Kind: kind, Turn: turn, Seat: seat}
}

func seqs(log []protocol.LogEvent) []uint64 {
	var out []uint64
	for _, e := range log {
		out = append(out, e.Seq)
	}
	return out
}

func fullWindow(from uint64) []protocol.LogEvent {
	w := make([]protocol.LogEvent, protocol.PublicLogMax)
	for i := range w {
		w[i] = ev(from+uint64(i), protocol.LogLife, 1, 0)
	}
	return w
}

func TestMergeLatestSeqWins(t *testing.T) {
	draw := func(n int) protocol.LogEvent {
		e := ev(2, protocol.LogDraw, 1, 0)
		e.Amount = n
		return e
	}
	in := jsonl(t,
		frame(1, ev(1, protocol.LogStep, 1, 0), draw(1)),
		frame(2, ev(1, protocol.LogStep, 1, 0), draw(7), ev(9, protocol.LogCast, 1, 0)),
	)
	m, err := mergeReplay(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := seqs(m.log), []uint64{1, 2, 9}; !reflect.DeepEqual(got, want) {
		t.Fatalf("seqs = %v, want %v", got, want)
	}
	if m.log[1].Amount != 7 {
		t.Errorf("draw amount = %d, want the later 7", m.log[1].Amount)
	}
	if !m.complete {
		t.Error("a short window holds the whole log: want complete")
	}
	if m.view.Log != nil {
		t.Error("the merged view should not carry a log window")
	}
}

func TestMergeDropsEntriesAnUndoRewound(t *testing.T) {
	in := jsonl(t,
		frame(1, ev(1, protocol.LogStep, 1, 0), ev(2, protocol.LogCast, 1, 0), ev(3, protocol.LogCast, 1, 0)),
		// Undo: Seq 2 is regrown as a different event and 3 is gone.
		frame(2, ev(1, protocol.LogStep, 1, 0), ev(2, protocol.LogAttack, 1, 0)),
	)
	m, err := mergeReplay(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := seqs(m.log), []uint64{1, 2}; !reflect.DeepEqual(got, want) {
		t.Fatalf("seqs = %v, want %v", got, want)
	}
	if m.log[1].Kind != protocol.LogAttack {
		t.Errorf("seq 2 kind = %q, want the regrown attack", m.log[1].Kind)
	}
}

func TestMergeDetectsAGap(t *testing.T) {
	n := uint64(protocol.PublicLogMax)
	cases := []struct {
		name   string
		frames []protocol.SnapshotPayload
		want   bool
	}{
		{"windows overlap", []protocol.SnapshotPayload{frame(1, ev(1, protocol.LogStep, 1, 0)), frame(2, fullWindow(1)...), frame(3, fullWindow(2)...)}, true},
		// The first frame is already full: older entries were dropped.
		{"first window full", []protocol.SnapshotPayload{frame(1, fullWindow(50)...)}, false},
		// Frames were lost: the next window starts past the last one's end.
		{"lost frames", []protocol.SnapshotPayload{frame(1, ev(1, protocol.LogStep, 1, 0)), frame(2, fullWindow(n+10)...)}, false},
		// Non-consecutive Seqs inside a short window are not a gap.
		{"collapsed run", []protocol.SnapshotPayload{frame(1, ev(1, protocol.LogStep, 1, 0), ev(9, protocol.LogCast, 1, 0))}, true},
	}
	for _, c := range cases {
		m, err := mergeReplay(strings.NewReader(jsonl(t, c.frames...)))
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if m.complete != c.want {
			t.Errorf("%s: complete = %v, want %v", c.name, m.complete, c.want)
		}
	}
}

func TestMergeToleratesATornLastLine(t *testing.T) {
	in := jsonl(t, frame(1, ev(1, protocol.LogStep, 1, 0))) + `{"seq":2,"game":{"id":`
	m, err := mergeReplay(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if m.frames != 1 || len(m.log) != 1 {
		t.Errorf("frames %d, log %d; want the whole first line only", m.frames, len(m.log))
	}
	if _, err := mergeReplay(strings.NewReader("not json\n")); err == nil {
		t.Error("garbage in the middle of a replay should fail")
	}
}

// Fixture for the end-to-end tests.

type env struct {
	dir string
	d   *db.DB
}

func newEnv(t *testing.T) env {
	t.Helper()
	dir := t.TempDir()
	d, err := db.Open(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return env{dir: dir, d: d}
}

func (e env) addGame(t *testing.T, id uuid.UUID, state string) {
	t.Helper()
	if _, err := e.d.Exec(
		`INSERT INTO games (id, name, state, created_at, ended_at, winner_seat, outcome) VALUES (?, 'g', ?, 1, 2, 0, 'win')`,
		id.String(), state); err != nil {
		t.Fatal(err)
	}
}

func (e env) writeReplay(t *testing.T, id uuid.UUID, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(e.dir, "replays"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(e.dir, "replays", id.String()+".jsonl"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (e env) run(t *testing.T, dry bool) (summary, string) {
	t.Helper()
	var out bytes.Buffer
	sum, err := run(context.Background(), e.d, e.dir, dry, time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC), &out)
	if err != nil {
		t.Fatal(err)
	}
	return sum, out.String()
}

func standardReplay(t *testing.T) string {
	return jsonl(t,
		frame(1, ev(1, protocol.LogStep, 1, 0)),
		frame(2, ev(1, protocol.LogStep, 1, 0), ev(2, protocol.LogCast, 1, 0), ev(3, protocol.LogEliminated, 2, 1)),
	)
}

func TestBackfillEndToEnd(t *testing.T) {
	e := newEnv(t)
	e.addGame(t, gameID, "ended")
	e.writeReplay(t, gameID, standardReplay(t))

	// Not backfilled: still active, no replay, and a practice table (no row).
	active := uuid.New()
	e.addGame(t, active, "active")
	e.writeReplay(t, active, standardReplay(t))
	noReplay := uuid.New()
	e.addGame(t, noReplay, "ended")
	e.writeReplay(t, uuid.New(), standardReplay(t))

	sum, out := e.run(t, true)
	if sum != (summary{Computed: 1, Skipped: 1}) {
		t.Fatalf("dry run: %v\n%s", sum, out)
	}
	store := gamestats.NewSQLStore(e.d)
	if _, err := store.Get(context.Background(), gameID); err != gamestats.ErrNotFound {
		t.Fatalf("dry run wrote a record: %v", err)
	}

	sum, out = e.run(t, false)
	if sum != (summary{Computed: 1, Skipped: 1}) {
		t.Fatalf("run: %v\n%s", sum, out)
	}
	rec, err := store.Get(context.Background(), gameID)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Source != gamestats.SourceReplay || rec.Version != gamestats.Version || !rec.LogComplete {
		t.Errorf("record header: %+v", rec)
	}
	if rec.Turns != 2 || len(rec.Seats) != 2 || rec.Seats[0].Result != gamestats.ResultWin {
		t.Errorf("record: %+v", rec)
	}
	if rec.Seats[0].SpellsCast != 1 || rec.Seats[1].EliminatedTurn == nil || *rec.Seats[1].EliminatedTurn != 2 {
		t.Errorf("seats: %+v", rec.Seats)
	}

	// A second run finds the current-version record and skips the game.
	sum, out = e.run(t, false)
	if sum != (summary{Skipped: 2}) || !strings.Contains(out, "already has a version") {
		t.Fatalf("rerun: %v\n%s", sum, out)
	}
}

func TestBackfillRecomputesOlderVersion(t *testing.T) {
	e := newEnv(t)
	e.addGame(t, gameID, "ended")
	e.writeReplay(t, gameID, standardReplay(t))
	store := gamestats.NewSQLStore(e.d)

	old := gamestats.Compute(gamestats.Input{GameID: gameID, View: frame(1).Game, Source: gamestats.SourceLive, LogComplete: true, Now: time.Now()})
	old.Version = gamestats.Version - 1
	if err := store.Put(context.Background(), old); err != nil {
		t.Fatal(err)
	}

	sum, out := e.run(t, false)
	if sum.Computed != 1 {
		t.Fatalf("%v\n%s", sum, out)
	}
	rec, err := store.Get(context.Background(), gameID)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Version != gamestats.Version || rec.Source != gamestats.SourceReplay || rec.Turns != 2 {
		t.Errorf("not recomputed: %+v", rec)
	}
}

func TestBackfillIncompleteLogAndFailures(t *testing.T) {
	e := newEnv(t)
	e.addGame(t, gameID, "ended")
	e.writeReplay(t, gameID, jsonl(t, frame(1, fullWindow(50)...)))
	other := uuid.New()
	e.addGame(t, other, "ended")
	e.writeReplay(t, other, standardReplay(t)) // holds gameID's frames, not other's

	sum, out := e.run(t, false)
	if sum.Computed != 1 || sum.Failed != 1 {
		t.Fatalf("%v\n%s", sum, out)
	}
	rec, err := gamestats.NewSQLStore(e.d).Get(context.Background(), gameID)
	if err != nil {
		t.Fatal(err)
	}
	if rec.LogComplete {
		t.Error("a full first window must store log_complete = 0")
	}
	if _, err := gamestats.NewSQLStore(e.d).Get(context.Background(), other); err != gamestats.ErrNotFound {
		t.Errorf("a failed game stored a record: %v", err)
	}
}
