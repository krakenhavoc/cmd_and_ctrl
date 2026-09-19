package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Tests for migration 0003 (users + identities, and the rebuild of
// games / seats / invites that makes their person columns real foreign
// keys). The interesting case is a database an older binary left at
// v2 with rows in it: the rebuild must keep every row and every index,
// must not let DROP TABLE games cascade into seats and invites, and
// must leave foreign keys enforced on the pool afterwards.

// openAtVersion opens a fresh database file with the production DSN
// and migrates it only as far as target, the way a binary that shipped
// with that many migrations would have left it.
func openAtVersion(t *testing.T, target int) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cmdctrl.sqlite")
	sqlDB, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := migrateTo(context.Background(), sqlDB, target); err != nil {
		t.Fatalf("migrateTo(%d): %v", target, err)
	}
	return sqlDB
}

func mustExec(t *testing.T, d *sql.DB, q string, args ...any) {
	t.Helper()
	if _, err := d.Exec(q, args...); err != nil {
		t.Fatalf("exec %q: %v", q, err)
	}
}

func countRows(t *testing.T, d *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

// indexesOn lists the named (non-automatic) indexes on table.
func indexesOn(t *testing.T, d *sql.DB, table string) []string {
	t.Helper()
	rows, err := d.Query(`SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = ? AND sql IS NOT NULL`, table)
	if err != nil {
		t.Fatalf("list indexes: %v", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// referencesOf maps each foreign-key column of table to the table it
// references, from PRAGMA foreign_key_list.
func referencesOf(t *testing.T, d *sql.DB, table string) map[string]string {
	t.Helper()
	rows, err := d.Query(`SELECT "from", "table" FROM pragma_foreign_key_list(?)`, table)
	if err != nil {
		t.Fatalf("foreign_key_list(%s): %v", table, err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var from, to string
		if err := rows.Scan(&from, &to); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out[from] = to
	}
	return out
}

func TestMigration0003FromEmpty(t *testing.T) {
	// Pinned to v3, not "latest" (openAtVersion(t, 0)): this test locks
	// down what migration 0003 itself leaves behind — in particular
	// that seats.deck_id stays unenforced — and migration 0005 (S34
	// sub-PR 5) changes exactly that once it exists. Its own tests
	// (migration_0005_test.go) cover the schema from there.
	d := openAtVersion(t, 3)

	for _, table := range []string{"users", "identities", "games", "seats", "invites"} {
		if n := countRows(t, d, table); n != 0 {
			t.Errorf("%s: %d rows in a fresh database", table, n)
		}
	}
	assertPersonForeignKeys(t, d)
}

func TestMigration0003PreservesAPopulatedV2Database(t *testing.T) {
	d := openAtVersion(t, 2)

	// Two games, one archived and finished, with seats (one pending a
	// Discord link, one a bot, one carrying an unenforced deck id) and
	// both kinds of invite each. Everything 0002 can hold.
	mustExec(t, d, `INSERT INTO games (id, name, created_by, state, created_at, started_at, ended_at, archived_at, winner_seat)
		VALUES ('g1', 'FNM', NULL, 'lobby', 1000, NULL, NULL, NULL, NULL),
		       ('g2', 'Old', NULL, 'ended', 2000, 2100, 2200, 2300, 1)`)
	mustExec(t, d, `INSERT INTO seats (game_id, seat, player_id, user_id, guest_name, bot_tier, deck_id, deck_name, pending_discord_id)
		VALUES ('g1', 0, 'p1', NULL, 'Alice', NULL, NULL, 'Atraxa', '123456789'),
		       ('g1', 1, 'p2', NULL, 'Bot', 'heuristic', NULL, NULL, NULL),
		       ('g2', 0, 'p3', NULL, 'Carol', NULL, 'deck-x', 'Krenko', NULL)`)
	mustExec(t, d, `INSERT INTO invites (token_hash, game_id, kind, created_by, created_at, expires_at, revoked_at)
		VALUES (x'01', 'g1', 'player', NULL, 1000, NULL, NULL),
		       (x'02', 'g1', 'spectator', NULL, 1000, NULL, NULL),
		       (x'03', 'g2', 'player', NULL, 2000, 9000, 2400)`)

	// Pinned to v3 for the same reason as TestMigration0003FromEmpty
	// above: migration 0005 (S34 sub-PR 5) enforces seats.deck_id,
	// which this test's fixture data (an unenforced deck_id string)
	// predates on purpose.
	if err := migrateTo(context.Background(), d, 3); err != nil {
		t.Fatalf("migrate v2 -> v3: %v", err)
	}

	// Every row survived — in particular DROP TABLE games did not
	// cascade into seats and invites.
	if got := countRows(t, d, "games"); got != 2 {
		t.Errorf("games: %d rows, want 2", got)
	}
	if got := countRows(t, d, "seats"); got != 3 {
		t.Errorf("seats: %d rows, want 3", got)
	}
	if got := countRows(t, d, "invites"); got != 3 {
		t.Errorf("invites: %d rows, want 3", got)
	}

	// Values came across column for column.
	var (
		name                     string
		started, ended, archived int64
		winner                   int
		pending, deckID, botTier sql.NullString
		expires, revoked         sql.NullInt64
		kind                     string
	)
	if err := d.QueryRow(`SELECT name, started_at, ended_at, archived_at, winner_seat FROM games WHERE id = 'g2'`).
		Scan(&name, &started, &ended, &archived, &winner); err != nil {
		t.Fatalf("read g2: %v", err)
	}
	if name != "Old" || started != 2100 || ended != 2200 || archived != 2300 || winner != 1 {
		t.Errorf("g2 = %q %d %d %d %d", name, started, ended, archived, winner)
	}
	if err := d.QueryRow(`SELECT pending_discord_id FROM seats WHERE game_id = 'g1' AND seat = 0`).Scan(&pending); err != nil {
		t.Fatalf("read seat: %v", err)
	}
	if pending.String != "123456789" {
		t.Errorf("pending_discord_id = %q, want 123456789", pending.String)
	}
	if err := d.QueryRow(`SELECT bot_tier FROM seats WHERE game_id = 'g1' AND seat = 1`).Scan(&botTier); err != nil {
		t.Fatalf("read bot seat: %v", err)
	}
	if botTier.String != "heuristic" {
		t.Errorf("bot_tier = %q", botTier.String)
	}
	if err := d.QueryRow(`SELECT deck_id FROM seats WHERE game_id = 'g2'`).Scan(&deckID); err != nil {
		t.Fatalf("read deck seat: %v", err)
	}
	if deckID.String != "deck-x" {
		t.Errorf("deck_id = %q, want deck-x", deckID.String)
	}
	if err := d.QueryRow(`SELECT kind, expires_at, revoked_at FROM invites WHERE token_hash = x'03'`).Scan(&kind, &expires, &revoked); err != nil {
		t.Fatalf("read invite: %v", err)
	}
	if kind != "player" || expires.Int64 != 9000 || revoked.Int64 != 2400 {
		t.Errorf("invite 03 = %q %v %v", kind, expires, revoked)
	}

	assertPersonForeignKeys(t, d)

	// The cascade from games is still wired after the rebuild.
	mustExec(t, d, `DELETE FROM games WHERE id = 'g1'`)
	for _, table := range []string{"seats", "invites"} {
		var n int
		if err := d.QueryRow(`SELECT COUNT(*) FROM ` + table + ` WHERE game_id = 'g1'`).Scan(&n); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if n != 0 {
			t.Errorf("after deleting g1, %s still has %d of its rows (ON DELETE CASCADE lost)", table, n)
		}
	}
}

// TestMigrationWithDanglingReferenceRollsBack pins the runner's
// foreign_key_check: with enforcement switched off for the rebuild, a
// value that references nothing must still stop the migration, and
// leave the database at the version it started from.
func TestMigrationWithDanglingReferenceRollsBack(t *testing.T) {
	d := openAtVersion(t, 2)
	mustExec(t, d, `INSERT INTO games (id, name, created_by, state, created_at) VALUES ('g1', 'FNM', 'no-such-user', 'lobby', 1)`)

	err := migrate(context.Background(), d)
	if err == nil || !strings.Contains(err.Error(), "dangling foreign key") {
		t.Fatalf("migrate: got %v, want a dangling-foreign-key refusal", err)
	}
	v, err := currentSchemaVersion(context.Background(), d)
	if err != nil {
		t.Fatalf("currentSchemaVersion: %v", err)
	}
	if v != 2 {
		t.Errorf("schema version after a refused migration = %d, want 2", v)
	}
	if got := countRows(t, d, "games"); got != 1 {
		t.Errorf("games: %d rows after rollback, want 1", got)
	}
	assertForeignKeysOnEveryConn(t, d)
}

// assertPersonForeignKeys checks the shape 0003 promises: the person
// columns reference users, the game columns still reference games,
// deck_id references nothing yet, the indexes are back, and the
// constraint is actually enforced on the pool's connections.
func assertPersonForeignKeys(t *testing.T, d *sql.DB) {
	t.Helper()

	if refs := referencesOf(t, d, "games"); refs["created_by"] != "users" {
		t.Errorf("games FKs = %v, want created_by -> users", refs)
	}
	refs := referencesOf(t, d, "seats")
	if refs["user_id"] != "users" || refs["game_id"] != "games" {
		t.Errorf("seats FKs = %v, want user_id -> users, game_id -> games", refs)
	}
	if _, ok := refs["deck_id"]; ok {
		t.Errorf("seats.deck_id must stay unenforced until sub-PR 5; FKs = %v", refs)
	}
	if refs := referencesOf(t, d, "invites"); refs["created_by"] != "users" || refs["game_id"] != "games" {
		t.Errorf("invites FKs = %v, want created_by -> users, game_id -> games", refs)
	}
	if refs := referencesOf(t, d, "identities"); refs["user_id"] != "users" {
		t.Errorf("identities FKs = %v, want user_id -> users", refs)
	}

	if got := strings.Join(indexesOn(t, d, "seats"), ","); got != "seats_pending_discord_id,seats_user_id" {
		t.Errorf("seats indexes = %s", got)
	}
	if got := strings.Join(indexesOn(t, d, "invites"), ","); got != "invites_game_id" {
		t.Errorf("invites indexes = %s", got)
	}
	// No temporary table left behind.
	var leftovers int
	if err := d.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE name LIKE '%\_new' ESCAPE '\'`).Scan(&leftovers); err != nil {
		t.Fatalf("leftovers: %v", err)
	}
	if leftovers != 0 {
		t.Errorf("%d *_new objects left in the schema", leftovers)
	}

	if _, err := d.Exec(`INSERT INTO games (id, name, created_by, state, created_at) VALUES ('gx', 'x', 'nobody', 'lobby', 1)`); err == nil {
		t.Error("games.created_by accepted a user that does not exist")
	}
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'Alice', 1, 1)`)
	mustExec(t, d, `INSERT INTO games (id, name, created_by, state, created_at) VALUES ('gy', 'y', 'u1', 'lobby', 1)`)
	if _, err := d.Exec(`INSERT INTO seats (game_id, seat, player_id, user_id) VALUES ('gy', 0, 'p', 'nobody')`); err == nil {
		t.Error("seats.user_id accepted a user that does not exist")
	}
	if _, err := d.Exec(`INSERT INTO invites (token_hash, game_id, kind, created_by, created_at) VALUES (x'ff', 'gy', 'player', 'nobody', 1)`); err == nil {
		t.Error("invites.created_by accepted a user that does not exist")
	}
	// deck_id is deliberately still free text.
	mustExec(t, d, `INSERT INTO seats (game_id, seat, player_id, user_id, deck_id) VALUES ('gy', 0, 'p', 'u1', 'no-such-deck')`)

	assertForeignKeysOnEveryConn(t, d)
}

// assertForeignKeysOnEveryConn holds several connections open at once
// so the pool cannot hand the same one back, and checks each has
// foreign_keys on: the migration's pinned connection must not return
// to the pool with enforcement off.
func assertForeignKeysOnEveryConn(t *testing.T, d *sql.DB) {
	t.Helper()
	ctx := context.Background()
	var conns []*sql.Conn
	defer func() {
		for _, c := range conns {
			_ = c.Close()
		}
	}()
	for i := 0; i < 4; i++ {
		c, err := d.Conn(ctx)
		if err != nil {
			t.Fatalf("Conn: %v", err)
		}
		conns = append(conns, c)
		var on int
		if err := c.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&on); err != nil {
			t.Fatalf("PRAGMA foreign_keys: %v", err)
		}
		if on != 1 {
			t.Errorf("connection %d has foreign_keys = %d after migrating, want 1", i, on)
		}
	}
}
