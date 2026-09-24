package ws

// persist.go is the read half of game persistence — the half that
// never shipped.
//
// The crash-recovery dump has always written <dumpDir>/games/<id>.json
// on every Apply, and nothing has ever read it back, because it holds
// a protocol.GameView: a lossy projection built for a browser. It is
// forensics, not state.
//
// This file writes a second artifact alongside it —
// <dumpDir>/restore/<id>.json, a game.GameSnapshot — and reads it at
// boot. That is what lets a deploy cost the table a refresh instead
// of the game.
//
// THE RESTORE POINT RULE
//
// A snapshot is only written when game.GameSnapshot.Restorable() is
// true, i.e. when the game holds no live Go continuations (see the
// header of internal/game/snapshot.go). When a game IS holding one —
// mid-prompt, an ability on the stack, a Fog in the air — the write
// is SKIPPED and the previous file is left in place.
//
// So the file on disk is always the most recent state that can be
// rebuilt exactly. A restart rewinds the table to that point rather
// than resurrecting it subtly wrong, and as continuations become
// data-driven the rewind distance shrinks to nothing. Silently
// dropping a Fog would be worse than a rewind the players can see.

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// restorePointPath is the on-disk location of a game's restore point.
// Kept beside snapshotPath / replayPath so the artifact layout has one
// definition and the reaper cannot drift from the writer.
func restorePointPath(dumpDir string, id uuid.UUID) string {
	return filepath.Join(dumpDir, "restore", id.String()+".json")
}

// restorePointFile is the on-disk envelope: the game snapshot plus the
// room-level sequence number.
//
// Seq is here because the client resyncs by it. A restored room that
// restarted its counter at zero would hand reconnecting clients a seq
// lower than the one they already hold.
//
// Generation is the restore generation this file's Seq belongs to
// (#523, ADR 0044 decision 5). It is bumped by restoreOne every time
// this file is actually used to rebuild a room, then carried forward
// verbatim by every later write until the next restore bumps it
// again — so it answers "how many times has this room been rebuilt
// from disk", not "how many times has this file been written". A
// file with no Generation field decodes it as the Go zero value, 0,
// which is exactly right for every restore point ever written before
// this field existed: those rooms have never been restored, so
// generation 0 is their true generation too. See TestRestorePointFileFields
// for the field-drift guard this struct sits outside of
// (snapshot_drift_test.go only covers game.GameSnapshot).
type restorePointFile struct {
	Seq        uint64             `json:"seq"`
	Generation uint64             `json:"generation"`
	Snapshot   *game.GameSnapshot `json:"snapshot"`
}

// writeRestorePointLocked captures the game and publishes it as this
// room's restore point, atomically, IF the capture is restorable.
//
// Returns (false, nil) when the game is holding continuations — that
// is the ordinary, expected case mid-prompt and is not an error. The
// previous restore point stays on disk.
//
// Caller MUST hold r.mu.
func (r *Room) writeRestorePointLocked(seq uint64) (bool, error) {
	if r.dumpDir == "" {
		return false, nil
	}
	snap := r.Game.CaptureSnapshot()
	if !snap.Restorable() {
		// ADR 0041 P7 (#1497): count what blocked it, by kind, so the
		// shutdown census can say what keeps restore points stale
		// between deploys rather than only at the instant of one.
		r.skips.noteSkip(snap.Continuations)
		return false, nil
	}

	// r.generation is read under r.mu (the caller already holds it) —
	// every later write in this room's life carries forward the
	// generation the room is CURRENTLY serving, so the file always
	// answers "if the process died and came back right now, what
	// generation would clients already holding this room's frames be
	// told about". Only restoreOne ever advances it.
	payload, err := json.Marshal(restorePointFile{Seq: seq, Generation: r.generation, Snapshot: snap})
	if err != nil {
		return false, fmt.Errorf("marshal restore point: %w", err)
	}
	if err := writeFileAtomic(restorePointPath(r.dumpDir, r.Game.ID), payload); err != nil {
		return false, err
	}
	// Bookkeeping only, for the shutdown census (#524) — the write
	// above already succeeded, so this cannot be the thing that fails
	// a caller's Apply.
	r.lastRestorePoint = restorePointRecord{Seq: seq, At: time.Now().UTC()}
	r.skips.noteWrite()
	return true, nil
}

// RemoveRestorePoint deletes this room's restore point. Called when a
// game ends — a finished game has nothing to resume, and leaving the
// file would have every subsequent boot rebuild a dead table.
func (r *Room) RemoveRestorePoint() {
	if r.dumpDir == "" {
		return
	}
	if err := os.Remove(restorePointPath(r.dumpDir, r.Game.ID)); err != nil && !os.IsNotExist(err) {
		r.log.Warn("removing restore point failed", "game_id", r.Game.ID, "err", err)
	}
}

// writeFileAtomic publishes payload at path via CreateTemp + rename,
// the same way dumpSnapshotLocked does: a crash mid-write leaves the
// previous file intact rather than a truncated one.
func writeFileAtomic(path string, payload []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		if tmp != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
		}
	}()
	if _, err := tmp.Write(payload); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close tmp: %w", err)
	}
	tmp = nil
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("publish %s: %w", path, err)
	}
	return nil
}

// RestoreOutcome describes what happened to one restore-point file.
// Returned so the boot path can log a per-game verdict and the lobby
// can rebuild metadata only for the games that actually came back.
type RestoreOutcome struct {
	GameID  uuid.UUID
	Room    *Room // nil unless Restored
	Skipped string
	Err     error

	// Reason categorises Err for LogRestoreSummary's tally (#524), so
	// an operator can see AT A GLANCE how many abandonments were a
	// rollback (schema_too_new — expected, rolls forward and comes
	// back) versus something that needs attention. Empty unless Err
	// is set.
	Reason string

	// AbilityShortfalls are the cards this game came back with fewer
	// catalog abilities than its restore point recorded (#522). The
	// game is restored anyway and each card is flagged as not
	// automated; this is the census of them the summary line counts.
	AbilityShortfalls []game.AbilityShortfall
}

// Restore-abandonment reasons. See RestoreOutcome.Reason.
const (
	ReasonReadError         = "read_error"
	ReasonDecodeError       = "decode_error"
	ReasonEmptySnapshot     = "empty_snapshot"
	ReasonSchemaTooNew      = "schema_too_new"
	ReasonSchemaUnsupported = "schema_unsupported"
	ReasonNotRestorable     = "not_restorable"
	ReasonUnknownEffectKey  = "unknown_effect_key"
	ReasonOther             = "other"
)

// Restored reports whether this game is live again.
func (o RestoreOutcome) Restored() bool { return o.Room != nil }

// RestoreRooms reads every restore point under the manager's dumpDir,
// rebuilds the games, and registers the resulting rooms.
//
// VERSION-SKEW POLICY, as implemented here: a file this binary does
// not understand is ABANDONED — logged loudly, left on disk, and its
// game simply does not come back. It is neither guessed at nor
// allowed to block the boot.
//
// The alternatives were considered and rejected:
//
//   - Blocking the deploy on an unreadable game makes correctness of
//     a card change hostage to whatever one table happens to be
//     holding. It converts "one game rewinds" into "nobody ships".
//   - Freezing the game read-only needs a whole second lifecycle
//     state (and a UI for it) to avoid being a table that looks live
//     and silently refuses every action.
//
// Abandonment is the honest one: the players see "game not found",
// which is exactly what they see today on every single restart, so
// the worst case of this feature is the current behaviour. The file
// is deliberately NOT deleted, so an operator can roll the binary
// back and get the game returned.
//
// Errors are never fatal: a broken restore point must not stop the
// server from starting.
func (m *RoomManager) RestoreRooms() []RestoreOutcome {
	if m.dumpDir == "" {
		return nil
	}
	dir := filepath.Join(m.dumpDir, "restore")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			m.log.Warn("reading restore directory failed", "dir", dir, "err", err)
		}
		return nil
	}

	var out []RestoreOutcome
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".json") {
			continue
		}
		out = append(out, m.restoreOne(filepath.Join(dir, name)))
	}
	return out
}

func (m *RoomManager) restoreOne(path string) RestoreOutcome {
	var res RestoreOutcome

	raw, err := os.ReadFile(path)
	if err != nil {
		res.Err = fmt.Errorf("read %s: %w", path, err)
		res.Reason = ReasonReadError
		m.log.Error("restore point unreadable; game abandoned", "path", path, "err", err)
		return res
	}
	var file restorePointFile
	if err := json.Unmarshal(raw, &file); err != nil {
		res.Err = fmt.Errorf("decode %s: %w", path, err)
		res.Reason = ReasonDecodeError
		m.log.Error("restore point undecodable; game abandoned", "path", path, "err", err)
		return res
	}
	if file.Snapshot == nil {
		res.Err = fmt.Errorf("%s: no snapshot in file", path)
		res.Reason = ReasonEmptySnapshot
		m.log.Error("restore point empty; game abandoned", "path", path)
		return res
	}
	res.GameID = file.Snapshot.ID

	// A finished game is not worth resurrecting; drop the file so the
	// next boot does not reconsider it.
	if file.Snapshot.State == game.StateEnded {
		res.Skipped = "game already ended"
		_ = os.Remove(path)
		return res
	}

	g, err := file.Snapshot.RestoreStrict()
	if err != nil {
		res.Err = err
		switch {
		case errors.Is(err, game.ErrSchemaTooNew):
			// A rollback, or a mixed-version fleet. The file stays
			// put: roll the binary forward again and the game
			// returns.
			res.Reason = ReasonSchemaTooNew
			m.log.Error("restore point written by a newer server; game abandoned (file kept for roll-forward)",
				"game_id", res.GameID, "path", path, "err", err)
		case errors.Is(err, game.ErrUnknownEffectKey):
			// ADR 0041 P4 (#1497): the file names an effect this
			// binary cannot interpret, which only a newer binary can
			// have written. The rollback case, so ErrSchemaTooNew's
			// answer: keep the file for the roll-forward.
			res.Reason = ReasonUnknownEffectKey
			m.log.Error("restore point names an effect this build cannot interpret; game abandoned (file kept for roll-forward)",
				"game_id", res.GameID, "path", path, "err", err)
		case errors.Is(err, game.ErrSchemaUnsupported):
			res.Reason = ReasonSchemaUnsupported
			m.log.Error("restore point too old to migrate; game abandoned",
				"game_id", res.GameID, "path", path, "err", err)
		case errors.Is(err, game.ErrSnapshotNotRestorable):
			// Should not happen — the writer only publishes
			// restorable captures — so this means the file predates
			// a census counter that this binary added. Loud, because
			// it is the shape of bug this whole design exists to
			// make visible.
			res.Reason = ReasonNotRestorable
			m.log.Error("restore point holds continuations this build cannot rebuild; game abandoned",
				"game_id", res.GameID, "path", path, "continuations", file.Snapshot.Continuations.Labels)
		default:
			res.Reason = ReasonOther
			m.log.Error("restore failed; game abandoned", "game_id", res.GameID, "path", path, "err", err)
		}
		return res
	}

	// #522: the rebuilt-parity check. A card whose catalog entry this
	// binary no longer has in full comes back anyway — the owner's
	// call on #515 is to restore the table and say so loudly, not to
	// abandon it — and restore has already flagged it as not
	// automated. Each one is an ERROR, because it is a card that will
	// silently stop doing something mid-game unless somebody looks.
	res.AbilityShortfalls = file.Snapshot.AbilityShortfalls()
	for _, sf := range res.AbilityShortfalls {
		m.log.Error("restored card has fewer catalog abilities than its restore point recorded; flagged manual for this game",
			"game_id", res.GameID,
			"card_id", sf.CardID,
			"card", sf.Name,
			"oracle_id", sf.OracleID,
			"token_key", sf.TokenKey,
			"zone", sf.Zone,
			"entry_missing", sf.EntryMissing,
			"captured", sf.Captured,
			"restored", sf.Restored,
		)
	}

	room := NewRoom(g, m.log, m.dumpDir)
	room.seq = file.Seq
	// This IS a restore: the room this process is about to serve is
	// one generation past what the file recorded, whether the file's
	// own Generation is a real prior value or the zero-valued
	// "never restored before" default (#523, ADR 0044 decision 5).
	// Every client dialling in gets this generation on its first
	// snapshot frame, and any client that already held a HIGHER seq
	// under a lower generation number knows — from the generation
	// change alone — that it just got rewound.
	room.generation = file.Generation + 1
	// Seed the bookkeeping this restore point itself represents (#524)
	// — nothing has run yet, so the room's own last-written restore
	// point IS this file, and its age is read straight off the
	// snapshot's own capture time rather than invented.
	room.lastRestorePoint = restorePointRecord{Seq: file.Seq, At: file.Snapshot.TakenAt}
	m.Register(room)
	res.Room = room
	restorePointAge := time.Since(file.Snapshot.TakenAt)
	m.log.Info("game restored",
		"game_id", g.ID,
		"seq", file.Seq,
		"generation", room.generation,
		"state", file.Snapshot.State,
		"turn", file.Snapshot.Turn.Round,
		"seats", len(file.Snapshot.Seats),
		"captured_at", file.Snapshot.TakenAt,
		"restore_point_age", restorePointAge.Round(time.Second).String(),
	)
	return res
}

// LogRestoreSummary writes one line an operator can read after a
// deploy: how many tables came back, how many did not, and — for the
// ones that didn't — WHY, so a fleet of ordinary schema_too_new
// rollback games doesn't read the same as a fleet of decode failures
// that need attention (#524).
func LogRestoreSummary(log *slog.Logger, outcomes []RestoreOutcome) {
	if len(outcomes) == 0 {
		return
	}
	var restored, ended, failed, degradedGames, degradedCards int
	reasons := map[string]int{}
	for _, o := range outcomes {
		switch {
		case o.Restored():
			restored++
		case o.Skipped != "":
			ended++
		default:
			failed++
			reason := o.Reason
			if reason == "" {
				reason = ReasonOther
			}
			reasons[reason]++
		}
		if n := len(o.AbilityShortfalls); n > 0 {
			degradedGames++
			degradedCards += n
		}
	}
	attrs := []any{"restored", restored, "ended", ended, "abandoned", failed}
	if failed > 0 {
		attrs = append(attrs, "abandoned_reasons", reasons)
	}
	// #522: a restore that stripped abilities from a card is a
	// restored table, not an abandoned one, so it is counted beside
	// the verdicts rather than as one — and escalated to ERROR,
	// because each per-card line above it is one.
	if degradedCards > 0 {
		attrs = append(attrs, "games_with_lost_abilities", degradedGames, "cards_with_lost_abilities", degradedCards)
		log.Error("restore pass complete; some cards came back with fewer abilities than captured", attrs...)
		return
	}
	log.Info("restore pass complete", attrs...)
}
