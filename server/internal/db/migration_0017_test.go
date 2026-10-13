package db

import (
	"context"
	"testing"
)

// Tests for migration 0017 (table_blocks.entry_id / label and the
// 'spectator' kind, #3113): the rows 0016 wrote survive the rebuild, are
// labelled by what the list shows, and a signed-in person's two rows
// read as one entry.

func TestMigration0017BackfillsEntriesAndLabels(t *testing.T) {
	d := openAtVersion(t, 16)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'Alice', 1, 1)`)
	mustExec(t, d, `INSERT INTO identities (provider, subject, user_id, display_name, scopes, linked_at)
		VALUES ('discord', '555', 'u1', 'alice#1', 'identify', 1)`)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'Table', 'lobby', 1)`)
	mustExec(t, d, `INSERT INTO table_blocks (game_id, kind, subject, created_at) VALUES ('g1', 'user', 'u1', 10)`)
	mustExec(t, d, `INSERT INTO table_blocks (game_id, kind, subject, created_at) VALUES ('g1', 'discord', '555', 10)`)
	mustExec(t, d, `INSERT INTO table_blocks (game_id, kind, subject, created_at) VALUES ('g1', 'discord', '777', 11)`)
	mustExec(t, d, `INSERT INTO table_blocks (game_id, kind, subject, created_at) VALUES ('g1', 'name', 'bob', 12)`)

	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v16 -> latest: %v", err)
	}
	if n := countRows(t, d, "table_blocks"); n != 4 {
		t.Fatalf("rows after the rebuild: %d, want 4", n)
	}

	type row struct{ entry, label string }
	get := func(kind, subject string) row {
		var r row
		if err := d.QueryRow(`SELECT entry_id, label FROM table_blocks WHERE game_id = 'g1' AND kind = ? AND subject = ?`,
			kind, subject).Scan(&r.entry, &r.label); err != nil {
			t.Fatalf("read %s/%s: %v", kind, subject, err)
		}
		return r
	}
	user, disc, orphan, name := get("user", "u1"), get("discord", "555"), get("discord", "777"), get("name", "bob")
	if user.label != "Alice" || disc.label != "Alice" {
		t.Errorf("account labels = %q / %q, want both Alice (a snowflake is never a label)", user.label, disc.label)
	}
	if user.entry == "" || user.entry != disc.entry {
		t.Errorf("one removal's account and Discord rows have entries %q and %q, want the same", user.entry, disc.entry)
	}
	if orphan.label != "Discord account" || orphan.entry == user.entry || orphan.entry == "" {
		t.Errorf("a Discord row that joins no account = %+v, want label %q and an entry of its own", orphan, "Discord account")
	}
	if name.label != "bob" || name.entry == "" || name.entry == user.entry {
		t.Errorf("guest row = %+v, want label bob and an entry of its own", name)
	}
}

func TestMigration0017TakesSpectatorMarkersAndStillCascades(t *testing.T) {
	d := openAtVersion(t, 17)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'Table', 'lobby', 1)`)
	mustExec(t, d, `INSERT INTO table_blocks (game_id, kind, subject, created_at, entry_id)
		VALUES ('g1', 'spectator', 'abc', 1, 'e1')`)
	if _, err := d.Exec(`INSERT INTO table_blocks (game_id, kind, subject, created_at) VALUES ('g1', 'name', 'x', 1)`); err == nil {
		t.Error("a block with no entry_id was accepted")
	}
	mustExec(t, d, `DELETE FROM games WHERE id = 'g1'`)
	if n := countRows(t, d, "table_blocks"); n != 0 {
		t.Errorf("blocks after the table went: %d, want 0 (ON DELETE CASCADE)", n)
	}
}
