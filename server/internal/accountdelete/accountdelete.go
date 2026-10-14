// Package accountdelete is self-serve account deletion (ADR 0149,
// #2983): the one transaction that removes or anonymises everything the
// user database holds about a person, and the inventory that says what
// it does to every column.
//
// # What it does
//
// One SQL transaction, in this order, all or nothing:
//
//  1. Refuse if the person holds a seat at a table that has not ended
//     (a lobby or an active game). SeatedError names the table.
//  2. Delete: decks, user_settings, table_setups, notification_prefs,
//     model_keys, user_playmats, the
//     identities row (Discord id, name, avatar hash and the encrypted
//     refresh token), the person's deck_request_asks rows, the player
//     reports they filed or are the subject of, and the table blocks
//     keyed on their account or Discord id (#2984).
//  3. Anonymise: the person's seats keep their place in the game but
//     take the name DeletedPlayerName and lose any pending Discord id;
//     games.host_discord_id is cleared if it was theirs.
//  4. Scrub the users row to a tombstone and stamp deleted_at and the
//     session watermark. The row is NOT deleted: seats, games and
//     invites keep their foreign keys, and the watermark keeps every
//     session token issued before the deletion refused after a restart.
//
// It returns what the caller must finish outside the database: the
// playmat image files, the engine files of the games the person sat in
// (replays, crash dumps, restore points), the live session sockets and
// the in-memory lobby. Those are not transactional, so they are the
// caller's, logged and best effort; the database half is the promise.
//
// # The inventory
//
// Inventory lists every table and column in the schema with what
// deletion does to it. TestInventoryCoversTheSchema fails when a
// migration adds a table or column that is not listed, which is the
// maintenance rule of ADR 0149: a new place that holds user data needs
// a deletion decision and a line on the privacy page before it merges.
package accountdelete

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// DeletedPlayerName is the stable name an anonymised seat carries, in
// the seats table, in the users tombstone and in the engine files of
// every past game.
const DeletedPlayerName = "Deleted player"

var (
	// ErrNotFound: there is no such user.
	ErrNotFound = errors.New("accountdelete: user not found")
	// ErrAlreadyDeleted: the account was already deleted. The caller's
	// session should have been refused first; this is the backstop.
	ErrAlreadyDeleted = errors.New("accountdelete: account already deleted")
)

// SeatedError means the person still holds a seat at a table that has
// not ended, so the account cannot be deleted yet.
type SeatedError struct {
	GameID   uuid.UUID
	GameName string
	State    string
}

func (e *SeatedError) Error() string {
	return fmt.Sprintf("accountdelete: seated at %s table %s", e.State, e.GameID)
}

// SeatRef is one seat the deletion anonymised: the game and the
// engine's player id, which is how the game's files are found.
type SeatRef struct {
	GameID   uuid.UUID
	PlayerID uuid.UUID
}

// Counts is what a deletion removed, for the audit line. It carries no
// name, no Discord id and nothing else that identifies the person.
type Counts struct {
	Decks           int
	Playmats        int
	DeckRequestAsks int
	Seats           int
	SettingsRows    int
	SetupRows       int
	NotifyPrefRows  int
	ModelKeyRows    int
	Identities      int
	Reports         int
	Blocks          int
}

// Result is what Delete hands back after it commits.
type Result struct {
	// Seats are the seats that were anonymised, one per (game, seat).
	Seats []SeatRef
	// PlaymatIDs are the image files (<data dir>/playmats/<id>.jpg)
	// whose rows were removed.
	PlaymatIDs []string
	// DiscordSubject is the Discord id the person had, "" if none. The
	// caller uses it to forget the cached avatar and must never log it.
	DiscordSubject string
	// AvatarHash is the avatar hash the identity row held, "" if none.
	AvatarHash string
	Counts     Counts
	DeletedAt  time.Time
}

// Store deletes accounts. SQLStore is the production one.
type Store interface {
	// Delete runs the deletion described in the package comment.
	// *SeatedError if the person holds a seat at an unended table;
	// ErrNotFound; ErrAlreadyDeleted.
	Delete(ctx context.Context, id uuid.UUID) (Result, error)
}

// SQLStore is the Store over internal/db.
type SQLStore struct {
	db  *sql.DB
	now func() time.Time
	// ledger, when set, records each committed deletion (ledger.go).
	ledger *Ledger
}

// NewSQLStore wraps an opened, migrated database. The caller owns d.
func NewSQLStore(d *db.DB) *SQLStore {
	return &SQLStore{db: d.DB, now: func() time.Time { return time.Now().UTC() }}
}

// WithLedger makes every committed deletion also append to l, so a
// restore from an older database copy can re-apply it (ledger.go).
func (s *SQLStore) WithLedger(l *Ledger) *SQLStore {
	s.ledger = l
	return s
}

var _ Store = (*SQLStore)(nil)

// Delete implements Store.
func (s *SQLStore) Delete(ctx context.Context, id uuid.UUID) (Result, error) {
	return s.delete(ctx, id, false)
}

// Reapply runs the deletion for an account a restore brought back
// (ledger.go). It differs from Delete in two ways: the active-table
// rule is skipped, because the person asked to be deleted before the
// restore point and the table may only look live in an old copy; and it
// writes no ledger entry, because the entry is already in the ledger.
func (s *SQLStore) Reapply(ctx context.Context, id uuid.UUID) (Result, error) {
	return s.delete(ctx, id, true)
}

func (s *SQLStore) delete(ctx context.Context, id uuid.UUID, reapply bool) (Result, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, fmt.Errorf("accountdelete: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	uid := id.String()
	at := s.now().Truncate(time.Millisecond)
	ms := at.UnixMilli()
	res := Result{DeletedAt: at}

	var deletedAt sql.NullInt64
	err = tx.QueryRowContext(ctx, `SELECT deleted_at FROM users WHERE id = ?`, uid).Scan(&deletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Result{}, ErrNotFound
	}
	if err != nil {
		return Result{}, fmt.Errorf("accountdelete: read user: %w", err)
	}
	if deletedAt.Valid {
		return Result{}, ErrAlreadyDeleted
	}

	// The Discord identity first: its subject also finds the seats that
	// were claimed through OAuth before a users row linked them.
	var subject, hash sql.NullString
	err = tx.QueryRowContext(ctx,
		`SELECT subject, avatar_hash FROM identities WHERE user_id = ? AND provider = 'discord'`, uid).Scan(&subject, &hash)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Result{}, fmt.Errorf("accountdelete: read identity: %w", err)
	}
	res.DiscordSubject, res.AvatarHash = subject.String, hash.String

	// 1. The active-table rule (not for a re-application).
	var (
		gid, gname, gstate string
	)
	err = tx.QueryRowContext(ctx,
		`SELECT g.id, g.name, g.state FROM seats s JOIN games g ON g.id = s.game_id
		 WHERE (s.user_id = ? OR (? <> '' AND s.pending_discord_id = ?)) AND g.state <> 'ended'
		 ORDER BY g.created_at DESC LIMIT 1`,
		uid, res.DiscordSubject, res.DiscordSubject).Scan(&gid, &gname, &gstate)
	switch {
	case err == nil && !reapply:
		parsed, _ := uuid.Parse(gid)
		return Result{}, &SeatedError{GameID: parsed, GameName: gname, State: gstate}
	case err != nil && !errors.Is(err, sql.ErrNoRows):
		return Result{}, fmt.Errorf("accountdelete: check active seats: %w", err)
	}

	// 2a. The seats to anonymise, read before they are touched.
	rows, err := tx.QueryContext(ctx,
		`SELECT game_id, player_id FROM seats WHERE user_id = ? OR (? <> '' AND pending_discord_id = ?)`,
		uid, res.DiscordSubject, res.DiscordSubject)
	if err != nil {
		return Result{}, fmt.Errorf("accountdelete: read seats: %w", err)
	}
	for rows.Next() {
		var g, p string
		if err := rows.Scan(&g, &p); err != nil {
			_ = rows.Close()
			return Result{}, fmt.Errorf("accountdelete: read seats: %w", err)
		}
		gu, gerr := uuid.Parse(g)
		pu, perr := uuid.Parse(p)
		if gerr != nil || perr != nil {
			continue
		}
		res.Seats = append(res.Seats, SeatRef{GameID: gu, PlayerID: pu})
	}
	if err := rows.Close(); err != nil {
		return Result{}, fmt.Errorf("accountdelete: read seats: %w", err)
	}

	// 2b. The playmat files to remove once the rows are gone.
	prows, err := tx.QueryContext(ctx, `SELECT playmat_id FROM user_playmats WHERE user_id = ?`, uid)
	if err != nil {
		return Result{}, fmt.Errorf("accountdelete: read playmats: %w", err)
	}
	for prows.Next() {
		var p string
		if err := prows.Scan(&p); err != nil {
			_ = prows.Close()
			return Result{}, fmt.Errorf("accountdelete: read playmats: %w", err)
		}
		res.PlaymatIDs = append(res.PlaymatIDs, p)
	}
	if err := prows.Close(); err != nil {
		return Result{}, fmt.Errorf("accountdelete: read playmats: %w", err)
	}

	exec := func(n *int, q string, args ...any) error {
		r, err := tx.ExecContext(ctx, q, args...)
		if err != nil {
			return fmt.Errorf("accountdelete: %s: %w", q, err)
		}
		if n != nil {
			if c, cerr := r.RowsAffected(); cerr == nil {
				*n = int(c)
			}
		}
		return nil
	}

	// 3. Anonymise the seats and the pending host pointer.
	if err := exec(&res.Counts.Seats,
		`UPDATE seats SET guest_name = ?, pending_discord_id = NULL
		 WHERE user_id = ? OR (? <> '' AND pending_discord_id = ?)`,
		DeletedPlayerName, uid, res.DiscordSubject, res.DiscordSubject); err != nil {
		return Result{}, err
	}
	if res.DiscordSubject != "" {
		if err := exec(nil, `UPDATE games SET host_discord_id = NULL WHERE host_discord_id = ?`, res.DiscordSubject); err != nil {
			return Result{}, err
		}
		// 2c. The deck-request rate-limit rows are keyed by the Discord id.
		if err := exec(&res.Counts.DeckRequestAsks, `DELETE FROM deck_request_asks WHERE requester = ?`,
			"discord:"+res.DiscordSubject); err != nil {
			return Result{}, err
		}
	}

	// 2. Delete. decks last of the pointer-holders: seats.deck_id is
	// ON DELETE SET NULL (migration 0008), so the seats keep their row.
	if err := exec(&res.Counts.Decks, `DELETE FROM decks WHERE owner_id = ?`, uid); err != nil {
		return Result{}, err
	}
	if err := exec(&res.Counts.SettingsRows, `DELETE FROM user_settings WHERE user_id = ?`, uid); err != nil {
		return Result{}, err
	}
	if err := exec(&res.Counts.SetupRows, `DELETE FROM table_setups WHERE user_id = ?`, uid); err != nil {
		return Result{}, err
	}
	if err := exec(&res.Counts.NotifyPrefRows, `DELETE FROM notification_prefs WHERE user_id = ?`, uid); err != nil {
		return Result{}, err
	}
	// The person's own model key (ADR 0182, migration 0025), sealed
	// key and all. A bot seat that ran on it keeps pointing at the
	// tombstone (seats.bot_key_user_id), which carries no identity and
	// has no key, so a restored seat plays the heuristic.
	if err := exec(&res.Counts.ModelKeyRows, `DELETE FROM model_keys WHERE user_id = ?`, uid); err != nil {
		return Result{}, err
	}
	if err := exec(&res.Counts.Playmats, `DELETE FROM user_playmats WHERE user_id = ?`, uid); err != nil {
		return Result{}, err
	}
	if err := exec(&res.Counts.Identities, `DELETE FROM identities WHERE user_id = ?`, uid); err != nil {
		return Result{}, err
	}
	// Moderation (migration 0016, #2984): the reports the person filed or
	// is the subject of, and the table blocks keyed on their account or
	// Discord id. A block on a folded guest name names no account and
	// stays; so does a report that merely quotes them in an excerpt.
	if err := exec(&res.Counts.Reports,
		`DELETE FROM player_reports WHERE reporter_user_id = ? OR reported_user_id = ?`, uid, uid); err != nil {
		return Result{}, err
	}
	if err := exec(&res.Counts.Blocks,
		`DELETE FROM table_blocks WHERE entry_id IN
		   (SELECT entry_id FROM table_blocks
		     WHERE (kind = 'user' AND subject = ?) OR (kind = 'discord' AND ? <> '' AND subject = ?))`,
		uid, res.DiscordSubject, res.DiscordSubject); err != nil {
		return Result{}, err
	}

	// 4. The tombstone. created_at stays (it names no one and orders
	// the account list); last_seen_at is activity and is replaced.
	if err := exec(nil,
		`UPDATE users SET display_name = ?, avatar_url = NULL, last_deck = NULL,
		   playmat_id = NULL, playmat_wash = NULL, admin_mode_at = 0,
		   access_checked_at = NULL, access_result = NULL,
		   last_seen_at = ?, sessions_invalid_before = MAX(sessions_invalid_before, ?),
		   deleted_at = ?
		 WHERE id = ?`,
		DeletedPlayerName, ms, ms, ms, uid); err != nil {
		return Result{}, err
	}

	if err := tx.Commit(); err != nil {
		return Result{}, fmt.Errorf("accountdelete: commit: %w", err)
	}
	if s.ledger != nil && !reapply {
		// After the commit, never before: an entry for a deletion that
		// rolled back would delete a live account at the next restore.
		// A failure here is logged, not returned: the database half of
		// the promise is kept, and only the next restore would miss it.
		if err := s.ledger.Append(id, at); err != nil {
			slog.Error("account deletion: ledger append failed; a restore would bring this account back",
				"user_id", id, "err", err)
		}
	}
	return res, nil
}
