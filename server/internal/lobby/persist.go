package lobby

// persist.go keeps the lobby's half of a game on disk so that a
// restored game.Game is actually reachable.
//
// The engine snapshot (internal/game/snapshot.go) rebuilds the table.
// It does not rebuild the INVITE — and without the invite token, a
// restored game is a room nobody can open. Sessions do not survive a
// restart either (auth.MemoryAuthenticator is explicit about that), so
// after a deploy every player re-authenticates through the invite
// link. That link has to still work, which means the token has to be
// on disk.
//
// What is written: GameMeta — display name, created-at, the two invite
// tokens, and the seat list. Note that this file contains SECRETS (the
// player and spectator invites). It sits under the same data dir that
// already holds the unfiltered crash dumps — which contain every
// player's hand — so it does not widen the blast radius of that
// directory, but it does mean the directory's permissions are load-
// bearing. It is written 0600 rather than 0644 for that reason.

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// metaPath is the on-disk location of one game's lobby metadata.
func metaPath(dumpDir string, id uuid.UUID) string {
	return filepath.Join(dumpDir, "lobby", id.String()+".json")
}

// dumpDir returns the data root, or "" when persistence is disabled.
// Sourced from the RoomManager so there is exactly one knob
// (CMDCTRL_DATA_DIR) rather than two that can disagree.
func (l *Lobby) dumpDir() string {
	if l.mgr == nil {
		return ""
	}
	return l.mgr.DumpDir()
}

// persistMetaLocked writes an entry's metadata. Called from the lobby
// mutations that change it: Create, Join, SetDeck, Start.
//
// It writes under l.mu. applyLocked deliberately defers the hub
// fan-out until after the unlock, and the same instinct applies here —
// but the cases are not alike: a fan-out marshals one filtered view
// per connected client, while this is a single sub-kilobyte file
// written a handful of times over a game's whole lifetime (create,
// each join, each deck upload, start). Deferring it would mean
// threading a thunk through four call sites to save microseconds on
// a cold path.
//
// Failure is logged, never returned: losing the metadata costs a
// rewind on the next deploy, while failing the player's join costs
// them the game.
func (l *Lobby) persistMetaLocked(entry *gameEntry) {
	dir := l.dumpDir()
	if dir == "" || entry == nil {
		return
	}
	payload, err := json.Marshal(entry.meta)
	if err != nil {
		slog.Default().Warn("marshal lobby metadata failed", "game_id", entry.meta.ID, "err", err)
		return
	}
	if err := writeFileAtomic0600(metaPath(dir, entry.meta.ID), payload); err != nil {
		slog.Default().Warn("persist lobby metadata failed", "game_id", entry.meta.ID, "err", err)
	}
}

// removeMeta deletes a game's metadata file. Called from Delete, so a
// removed game does not reappear at the next boot.
func (l *Lobby) removeMeta(id uuid.UUID) {
	dir := l.dumpDir()
	if dir == "" {
		return
	}
	if err := os.Remove(metaPath(dir, id)); err != nil && !os.IsNotExist(err) {
		slog.Default().Warn("removing lobby metadata failed", "game_id", id, "err", err)
	}
}

// writeFileAtomic0600 publishes payload via CreateTemp + rename with
// owner-only permissions — the file holds invite tokens.
func writeFileAtomic0600(path string, payload []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
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
	if err := tmp.Chmod(0o600); err != nil {
		return err
	}
	if _, err := tmp.Write(payload); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	tmp = nil
	return os.Rename(tmpPath, path)
}

// RestoreFromDisk rebuilds the lobby from the artifacts on disk. Call
// it once at boot, before the server starts serving.
//
// The engine snapshots are the source of truth for WHICH games come
// back: RoomManager.RestoreRooms decides that, applying the
// version-skew policy (see ws/persist.go). This function then pairs
// each restored room with its metadata. A room whose metadata is
// missing or unreadable is dropped from the lobby and unregistered —
// a game nobody can produce an invite for is not a game, and leaving
// it registered would make it visible to admins while being unjoinable
// by anyone else.
//
// Returns the number of games that came back.
func (l *Lobby) RestoreFromDisk(log *slog.Logger) int {
	if log == nil {
		log = slog.Default()
	}
	dir := l.dumpDir()
	if dir == "" {
		return 0
	}

	outcomes := l.mgr.RestoreRooms()
	ws.LogRestoreSummary(log, outcomes)

	restored := 0
	for _, o := range outcomes {
		if !o.Restored() {
			continue
		}
		meta, err := readMeta(dir, o.GameID)
		if err != nil {
			log.Error("restored game has no usable lobby metadata; dropping it (nobody could be invited back in)",
				"game_id", o.GameID, "err", err)
			l.mgr.Delete(o.GameID)
			continue
		}
		// The engine is authoritative for lifecycle state; the meta
		// copy is a cache for the listing and can be stale if the
		// process died between the two writes.
		meta.State = string(o.Room.Game.CurrentState())

		entry := &gameEntry{meta: meta, room: o.Room}
		l.mu.Lock()
		l.games[o.GameID] = entry
		// Bot seats do not come back on their own. The seat itself is
		// carried — Player.IsBot / BotTier ride the engine snapshot
		// and SeatInfo.IsBot rides the metadata written above — but
		// the runner was a goroutine in the process that died, so an
		// active game with a bot seat would otherwise resume with the
		// bot's chair occupied and nobody in it, and the table would
		// hang the first time priority reached it. Relaunch here,
		// under the same guard Start uses.
		var startBots func()
		if o.Room.Game.CurrentState() == game.StateActive {
			startBots = l.botStartLocked(entry)
		}
		l.mu.Unlock()
		if startBots != nil {
			startBots()
			log.Info("relaunched bot runners for a restored game", "game_id", o.GameID)
		}
		restored++
	}

	// Metadata for games whose engine snapshot did not come back is
	// dead weight; drop it so the directory does not grow forever.
	l.pruneOrphanMeta(dir, log)

	if restored > 0 {
		log.Info("lobby restored from disk", "games", restored)
	}
	return restored
}

func readMeta(dumpDir string, id uuid.UUID) (GameMeta, error) {
	raw, err := os.ReadFile(metaPath(dumpDir, id))
	if err != nil {
		return GameMeta{}, err
	}
	var m GameMeta
	if err := json.Unmarshal(raw, &m); err != nil {
		return GameMeta{}, err
	}
	m.ID = id
	if m.Players == nil {
		m.Players = []SeatInfo{}
	}
	return m, nil
}

// pruneOrphanMeta removes metadata files with no live game behind
// them.
func (l *Lobby) pruneOrphanMeta(dumpDir string, log *slog.Logger) {
	entries, err := os.ReadDir(filepath.Join(dumpDir, "lobby"))
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".json") {
			continue
		}
		id, err := uuid.Parse(strings.TrimSuffix(name, ".json"))
		if err != nil {
			continue
		}
		l.mu.Lock()
		_, live := l.games[id]
		l.mu.Unlock()
		if live {
			continue
		}
		if err := os.Remove(metaPath(dumpDir, id)); err != nil && !os.IsNotExist(err) {
			log.Warn("pruning orphan lobby metadata failed", "game_id", id, "err", err)
		}
	}
}
