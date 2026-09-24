package ws

// shutdown_report.go answers the question ADR 0044 decision 1 says
// the shutdown path should gain: "which tables are at a clean
// boundary, and how far back is each one's restore point" — the
// instant before the process dies is the only moment that is
// knowable, and until #524 it was thrown away.
//
// This is deliberately NOT a write. It is a read of already-held
// state — one CaptureSnapshot per room (a read lock plus the same
// in-memory deep copy writeRestorePointLocked already pays on every
// action) and the room's own bookkeeping about its last WRITTEN
// restore point. Nothing here touches disk, so it cannot make the 5s
// shutdown grace period any less safe.

import (
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ShutdownGameReport is one room's answer to "what would a restart
// cost this table right now".
type ShutdownGameReport struct {
	GameID uuid.UUID

	// LiveSeq is the room's current sequence number — the value a
	// connected client has already rendered.
	LiveSeq uint64

	// Clean reports whether the CURRENT game state is itself a usable
	// restore point: capturing it right now would produce an empty
	// ContinuationCensus. When true, HasRestorePoint's numbers below
	// describe a table that a restart would not rewind at all (the
	// live state and the last restore point coincide).
	Clean bool

	// Census is why Clean is false. Empty when Clean is true.
	Census game.ContinuationCensus

	// HasRestorePoint reports whether this room has ever written a
	// restore point — in this process's lifetime, or inherited one
	// at boot (RestoreRooms). False only for a table that started
	// this process and has not yet reached one clean boundary.
	HasRestorePoint bool

	// RestoreSeq / RestoreAge describe the LAST WRITTEN restore
	// point. Zero unless HasRestorePoint.
	RestoreSeq uint64
	RestoreAge time.Duration

	// SeqBehind is LiveSeq - RestoreSeq: how many actions have
	// happened since the last state a restart could rebuild exactly.
	// Zero unless HasRestorePoint (and, in particular, zero when
	// Clean is true, since a clean capture IS the restore point that
	// would be written next).
	SeqBehind uint64
}

// ShutdownReport captures r's current state and pairs it with the
// room's restore-point bookkeeping. Safe to call at any time; it
// takes r.mu (the same lock Apply holds across mutate → capture →
// dump) and, inside that, CaptureSnapshot's read lock on the game —
// the same nesting order every other Room method already uses.
func (r *Room) ShutdownReport() ShutdownGameReport {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.shutdownReportLocked()
}

// shutdownReportLocked is the unlocked variant. Caller must hold r.mu.
func (r *Room) shutdownReportLocked() ShutdownGameReport {
	snap := r.Game.CaptureSnapshot()
	rep := ShutdownGameReport{
		GameID:  r.Game.ID,
		LiveSeq: r.seq,
		Clean:   snap.Restorable(),
		Census:  snap.Continuations,
	}
	if !r.lastRestorePoint.At.IsZero() {
		rep.HasRestorePoint = true
		rep.RestoreSeq = r.lastRestorePoint.Seq
		rep.RestoreAge = time.Since(r.lastRestorePoint.At)
		if rep.LiveSeq > rep.RestoreSeq {
			rep.SeqBehind = rep.LiveSeq - rep.RestoreSeq
		}
	}
	return rep
}

// LogShutdownCensus logs one line per room describing what a restart
// would cost it right now, plus one summary line with the totals.
// Called once, at SIGTERM, before the hub closes (main.go) — see
// ADR 0044 decision 1 and its 2026-09-24 amendment.
//
// archived reports whether the lobby has retired a table (ADR 0044
// amendment, decision 7 / #531's note on #515): a lobby-level fact ws
// does not otherwise know, since Room and RoomManager know nothing of
// lobby.GameMeta. The caller supplies it rather than ws importing the
// lobby package, which also keeps this testable with a fake.
func LogShutdownCensus(log *slog.Logger, rooms []*Room, archived func(gameID uuid.UUID) bool) {
	if log == nil {
		log = slog.Default()
	}
	if archived == nil {
		archived = func(uuid.UUID) bool { return false }
	}

	var cleanCount, rewindCount int
	var actionsRewound uint64
	for _, r := range rooms {
		rep := r.ShutdownReport()
		isArchived := archived(rep.GameID)

		args := []any{
			"game_id", rep.GameID,
			"archived", isArchived,
			"clean", rep.Clean,
			"seq", rep.LiveSeq,
			"has_restore_point", rep.HasRestorePoint,
		}
		if rep.HasRestorePoint {
			args = append(args,
				"restore_seq", rep.RestoreSeq,
				"seq_behind", rep.SeqBehind,
				"restore_age", rep.RestoreAge.Round(time.Second).String(),
			)
		}

		if rep.Clean {
			cleanCount++
			log.Info("shutdown census: game", args...)
			continue
		}

		rewindCount++
		actionsRewound += rep.SeqBehind
		args = append(args,
			"continuations", rep.Census.Total(),
			"continuation_labels", rep.Census.Labels,
		)
		log.Warn("shutdown census: game", args...)
	}

	log.Info("shutdown census: summary",
		"games", len(rooms),
		"clean", cleanCount,
		"would_rewind", rewindCount,
		"actions_rewound", actionsRewound,
	)
}
