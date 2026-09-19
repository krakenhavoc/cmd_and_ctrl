package lobby

// persist.go keeps the lobby's half of a game where a restart can find
// it, so that a restored game.Game is actually reachable.
//
// The engine snapshot (internal/game/snapshot.go) rebuilds the table.
// It does not rebuild the INVITE, and without one a restored game is a
// room nobody can open. Signed sessions survive a restart
// (auth.HMACAuthenticator, #517), but a player whose token is gone —
// cleared storage, a new device, an expired TTL, or a server running
// without CMDCTRL_SESSION_KEY — re-authenticates through the invite
// link, so that link has to keep working.
//
// Until S34 that half lived in <dumpDir>/lobby/<id>.json (ADR 0041).
// ADR 0051 decision 4 moved it into the games / seats / invites tables
// behind the Store seam (store.go); the engine's own files —
// restore/, replays/, games/ — stay on disk and are not touched here.
//
// What the rows do NOT hold is the invite plaintext. invites keeps the
// SHA-256 of each token. The plaintext lives in the gameEntry of the
// process that minted it, so the create response and GET /games/{id}
// (admin, seated players) can still show the link. After a restart
// only the hash is left: every link keeps working, and the tokens are
// no longer displayable (docs/lobby.md).
//
// Write failures are logged, never returned, except at Create: losing
// a seat row costs a stale seat list after the next deploy, while
// failing the player's join costs them the game. Create is different
// because an invite that never reached the store cannot be redeemed at
// all, even in the same process.

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// storeTimeout bounds one store call. The store is a local SQLite file
// (or a map), so this only matters if the disk has stopped answering.
const storeTimeout = 5 * time.Second

func storeCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), storeTimeout)
}

// legacyMetaPath is where ADR 0041 kept one game's lobby metadata.
// Read only by the importer (import.go) and reaped by Delete.
func legacyMetaPath(dumpDir string, id uuid.UUID) string {
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

// seatRecords projects the seat list onto seats rows.
func seatRecords(seats []SeatInfo) []SeatRecord {
	out := make([]SeatRecord, 0, len(seats))
	for _, s := range seats {
		rec := SeatRecord{
			Seat:             s.Seat,
			PlayerID:         s.PlayerID,
			GuestName:        s.Name,
			DeckName:         s.DeckName,
			PendingDiscordID: s.DiscordID,
		}
		if s.IsBot {
			rec.BotTier = s.BotTier
		}
		out = append(out, rec)
	}
	return out
}

// gameRecordLocked projects an entry onto its games row. Callers hold
// l.mu.
func gameRecordLocked(entry *gameEntry) GameRecord {
	rec := GameRecord{
		ID:         entry.meta.ID,
		Name:       entry.meta.Name,
		CreatedBy:  entry.createdBy,
		State:      entry.meta.State,
		CreatedAt:  entry.meta.CreatedAt,
		StartedAt:  entry.startedAt,
		EndedAt:    entry.endedAt,
		ArchivedAt: entry.meta.ArchivedAt,
		WinnerSeat: entry.winnerSeat,
		// ADR 0075 §2.1. syncHostLocked keeps HostPlayerID equal to the
		// room's effective host before any write that matters.
		HostPlayerID:  entry.meta.HostPlayerID,
		HostDiscordID: entry.meta.HostDiscordID,
	}
	return rec
}

// persistSeatsLocked writes the entry's seat list. Called from the
// mutations that change it: Join, SetDeck, AddBot, RemoveBot.
//
// It runs under l.mu, like the file write it replaces: a handful of
// small writes over a game's whole lifetime, and deferring it would
// mean threading a thunk through every call site for no gain.
func (l *Lobby) persistSeatsLocked(entry *gameEntry) {
	ctx, cancel := storeCtx()
	defer cancel()
	if err := l.store.ReplaceSeats(ctx, entry.meta.ID, seatRecords(entry.meta.Players)); err != nil {
		slog.Default().Warn("persist lobby seats failed", "game_id", entry.meta.ID, "err", err)
	}
}

// persistGameLocked writes the entry's games row. Called when the
// lifecycle columns change: Start, archive/unarchive, and whenever
// syncStateLocked observes a transition.
func (l *Lobby) persistGameLocked(entry *gameEntry) {
	ctx, cancel := storeCtx()
	defer cancel()
	if err := l.store.UpdateGame(ctx, gameRecordLocked(entry)); err != nil {
		slog.Default().Warn("persist lobby game failed", "game_id", entry.meta.ID, "err", err)
	}
}

// persistGameLocked is also how the host columns (ADR 0075 §2.1) are
// written: by syncHostLocked when hosting moves, and by Join /
// TransferHost when a pending named host is spent.
//
// syncStateLocked copies the engine's lifecycle state onto the entry
// and, when it moved, onto the games row: started_at on the first
// sight of active, ended_at and winner_seat on the first sight of
// ended. The engine is authoritative; meta.State is a cache of it.
//
// The lobby sees a transition wherever it already reads the live
// state — Start, Get, List, restore — and through watchEnd, which is
// how a game that ends over the WebSocket (concede, a state-based
// loss) gets its ended_at without anybody opening the lobby.
func (l *Lobby) syncStateLocked(entry *gameEntry) {
	live := entry.room.Game.CurrentState()
	if string(live) == entry.meta.State && (live != game.StateEnded || entry.endedAt != nil) {
		return
	}
	now := time.Now().UTC()
	entry.meta.State = string(live)
	if (live == game.StateActive || live == game.StateEnded) && entry.startedAt == nil && entry.startedKnown {
		entry.startedAt = &now
	}
	if live == game.StateEnded && entry.endedAt == nil {
		entry.endedAt = &now
		if seat, ok := entry.room.Game.WinnerSeat(); ok {
			entry.winnerSeat = &seat
		}
	}
	l.persistGameLocked(entry)
}

// watchEnd follows one active room until its game ends, then records
// the end. It rides Room.Subscribe, the wake the bot runners already
// use, and exits when the game ends or the entry is dropped (stop is
// closed). Started for every game that becomes active in this process
// or comes back active from a restore point.
func (l *Lobby) watchEnd(entry *gameEntry) {
	wake, unsubscribe := entry.room.Subscribe()
	defer unsubscribe()
	for {
		if entry.room.Game.CurrentState() == game.StateEnded {
			l.mu.Lock()
			if l.games[entry.meta.ID] == entry {
				l.syncStateLocked(entry)
			}
			l.mu.Unlock()
			return
		}
		select {
		case <-wake:
		case <-entry.stop:
			return
		}
	}
}

// startWatchLocked launches watchEnd once per entry. Callers hold l.mu.
func (l *Lobby) startWatchLocked(entry *gameEntry) {
	if entry.watching {
		return
	}
	entry.watching = true
	go l.watchEnd(entry)
}

// dropEntryLocked removes an entry from the registry and stops its
// watcher. Callers hold l.mu.
func (l *Lobby) dropEntryLocked(id uuid.UUID) {
	if e, ok := l.games[id]; ok {
		close(e.stop)
		delete(l.games, id)
	}
}

// removeLegacyMeta reaps a deleted game's ADR 0041 file and the
// importer's renamed copy of it. Only Delete does this; the importer
// itself never removes a file.
func (l *Lobby) removeLegacyMeta(id uuid.UUID) {
	dir := l.dumpDir()
	if dir == "" {
		return
	}
	base := legacyMetaPath(dir, id)
	for _, p := range []string{base, base + importedSuffix} {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			slog.Default().Warn("removing legacy lobby metadata failed", "path", p, "err", err)
		}
	}
}

// legacyImporter is implemented by a Store that can take in ADR 0041's
// lobby/*.json files. SQLStore does; the memory store has nowhere to
// put them.
type legacyImporter interface {
	ImportLegacyLobby(ctx context.Context, dumpDir string, log *slog.Logger) (ImportResult, error)
}

// RestoreFromDisk rebuilds the lobby from what the last process left
// behind. Call it once at boot, before the server starts serving.
//
// Order:
//
//  1. Import any ADR 0041 lobby/*.json into the store (import.go), so
//     the first boot of this binary finds metadata for games the old
//     one was running.
//  2. RoomManager.RestoreRooms decides WHICH games come back,
//     applying the version-skew policy (see ws/persist.go).
//  3. Each restored room is paired with its games and seats rows. A
//     room with no row is dropped and unregistered — a game nobody
//     can produce an invite for is not a game.
//
// Rows whose game did not come back are kept. A finished game has no
// restore point by design (ws.Room drops it at the end), and its row
// is the history "my games" reads (ADR 0051 decision 4). ADR 0041
// pruned the equivalent files; that is the one behaviour this does not
// carry over.
//
// A store that is not durable has no metadata to pair with, so the
// restore is skipped and the restore points are left untouched on
// disk for a boot that has one.
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
	if !l.store.Durable() {
		log.Warn("lobby has no durable store; not restoring games (restore points left on disk)")
		return 0
	}
	if imp, ok := l.store.(legacyImporter); ok {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		res, err := imp.ImportLegacyLobby(ctx, dir, log)
		cancel()
		if err != nil {
			// Carrying on would drop every restored room whose
			// metadata is still in an unimported file. Skip the
			// restore instead: nothing is deleted, and the next boot
			// retries the import.
			log.Error("importing lobby/*.json failed; not restoring games this boot (nothing was deleted)", "err", err)
			return 0
		}
		if res.Imported > 0 || res.AlreadyPresent > 0 || res.Skipped > 0 {
			log.Info("imported legacy lobby metadata into the database",
				"imported", res.Imported, "already_present", res.AlreadyPresent, "skipped", res.Skipped)
		}
	}

	outcomes := l.mgr.RestoreRooms()
	ws.LogRestoreSummary(log, outcomes)

	restored := 0
	for _, o := range outcomes {
		if !o.Restored() {
			continue
		}
		entry, err := l.loadEntry(o.GameID, o.Room)
		if err != nil {
			log.Error("restored game has no usable lobby metadata; dropping it (nobody could be invited back in)",
				"game_id", o.GameID, "err", err)
			l.mgr.Delete(o.GameID)
			continue
		}

		// The host rides the games row, not the engine snapshot; hand it
		// back to the room so is_host and the host gates survive the
		// restart (ADR 0075 §2.1).
		o.Room.SetHost(entry.meta.HostPlayerID)
		l.mu.Lock()
		l.games[o.GameID] = entry
		// The engine is authoritative for lifecycle state; the row can
		// be stale if the process died between the two writes.
		l.syncStateLocked(entry)
		l.syncHostLocked(entry)
		// Bot seats do not come back on their own. The seat itself is
		// carried — Player.IsBot / BotTier ride the engine snapshot
		// and seats.bot_tier rides the row — but the runner was a
		// goroutine in the process that died, so an active game with a
		// bot seat would otherwise resume with the bot's chair occupied
		// and nobody in it, and the table would hang the first time
		// priority reached it. Relaunch here, under the same guard
		// Start uses.
		// An archived table is retired: it is hidden from the listing
		// and SetArchived stopped its runners on the way out. Booting
		// them again here would resurrect exactly the goroutines
		// archiving exists to stop.
		var startBots func()
		if o.Room.Game.CurrentState() == game.StateActive {
			l.startWatchLocked(entry)
			if !entry.meta.Archived() {
				startBots = l.botStartLocked(entry)
			}
		}
		l.mu.Unlock()
		if startBots != nil {
			startBots()
			log.Info("relaunched bot runners for a restored game", "game_id", o.GameID)
		}
		restored++
	}

	if restored > 0 {
		log.Info("lobby restored from disk", "games", restored)
	}
	return restored
}

// loadEntry rebuilds one restored game's lobby entry from its rows and
// its engine players.
//
// The rows hold what ADR 0051 decision 4 gives a column. The rest of
// SeatInfo — whether a real deck is loaded, the Discord avatar and
// display name, the curated bot deck — rides the engine snapshot on
// game.Player already, so it is read back from there rather than
// duplicated into seats.
func (l *Lobby) loadEntry(id uuid.UUID, room *ws.Room) (*gameEntry, error) {
	ctx, cancel := storeCtx()
	defer cancel()
	rec, seats, err := l.store.LoadGame(ctx, id)
	if err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			return nil, errors.New("no games row")
		}
		return nil, err
	}
	meta := GameMeta{
		ID:         id,
		Name:       rec.Name,
		CreatedAt:  rec.CreatedAt,
		Players:    make([]SeatInfo, 0, len(seats)),
		State:      rec.State,
		ArchivedAt: rec.ArchivedAt,
		// The host (ADR 0075 §2.1). RestoreFromDisk hands it back to
		// the room, which decides whether it still stands.
		HostPlayerID:  rec.HostPlayerID,
		HostDiscordID: rec.HostDiscordID,
	}
	for _, s := range seats {
		info := SeatInfo{
			PlayerID:  s.PlayerID,
			Name:      s.GuestName,
			Seat:      s.Seat,
			DeckName:  s.DeckName,
			DiscordID: s.PendingDiscordID,
			IsBot:     s.BotTier != "",
			BotTier:   s.BotTier,
		}
		if p := room.Game.PlayerByID(s.PlayerID); p != nil {
			info.DeckUploaded = p.DeckImported
			if p.DiscordID != "" {
				info.DiscordID = p.DiscordID
			}
			info.DiscordAvatarHash = p.DiscordAvatarHash
			info.DisplayName = p.DisplayName
			if p.IsBot {
				info.IsBot = true
				info.BotTier = p.BotTier
				info.BotDeck = p.BotDeck
			}
		}
		meta.Players = append(meta.Players, info)
	}
	return &gameEntry{
		meta:       meta,
		room:       room,
		stop:       make(chan struct{}),
		createdBy:  rec.CreatedBy,
		startedAt:  rec.StartedAt,
		endedAt:    rec.EndedAt,
		winnerSeat: rec.WinnerSeat,
		// A game imported from lobby/*.json, or one that was already
		// running when this binary first booted, has no start time on
		// record. Stamping "now" on it would be a made-up date.
		startedKnown: rec.State == string(game.StateLobby),
	}, nil
}
