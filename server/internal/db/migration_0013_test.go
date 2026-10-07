package db

import (
	"context"
	"database/sql"
	"testing"
)

// Tests for migration 0013 (user_playmats, ADR 0128 §11): the table,
// the slot check, and the backfill that makes every existing active
// pointer a saved slot 1.

// unmatchedPointers counts users whose active playmat is not one of
// their own saved slots: the state ADR 0128 §11 makes impossible.
func unmatchedPointers(t *testing.T, d *sql.DB) int {
	t.Helper()
	var n int
	err := d.QueryRow(`SELECT COUNT(*) FROM users u WHERE u.playmat_id IS NOT NULL AND NOT EXISTS
		(SELECT 1 FROM user_playmats p WHERE p.user_id = u.id AND p.playmat_id = u.playmat_id)`).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestMigration0013BackfillsEveryExistingPlaymatIntoSlotOne(t *testing.T) {
	d := openAtVersion(t, 12)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at, playmat_id, playmat_wash)
		VALUES ('u1', 'Alice', 1, 1, 'mat-1', 70), ('u2', 'Bob', 1, 1, NULL, NULL),
		       ('u3', 'Cy', 1, 1, 'mat-3', NULL), ('u4', 'Di', 1, 1, '', NULL)`)

	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v12 -> latest: %v", err)
	}

	rows, err := d.Query(`SELECT user_id, slot, playmat_id, width, height, created_at FROM user_playmats ORDER BY user_id`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	got := map[string]string{}
	for rows.Next() {
		var (
			user, mat string
			slot      int
			w, h      *int
			at        int64
		)
		if err := rows.Scan(&user, &slot, &mat, &w, &h, &at); err != nil {
			t.Fatal(err)
		}
		if slot != 1 {
			t.Errorf("%s backfilled into slot %d, want 1", user, slot)
		}
		if w != nil || h != nil {
			t.Errorf("%s has stored dimensions %v x %v, want NULL (read from the file)", user, w, h)
		}
		if at <= 0 {
			t.Errorf("%s created_at = %d, want a Unix-millisecond time", user, at)
		}
		got[user] = mat
	}
	if len(got) != 2 || got["u1"] != "mat-1" || got["u3"] != "mat-3" {
		t.Errorf("backfilled rows = %v, want u1 and u3 only", got)
	}
	if n := unmatchedPointers(t, d); n != 0 {
		t.Errorf("%d users have an active pointer matching no saved row after the migration", n)
	}
	// What the migration does not touch.
	var active, wash sql.NullString
	if err := d.QueryRow(`SELECT playmat_id, CAST(playmat_wash AS TEXT) FROM users WHERE id = 'u1'`).Scan(&active, &wash); err != nil {
		t.Fatal(err)
	}
	if active.String != "mat-1" || wash.String != "70" {
		t.Errorf("u1 active = %q wash = %q, want the pointer and the wash kept", active.String, wash.String)
	}
	// An empty-string pointer cannot survive as an unmatched pointer.
	if err := d.QueryRow(`SELECT playmat_id FROM users WHERE id = 'u4'`).Scan(&active); err != nil || active.Valid {
		t.Errorf("u4 playmat_id = %v (%v), want NULL", active, err)
	}
}

func TestMigration0013FromEmptyHasTheTableAndTheSlotCheck(t *testing.T) {
	d := openAtVersion(t, 0)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	for _, slot := range []int{1, 2, 3} {
		mustExec(t, d, `INSERT INTO user_playmats (user_id, slot, playmat_id, created_at) VALUES ('u1', ?, ?, 1)`, slot, "m"+string(rune('0'+slot)))
	}
	for _, slot := range []int{0, 4, -1} {
		if _, err := d.Exec(`INSERT INTO user_playmats (user_id, slot, playmat_id, created_at) VALUES ('u1', ?, 'x', 1)`, slot); err == nil {
			t.Errorf("slot %d was accepted", slot)
		}
	}
	// One row per (user, slot), and one slot per image.
	if _, err := d.Exec(`INSERT INTO user_playmats (user_id, slot, playmat_id, created_at) VALUES ('u1', 1, 'again', 1)`); err == nil {
		t.Error("a second row for the same slot was accepted")
	}
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u2', 'B', 1, 1)`)
	if _, err := d.Exec(`INSERT INTO user_playmats (user_id, slot, playmat_id, created_at) VALUES ('u2', 1, 'm1', 1)`); err == nil {
		t.Error("one image in two accounts' slots was accepted")
	}
	if _, err := d.Exec(`INSERT INTO user_playmats (user_id, slot, playmat_id, created_at) VALUES ('nobody', 1, 'm9', 1)`); err == nil {
		t.Error("a row for a user that does not exist was accepted")
	}
}

func TestMigration0013RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, 13)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at, playmat_id) VALUES ('u1', 'A', 1, 1, 'mat-1')`)
	mustExec(t, d, `INSERT INTO user_playmats (user_id, slot, playmat_id, created_at) VALUES ('u1', 1, 'mat-1', 1), ('u1', 2, 'mat-2', 1)`)
	for _, q := range []string{
		`DROP TABLE user_playmats`,
		`DELETE FROM schema_migrations WHERE version = 13`,
	} {
		mustExec(t, d, q)
	}
	if v, _ := currentSchemaVersion(ctx, d); v != 12 {
		t.Fatalf("after the rollback the version is %d, want 12", v)
	}
	var active string
	if err := d.QueryRow(`SELECT playmat_id FROM users WHERE id = 'u1'`).Scan(&active); err != nil || active != "mat-1" {
		t.Errorf("after the rollback the active pointer = %q (%v), want mat-1: an older binary must find its playmat", active, err)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	// Rolling forward backfills the pointer into slot 1 again; slot 2 was
	// dropped with the table.
	if n := countRows(t, d, "user_playmats"); n != 1 {
		t.Errorf("user_playmats has %d rows after the round trip, want 1 (the active mat)", n)
	}
	if n := unmatchedPointers(t, d); n != 0 {
		t.Errorf("%d unmatched pointers after the round trip", n)
	}
}
