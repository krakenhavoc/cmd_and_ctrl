package ws

// scrub.go removes a person's identity from a finished game's files on
// disk (account deletion, ADR 0149, #2983).
//
// A game leaves three files under the data root: games/<id>.json (the
// last view, written for crash forensics), replays/<id>.jsonl (one full
// unfiltered view per commit) and restore/<id>.json (the engine
// snapshot, kept for EndedTableRetention). All three are JSON, and all
// three name a seat by its player id beside its name and Discord
// metadata. The game log inside them is structured (seat indexes, card
// ids), not rendered text, so a person's name appears only on the
// player objects and a few annotations that repeat it; there is no
// free text to search.
//
// The rewrite is therefore structural, not textual: find every object
// that is about one of the given players and blank the identity fields
// on it. Nothing else in the file changes, so a replay still reads, and
// a restore point still restores, with the same game in it.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// identityKeys are the keys on a player object that identify a person
// beyond the seat itself. The snake_case spellings are the wire view's;
// the camelCase ones are the engine snapshot's (game.PlayerSnapshot).
var identityKeys = []string{
	"discord_id", "discord_avatar_hash", "display_name", "playmat_url", "playmat_wash",
	"discordId", "discordAvatarHash", "displayName",
}

// scrubPlayers rewrites v in place so that every object about one of
// ids loses its identity, and reports whether it changed anything.
//
// An object is "about" a player when its id, seat, target_id or
// source_id is that player's id (the first is a player object, the
// others are annotations and shield entries that repeat the name beside
// the id).
func scrubPlayers(v any, ids map[string]bool, name string) bool {
	changed := false
	switch x := v.(type) {
	case map[string]any:
		set := func(idKey, nameKey string) {
			id, _ := x[idKey].(string)
			if !ids[id] {
				return
			}
			if cur, ok := x[nameKey]; ok && cur != name {
				x[nameKey] = name
				changed = true
			}
		}
		set("id", "name")
		set("seat", "seat_name")
		set("target_id", "target_name")
		set("source_id", "source_name")
		if id, _ := x["id"].(string); ids[id] {
			for _, k := range identityKeys {
				if _, ok := x[k]; ok {
					delete(x, k)
					changed = true
				}
			}
		}
		for _, c := range x {
			if scrubPlayers(c, ids, name) {
				changed = true
			}
		}
	case []any:
		for _, c := range x {
			if scrubPlayers(c, ids, name) {
				changed = true
			}
		}
	}
	return changed
}

// scrubLine decodes one JSON document, scrubs it and encodes it again.
// A document that does not change is returned byte for byte, so a file
// with nothing to scrub is not rewritten at all.
func scrubLine(line []byte, ids map[string]bool, name string) ([]byte, bool, error) {
	dec := json.NewDecoder(bytes.NewReader(line))
	dec.UseNumber() // keep 64-bit ids and counters exact
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, false, err
	}
	if !scrubPlayers(v, ids, name) {
		return line, false, nil
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, false, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), true, nil
}

// scrubFile rewrites one file line by line, atomically, and does
// nothing if the file is absent or no line changes. A line that is not
// JSON is an error: the file is left as it was.
func scrubFile(path string, ids map[string]bool, name string) error {
	src, err := os.Open(path) //nolint:gosec // path is built from a uuid and a fixed layout
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer func() { _ = src.Close() }()

	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	done := false
	defer func() {
		if !done {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	w := bufio.NewWriter(tmp)
	r := bufio.NewReaderSize(src, 1<<20)
	rewrote := false
	for {
		line, rerr := r.ReadBytes('\n')
		hadNL := bytes.HasSuffix(line, []byte("\n"))
		body := bytes.TrimRight(line, "\r\n")
		if len(bytes.TrimSpace(body)) > 0 {
			out, changed, serr := scrubLine(body, ids, name)
			if serr != nil {
				return fmt.Errorf("scrub %s: %w", filepath.Base(path), serr)
			}
			if changed {
				rewrote = true
			}
			body = out
		}
		if _, werr := w.Write(body); werr != nil {
			return werr
		}
		if hadNL {
			if _, werr := w.WriteString("\n"); werr != nil {
				return werr
			}
		}
		if errors.Is(rerr, io.EOF) {
			break
		}
		if rerr != nil {
			return rerr
		}
	}
	if !rewrote {
		return nil // nothing to scrub; leave the file alone
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if info, serr := os.Stat(path); serr == nil {
		_ = os.Chmod(tmpPath, info.Mode().Perm())
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("publish %s: %w", path, err)
	}
	done = true
	return nil
}

// ScrubPlayerFiles removes the identity of the given players from a
// game's files on disk (the crash dump, the replay and the restore
// point) and replaces their name with name. A file that does not exist
// is fine; the first real error is returned after every file has been
// tried. With no data directory it does nothing.
//
// A resident room is locked while its files are rewritten, so a commit
// cannot append to the replay in the middle of it. The caller scrubs
// the live game first (Game.AnonymisePlayer through ApplyExternal) so
// the lines that commit writes are already anonymous.
func (m *RoomManager) ScrubPlayerFiles(gameID uuid.UUID, players []uuid.UUID, name string) error {
	dir := m.DumpDir()
	if dir == "" || len(players) == 0 {
		return nil
	}
	ids := make(map[string]bool, len(players))
	for _, p := range players {
		ids[p.String()] = true
	}
	if room := m.Get(gameID); room != nil {
		room.mu.Lock()
		defer room.mu.Unlock()
	}
	var first error
	for _, path := range []string{
		snapshotPath(dir, gameID),
		replayPath(dir, gameID),
		restorePointPath(dir, gameID),
	} {
		if err := scrubFile(path, ids, name); err != nil && first == nil {
			first = err
		}
	}
	return first
}
