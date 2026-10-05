package ws

// action_metrics.go is the room's half of ADR 0123 §3's engine health
// metrics: what cmdctrl_actions_total and cmdctrl_action_apply_seconds
// record about each commit, and the restore-point lag the collector
// reads at scrape time.
//
// HOW EACH COMMIT IS LABELLED
//
//   - type: ApplyAction and ApplyExternalAction carry the action's type
//     from the actions enum; a type outside it (a client can send any
//     string) is "other", and so is plain Apply, which is not told.
//     ApplyBundle is always "bundle", one count for the whole bundle,
//     and plain ApplyExternal (the lobby's joins, decks, start, bot
//     seats, practice tables) is always "lobby".
//   - seat_kind: "admin" for an admin connection (Binding.Admin, seated
//     or not) and for an actor of uuid.Nil, which the room already
//     reads as the admin; otherwise the actor's seat: "bot" for an
//     aiseat seat, "agent" for an MCP seat, "human" for anyone else.
//     A lobby setup step is "human": the lobby does not say who asked,
//     and the HTTP metrics split those requests by route.
//   - result: "applied" when the commit changed the game, "rejected"
//     when fn (or a bundle step) refused and nothing changed. A state
//     capture that fails after fn committed is still "applied".
//
// The seat kind is read before the room lock is taken (Game.SeatDriver
// takes only the game's read lock), and the counter is recorded after
// it is released, so the lock is held for the dispatch alone, plus two
// clock reads.

import (
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// ActionTag says what one commit is, for cmdctrl_actions_total.
type ActionTag struct {
	// Type is the action's type. One outside the actions enum counts
	// as "other".
	Type actions.Type
	// Actor is the seat that acted, or uuid.Nil for the admin. On the
	// undo-minting path it is also the undo entry's caller.
	Actor uuid.UUID
	// Admin marks an admin connection, which counts as "admin" even
	// while it plays as a seat.
	Admin bool

	// fixedType and seatKind are set by the room itself for the commits
	// that are not one action: a bundle and a lobby setup step.
	fixedType string
	seatKind  string
}

// knownActionTypes is the actions enum as a set, for the type label.
var knownActionTypes = map[actions.Type]bool{}

func init() {
	for _, t := range actions.Types() {
		knownActionTypes[t] = true
		metrics.RegisterActionTypes(string(t))
	}
}

func (t ActionTag) typeLabel() string {
	if t.fixedType != "" {
		return t.fixedType
	}
	if knownActionTypes[t.Type] {
		return string(t.Type)
	}
	return metrics.ActionTypeOther
}

// commitStat is one commit's metric, filled in while it holds the lock
// and recorded after it lets go.
type commitStat struct {
	typ, kind string
	applied   bool
	held      time.Duration
}

// startCommit labels a commit before it takes the room lock. It reads
// the actor's seat under the game's read lock, never the room's.
func (r *Room) startCommit(tag ActionTag) commitStat {
	return commitStat{typ: tag.typeLabel(), kind: r.seatKind(tag)}
}

func (r *Room) seatKind(tag ActionTag) string {
	switch {
	case tag.seatKind != "":
		return tag.seatKind
	case tag.Admin, tag.Actor == uuid.Nil:
		return metrics.SeatKindAdmin
	}
	bot, agent := r.Game.SeatDriver(tag.Actor)
	switch {
	case bot:
		return metrics.SeatKindBot
	case agent:
		return metrics.SeatKindAgent
	}
	return metrics.SeatKindHuman
}

// hold starts the lock-held clock and returns the func that stops it.
// Deferred right after the room lock's own Unlock is deferred, so it
// runs first, while the lock is still held.
func (st *commitStat) hold() func() {
	start := time.Now()
	return func() { st.held = time.Since(start) }
}

func (st *commitStat) record() {
	metrics.RecordAction(st.typ, st.kind, st.applied, st.held)
}

// roomLag is one room's restore-point bookkeeping, copied under r.mu.
type roomLag struct {
	persisted  bool
	seq        uint64
	restoreSeq uint64
	restoreAt  time.Time
}

func (r *Room) restoreLag() roomLag {
	r.mu.Lock()
	defer r.mu.Unlock()
	return roomLag{
		persisted:  r.dumpDir != "",
		seq:        r.seq,
		restoreSeq: r.lastRestorePoint.Seq,
		restoreAt:  r.lastRestorePoint.At,
	}
}

// RestorePointLag is what cmdctrl_restore_point_age_seconds and
// cmdctrl_rooms_behind_restore_point report (metrics.RestorePointReader).
// Only ACTIVE rooms that write restore points count: a lobby-state or
// ended table, and a practice table (which never writes one), have
// nothing a deploy could rewind.
//
// A room is behind when its seq is past the seq of its last written
// restore point. The age is how long ago that restore point was
// written, so a table idling at its restore point adds nothing however
// old the point is. A room that has never written one (it started
// mid-continuation) is aged from its game's creation: a deploy would
// lose all of it.
//
// Lock safety: it never holds two locks. The manager's lock is held
// only to copy the room list (List), each room's lock only to copy its
// numbers (restoreLag), and the game's read lock only for its state.
func (m *RoomManager) RestorePointLag(now time.Time) metrics.RestorePointLag {
	var lag metrics.RestorePointLag
	for _, r := range m.List() {
		rl := r.restoreLag()
		if !rl.persisted || rl.seq <= rl.restoreSeq {
			continue
		}
		if r.Game.CurrentState() != game.StateActive {
			continue
		}
		since := rl.restoreAt
		if since.IsZero() {
			since = r.Game.CreatedAt
		}
		lag.Behind++
		if age := now.Sub(since); age > lag.Oldest {
			lag.Oldest = age
		}
	}
	return lag
}
