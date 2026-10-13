package lobby

// Games the database knows and the lobby does not (#2987).
//
// A games row is written when a table is created and updated as it
// moves through its lifecycle, but a table only lives in l.games if
// the restore pass brought its room back. After a crash that can miss
// a row for three reasons: the process died before the table's first
// restore point reached disk (so RestoreRooms, which walks restore
// FILES, never visits it and the boot census never counts it); the
// file was unusable (the census counts it as abandoned); or the file
// came back but its row did not pair (loadEntry fails and the room is
// dropped). In every case the row kept saying `active`, GET
// /admin/games listed a table nobody could reach, and DELETE and
// archive answered 404 because both looked only at l.games.
//
// Two halves fix it: closeUnresidentRows retires such a row at boot,
// and Delete / SetArchived reach a row that is not resident anyway.

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// closeUnresidentRows marks every unfinished (lobby or active) row the
// restore pass did not bring back as ended. The row is not deleted:
// "my games" and the admin views read it, and an operator may still
// want to see what was lost. It carries no outcome, which is how a
// table an admin closed is recorded too (GameRecord.Outcome).
//
// keep names the games whose restore point was deliberately left on
// disk for a roll-forward (schema_too_new, unknown_effect_key). Their
// rows stay as they are, so the binary that can read the file brings
// the table back with its lifecycle intact.
//
// Returns how many rows it closed. A store error is logged and skips
// that row: a boot must not stop because one row is unreadable.
func (l *Lobby) closeUnresidentRows(log *slog.Logger, keep map[uuid.UUID]bool) int {
	ctx, cancel := storeCtx()
	ids, err := l.store.GameIDs(ctx)
	cancel()
	if err != nil {
		log.Warn("listing stored games to close unrestored ones failed", "err", err)
		return 0
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	closed := 0
	for _, id := range ids {
		if _, resident := l.games[id]; resident || keep[id] {
			continue
		}
		ctx, cancel := storeCtx()
		rec, _, err := l.store.LoadGame(ctx, id)
		cancel()
		if err != nil {
			log.Warn("loading a stored game to close it failed", "game_id", id, "err", err)
			continue
		}
		if rec.State != string(game.StateLobby) && rec.State != string(game.StateActive) {
			continue
		}
		now := time.Now().UTC()
		prev := rec.State
		rec.State = string(game.StateEnded)
		if rec.EndedAt == nil {
			rec.EndedAt = &now
		}
		ctx, cancel = storeCtx()
		err = l.store.UpdateGame(ctx, rec)
		cancel()
		if err != nil {
			log.Warn("closing an unrestored game row failed", "game_id", id, "err", err)
			continue
		}
		log.Warn("stored game was not restored; marked it ended so it stops reading as active",
			"game_id", id, "name", rec.Name, "was", prev)
		closed++
	}
	return closed
}

// keepsRestorePointForRollForward reports whether a failed restore
// left its file on disk on purpose, for a binary that can read it.
func keepsRestorePointForRollForward(o ws.RestoreOutcome) bool {
	return o.Reason == ws.ReasonSchemaTooNew || o.Reason == ws.ReasonUnknownEffectKey
}

// loadUnresident reads a games row that has no lobby entry. It maps a
// missing row to ErrGameNotFound, the same answer a wholly unknown id
// gets, and anything else to a wrapped store error. Callers hold l.mu.
func (l *Lobby) loadUnresident(id uuid.UUID) (GameRecord, []SeatRecord, error) {
	ctx, cancel := storeCtx()
	defer cancel()
	rec, seats, err := l.store.LoadGame(ctx, id)
	if err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			return GameRecord{}, nil, ErrGameNotFound
		}
		return GameRecord{}, nil, fmt.Errorf("lobby: loading game %s: %w", id, err)
	}
	return rec, seats, nil
}

// deleteUnresidentLocked is Delete for a row with no lobby entry:
// the rows, the ADR 0041 file and any restore point or snapshot the
// room manager's reaper knows the paths of. No bots run for it, so
// there is nothing to stop. Callers hold l.mu.
func (l *Lobby) deleteUnresidentLocked(id uuid.UUID) error {
	if _, _, err := l.loadUnresident(id); err != nil {
		return err
	}
	ctx, cancel := storeCtx()
	err := l.store.DeleteGame(ctx, id)
	cancel()
	if err != nil {
		return fmt.Errorf("lobby: deleting game %s: %w", id, err)
	}
	// mgr.Delete is the one place that knows the artifact paths, and it
	// is safe for an id with no room. A kept restore point left behind
	// would be dropped at the next boot anyway (no games row), but
	// there is no reason to leave it for then.
	l.mgr.Delete(id)
	l.dropReclaimsLocked(id)
	l.removeLegacyMeta(id)
	return nil
}

// setArchivedUnresidentLocked is SetArchived for a row with no lobby
// entry: archived_at flips and nothing else does. The returned meta is
// built from the rows alone, so it has no invite tokens (the lobby
// never held them for this table) and Players carries only what a
// seats row records. Callers hold l.mu.
func (l *Lobby) setArchivedUnresidentLocked(id uuid.UUID, archived bool) (GameMeta, error) {
	rec, seats, err := l.loadUnresident(id)
	if err != nil {
		return GameMeta{}, err
	}
	switch {
	case archived && rec.ArchivedAt == nil:
		now := time.Now().UTC()
		rec.ArchivedAt = &now
	case !archived && rec.ArchivedAt != nil:
		rec.ArchivedAt = nil
	default:
		return metaFromRecords(rec, seats), nil
	}
	ctx, cancel := storeCtx()
	err = l.store.UpdateGame(ctx, rec)
	cancel()
	if err != nil {
		return GameMeta{}, fmt.Errorf("lobby: archiving game %s: %w", id, err)
	}
	return metaFromRecords(rec, seats), nil
}

// metaFromRecords is the listing row for a game that has rows and no
// room.
func metaFromRecords(rec GameRecord, seats []SeatRecord) GameMeta {
	meta := GameMeta{
		ID:            rec.ID,
		Name:          rec.Name,
		CreatedAt:     rec.CreatedAt,
		Players:       make([]SeatInfo, 0, len(seats)),
		State:         rec.State,
		ArchivedAt:    rec.ArchivedAt,
		HostPlayerID:  rec.HostPlayerID,
		HostDiscordID: rec.HostDiscordID,
		CreatedBy:     parseUUIDOrNil(rec.CreatedBy),
		VoiceRoom:     rec.VoiceRoom,
	}
	for _, s := range seats {
		meta.Players = append(meta.Players, SeatInfo{
			PlayerID: s.PlayerID,
			Name:     s.GuestName,
			Seat:     s.Seat,
			DeckID:   s.DeckID,
			DeckName: s.DeckName,
			UserID:   s.UserID,
			IsBot:    s.BotTier != "",
			BotTier:  s.BotTier,
		})
	}
	return meta
}
