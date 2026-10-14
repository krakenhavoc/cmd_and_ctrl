package lobby

// Game statistics (ADR 0165 §2): when a stored game ends, derive its
// record from the engine's whole log and write it, off l.mu.

import (
	"context"
	"log/slog"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// statsTimeout bounds one record's write.
const statsTimeout = 10 * time.Second

// SetGameStats wires the game statistics store in after construction.
// Without one (a server with no database) no record is computed.
func (l *Lobby) SetGameStats(s gamestats.Store) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.gameStats = s
}

// recordStatsLocked starts computing entry's record. syncStateLocked
// calls it once, on the game's first sight of ended, after the games
// row has its outcome. A practice table has no row, so no record.
// Callers hold l.mu; the work runs on its own goroutine, and a failure
// is logged and touches nothing else.
func (l *Lobby) recordStatsLocked(entry *gameEntry) {
	if l.gameStats == nil || entry.practice || entry.room == nil {
		return
	}
	store, room, id := l.gameStats, entry.room, entry.meta.ID
	outcome, winner := entry.outcome, entry.winnerSeat
	l.statsWG.Add(1)
	go func() {
		defer l.statsWG.Done()
		view, log := protocol.FullLogOf(room.Game)
		rec := gamestats.Compute(gamestats.Input{
			GameID:      id,
			View:        view,
			Log:         log,
			Outcome:     outcome,
			WinnerSeat:  winner,
			Source:      gamestats.SourceLive,
			LogComplete: true,
			Now:         time.Now(),
		})
		ctx, cancel := context.WithTimeout(context.Background(), statsTimeout)
		defer cancel()
		if err := store.Put(ctx, rec); err != nil {
			slog.Default().Warn("game stats: write failed", "game_id", id, "err", err)
		}
	}()
}

// WaitGameStats blocks until every record started so far is written.
// For tests and shutdown.
func (l *Lobby) WaitGameStats() { l.statsWG.Wait() }
