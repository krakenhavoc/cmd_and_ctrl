package lobby

// Game statistics (ADR 0165 §2, ADR 0217 §1): when a stored game ends,
// derive its record from the engine's whole log and the room's replay
// and write it, off l.mu, on one worker.

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// statsTimeout bounds one record's write.
const statsTimeout = 10 * time.Second

// statsQueueWarn is the queue length that logs a warning. Games end a
// few at a time, so a queue this long means the worker is falling
// behind; nothing is dropped.
const statsQueueWarn = 16

// SetGameStats wires the game statistics store in after construction.
// Without one (a server with no database) no record is computed.
func (l *Lobby) SetGameStats(s gamestats.Store) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.gameStats = s
}

// statsJob is one ended game waiting for its record.
type statsJob struct {
	room    *ws.Room
	id      uuid.UUID
	outcome string
	winner  *int
	store   gamestats.Store
	results *resultsJob
	cards   gamestats.CardIndex
}

// recordStatsLocked queues entry's record. syncStateLocked calls it
// once, on the game's first sight of ended, after the games row has its
// outcome. A practice table has no row, so no record. Callers hold
// l.mu; the work runs on the one stats worker (ADR 0217 §1), because
// reading a replay costs more than folding the log, and a failure is
// logged and touches nothing else.
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
	job := statsJob{room: entry.room, id: entry.meta.ID, outcome: entry.outcome, winner: entry.winnerSeat,
		store: store, results: results}
	if idx := l.cardIndex.Load(); idx != nil {
		job.cards = idx
	}
	l.statsWG.Add(1)
	l.statsMu.Lock()
	l.statsQueue = append(l.statsQueue, job)
	queued, start := len(l.statsQueue), !l.statsRunning
	l.statsRunning = true
	l.statsMu.Unlock()
	if queued >= statsQueueWarn {
		slog.Default().Warn("game stats: queue is long", "queued", queued)
	}
	if start {
		go l.statsWorker()
	}
}

// statsWorker computes queued records one at a time until the queue is
// empty, then stops; the next ended game starts it again.
func (l *Lobby) statsWorker() {
	for {
		l.statsMu.Lock()
		if len(l.statsQueue) == 0 {
			l.statsRunning = false
			l.statsMu.Unlock()
			return
		}
		job := l.statsQueue[0]
		l.statsQueue[0] = statsJob{}
		l.statsQueue = l.statsQueue[1:]
		l.statsMu.Unlock()
		job.run()
		l.statsWG.Done()
	}
}

// run computes and writes one record from the engine's whole log and
// the room's replay.
func (j statsJob) run() {
	// The room writes each frame under its own lock as it commits, so
	// taking that lock once means the final frame is on disk.
	_ = j.room.Seq()
	view, log := protocol.FullLogOf(j.room.Game, protocol.StatsLog)
	in := gamestats.Input{
		GameID:      j.id,
		View:        view,
		Log:         log,
		Outcome:     j.outcome,
		WinnerSeat:  j.winner,
		Source:      gamestats.SourceLive,
		LogComplete: true,
		Now:         time.Now(),
		Cards:       j.cards,
	}
	if path := j.room.ReplayPath(); path != "" {
		in.Frames = gamestats.ReplayFile(path, j.id.String())
	}
	rec := gamestats.Compute(in)
	if j.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), statsTimeout)
		defer cancel()
		if err := j.store.Put(ctx, rec); err != nil {
			slog.Default().Warn("game stats: write failed", "game_id", j.id, "err", err)
		}
	}
	if j.results != nil {
		j.results.run(view, rec)
	}
}

// WaitGameStats blocks until every record queued so far is written.
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

// GameCardsSeat is one opponent's line of GameCards: the commanders and
// the cards that seat cast. It names no one (ADR 0165 §9, ADR 0179).
type GameCardsSeat struct {
	Commanders []string   `json:"commanders"`
	Cards      []GameCard `json:"cards"`
}

// GameCard is a card name and how often it was cast.
type GameCard struct {
	Name  string `json:"name"`
	Casts int    `json:"casts"`
}

// GameCards is the body of GET /me/games/{id}/cards.
type GameCards struct {
	// Ready is false while the game's record is not written yet (it is
	// derived a moment after the game ends), or a server keeps none.
	Ready     bool            `json:"ready"`
	Opponents []GameCardsSeat `json:"opponents"`
}

// GameCards lists the commanders and the cards the caller's opponents
// cast in a game that has ended (ADR 0179 §7: the shop links after a
// game). The caller must hold a seat at the game (ErrPlayerNotInGame
// otherwise) and the game must have ended (409 otherwise): the answer
// is never given while a game runs. Their own seat is left out, and no
// seat is named.
func (l *Lobby) GameCards(userID, gameID uuid.UUID) (GameCards, error) {
	games, err := l.MyGames(userID)
	if err != nil {
		return GameCards{}, err
	}
	var mine *MyGame
	for i := range games {
		if games[i].ID == gameID {
			mine = &games[i]
			break
		}
	}
	if mine == nil {
		return GameCards{}, ErrPlayerNotInGame
	}
	if mine.State != "ended" {
		return GameCards{}, httpError(http.StatusConflict, "this game has not ended yet")
	}
	out := GameCards{Opponents: []GameCardsSeat{}}
	l.mu.Lock()
	store := l.gameStats
	l.mu.Unlock()
	if store == nil {
		return out, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), statsTimeout)
	defer cancel()
	rec, err := store.Get(ctx, gameID)
	if errors.Is(err, gamestats.ErrNotFound) {
		return out, nil
	}
	if err != nil {
		return GameCards{}, err
	}
	out.Ready = true
	for _, st := range rec.Seats {
		if st.Seat == mine.Seat {
			continue
		}
		o := GameCardsSeat{Commanders: st.Commanders, Cards: make([]GameCard, 0, len(st.Cards))}
		if o.Commanders == nil {
			o.Commanders = []string{}
		}
		for _, c := range st.Cards {
			o.Cards = append(o.Cards, GameCard{Name: c.Name, Casts: c.Casts})
		}
		out.Opponents = append(out.Opponents, o)
	}
	return out, nil
}
