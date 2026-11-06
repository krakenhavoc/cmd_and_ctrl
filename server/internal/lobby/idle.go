package lobby

// idle.go: closing idle tables (ADR 0211, #4562). The load cap (ADR
// 0148) adds up every unended, unarchived table, and each account may
// hold three open ones, so a table nobody starts, or a game everyone
// walked away from, would hold the cap until an admin archived it. The
// sweep here closes them on its own:
//
//   - an unstarted table with no activity for IdleLobbyAfter (24 hours);
//   - a started game with no human move for IdleGameAfter (48 hours),
//     warned 12 hours before (§6);
//   - a started game left to its bots, with nobody at the table, for
//     IdleBotsAfter (1 hour, §2).
//
// What counts as activity is §1: touchLocked at the lobby's entry
// points, and the room's LastHumanAt and LastSeatedChatAt (ws/idle.go).
// The close is an archive (§4), recorded exactly as an admin's, plus a
// reason (games.closed_reason), a log line, and, after IdleClosedKeep,
// the room and its restore point leave memory and disk.
//
// Practice tables are not this file's: they keep their own 20-minute
// reap (practice.go). Deferred tables are not resident, so the sweep
// never sees them (§3).

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/capacity"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/schedule"
)

// The constants of ADR 0211 §7. The thresholds themselves are settings
// (capacity.Limits); these are not.
const (
	// IdleSweepEvery is how often RunIdle looks at the tables. A close is
	// at most this late.
	IdleSweepEvery = 5 * time.Minute
	// IdleResidencyGrace is how long a table that just came into this
	// process is never closed: wall-clock time counts while the server
	// is down, and a long outage must not close everything at boot.
	IdleResidencyGrace = time.Hour
	// IdleWarnLead is how long before its deadline a started game is
	// warned, or a quarter of the threshold when that is shorter.
	IdleWarnLead = 12 * time.Hour
	// IdleWarnHold is the least time between a warning and the close, so
	// a warning repeated after a restart still gives an hour.
	IdleWarnHold = time.Hour
	// IdleClosedKeep is how long an idle-closed table keeps its room and
	// restore point, so an admin can unarchive it (owner decision 6).
	IdleClosedKeep = 48 * time.Hour
)

// The games.closed_reason values (migration 0031). They are the
// EventIdleClose labels, so the row and the log line agree.
const (
	ClosedIdleLobby = game.IdleCloseLobby
	ClosedIdleGame  = game.IdleCloseGame
	ClosedIdleBots  = game.IdleCloseBots
)

// idleState is the sweep's clock. Guarded by its own mutex, a leaf: it
// is read under l.mu.
type idleState struct {
	mu  sync.Mutex
	now func() time.Time
}

// SetIdleClock replaces the idle clock. Tests only; nil is time.Now.
func (l *Lobby) SetIdleClock(now func() time.Time) {
	l.idle.mu.Lock()
	defer l.idle.mu.Unlock()
	l.idle.now = now
}

func (l *Lobby) idleNow() time.Time {
	l.idle.mu.Lock()
	defer l.idle.mu.Unlock()
	if l.idle.now != nil {
		return l.idle.now()
	}
	return time.Now()
}

// touchLocked records lobby-side activity on entry (ADR 0211 §1): a
// seat joining or leaving, a deck, a bot, a setting, the host passing
// on, an invite, the start, an away answer or a take-back. Callers hold
// l.mu, and call it only for what a seated person, the host or the
// creator did; never for a spectator, a preview, a reconnect or the
// voice and #lfg managers.
func (l *Lobby) touchLocked(entry *gameEntry) {
	if entry == nil || entry.practice {
		return
	}
	entry.lastTouched = l.idleNow()
}

// TouchTable is touchLocked for an HTTP route that acts on a table
// without a lobby method of its own (a seated person's DM invite).
func (l *Lobby) TouchTable(id uuid.UUID) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if entry, ok := l.games[id]; ok {
		l.touchLocked(entry)
	}
}

// resetIdleLocked forgets entry's in-memory idle state: the bot-only
// clock and any warning. Callers hold l.mu.
func (l *Lobby) resetIdleLocked(entry *gameEntry) {
	entry.botOnlySince = time.Time{}
	entry.idleWarnedAt = time.Time{}
	entry.idleCloseAt = time.Time{}
}

// lastActivityLocked is the newest of everything that counts as entry's
// activity (ADR 0211 §1): the lobby's touch, the room's last human move
// (and, before the start, its last seated chat line), what the row
// last recorded, and the table's creation or its start. Callers hold
// l.mu.
func (l *Lobby) lastActivityLocked(entry *gameEntry, unstarted bool) time.Time {
	last := entry.meta.CreatedAt
	if !unstarted && entry.startedAt != nil {
		last = *entry.startedAt
	}
	later := func(t time.Time) {
		if t.After(last) {
			last = t
		}
	}
	later(entry.lastTouched)
	if entry.lastActiveStored != nil {
		later(*entry.lastActiveStored)
	}
	if entry.room != nil {
		later(entry.room.LastHumanAt())
		if unstarted {
			later(entry.room.LastSeatedChatAt())
		}
	}
	return last
}

// idleGameAfter is entry's game threshold. One function, so a per-table
// "slow game" setting can lengthen it later (ADR 0211 §9).
func idleGameAfter(_ *gameEntry, lim capacity.Limits) time.Duration {
	return lim.IdleGameAfter
}

// idleWarnLead is the warning's lead for a game threshold of after: 12
// hours, or a quarter of after when that is shorter (§6).
func idleWarnLead(after time.Duration) time.Duration {
	return min(IdleWarnLead, after/4)
}

// botOnlyLocked reports whether no seat still in entry's game is played
// by a person: every seat not out of the game is a bot or a stand-in
// (ADR 0211 §2). Callers hold l.mu.
func botOnlyLocked(entry *gameEntry) bool {
	out := map[uuid.UUID]bool{}
	entry.room.Game.ReadSnapshot(func() {
		for _, pl := range entry.room.Game.Seats {
			if pl != nil && pl.Eliminated {
				out[pl.ID] = true
			}
		}
	})
	for _, s := range entry.meta.Players {
		if out[s.PlayerID] || s.IsBot || s.StandIn != "" {
			continue
		}
		return false
	}
	return true
}

// idleCandidate is one table the sweep looks at, copied under l.mu.
type idleCandidate struct {
	entry *gameEntry
	// people are the table's non-bot seats, still in the game or not:
	// a conceded player watching the bots finish is at the table.
	people []uuid.UUID
}

// RunIdle ticks the idle sweep every IdleSweepEvery until ctx ends.
// Started at boot after RestoreFromDisk, beside RunAway.
func (l *Lobby) RunIdle(ctx context.Context) {
	t := time.NewTicker(IdleSweepEvery)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			l.TickIdle()
		}
	}
}

// TickIdle runs the sweep once (ADR 0211 §3): for every resident table
// that is not practice, not archived and not ended, it writes the last
// activity if it moved, warns a started game that is close to its
// deadline, and closes what is due; and it unloads every idle-closed
// table past its keep window. Exported for tests, which drive it with
// SetIdleClock as TickAway is driven.
func (l *Lobby) TickIdle() {
	now := l.idleNow()
	p := l.awayHub()

	l.mu.Lock()
	lim := l.limits
	var cands []idleCandidate
	var unload []*gameEntry
	for _, e := range l.games {
		if e.practice || e.room == nil {
			continue
		}
		if e.meta.Archived() {
			if e.closedReason != "" && !now.Before(e.meta.ArchivedAt.Add(IdleClosedKeep)) {
				unload = append(unload, e)
			}
			continue
		}
		c := idleCandidate{entry: e}
		for _, s := range e.meta.Players {
			if !s.IsBot {
				c.people = append(c.people, s.PlayerID)
			}
		}
		cands = append(cands, c)
	}
	l.mu.Unlock()

	for _, c := range cands {
		// The hub is asked outside l.mu, as the away clock asks it.
		watched := false
		if p != nil {
			for _, seat := range c.people {
				if state, _ := p.SeatPresence(c.entry.meta.ID, seat); state != protocol.PresenceOffline {
					watched = true
					break
				}
			}
		}
		l.mu.Lock()
		after := l.idleStepLocked(c.entry, lim, now, watched)
		l.mu.Unlock()
		if after != nil {
			after()
		}
	}
	for _, e := range unload {
		l.mu.Lock()
		after := l.unloadIdleLocked(e, now)
		l.mu.Unlock()
		if after != nil {
			after()
		}
	}
}

// idleStepLocked is one table's turn of the sweep. It returns what must
// run after l.mu is released (stopping bots, closing sockets, a
// broadcast), or nil. Callers hold l.mu.
func (l *Lobby) idleStepLocked(e *gameEntry, lim capacity.Limits, now time.Time, watched bool) func() {
	id := e.meta.ID
	if l.games[id] != e || e.meta.Archived() {
		return nil
	}
	state := e.room.Game.CurrentState()
	if state != game.StateLobby && state != game.StateActive {
		return nil
	}
	unstarted := state == game.StateLobby
	last := l.lastActivityLocked(e, unstarted)
	l.storeLastActiveLocked(e, last)

	// Any activity since the warning takes it back.
	if !e.idleWarnedAt.IsZero() && last.After(e.idleWarnedAt) {
		e.idleWarnedAt, e.idleCloseAt = time.Time{}, time.Time{}
	}
	graceEnds := e.residentSince.Add(IdleResidencyGrace)
	inGrace := now.Before(graceEnds)
	idleFor := now.Sub(last)

	if unstarted {
		e.botOnlySince = time.Time{}
		// ADR 0215 §7, this ADR's amendment: a scheduled table nobody
		// came to, with no person seated an hour after its start, is
		// closed as an idle lobby rather than holding load for a day.
		if e.sched != nil && !inGrace && !now.Before(e.sched.start.Add(schedule.NoShowAfter)) && !personSeatedLocked(e) {
			return l.closeIdleLocked(e, ClosedIdleLobby, int(now.Sub(e.sched.start)/time.Hour), now)
		}
		if lim.IdleLobbyAfter > 0 && idleFor >= lim.IdleLobbyAfter && !inGrace {
			return l.closeIdleLocked(e, ClosedIdleLobby, int(idleFor/time.Hour), now)
		}
		return nil
	}

	// §2: a game left to its bots, with nobody at the table.
	if lim.IdleBotsAfter > 0 && !watched && botOnlyLocked(e) {
		if e.botOnlySince.IsZero() {
			e.botOnlySince = now
		}
		if alone := now.Sub(e.botOnlySince); alone >= lim.IdleBotsAfter && !inGrace {
			return l.closeIdleLocked(e, ClosedIdleBots, int(alone/time.Minute), now)
		}
	} else {
		e.botOnlySince = time.Time{}
	}

	after := idleGameAfter(e, lim)
	if after <= 0 {
		return nil
	}
	deadline := last.Add(after)
	if e.idleWarnedAt.IsZero() {
		if now.Before(deadline.Add(-idleWarnLead(after))) {
			return nil
		}
		// §6: warn, and close no sooner than an hour from now, or than
		// the end of the residency grace.
		closesAt := deadline
		if t := now.Add(IdleWarnHold); t.After(closesAt) {
			closesAt = t
		}
		if graceEnds.After(closesAt) {
			closesAt = graceEnds
		}
		return l.warnIdleLocked(e, int(idleFor/time.Hour), now, closesAt)
	}
	if now.Before(e.idleCloseAt) || inGrace {
		return nil
	}
	return l.closeIdleLocked(e, ClosedIdleGame, int(idleFor/time.Hour), now)
}

// storeLastActiveLocked writes games.last_active_at when last is newer
// than what the row holds (ADR 0211 §1). Nothing on the commit path
// writes it, so it is at most one sweep stale. Callers hold l.mu.
func (l *Lobby) storeLastActiveLocked(e *gameEntry, last time.Time) {
	last = last.UTC().Truncate(time.Millisecond)
	if e.lastActiveStored != nil && !last.After(*e.lastActiveStored) {
		return
	}
	ctx, cancel := storeCtx()
	err := l.store.SetLastActive(ctx, e.meta.ID, last)
	cancel()
	if err != nil {
		slog.Default().Warn("persist last activity failed", "game_id", e.meta.ID, "err", err)
		return
	}
	e.lastActiveStored = &last
}

// warnIdleLocked warns a started game that it will be closed at
// closesAt (ADR 0211 §6): a line in its log, and the warning held in
// memory until a move clears it. Callers hold l.mu.
func (l *Lobby) warnIdleLocked(e *gameEntry, hours int, now, closesAt time.Time) func() {
	broadcast, err := l.applyLocked(e.meta.ID, e, func() error {
		e.room.Game.NoteIdleClose(game.IdleCloseWarning, hours, closesAt)
		return nil
	})
	if err != nil {
		slog.Default().Warn("idle warning failed", "game_id", e.meta.ID, "err", err)
		return nil
	}
	e.idleWarnedAt, e.idleCloseAt = now, closesAt
	metrics.IdleWarned()
	slog.Default().Info("idle table warned", "game_id", e.meta.ID, "idle_hours", hours,
		"closes_at", closesAt.UTC().Format(time.RFC3339))
	return broadcast
}

// idleMetricKind is a closed_reason's metrics kind.
func idleMetricKind(reason string) string {
	switch reason {
	case ClosedIdleLobby:
		return metrics.IdleKindLobby
	case ClosedIdleBots:
		return metrics.IdleKindBots
	}
	return metrics.IdleKindGame
}

// closeIdleLocked closes entry for being idle (ADR 0211 §4): one log
// line, committed with the undo ring emptied, then the admin close's own
// steps (archiveLocked), with the reason recorded and the sockets
// evicted. amount is the hours idle (minutes for ClosedIdleBots), for
// the log line. Callers hold l.mu and run the result after releasing
// it.
func (l *Lobby) closeIdleLocked(e *gameEntry, reason string, amount int, now time.Time) func() {
	id := e.meta.ID
	if _, _, err := e.room.ApplyExternalClearingUndo(func() error {
		e.room.Game.NoteIdleClose(reason, amount, time.Time{})
		return nil
	}); err != nil {
		// The line is for the history; the close matters more.
		slog.Default().Warn("idle close log line failed", "game_id", id, "err", err)
	}
	stopBots := l.archiveLocked(e, now.UTC())
	e.closedReason = reason
	l.resetIdleLocked(e)
	l.syncStateLocked(e)
	l.persistGameLocked(e)
	metrics.IdleClosed(idleMetricKind(reason))
	slog.Default().Info("idle table closed", "game_id", id, "reason", reason, "idle", amount)
	evictor := l.evictor
	return func() {
		if stopBots != nil {
			stopBots()
		}
		if evictor != nil {
			evictor.EvictGame(id)
		}
	}
}

// unloadIdleLocked lets an idle-closed table go once its keep window has
// passed (ADR 0211 §4): the row is written ended, with no outcome, as
// closeUnresidentRows records a table the server let go; the entry
// leaves the registry; and the room, its restore point and its snapshot
// leave memory and disk. The replay is kept. Callers hold l.mu.
func (l *Lobby) unloadIdleLocked(e *gameEntry, now time.Time) func() {
	id := e.meta.ID
	if l.games[id] != e || !e.meta.Archived() || e.closedReason == "" ||
		now.Before(e.meta.ArchivedAt.Add(IdleClosedKeep)) {
		return nil
	}
	e.meta.State = string(game.StateEnded)
	if e.endedAt == nil {
		at := now.UTC()
		e.endedAt = &at
	}
	l.persistGameLocked(e)
	l.dropEntryLocked(id)
	l.mgr.Unload(id)
	l.dropReclaimsLocked(id)
	metrics.IdleUnloaded()
	slog.Default().Info("idle-closed table unloaded", "game_id", id, "reason", e.closedReason)
	evictor := l.evictor
	return func() {
		if evictor != nil {
			evictor.EvictGame(id)
		}
	}
}
