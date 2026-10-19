package accountdelete

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

func openDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return d
}

func exec(t *testing.T, d *db.DB, q string, args ...any) {
	t.Helper()
	if _, err := d.Exec(q, args...); err != nil {
		t.Fatalf("exec %q: %v", q, err)
	}
}

func count(t *testing.T, d *db.DB, q string, args ...any) int {
	t.Helper()
	var n int
	if err := d.QueryRow(q, args...).Scan(&n); err != nil {
		t.Fatalf("count %q: %v", q, err)
	}
	return n
}

// schemaColumns reads every table.column the migrated database has.
func schemaColumns(t *testing.T, d *db.DB) map[string]bool {
	t.Helper()
	rows, err := d.Query(`SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		t.Fatal(err)
	}
	var tables []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatal(err)
		}
		tables = append(tables, n)
	}
	_ = rows.Close()
	out := map[string]bool{}
	for _, tb := range tables {
		cr, err := d.Query(fmt.Sprintf(`SELECT name FROM pragma_table_info('%s')`, tb))
		if err != nil {
			t.Fatal(err)
		}
		for cr.Next() {
			var c string
			if err := cr.Scan(&c); err != nil {
				t.Fatal(err)
			}
			out[tb+"."+c] = true
		}
		_ = cr.Close()
	}
	return out
}

// TestInventoryCoversTheSchema is the maintenance rule of ADR 0149: a
// migration that adds a table or a column must say what account
// deletion does to it. When this fails, add the new columns to
// Inventory (inventory.go), make Delete honour the treatment, and
// update the privacy page (client/src/routes/Privacy.svelte).
func TestInventoryCoversTheSchema(t *testing.T) {
	d := openDB(t)
	schema := schemaColumns(t, d)
	listed := map[string]bool{}
	for _, e := range Inventory {
		key := e.Table + "." + e.Column
		if listed[key] {
			t.Errorf("Inventory lists %s twice", key)
		}
		listed[key] = true
		if e.Treatment == "" {
			t.Errorf("Inventory entry %s has no treatment", key)
		}
		if (e.Treatment == Kept || e.Treatment == Anonymised) && strings.TrimSpace(e.Note) == "" {
			t.Errorf("Inventory entry %s is %s and needs a note saying why", key, e.Treatment)
		}
	}
	var missing, stale []string
	for k := range schema {
		if !listed[k] {
			missing = append(missing, k)
		}
	}
	for k := range listed {
		if !schema[k] {
			stale = append(stale, k)
		}
	}
	sort.Strings(missing)
	sort.Strings(stale)
	if len(missing) > 0 {
		t.Errorf("the schema has columns account deletion has no decision for: %v\n"+
			"add them to accountdelete.Inventory, make Delete honour the treatment, and update the privacy page "+
			"(client/src/routes/Privacy.svelte); see ADR 0149", missing)
	}
	if len(stale) > 0 {
		t.Errorf("Inventory lists columns the schema no longer has: %v", stale)
	}
}

type fixture struct {
	d       *db.DB
	user    uuid.UUID
	other   uuid.UUID
	subject string
	hash    string

	endedGame, otherGame uuid.UUID
	endedPlayer          uuid.UUID
	deckID               uuid.UUID
}

const (
	userSubject  = "111111111111111111"
	otherSubject = "222222222222222222"
)

// seed builds a person with a row in every table that holds anything
// about them, an ended table they sat at beside another person, and
// that other person's own rows (which a deletion must not touch).
func seed(t *testing.T) fixture {
	t.Helper()
	d := openDB(t)
	f := fixture{d: d, user: uuid.New(), other: uuid.New(), subject: userSubject, hash: "abcdef0123456789abcdef0123456789"}
	f.endedGame, f.otherGame, f.endedPlayer, f.deckID = uuid.New(), uuid.New(), uuid.New(), uuid.New()
	otherDeck, otherPlayer := uuid.New(), uuid.New()

	for _, u := range []struct {
		id   uuid.UUID
		name string
	}{{f.user, "Ian"}, {f.other, "Bea"}} {
		exec(t, d, `INSERT INTO users (id, display_name, avatar_url, created_at, last_seen_at, last_deck, playmat_id, playmat_wash, admin_mode_at,
				access_checked_at, access_result)
			VALUES (?, ?, '/avatars/x/y.png', 100, 200, '{"kind":"prebuilt","id":"x"}', ?, 60, 300, 400, 'admitted')`,
			u.id.String(), u.name, "mat-"+u.name)
	}
	exec(t, d, `INSERT INTO identities (provider, subject, user_id, display_name, avatar_hash, refresh_token, scopes, linked_at)
		VALUES ('discord', ?, ?, 'ian#1', ?, x'0102', 'identify', 1)`, userSubject, f.user.String(), f.hash)
	exec(t, d, `INSERT INTO identities (provider, subject, user_id, display_name, avatar_hash, refresh_token, scopes, linked_at)
		VALUES ('discord', ?, ?, 'bea#1', 'bbbb', x'0304', 'identify', 1)`, otherSubject, f.other.String())

	for _, o := range []struct {
		owner uuid.UUID
		id    uuid.UUID
	}{{f.user, f.deckID}, {f.other, otherDeck}} {
		exec(t, d, `INSERT INTO decks (id, owner_id, name, source_format, source_text, commanders, card_count, created_at, updated_at, source_url)
			VALUES (?, ?, 'Deck', 'text', '1 Sol Ring', '[]', 1, 1, 1, 'https://example.test/d')`, o.id.String(), o.owner.String())
	}
	for _, u := range []uuid.UUID{f.user, f.other} {
		exec(t, d, `INSERT INTO user_settings (user_id, version, revision, body, updated_at) VALUES (?, 1, 1, '{}', 1)`, u.String())
		exec(t, d, `INSERT INTO table_setups (user_id, body, game_id, updated_at) VALUES (?, '{}', NULL, 1)`, u.String())
		exec(t, d, `INSERT INTO notification_prefs (user_id, turn, table_ready, table_filled, dm_failed_at, updated_at) VALUES (?, 1, 1, 0, 5, 1)`, u.String())
		exec(t, d, `INSERT INTO model_keys (user_id, provider, base_url, model, sealed_key, last4, games_per_day, calls_per_game,
				day, day_games, created_at, updated_at, last_used_at, last_error, last_error_at)
			VALUES (?, 'custom', 'https://llm.example.test/v1', 'm', ?, 'abcd', 5, 50, '2026-10-14', 1, 1, 1, 2, 'refused', 3)`,
			u.String(), make([]byte, 40))
		exec(t, d, `INSERT INTO user_playmats (user_id, slot, playmat_id, created_at) VALUES (?, 1, ?, 1)`, u.String(), "mat-"+map[bool]string{true: "Ian", false: "Bea"}[u == f.user])
	}
	exec(t, d, `INSERT INTO deck_request_asks (deck_key, requester, at) VALUES ('moxfield:a', ?, 1), ('moxfield:b', ?, 1)`,
		"discord:"+userSubject, "discord:"+otherSubject)
	exec(t, d, `INSERT INTO deck_requests (deck_key, issue_number, issue_url, created_at) VALUES ('moxfield:a', 7, 'https://example.test/7', 1)`)
	exec(t, d, `INSERT INTO deck_request_cards (deck_key, card_key, card_name, issue_number) VALUES ('moxfield:a', 'oracle-1', 'Sol Ring', 7)`)

	exec(t, d, `INSERT INTO games (id, name, created_by, state, created_at, ended_at, winner_seat, outcome, host_player_id, host_discord_id)
		VALUES (?, 'Friday pod', ?, 'ended', 1, 2, 0, 'win', ?, ?)`, f.endedGame.String(), f.user.String(), f.endedPlayer.String(), userSubject)
	exec(t, d, `INSERT INTO seats (game_id, seat, player_id, user_id, guest_name, deck_id, deck_name) VALUES (?, 0, ?, ?, 'Ian', ?, 'Deck')`,
		f.endedGame.String(), f.endedPlayer.String(), f.user.String(), f.deckID.String())
	exec(t, d, `INSERT INTO seats (game_id, seat, player_id, user_id, guest_name, deck_id, deck_name) VALUES (?, 1, ?, ?, 'Bea', ?, 'Deck')`,
		f.endedGame.String(), otherPlayer.String(), f.other.String(), otherDeck.String())
	exec(t, d, `INSERT INTO invites (token_hash, game_id, kind, created_by, created_at) VALUES (x'aa', ?, 'player', ?, 1)`,
		f.endedGame.String(), f.user.String())
	// A bot the person ran on their own model key (ADR 0182).
	exec(t, d, `INSERT INTO seats (game_id, seat, player_id, guest_name, bot_tier, bot_key_user_id, bot_key_model) VALUES (?, 3, ?, 'Bot 1', 'assisted', ?, 'm')`,
		f.endedGame.String(), uuid.NewString(), f.user.String())

	// An ended table only the other person sat at, to prove it is untouched.
	exec(t, d, `INSERT INTO games (id, name, created_by, state, created_at, ended_at) VALUES (?, 'Other pod', ?, 'ended', 1, 2)`,
		f.otherGame.String(), f.other.String())
	exec(t, d, `INSERT INTO seats (game_id, seat, player_id, user_id, guest_name) VALUES (?, 0, ?, ?, 'Bea')`,
		f.otherGame.String(), uuid.NewString(), f.other.String())

	// Moderation (migration 0016, #2984): a report the person filed, one
	// about them, one between strangers to them; blocks keyed on their
	// account, their Discord id and a guest name, plus one on the other
	// person's account.
	for _, r := range []struct{ id, by, about string }{
		{"r-filed", f.user.String(), f.other.String()},
		{"r-about", f.other.String(), f.user.String()},
		{"r-other", f.other.String(), uuid.NewString()},
	} {
		exec(t, d, `INSERT INTO player_reports (id, created_at, game_id, game_name, reporter_player_id, reporter_user_id, reporter_name,
			reported_player_id, reported_user_id, reported_name, reason, excerpt)
			VALUES (?, 1, ?, 'Friday pod', ?, ?, 'A', ?, ?, 'B', 'rude', '[]')`,
			r.id, f.endedGame.String(), uuid.NewString(), r.by, uuid.NewString(), r.about)
	}
	// One removal is one entry (migration 0017): the deleted person's
	// account and Discord id share "e-user" with the display-name label,
	// and a spectator-session marker rides under it; a guest's name and
	// the other person's pair are entries of their own.
	for _, b := range []struct{ kind, subject, entry, label string }{
		{"user", f.user.String(), "e-user", "Alice"}, {"discord", userSubject, "e-user", "Alice"},
		{"spectator", "sess-1", "e-user", ""}, {"name", "bob", "e-bob", "bob"},
		{"user", f.other.String(), "e-other", "Zed"}, {"discord", otherSubject, "e-other", "Zed"},
	} {
		exec(t, d, `INSERT INTO table_blocks (game_id, kind, subject, created_at, entry_id, label) VALUES (?, ?, ?, 1, ?, ?)`,
			f.endedGame.String(), b.kind, b.subject, b.entry, b.label)
	}
	return f
}

func TestDeleteRemovesOrAnonymisesEveryTable(t *testing.T) {
	f := seed(t)
	s := NewSQLStore(f.d)
	at := time.UnixMilli(1_800_000_000_000).UTC()
	s.now = func() time.Time { return at }

	res, err := s.Delete(context.Background(), f.user)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	u := f.user.String()

	// Deleted.
	for table, q := range map[string]string{
		"decks":              `SELECT COUNT(*) FROM decks WHERE owner_id = ?`,
		"user_settings":      `SELECT COUNT(*) FROM user_settings WHERE user_id = ?`,
		"table_setups":       `SELECT COUNT(*) FROM table_setups WHERE user_id = ?`,
		"notification_prefs": `SELECT COUNT(*) FROM notification_prefs WHERE user_id = ?`,
		"model_keys":         `SELECT COUNT(*) FROM model_keys WHERE user_id = ?`,
		"user_playmats":      `SELECT COUNT(*) FROM user_playmats WHERE user_id = ?`,
		"identities":         `SELECT COUNT(*) FROM identities WHERE user_id = ?`,
		"deck_request_asks":  `SELECT COUNT(*) FROM deck_request_asks WHERE requester = 'discord:` + userSubject + `' AND ? <> ''`,
	} {
		if n := count(t, f.d, q, u); n != 0 {
			t.Errorf("%s: %d rows left for the deleted user", table, n)
		}
	}
	if len(res.PlaymatIDs) != 1 || res.PlaymatIDs[0] != "mat-Ian" {
		t.Errorf("PlaymatIDs = %v, want [mat-Ian]", res.PlaymatIDs)
	}
	if res.DiscordSubject != userSubject || res.AvatarHash != f.hash {
		t.Errorf("identity handed back = %q / %q", res.DiscordSubject, res.AvatarHash)
	}

	// Anonymised: the seat keeps its place, the name and the pointers go.
	var name string
	var deckID sql.NullString
	if err := f.d.QueryRow(`SELECT guest_name, deck_id FROM seats WHERE game_id = ? AND seat = 0`, f.endedGame.String()).Scan(&name, &deckID); err != nil {
		t.Fatal(err)
	}
	if name != DeletedPlayerName || deckID.Valid {
		t.Errorf("seat = %q deck %v, want %q and no deck", name, deckID, DeletedPlayerName)
	}
	if len(res.Seats) != 1 || res.Seats[0].GameID != f.endedGame || res.Seats[0].PlayerID != f.endedPlayer {
		t.Errorf("Seats = %+v", res.Seats)
	}
	var hostDiscord sql.NullString
	if err := f.d.QueryRow(`SELECT host_discord_id FROM games WHERE id = ?`, f.endedGame.String()).Scan(&hostDiscord); err != nil || hostDiscord.Valid {
		t.Errorf("host_discord_id = %v, %v: want NULL", hostDiscord, err)
	}

	// Moderation: reports the person filed or is the subject of go whole;
	// a stranger's stays. Blocks on their account and Discord id go; a
	// folded guest name and the other person's blocks stay.
	if n := count(t, f.d, `SELECT COUNT(*) FROM player_reports WHERE reporter_user_id = ? OR reported_user_id = ?`, u, u); n != 0 {
		t.Errorf("%d reports left for the deleted user", n)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM player_reports`); n != 1 {
		t.Errorf("player_reports = %d, want only the stranger's report", n)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM table_blocks WHERE (kind = 'user' AND subject = ?) OR (kind = 'discord' AND subject = ?)`, u, userSubject); n != 0 {
		t.Errorf("%d blocks left on the deleted user's account or Discord id", n)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM table_blocks`); n != 3 {
		t.Errorf("table_blocks = %d, want the name block and the other person's two", n)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM table_blocks WHERE entry_id = 'e-user' OR label = 'Alice'`); n != 0 {
		t.Errorf("%d rows of the deleted person's block entry remain (their display-name label, or the spectator marker under it)", n)
	}

	// The tombstone.
	var dn string
	var avatar, lastDeck, playmat sql.NullString
	var wash sql.NullInt64
	var lastSeen, mode, invalid, deleted int64
	if err := f.d.QueryRow(`SELECT display_name, avatar_url, last_deck, playmat_id, playmat_wash, last_seen_at, admin_mode_at, sessions_invalid_before, deleted_at FROM users WHERE id = ?`, u).
		Scan(&dn, &avatar, &lastDeck, &playmat, &wash, &lastSeen, &mode, &invalid, &deleted); err != nil {
		t.Fatal(err)
	}
	if dn != DeletedPlayerName || avatar.Valid || lastDeck.Valid || playmat.Valid || wash.Valid || mode != 0 {
		t.Errorf("tombstone not scrubbed: %q %v %v %v %v mode=%d", dn, avatar, lastDeck, playmat, wash, mode)
	}
	var checkedAt sql.NullInt64
	var access sql.NullString
	if err := f.d.QueryRow(`SELECT access_checked_at, access_result FROM users WHERE id = ?`, u).Scan(&checkedAt, &access); err != nil {
		t.Fatal(err)
	}
	if checkedAt.Valid || access.Valid {
		t.Errorf("tombstone keeps the sign-up gate's answer: %v %v", checkedAt, access)
	}
	if invalid != at.UnixMilli() || deleted != at.UnixMilli() || lastSeen != at.UnixMilli() {
		t.Errorf("tombstone clocks = %d %d %d, want %d", invalid, deleted, lastSeen, at.UnixMilli())
	}

	// Kept: the game, the other player's seat, the invite and the
	// foreign keys that point at the tombstone.
	if n := count(t, f.d, `SELECT COUNT(*) FROM seats WHERE game_id = ? AND user_id = ?`, f.endedGame.String(), u); n != 1 {
		t.Errorf("the seat must still point at the tombstone, got %d", n)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM seats WHERE game_id = ? AND bot_key_user_id = ? AND bot_key_model = 'm'`, f.endedGame.String(), u); n != 1 {
		t.Errorf("the own-key bot seat must still point at the tombstone, got %d", n)
	}
	if res.Counts.ModelKeyRows != 1 {
		t.Errorf("ModelKeyRows = %d, want 1", res.Counts.ModelKeyRows)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM invites WHERE created_by = ?`, u); n != 1 {
		t.Errorf("the invite must stay, got %d", n)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM games WHERE created_by = ?`, u); n != 1 {
		t.Errorf("the game must stay, got %d", n)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM deck_requests`); n != 1 {
		t.Errorf("deck_requests are not personal and stay, got %d", n)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM deck_request_cards`); n != 1 {
		t.Errorf("deck_request_cards name a deck and a card, not a person, and stay, got %d", n)
	}

	// The other person is untouched.
	o := f.other.String()
	for _, q := range []string{
		`SELECT COUNT(*) FROM decks WHERE owner_id = ?`, `SELECT COUNT(*) FROM user_settings WHERE user_id = ?`,
		`SELECT COUNT(*) FROM table_setups WHERE user_id = ?`, `SELECT COUNT(*) FROM user_playmats WHERE user_id = ?`,
		`SELECT COUNT(*) FROM identities WHERE user_id = ?`, `SELECT COUNT(*) FROM notification_prefs WHERE user_id = ?`,
		`SELECT COUNT(*) FROM model_keys WHERE user_id = ?`,
	} {
		if n := count(t, f.d, q, o); n != 1 {
			t.Errorf("%s = %d for the other person, want 1", q, n)
		}
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM deck_request_asks WHERE requester = 'discord:`+otherSubject+`'`); n != 1 {
		t.Errorf("the other person's ask was removed")
	}
	var bea string
	if err := f.d.QueryRow(`SELECT guest_name FROM seats WHERE game_id = ? AND seat = 1`, f.endedGame.String()).Scan(&bea); err != nil || bea != "Bea" {
		t.Errorf("other seat = %q, %v", bea, err)
	}
	var beaUser string
	if err := f.d.QueryRow(`SELECT display_name FROM users WHERE id = ?`, o).Scan(&beaUser); err != nil || beaUser != "Bea" {
		t.Errorf("other user = %q, %v", beaUser, err)
	}
}

func TestDeleteRefusesWhileSeatedAtAnUnendedTable(t *testing.T) {
	for _, state := range []string{"lobby", "active"} {
		t.Run(state, func(t *testing.T) {
			f := seed(t)
			live := uuid.New()
			exec(t, f.d, `INSERT INTO games (id, name, created_by, state, created_at) VALUES (?, 'Live pod', ?, ?, 5)`, live.String(), f.other.String(), state)
			exec(t, f.d, `INSERT INTO seats (game_id, seat, player_id, user_id, guest_name) VALUES (?, 0, ?, ?, 'Ian')`, live.String(), uuid.NewString(), f.user.String())

			_, err := NewSQLStore(f.d).Delete(context.Background(), f.user)
			var se *SeatedError
			if !errors.As(err, &se) || se.GameID != live || se.GameName != "Live pod" || se.State != state {
				t.Fatalf("err = %v, want SeatedError for %s", err, live)
			}
			// Nothing happened: the transaction rolled back.
			if n := count(t, f.d, `SELECT COUNT(*) FROM decks WHERE owner_id = ?`, f.user.String()); n != 1 {
				t.Errorf("a refused deletion removed the deck")
			}
			if n := count(t, f.d, `SELECT COUNT(*) FROM users WHERE id = ? AND deleted_at IS NULL`, f.user.String()); n != 1 {
				t.Errorf("a refused deletion marked the account deleted")
			}
		})
	}
}

func TestDeleteRefusesAPendingSeatAtAnUnendedTable(t *testing.T) {
	// A Discord seat claimed before the person had a users row waits on
	// the snowflake (seats.pending_discord_id). It counts as theirs.
	f := seed(t)
	live := uuid.New()
	exec(t, f.d, `INSERT INTO games (id, name, state, created_at) VALUES (?, 'Pending pod', 'active', 5)`, live.String())
	exec(t, f.d, `INSERT INTO seats (game_id, seat, player_id, guest_name, pending_discord_id) VALUES (?, 0, ?, 'Ian', ?)`,
		live.String(), uuid.NewString(), userSubject)
	if _, err := NewSQLStore(f.d).Delete(context.Background(), f.user); err == nil {
		t.Fatal("a pending seat at an active table must block deletion")
	}
}

func TestDeleteAnonymisesAPendingSeatInAnEndedGame(t *testing.T) {
	f := seed(t)
	p := uuid.New()
	exec(t, f.d, `INSERT INTO seats (game_id, seat, player_id, guest_name, pending_discord_id) VALUES (?, 2, ?, 'Ian', ?)`,
		f.endedGame.String(), p.String(), userSubject)
	res, err := NewSQLStore(f.d).Delete(context.Background(), f.user)
	if err != nil {
		t.Fatal(err)
	}
	var name string
	var pending sql.NullString
	if err := f.d.QueryRow(`SELECT guest_name, pending_discord_id FROM seats WHERE player_id = ?`, p.String()).Scan(&name, &pending); err != nil {
		t.Fatal(err)
	}
	if name != DeletedPlayerName || pending.Valid {
		t.Errorf("pending seat = %q %v", name, pending)
	}
	if len(res.Seats) != 2 {
		t.Errorf("Seats = %+v, want both the linked and the pending seat", res.Seats)
	}
}

func TestDeleteTwiceAndUnknown(t *testing.T) {
	f := seed(t)
	s := NewSQLStore(f.d)
	if _, err := s.Delete(context.Background(), f.user); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Delete(context.Background(), f.user); !errors.Is(err, ErrAlreadyDeleted) {
		t.Errorf("second delete = %v, want ErrAlreadyDeleted", err)
	}
	if _, err := s.Delete(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown = %v, want ErrNotFound", err)
	}
}

func TestSigningInAgainAfterDeletionIsANewAccount(t *testing.T) {
	// The identity row is gone, so a later Discord sign-in finds nothing
	// to link to and mints a fresh user. Pinned here at the SQL level:
	// no identity points at the tombstone.
	f := seed(t)
	if _, err := NewSQLStore(f.d).Delete(context.Background(), f.user); err != nil {
		t.Fatal(err)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM identities WHERE provider = 'discord' AND subject = ?`, userSubject); n != 0 {
		t.Errorf("the Discord identity survived: %d", n)
	}
}
