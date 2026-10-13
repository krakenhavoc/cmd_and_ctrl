package lobby

// blocks.go: "remove and block" (#2984, ADR 0075 section 2.6).
//
// Removing a seat (seat_remove.go) takes a person off a table. It does
// not stop them sitting down again: the invite link they already hold
// still works on an unstarted table. A block closes that, for that one
// table, without rotating the invite for everyone else.
//
// WHAT A BLOCK KEYS ON, and so what it can and cannot hold against:
//
//   - a signed-in seat is blocked by its ACCOUNT (users id) and by its
//     Discord id. Signing in again as the same person is refused, from
//     any browser. A second Discord account is a different person to us.
//   - a guest seat is blocked by its NAME, folded (moderation.Key), so
//     "Bob", "b0b" and "B O B" are one guest. A guest has no account, no
//     stable id and no address we keep; the session behind a guest seat
//     dies with the seat. So a blocked guest who types a DIFFERENT name
//     gets in, and so does a blocked signed-in person who signs out and
//     joins as a guest under another name. A block is a speed bump for
//     those, not a wall. The wall is rotating the invite
//     (POST /games/{id}/invites/rotate), and the Remove control says so.
//   - a name block never touches a signed-in joiner: two people called
//     "Alex" on Discord are not one person, and the account is the
//     identity there.
//
// Blocks are per table, persisted with it (table_blocks, migration 0016)
// and read back when the table is restored, so a restart does not lift
// them. They leave with the table. There is no unblock: a block is made
// by an act the host chose, at one table, and a mistake costs a new
// invite or a new table, never a ban from the site.

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/moderation"
)

// ErrBlockedFromTable is what a blocked joiner is told. It does not say
// who blocked them or which of their identities matched.
var ErrBlockedFromTable = errors.New("you were removed from this table and can't take a seat here again")

// The stable `code` of the JSON error body for the two refusals a client
// shows in its own words (docs/lobby.md, "Moderation").
const (
	CodeBlockedFromTable = "blocked_from_table"
	CodeNameNotAllowed   = "name_not_allowed"
)

// blockKey is the in-memory form of a BlockRecord.
type blockKey struct {
	kind    BlockKind
	subject string
}

// blockKeysOf lists the identities a seat can be blocked by.
func blockKeysOf(s SeatInfo) []blockKey {
	var keys []blockKey
	if s.UserID != "" {
		keys = append(keys, blockKey{BlockUser, s.UserID})
	}
	if s.DiscordID != "" {
		keys = append(keys, blockKey{BlockDiscord, s.DiscordID})
	}
	if s.UserID == "" && s.DiscordID == "" && !s.IsBot {
		if k := moderation.Key(s.Name); k != "" {
			keys = append(keys, blockKey{BlockName, k})
		}
	}
	return keys
}

// blockedLocked reports whether a joiner is blocked at this table. userID
// is uuid.Nil for a guest and discordID is "" without a Discord identity;
// name is the seat label they are asking for.
func (e *gameEntry) blockedLocked(userID uuid.UUID, discordID, name string) bool {
	if len(e.blocks) == 0 {
		return false
	}
	if userID != uuid.Nil {
		if _, ok := e.blocks[blockKey{BlockUser, userID.String()}]; ok {
			return true
		}
	}
	if discordID != "" {
		if _, ok := e.blocks[blockKey{BlockDiscord, discordID}]; ok {
			return true
		}
	}
	if userID == uuid.Nil && discordID == "" {
		if _, ok := e.blocks[blockKey{BlockName, moderation.Key(name)}]; ok {
			return true
		}
	}
	return false
}

// addBlocksLocked remembers keys on the entry and writes them. A failed
// write is logged and the block still holds for the life of the process:
// the host asked for it, and refusing the removal over a database error
// would leave the person seated.
func (l *Lobby) addBlocksLocked(entry *gameEntry, keys []blockKey) {
	if len(keys) == 0 {
		return
	}
	if entry.blocks == nil {
		entry.blocks = make(map[blockKey]struct{}, len(keys))
	}
	now := time.Now().UTC()
	for _, k := range keys {
		entry.blocks[k] = struct{}{}
		if entry.practice {
			continue
		}
		ctx, cancel := storeCtx()
		err := l.store.AddBlock(ctx, entry.meta.ID, BlockRecord{Kind: k.kind, Subject: k.subject, CreatedAt: now})
		cancel()
		if err != nil {
			slog.Default().Warn("persist table block failed", "game_id", entry.meta.ID, "kind", k.kind, "err", err)
		}
	}
}

// loadBlocks reads a table's blocks back at restore.
func (l *Lobby) loadBlocks(id uuid.UUID) map[blockKey]struct{} {
	ctx, cancel := storeCtx()
	defer cancel()
	recs, err := l.store.Blocks(ctx, id)
	if err != nil {
		slog.Default().Warn("load table blocks failed", "game_id", id, "err", err)
		return nil
	}
	if len(recs) == 0 {
		return nil
	}
	out := make(map[blockKey]struct{}, len(recs))
	for _, r := range recs {
		out[blockKey{r.Kind, r.Subject}] = struct{}{}
	}
	return out
}

// BlockCount is how many identities the table has blocked, for tests and
// the admin views. It names none of them.
func (l *Lobby) BlockCount(id uuid.UUID) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	if e, ok := l.games[id]; ok {
		return len(e.blocks)
	}
	return 0
}

// cleanDiscordIdentity makes a Discord identity safe to show as a seat
// name (the name filter, #2984). The global name is shown by preference;
// if it is refused the username stands in, and if that is refused too
// the join is.
func cleanDiscordIdentity(d DiscordIdentity) (DiscordIdentity, error) {
	if !d.Populated() {
		return d, nil
	}
	if strings.TrimSpace(d.GlobalName) != "" && moderation.Name(d.GlobalName) != nil {
		d.GlobalName = ""
	}
	if moderation.Name(d.DisplayName()) != nil {
		return d, ErrDiscordNameNotAllowed
	}
	return d, nil
}

// ErrDiscordNameNotAllowed is the join refusal for a Discord account
// whose names are all on the filter's list.
var ErrDiscordNameNotAllowed = errors.New("your Discord display name isn't allowed here; change it on Discord and sign in again")

// forgetBlocksOf drops, from every resident table, the blocks keyed on a
// deleted account's id and Discord id, to match the rows the deletion
// removed. Without it the in-memory copy would go on refusing a Discord
// id that the person may now sign in with as a new account, until the
// next restart quietly lifted it.
func (l *Lobby) forgetBlocksOf(userID uuid.UUID, discordID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, e := range l.games {
		delete(e.blocks, blockKey{BlockUser, userID.String()})
		if discordID != "" {
			delete(e.blocks, blockKey{BlockDiscord, discordID})
		}
	}
}
