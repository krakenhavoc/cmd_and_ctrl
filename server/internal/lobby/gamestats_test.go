package lobby

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
)

// TestAFinishedGameGetsAStatsRecord (ADR 0164 §2): a stored game that
// ends over the room, with nobody calling the lobby, gets a record whose
// results follow the engine's outcome, one line per seat.
func TestAFinishedGameGetsAStatsRecord(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	d := openTestDB(t, dir)
	stats := gamestats.NewSQLStore(d)
	l.SetGameStats(stats)
	meta, alice, _ := startTwoSeatGame(t, l, "Stats")

	room := l.RoomOf(meta.ID)
	if _, _, err := room.Apply(alice, func() error { return room.Game.Concede(alice) }); err != nil {
		t.Fatalf("conceding: %v", err)
	}
	var rec gamestats.Game
	waitFor(t, func() bool {
		l.WaitGameStats()
		got, err := stats.Get(context.Background(), meta.ID)
		rec = got
		return err == nil
	})

	if rec.Source != gamestats.SourceLive || !rec.LogComplete || rec.Version != gamestats.Version {
		t.Errorf("record header %+v", rec)
	}
	if len(rec.Seats) != 2 {
		t.Fatalf("%d seat lines, want 2", len(rec.Seats))
	}
	var aliceSeat, bobSeat gamestats.Seat
	for _, s := range rec.Seats {
		if room.Game.Seats[s.Seat].ID == alice {
			aliceSeat = s
		} else {
			bobSeat = s
		}
	}
	if aliceSeat.Result != gamestats.ResultLoss || !aliceSeat.Conceded || aliceSeat.Place == nil || *aliceSeat.Place != 2 {
		t.Errorf("Alice: %s conceded=%v place=%v; want loss, conceded, 2nd", aliceSeat.Result, aliceSeat.Conceded, aliceSeat.Place)
	}
	if bobSeat.Result != gamestats.ResultWin || bobSeat.Place == nil || *bobSeat.Place != 1 {
		t.Errorf("Bob: %s place=%v; want win, 1st", bobSeat.Result, bobSeat.Place)
	}
	// Conceded and the elimination turn come only from the log's
	// eliminated line, so they prove the record read the whole log.
	if aliceSeat.EliminatedTurn == nil {
		t.Error("Alice has no elimination turn: the record did not read the game's log")
	}
}

// countingStats records how many records were written.
type countingStats struct{ puts atomic.Int32 }

func (c *countingStats) Put(context.Context, gamestats.Game) error { c.puts.Add(1); return nil }
func (c *countingStats) Get(context.Context, uuid.UUID) (gamestats.Game, error) {
	return gamestats.Game{}, gamestats.ErrNotFound
}

// A practice table has no games row, so it gets no record; and with no
// store wired in, nothing is computed at all.
func TestStatsSkipPracticeTablesAndServersWithoutAStore(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, _, _ := startTwoSeatGame(t, l, "Practice-ish")

	l.mu.Lock()
	entry := *l.games[meta.ID]
	l.mu.Unlock()
	entry.practice = true

	spy := &countingStats{}
	l.SetGameStats(spy)
	l.mu.Lock()
	l.recordStatsLocked(&entry)
	l.mu.Unlock()
	l.WaitGameStats()
	if n := spy.puts.Load(); n != 0 {
		t.Errorf("a practice table wrote %d records", n)
	}

	l.SetGameStats(nil)
	entry.practice = false
	l.mu.Lock()
	l.recordStatsLocked(&entry)
	l.mu.Unlock()
	l.WaitGameStats()
	if n := spy.puts.Load(); n != 0 {
		t.Errorf("with no store, %d records were written", n)
	}
}
