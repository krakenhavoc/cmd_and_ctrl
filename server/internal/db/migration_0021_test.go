package db

import (
	"context"
	"database/sql"
	"testing"
)

// Tests for migration 0021 (a table's #lfg post, ADR 0161 §6):
// games.lfg defaults to 0 on every row old and new and takes only 0 or
// 1; games.lfg_message_id is nullable; and the hand rollback in
// docs/environments.md.

func TestMigration0021PreservesAPopulatedV20DatabaseWithLFGOff(t *testing.T) {
	d := openAtVersion(t, 20)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at, voice_room, voice_channel_id)
		VALUES ('g1', 'Friday', 'lobby', 1, 1, '444')`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v20 -> latest: %v", err)
	}
	assertLFGShape(t, d)
	lfg, message := lfgColumns(t, d, "g1")
	if lfg != 0 || message.Valid {
		t.Errorf("an existing game after the migration: lfg %d, message %v; want 0 and NULL", lfg, message)
	}
	var (
		name    string
		voice   int
		channel string
	)
	if err := d.QueryRow(`SELECT name, voice_room, voice_channel_id FROM games WHERE id = 'g1'`).Scan(&name, &voice, &channel); err != nil {
		t.Fatal(err)
	}
	if name != "Friday" || voice != 1 || channel != "444" {
		t.Errorf("g1 lost its columns: %q %d %q", name, voice, channel)
	}
}

func TestMigration0021AcceptsOnlyOnOrOff(t *testing.T) {
	d := openAtVersion(t, 0)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'A', 'lobby', 1)`)
	mustExec(t, d, `UPDATE games SET lfg = 1, lfg_message_id = '900' WHERE id = 'g1'`)
	mustExec(t, d, `UPDATE games SET lfg = 0, lfg_message_id = NULL WHERE id = 'g1'`)
	for _, bad := range []any{2, -1, "yes"} {
		if _, err := d.Exec(`UPDATE games SET lfg = ? WHERE id = 'g1'`, bad); err == nil {
			t.Errorf("lfg %v accepted", bad)
		}
	}
	if _, err := d.Exec(`UPDATE games SET lfg = NULL WHERE id = 'g1'`); err == nil {
		t.Error("lfg NULL accepted")
	}
}

func TestMigration0021RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, 21)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at, lfg, lfg_message_id)
		VALUES ('g1', 'A', 'lobby', 1, 1, '900')`)
	for _, q := range []string{
		`ALTER TABLE games DROP COLUMN lfg`,
		`ALTER TABLE games DROP COLUMN lfg_message_id`,
		`DELETE FROM schema_migrations WHERE version = 21`,
	} {
		mustExec(t, d, q)
	}
	if v, _ := currentSchemaVersion(ctx, d); v != 20 {
		t.Fatalf("after the rollback the version is %d, want 20", v)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	assertLFGShape(t, d)
	if lfg, message := lfgColumns(t, d, "g1"); lfg != 0 || message.Valid {
		t.Errorf("after the round trip: %d %v, want lfg off and no message", lfg, message)
	}
}

func lfgColumns(t *testing.T, d *sql.DB, id string) (int, sql.NullString) {
	t.Helper()
	var (
		lfg     int
		message sql.NullString
	)
	if err := d.QueryRow(`SELECT lfg, lfg_message_id FROM games WHERE id = ?`, id).Scan(&lfg, &message); err != nil {
		t.Fatalf("read lfg columns for %s: %v", id, err)
	}
	return lfg, message
}

func assertLFGShape(t *testing.T, d *sql.DB) {
	t.Helper()
	for col, want := range map[string]struct {
		typ     string
		notNull int
	}{
		"lfg":            {"INTEGER", 1},
		"lfg_message_id": {"TEXT", 0},
	} {
		var (
			typ     string
			notNull int
		)
		if err := d.QueryRow(`SELECT type, "notnull" FROM pragma_table_info('games') WHERE name = ?`, col).Scan(&typ, &notNull); err != nil {
			t.Fatalf("games.%s missing: %v", col, err)
		}
		if typ != want.typ || notNull != want.notNull {
			t.Errorf("games.%s = %s notnull=%d, want %s notnull=%d", col, typ, notNull, want.typ, want.notNull)
		}
	}
	assertForeignKeysOnEveryConn(t, d)
}
