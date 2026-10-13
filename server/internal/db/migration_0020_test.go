package db

import (
	"context"
	"database/sql"
	"testing"
)

// Tests for migration 0020 (per-table Discord voice rooms, ADR 0159
// §5): games.voice_room defaults to 0 on every row old and new and
// takes only 0 or 1; games.voice_channel_id is nullable; and the hand
// rollback in docs/environments.md.

func TestMigration0020PreservesAPopulatedV19DatabaseWithVoiceOff(t *testing.T) {
	d := openAtVersion(t, 19)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at, host_discord_id) VALUES ('g1', 'Friday', 'active', 1, '501')`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v19 -> latest: %v", err)
	}
	assertVoiceShape(t, d)
	room, channel := voiceColumns(t, d, "g1")
	if room != 0 || channel.Valid {
		t.Errorf("an existing game after the migration: voice_room %d, channel %v; want 0 and NULL", room, channel)
	}
	var name, host string
	if err := d.QueryRow(`SELECT name, host_discord_id FROM games WHERE id = 'g1'`).Scan(&name, &host); err != nil {
		t.Fatal(err)
	}
	if name != "Friday" || host != "501" {
		t.Errorf("g1 lost its columns: %q %q", name, host)
	}
}

func TestMigration0020AcceptsOnlyOnOrOff(t *testing.T) {
	d := openAtVersion(t, 0)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'A', 'lobby', 1)`)
	mustExec(t, d, `UPDATE games SET voice_room = 1, voice_channel_id = '444' WHERE id = 'g1'`)
	mustExec(t, d, `UPDATE games SET voice_room = 0, voice_channel_id = NULL WHERE id = 'g1'`)
	for _, bad := range []any{2, -1, "yes"} {
		if _, err := d.Exec(`UPDATE games SET voice_room = ? WHERE id = 'g1'`, bad); err == nil {
			t.Errorf("voice_room %v accepted", bad)
		}
	}
	if _, err := d.Exec(`UPDATE games SET voice_room = NULL WHERE id = 'g1'`); err == nil {
		t.Error("voice_room NULL accepted")
	}
}

func TestMigration0020RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, 20)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at, voice_room, voice_channel_id)
		VALUES ('g1', 'A', 'active', 1, 1, '444')`)
	for _, q := range []string{
		`ALTER TABLE games DROP COLUMN voice_room`,
		`ALTER TABLE games DROP COLUMN voice_channel_id`,
		`DELETE FROM schema_migrations WHERE version = 20`,
	} {
		mustExec(t, d, q)
	}
	if v, _ := currentSchemaVersion(ctx, d); v != 19 {
		t.Fatalf("after the rollback the version is %d, want 19", v)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	assertVoiceShape(t, d)
	if room, channel := voiceColumns(t, d, "g1"); room != 0 || channel.Valid {
		t.Errorf("after the round trip: %d %v, want voice off and no channel", room, channel)
	}
}

func voiceColumns(t *testing.T, d *sql.DB, id string) (int, sql.NullString) {
	t.Helper()
	var (
		room    int
		channel sql.NullString
	)
	if err := d.QueryRow(`SELECT voice_room, voice_channel_id FROM games WHERE id = ?`, id).Scan(&room, &channel); err != nil {
		t.Fatalf("read voice columns for %s: %v", id, err)
	}
	return room, channel
}

func assertVoiceShape(t *testing.T, d *sql.DB) {
	t.Helper()
	for col, want := range map[string]struct {
		typ     string
		notNull int
	}{
		"voice_room":       {"INTEGER", 1},
		"voice_channel_id": {"TEXT", 0},
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
