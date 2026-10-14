package lobby

// Game statistics (ADR 0165 §2): when a stored game ends, derive its
// record from the engine's whole log and write it, off l.mu.

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

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
	if entry.practice || entry.room == nil {
		return
	}
	// The results post (ADR 0174) rides the same computation, so it is
	// decided here, under the lock, from the entry as it is now.
	store, results := l.gameStats, l.resultsJobLocked(entry)
	if store == nil && results == nil {
		return
	}
	room, id := entry.room, entry.meta.ID
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
		if store != nil {
			ctx, cancel := context.WithTimeout(context.Background(), statsTimeout)
			defer cancel()
			if err := store.Put(ctx, rec); err != nil {
				slog.Default().Warn("game stats: write failed", "game_id", id, "err", err)
			}
		}
		if results != nil {
			results.run(view, rec)
		}
	}()
}

// WaitGameStats blocks until every record started so far is written.
// For tests and shutdown.
func (l *Lobby) WaitGameStats() { l.statsWG.Wait() }

// UserStats is the caller's own game statistics (GET /me/stats, ADR
// 0165 §7). A server with no stats store has none to show.
func (l *Lobby) UserStats(userID uuid.UUID) (gamestats.UserStats, error) {
	l.mu.Lock()
	store := l.gameStats
	l.mu.Unlock()
	if store == nil || userID == uuid.Nil {
		return gamestats.NoStore{}.ForUser(context.Background(), "")
	}
	ctx, cancel := context.WithTimeout(context.Background(), statsTimeout)
	defer cancel()
	return store.ForUser(ctx, userID.String())
}

// StatsOverview is every game's statistics for an operator (GET
// /admin/stats, ADR 0165 §8). A server with no stats store has none.
func (l *Lobby) StatsOverview() (gamestats.Overview, error) {
	l.mu.Lock()
	store := l.gameStats
	l.mu.Unlock()
	if store == nil {
		return gamestats.EmptyOverview(), nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), statsTimeout)
	defer cancel()
	return store.Overview(ctx)
}
