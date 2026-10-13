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
// Blocks are per table, persisted with it (table_blocks, migrations 0016
// and 0017) and read back when the table is restored, so a restart does
// not lift them. They leave with the table.
//
// THE LIST AND THE UNBLOCK (#3113). A block is one ENTRY: the rows one
// removal wrote (an account and its Discord id are two rows, one entry),
// shown by a label (an account's display name, a guest's folded name,
// never a Discord id). The host and the creator may list and lift the
// entries of a table that has not started; an admin may on any table.
// That is the same authority split as removing a seat (seat_remove.go).
// A mistake costs one click, not a new invite.

import (
	"cmp"
	"errors"
	"log/slog"
	"sort"
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

// blockInfo is what a blockKey carries beside its identity: the entry it
// belongs to, how the list names it, and when it was made.
type blockInfo struct {
	entry string
	label string
	at    time.Time
}

// BlockEntry kinds, as the list reports them.
const (
	BlockEntryAccount = "account"
	BlockEntryGuest   = "guest"
)

// BlockEntry is one line of a table's block list: a person the table
// will not seat again.
type BlockEntry struct {
	// ID addresses the entry for the unblock.
	ID string `json:"id"`
	// Label is the account's display name, or the guest's folded name.
	// Never a Discord id.
	Label string `json:"label"`
	// Kind is "account" or "guest", so the list can say what the block
	// holds against and what it does not.
	Kind      string `json:"kind"`
	CreatedAt int64  `json:"created_at"` // Unix milliseconds
}

// blockLabelOf is how the block list names a seat: a signed-in person by
// the name the table knew them by, a guest by the folded name the block
// actually matches on. A Discord id is never a label.
func blockLabelOf(s SeatInfo) string {
	if s.UserID != "" || s.DiscordID != "" {
		return cmp.Or(s.DisplayName, s.Name)
	}
	return moderation.Key(s.Name)
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

// addBlocksLocked remembers keys on the entry as ONE block entry named
// label, and writes them. A failed write is logged and the block still holds
// for the life of the process: the host asked for it, and refusing the
// removal over a database error would leave the person seated.
func (l *Lobby) addBlocksLocked(entry *gameEntry, keys []blockKey, label string) {
	if len(keys) == 0 {
		return
	}
	if entry.blocks == nil {
		entry.blocks = make(map[blockKey]blockInfo, len(keys))
	}
	now := time.Now().UTC().Truncate(time.Millisecond)
	entryID := uuid.NewString()
	for _, k := range keys {
		lbl := label
		if k.kind == BlockSpectator {
			lbl = "" // a session marker names no one
		}
		entry.blocks[k] = blockInfo{entry: entryID, label: lbl, at: now}
		if entry.practice {
			continue
		}
		ctx, cancel := storeCtx()
		err := l.store.AddBlock(ctx, entry.meta.ID,
			BlockRecord{Kind: k.kind, Subject: k.subject, EntryID: entryID, Label: lbl, CreatedAt: now})
		cancel()
		if err != nil {
			slog.Default().Warn("persist table block failed", "game_id", entry.meta.ID, "kind", k.kind, "err", err)
		}
	}
}

// loadBlocks reads a table's blocks back at restore.
func (l *Lobby) loadBlocks(id uuid.UUID) map[blockKey]blockInfo {
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
	out := make(map[blockKey]blockInfo, len(recs))
	for _, r := range recs {
		out[blockKey{r.Kind, r.Subject}] = blockInfo{entry: r.EntryID, label: r.Label, at: r.CreatedAt}
	}
	return out
}

// ErrBlockNotFound is the unblock of an entry the table does not have.
var ErrBlockNotFound = errors.New("lobby: no such block at this table")

// BlockList is a table's block entries, oldest first. An entry that is
// only a spectator session marker (a removal without "also block", or of
// a guest with no name to match) is not listed: it names no one and there
// is nothing for a host to lift. The caller is authorised at the HTTP
// layer.
func (l *Lobby) BlockList(id uuid.UUID) ([]BlockEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.games[id]
	if !ok {
		return nil, ErrGameNotFound
	}
	byID := map[string]*BlockEntry{}
	var order []string
	for k, info := range entry.blocks {
		if k.kind == BlockSpectator {
			continue
		}
		e := byID[info.entry]
		if e == nil {
			e = &BlockEntry{ID: info.entry, Kind: BlockEntryGuest, CreatedAt: info.at.UnixMilli()}
			byID[info.entry] = e
			order = append(order, info.entry)
		}
		if k.kind == BlockUser || k.kind == BlockDiscord {
			e.Kind = BlockEntryAccount
		}
		if e.Label == "" || (k.kind == BlockUser && info.label != "") {
			e.Label = info.label
		}
		if ms := info.at.UnixMilli(); ms < e.CreatedAt {
			e.CreatedAt = ms
		}
	}
	out := make([]BlockEntry, 0, len(order))
	for _, o := range order {
		out = append(out, *byID[o])
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt != out[j].CreatedAt {
			return out[i].CreatedAt < out[j].CreatedAt
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

// Unblock lifts one block entry: the person can take a seat (or
// spectate) at this table again. It removes the entry's spectator-session
// marker too, so a session the table killed with the entry works again.
// ErrBlockNotFound when the table has no such entry. The caller is
// authorised at the HTTP layer.
func (l *Lobby) Unblock(id uuid.UUID, entryID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.games[id]
	if !ok {
		return ErrGameNotFound
	}
	var keys []blockKey
	for k, info := range entry.blocks {
		if info.entry == entryID && entryID != "" {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return ErrBlockNotFound
	}
	// The store first: a failed delete must not leave the table
	// believing a block is lifted that a restart would bring back.
	if !entry.practice {
		ctx, cancel := storeCtx()
		_, err := l.store.DeleteBlockEntry(ctx, id, entryID)
		cancel()
		if err != nil {
			return err
		}
	}
	for _, k := range keys {
		delete(entry.blocks, k)
	}
	return nil
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
		// The store deletes whole entries (accountdelete): the label is
		// the person's name, and a spectator marker under the same entry
		// goes with it. The in-memory copy does the same.
		gone := map[string]bool{}
		for _, k := range []blockKey{{BlockUser, userID.String()}, {BlockDiscord, discordID}} {
			if k.subject == "" {
				continue
			}
			if info, ok := e.blocks[k]; ok {
				gone[info.entry] = true
			}
		}
		for k, info := range e.blocks {
			if gone[info.entry] {
				delete(e.blocks, k)
			}
		}
	}
}
