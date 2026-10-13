package db

import (
	"context"
	"database/sql"
	"testing"
)

// Tests for migration 0015 (users.deleted_at, #2983): a v14 database
// gains the column, every existing user reads NULL (live), and the
// column takes a deletion time.

func TestMigration0015AddsDeletedAtAsLive(t *testing.T) {
	d := openAtVersion(t, 14)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'Alice', 1, 1)`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v14 -> latest: %v", err)
	}
	var deleted sql.NullInt64
	if err := d.QueryRow(`SELECT deleted_at FROM users WHERE id = 'u1'`).Scan(&deleted); err != nil {
		t.Fatalf("read deleted_at: %v", err)
	}
	if deleted.Valid {
		t.Errorf("an existing user reads deleted_at = %d, want NULL (live)", deleted.Int64)
	}
	mustExec(t, d, `UPDATE users SET deleted_at = 1700000000000 WHERE id = 'u1'`)
	if err := d.QueryRow(`SELECT deleted_at FROM users WHERE id = 'u1'`).Scan(&deleted); err != nil || deleted.Int64 != 1700000000000 {
		t.Errorf("deleted_at = %v, %v", deleted, err)
	}
}
