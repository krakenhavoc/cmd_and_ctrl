package ws

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shutdown_report_test.go is #524: the shutdown-time census that logs
// what a restart would cost each live table, without writing anything
// to disk. The two claims under test are ADR 0044 decision 8's own
// example (a Treasure on the battlefield must not make a clean table
// look dirty) and the ordinary case decision 1 exists for (a table
// mid-prompt names the continuation holding it open).

// newShutdownTestRoom builds a started 2-player game with a
// persistable RNG (so a clean capture really is restorable) wrapped
// in a Room with disk persistence enabled, so writeRestorePointLocked
// actually runs and Room.lastRestorePoint gets populated the same way
// production does it.
func newShutdownTestRoom(t *testing.T) (*Room, *game.Game, uuid.UUID) {
	t.Helper()
	g := newPersistGame(t)
	room := NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), t.TempDir())
	return room, g, g.Seats[0].ID
}

// TestShutdownReportCleanWithTreasure is ADR 0044 decision 8's claim,
// read from the other end: a Treasure token on the battlefield is
// catalog data, not a live closure, so a table holding one still
// reports clean at shutdown — the census must not regress into
// treating every token-bearing board as mid-prompt.
func TestShutdownReportCleanWithTreasure(t *testing.T) {
	room, g, seat0 := newShutdownTestRoom(t)

	if _, _, err := room.Apply(seat0, func() error {
		g.WithWriteLock(func() { g.Seats[0].Life = 30 })
		return nil
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if _, _, err := room.Apply(seat0, func() error {
		_, err := g.SpawnCards(seat0, seat0, game.ZoneBattlefield, effects.TreasureToken(), 1)
		return err
	}); err != nil {
		t.Fatalf("Apply spawn Treasure: %v", err)
	}

	rep := room.ShutdownReport()
	if !rep.Clean {
		t.Errorf("Clean = false, want true — a Treasure's mana ability is catalog data (ADR 0044 decision 8); census: %+v", rep.Census)
	}
	if rep.Census.Total() != 0 {
		t.Errorf("Census.Total() = %d, want 0; labels: %v", rep.Census.Total(), rep.Census.Labels)
	}
	if !rep.HasRestorePoint {
		t.Fatal("HasRestorePoint = false, want true — the Treasure spawn itself is restorable and should have written one")
	}
	if rep.SeqBehind != 0 {
		t.Errorf("SeqBehind = %d, want 0 — a clean table's restore point should be its live seq", rep.SeqBehind)
	}
	if rep.LiveSeq != room.Seq() {
		t.Errorf("LiveSeq = %d, want %d", rep.LiveSeq, room.Seq())
	}
}

// TestShutdownReportNotCleanNamesTheContinuation is the mid-prompt
// case: an ability on the stack whose resolution is a Go closure
// blocks the restore point, and the shutdown report must say so by
// name (the census label), not just with a boolean.
func TestShutdownReportNotCleanNamesTheContinuation(t *testing.T) {
	room, g, seat0 := newShutdownTestRoom(t)

	if _, _, err := room.Apply(seat0, func() error {
		g.WithWriteLock(func() { g.Seats[0].Life = 30 })
		return nil
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	cleanSeq := room.Seq()

	if _, _, err := room.Apply(seat0, func() error {
		g.WithWriteLock(func() {
			id := uuid.New()
			g.StackMeta = map[uuid.UUID]*game.StackItem{id: {
				ID:         id,
				Kind:       game.StackItemActivated,
				Controller: seat0,
				Label:      "the shutdown test ability",
				Effect:     func(*game.Game, *game.StackItem) error { return nil },
			}}
		})
		return nil
	}); err != nil {
		t.Fatalf("Apply stack item: %v", err)
	}

	rep := room.ShutdownReport()
	if rep.Clean {
		t.Fatalf("Clean = true, want false; census: %+v", rep.Census)
	}
	if rep.Census.StackEffects != 1 {
		t.Errorf("Census.StackEffects = %d, want 1", rep.Census.StackEffects)
	}
	found := false
	for _, l := range rep.Census.Labels {
		if strings.Contains(l, "the shutdown test ability") {
			found = true
		}
	}
	if !found {
		t.Errorf("Census.Labels = %v, want one naming %q", rep.Census.Labels, "the shutdown test ability")
	}
	if !rep.HasRestorePoint {
		t.Fatal("HasRestorePoint = false, want true — the earlier clean Apply should have written one")
	}
	if rep.RestoreSeq != cleanSeq {
		t.Errorf("RestoreSeq = %d, want %d (the last clean boundary)", rep.RestoreSeq, cleanSeq)
	}
	wantBehind := room.Seq() - cleanSeq
	if rep.SeqBehind != wantBehind {
		t.Errorf("SeqBehind = %d, want %d", rep.SeqBehind, wantBehind)
	}
	if rep.SeqBehind == 0 {
		t.Error("SeqBehind = 0, want > 0 — the stack item's Apply advanced seq without a new restore point")
	}
}

// TestShutdownReportNoRestorePointYet is the boundary case: a room
// that has never once reached a clean boundary (or has persistence
// disabled) must report HasRestorePoint = false rather than a
// misleading zero seq.
func TestShutdownReportNoRestorePointYet(t *testing.T) {
	g := newPersistGame(t)
	room := NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	seat0 := g.Seats[0].ID

	if _, _, err := room.Apply(seat0, func() error {
		g.WithWriteLock(func() { g.Seats[0].Life = 30 })
		return nil
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	rep := room.ShutdownReport()
	if rep.HasRestorePoint {
		t.Errorf("HasRestorePoint = true, want false — dumpDir is disabled, nothing was ever written")
	}
	if rep.RestoreSeq != 0 || rep.SeqBehind != 0 {
		t.Errorf("RestoreSeq/SeqBehind = %d/%d, want 0/0 when there is no restore point", rep.RestoreSeq, rep.SeqBehind)
	}
	if !rep.Clean {
		t.Errorf("Clean = false, want true — the live state itself holds no continuations")
	}
}

// TestLogShutdownCensusFormatsBothKinds is the operator-facing half:
// one distinguishable line per room, an archived table flagged as
// such, and a summary line with totals that agree with the per-room
// lines.
func TestLogShutdownCensusFormatsBothKinds(t *testing.T) {
	cleanRoom, cleanGame, cleanSeat := newShutdownTestRoom(t)
	if _, _, err := cleanRoom.Apply(cleanSeat, func() error {
		cleanGame.WithWriteLock(func() { cleanGame.Seats[0].Life = 30 })
		return nil
	}); err != nil {
		t.Fatalf("Apply clean: %v", err)
	}

	dirtyRoom, dirtyGame, dirtySeat := newShutdownTestRoom(t)
	if _, _, err := dirtyRoom.Apply(dirtySeat, func() error {
		dirtyGame.WithWriteLock(func() { dirtyGame.Seats[0].Life = 30 })
		return nil
	}); err != nil {
		t.Fatalf("Apply dirty clean boundary: %v", err)
	}
	if _, _, err := dirtyRoom.Apply(dirtySeat, func() error {
		dirtyGame.WithWriteLock(func() {
			id := uuid.New()
			dirtyGame.StackMeta = map[uuid.UUID]*game.StackItem{id: {
				ID:         id,
				Kind:       game.StackItemActivated,
				Controller: dirtySeat,
				Label:      "named continuation for the log test",
				Effect:     func(*game.Game, *game.StackItem) error { return nil },
			}}
		})
		return nil
	}); err != nil {
		t.Fatalf("Apply dirty stack item: %v", err)
	}

	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	LogShutdownCensus(log, []*Room{cleanRoom, dirtyRoom}, func(id uuid.UUID) bool {
		return id == dirtyGame.ID
	})

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")

	var cleanLine, dirtyLine, summaryLine string
	for _, l := range lines {
		switch {
		case strings.Contains(l, cleanGame.ID.String()):
			cleanLine = l
		case strings.Contains(l, dirtyGame.ID.String()):
			dirtyLine = l
		case strings.Contains(l, "shutdown census: summary"):
			summaryLine = l
		}
	}

	if cleanLine == "" {
		t.Fatalf("no log line for the clean game; full output:\n%s", out)
	}
	if !strings.Contains(cleanLine, "clean=true") {
		t.Errorf("clean line missing clean=true: %s", cleanLine)
	}
	if !strings.Contains(cleanLine, "archived=false") {
		t.Errorf("clean line missing archived=false: %s", cleanLine)
	}

	if dirtyLine == "" {
		t.Fatalf("no log line for the dirty game; full output:\n%s", out)
	}
	if !strings.Contains(dirtyLine, "clean=false") {
		t.Errorf("dirty line missing clean=false: %s", dirtyLine)
	}
	if !strings.Contains(dirtyLine, "archived=true") {
		t.Errorf("dirty line missing archived=true: %s", dirtyLine)
	}
	if !strings.Contains(dirtyLine, "named continuation for the log test") {
		t.Errorf("dirty line does not name the blocking continuation: %s", dirtyLine)
	}
	if !strings.Contains(dirtyLine, "seq_behind=1") {
		t.Errorf("dirty line missing seq_behind=1: %s", dirtyLine)
	}

	if summaryLine == "" {
		t.Fatalf("no summary line; full output:\n%s", out)
	}
	if !strings.Contains(summaryLine, "games=2") {
		t.Errorf("summary missing games=2: %s", summaryLine)
	}
	if !strings.Contains(summaryLine, "clean=1") {
		t.Errorf("summary missing clean=1: %s", summaryLine)
	}
	if !strings.Contains(summaryLine, "would_rewind=1") {
		t.Errorf("summary missing would_rewind=1: %s", summaryLine)
	}
	if !strings.Contains(summaryLine, "actions_rewound=1") {
		t.Errorf("summary missing actions_rewound=1: %s", summaryLine)
	}
}

// TestLogShutdownCensusNilSafety documents that a nil archived
// predicate is a supported call (main.go never passes one, but a
// future caller might forget) and defaults to "nothing is archived"
// rather than panicking.
func TestLogShutdownCensusNilSafety(t *testing.T) {
	room, g, seat0 := newShutdownTestRoom(t)
	if _, _, err := room.Apply(seat0, func() error {
		g.WithWriteLock(func() { g.Seats[0].Life = 30 })
		return nil
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, nil))
	LogShutdownCensus(log, []*Room{room}, nil)

	if !strings.Contains(buf.String(), "archived=false") {
		t.Errorf("nil archived predicate should default to false; output:\n%s", buf.String())
	}
}
