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
type restorePointFile struct {
	Seq      uint64             `json:"seq"`
	Snapshot *game.GameSnapshot `json:"snapshot"`
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
		return false, nil
	}

	payload, err := json.Marshal(restorePointFile{Seq: seq, Snapshot: snap})
	if err != nil {
		return false, fmt.Errorf("marshal restore point: %w", err)
	}
	if err := writeFileAtomic(restorePointPath(r.dumpDir, r.Game.ID), payload); err != nil {
		return false, err
	}
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
}

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
		m.log.Error("restore point unreadable; game abandoned", "path", path, "err", err)
		return res
	}
	var file restorePointFile
	if err := json.Unmarshal(raw, &file); err != nil {
		res.Err = fmt.Errorf("decode %s: %w", path, err)
		m.log.Error("restore point undecodable; game abandoned", "path", path, "err", err)
		return res
	}
	if file.Snapshot == nil {
		res.Err = fmt.Errorf("%s: no snapshot in file", path)
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
			m.log.Error("restore point written by a newer server; game abandoned (file kept for roll-forward)",
				"game_id", res.GameID, "path", path, "err", err)
		case errors.Is(err, game.ErrSchemaUnsupported):
			m.log.Error("restore point too old to migrate; game abandoned",
				"game_id", res.GameID, "path", path, "err", err)
		case errors.Is(err, game.ErrSnapshotNotRestorable):
			// Should not happen — the writer only publishes
			// restorable captures — so this means the file predates
			// a census counter that this binary added. Loud, because
			// it is the shape of bug this whole design exists to
			// make visible.
			m.log.Error("restore point holds continuations this build cannot rebuild; game abandoned",
				"game_id", res.GameID, "path", path, "continuations", file.Snapshot.Continuations.Labels)
		default:
			m.log.Error("restore failed; game abandoned", "game_id", res.GameID, "path", path, "err", err)
		}
		return res
	}

	room := NewRoom(g, m.log, m.dumpDir)
	room.seq = file.Seq
	m.Register(room)
	res.Room = room
	m.log.Info("game restored",
		"game_id", g.ID,
		"seq", file.Seq,
		"state", file.Snapshot.State,
		"turn", file.Snapshot.Turn.Round,
		"seats", len(file.Snapshot.Seats),
		"captured_at", file.Snapshot.TakenAt,
	)
	return res
}

// LogRestoreSummary writes one line an operator can read after a
// deploy: how many tables came back, and how many did not.
func LogRestoreSummary(log *slog.Logger, outcomes []RestoreOutcome) {
	if len(outcomes) == 0 {
		return
	}
	var restored, ended, failed int
	for _, o := range outcomes {
		switch {
		case o.Restored():
			restored++
		case o.Skipped != "":
			ended++
		default:
			failed++
		}
	}
	log.Info("restore pass complete", "restored", restored, "ended", ended, "abandoned", failed)
}
