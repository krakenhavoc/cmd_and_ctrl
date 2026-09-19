package lobby

// import.go moves ADR 0041's <dumpDir>/lobby/<id>.json files into the
// games / seats / invites tables, once, at boot (ADR 0051 "Migration").
//
//   - Each file becomes a games row, its seat list becomes seats rows,
//     and its two tokens become invites rows, hashed. A seat claimed
//     through Discord gets user_id NULL and pending_discord_id set,
//     because no users row exists yet.
//   - The whole import is one transaction. A file that cannot be read,
//     parsed or inserted is logged and skipped (a savepoint rolls back
//     just that file), and it is NOT renamed, so a fixed binary or an
//     operator can retry it.
//   - After the commit each imported file is renamed to
//     <id>.json.imported. The importer never deletes one: a rolled-back
//     binary reads lobby/*.json, and the rollback one-liner in
//     docs/environments.md renames them back.
//
// Idempotent. A file whose game already has a row is not inserted
// again — the row wins — and is renamed, which finishes an import
// whose renames were interrupted. So a re-run with nothing new is a
// directory listing and no writes.

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// importedSuffix is appended to a lobby/<id>.json once its content is
// in the database.
const importedSuffix = ".imported"

// ImportResult counts what one import pass did with lobby/*.json.
type ImportResult struct {
	Imported       int // new games rows written
	AlreadyPresent int // a row already existed; file renamed, row untouched
	Skipped        int // unreadable or unusable; left in place, not renamed
}

// legacyFile is one lobby/<id>.json, parsed and validated.
type legacyFile struct {
	path    string
	game    GameRecord
	seats   []SeatRecord
	invites []InviteRecord
}

// ImportLegacyLobby imports <dumpDir>/lobby/*.json. See the file
// comment. An error means nothing was committed and nothing renamed.
func (s *SQLStore) ImportLegacyLobby(ctx context.Context, dumpDir string, log *slog.Logger) (ImportResult, error) {
	var res ImportResult
	if log == nil {
		log = slog.Default()
	}
	dir := filepath.Join(dumpDir, "lobby")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return res, nil
		}
		return res, fmt.Errorf("read %s: %w", dir, err)
	}

	var files []legacyFile
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".json") {
			continue // .json.imported, .tmp leftovers, anything else
		}
		path := filepath.Join(dir, name)
		f, err := parseLegacyFile(path, strings.TrimSuffix(name, ".json"))
		if err != nil {
			log.Warn("skipping unreadable lobby metadata file; left in place, not imported", "path", path, "err", err)
			res.Skipped++
			continue
		}
		files = append(files, f)
	}
	if len(files) == 0 {
		return res, nil
	}

	var toRename []string
	err = s.withTx(ctx, func(tx *sql.Tx) error {
		for _, f := range files {
			exists, err := gameExists(ctx, tx, f.game.ID)
			if err != nil {
				return err
			}
			if exists {
				log.Warn("lobby metadata file is for a game already in the database; keeping the row, renaming the file",
					"path", f.path, "game_id", f.game.ID)
				res.AlreadyPresent++
				toRename = append(toRename, f.path)
				continue
			}
			if _, err := tx.ExecContext(ctx, `SAVEPOINT import_one`); err != nil {
				return err
			}
			if err := insertLegacy(ctx, tx, f); err != nil {
				if _, rbErr := tx.ExecContext(ctx, `ROLLBACK TO import_one`); rbErr != nil {
					return rbErr
				}
				if _, relErr := tx.ExecContext(ctx, `RELEASE import_one`); relErr != nil {
					return relErr
				}
				log.Warn("skipping lobby metadata file the database refused; left in place, not imported",
					"path", f.path, "game_id", f.game.ID, "err", err)
				res.Skipped++
				continue
			}
			if _, err := tx.ExecContext(ctx, `RELEASE import_one`); err != nil {
				return err
			}
			res.Imported++
			toRename = append(toRename, f.path)
		}
		return nil
	})
	if err != nil {
		return ImportResult{}, err
	}

	for _, p := range toRename {
		if err := os.Rename(p, p+importedSuffix); err != nil {
			// The rows are committed; the next boot sees the game
			// already present and retries just the rename.
			log.Warn("renaming imported lobby metadata file failed", "path", p, "err", err)
		}
	}
	return res, nil
}

func insertLegacy(ctx context.Context, tx *sql.Tx, f legacyFile) error {
	if err := insertGame(ctx, tx, f.game); err != nil {
		return fmt.Errorf("insert game: %w", err)
	}
	if err := insertSeats(ctx, tx, f.game.ID, f.seats); err != nil {
		return fmt.Errorf("insert seats: %w", err)
	}
	if err := insertInvites(ctx, tx, f.invites); err != nil {
		return fmt.Errorf("insert invites: %w", err)
	}
	return nil
}

// parseLegacyFile reads one ADR 0041 file into rows. The file name is
// the game ID, as readMeta treated it.
func parseLegacyFile(path, idStr string) (legacyFile, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return legacyFile{}, fmt.Errorf("file name is not a game id: %w", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return legacyFile{}, err
	}
	var m GameMeta
	if err := json.Unmarshal(raw, &m); err != nil {
		return legacyFile{}, fmt.Errorf("parse: %w", err)
	}
	playerHash, ok := hashInvite(m.InviteToken)
	if !ok {
		// Without the player invite the game cannot be joined, which
		// is the whole reason the file existed.
		return legacyFile{}, fmt.Errorf("no usable player invite token")
	}
	created := m.CreatedAt.UTC()
	if m.CreatedAt.IsZero() {
		created = time.Now().UTC()
	}
	state := m.State
	if state == "" {
		state = "lobby"
	}
	f := legacyFile{
		path: path,
		game: GameRecord{
			ID:         id,
			Name:       m.Name,
			State:      state,
			CreatedAt:  created,
			ArchivedAt: m.ArchivedAt,
		},
		seats: seatRecords(m.Players),
		invites: []InviteRecord{{
			Hash: playerHash, GameID: id, Kind: InvitePlayer, CreatedAt: created,
		}},
	}
	// A file from before S11 has no spectator invite; that game simply
	// has none, as it did before.
	if m.SpectatorInvite != "" {
		specHash, ok := hashInvite(m.SpectatorInvite)
		if !ok {
			return legacyFile{}, fmt.Errorf("spectator invite token is malformed")
		}
		f.invites = append(f.invites, InviteRecord{
			Hash: specHash, GameID: id, Kind: InviteSpectator, CreatedAt: created,
		})
	}
	return f, nil
}
