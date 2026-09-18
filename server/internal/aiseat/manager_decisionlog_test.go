package aiseat_test

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
)

// manager_decisionlog_test.go covers the wiring rather than the
// writer: one log per GAME (not per seat), every seat observed
// through it, and closed exactly once — including on the path nobody
// calls StopBots on, which is the ordinary one. A game that ends
// normally is never stopped; its runners see the state leave
// StateActive and return. A log closed only by StopBots would
// therefore be a log that is never flushed on a finished game.

type fakeGameLog struct {
	mu     sync.Mutex
	events int
	seats  map[uuid.UUID]bool
	closes int
}

func (f *fakeGameLog) Observe(ev aiseat.DecisionEvent) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events++
	if f.seats == nil {
		f.seats = map[uuid.UUID]bool{}
	}
	f.seats[ev.Seat] = true
}

func (f *fakeGameLog) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closes++
	return nil
}

func (f *fakeGameLog) snapshot() (events, seats, closes int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.events, len(f.seats), f.closes
}

type fakeDecisionLogger struct {
	mu    sync.Mutex
	opens []uuid.UUID
	log   *fakeGameLog
}

func (f *fakeDecisionLogger) OpenGame(gameID uuid.UUID) (aiseat.GameDecisionLog, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.opens = append(f.opens, gameID)
	return f.log, nil
}

func (f *fakeDecisionLogger) openCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.opens)
}

func TestManagerInstallsOneDecisionLogPerGame(t *testing.T) {
	room := newRoom(t, 2, 11)
	gl := &fakeGameLog{}
	dl := &fakeDecisionLogger{log: gl}

	mgr := aiseat.NewManagerWithConfig(nil, aiseat.Config{}, testLogger())
	mgr.SetDecisionLogger(dl)
	seats := []aiseat.SeatSpec{
		{PlayerID: room.Game.Seats[0].ID, Tier: string(aiseat.TierRandom)},
		{PlayerID: room.Game.Seats[1].ID, Tier: string(aiseat.TierRandom)},
	}
	mgr.StartBots(room, seats)

	waitFor(t, "both seats to be observed", func() bool {
		_, s, _ := gl.snapshot()
		return s == 2
	})
	if n := dl.openCount(); n != 1 {
		t.Errorf("opened %d logs for one game; it is one file per table", n)
	}

	mgr.StopBots(room.Game.ID)
	events, seatCount, closes := gl.snapshot()
	if events == 0 {
		t.Error("no decisions were observed")
	}
	if seatCount != 2 {
		t.Errorf("%d seats observed, want 2", seatCount)
	}
	if closes != 1 {
		t.Errorf("the log was closed %d times, want exactly 1", closes)
	}

	// Idempotent: a second stop, and the watcher goroutine that also
	// races to close when the runners exit, must not double-close.
	mgr.StopBots(room.Game.ID)
	mgr.Shutdown()
	time.Sleep(50 * time.Millisecond)
	if _, _, closes := gl.snapshot(); closes != 1 {
		t.Errorf("the log was closed %d times after repeated stops", closes)
	}
}

// StartBots twice replaces the runners; the first game's log must be
// closed rather than leaked, and the second set must get its own.
func TestManagerReplacingRunnersClosesTheOldDecisionLog(t *testing.T) {
	room := newRoom(t, 2, 12)
	first, second := &fakeGameLog{}, &fakeGameLog{}
	n := 0
	dl := aiseat.DecisionLoggerFunc(func(uuid.UUID) (aiseat.GameDecisionLog, error) {
		n++
		if n == 1 {
			return first, nil
		}
		return second, nil
	})

	mgr := aiseat.NewManagerWithConfig(nil, aiseat.Config{}, testLogger())
	mgr.SetDecisionLogger(dl)
	// Both seats, so the table keeps moving after the replacement:
	// one bot and one empty chair stalls as soon as priority crosses.
	seats := []aiseat.SeatSpec{
		{PlayerID: room.Game.Seats[0].ID, Tier: string(aiseat.TierRandom)},
		{PlayerID: room.Game.Seats[1].ID, Tier: string(aiseat.TierRandom)},
	}
	mgr.StartBots(room, seats)
	waitFor(t, "the first log to see a decision", func() bool {
		e, _, _ := first.snapshot()
		return e > 0
	})
	mgr.StartBots(room, seats)
	defer mgr.StopBots(room.Game.ID)

	waitFor(t, "the first log to be closed", func() bool {
		_, _, c := first.snapshot()
		return c == 1
	})
	waitFor(t, "the second log to see a decision", func() bool {
		e, _, _ := second.snapshot()
		return e > 0
	})
	if _, _, c := second.snapshot(); c != 0 {
		t.Errorf("the live log was closed %d times", c)
	}
}
