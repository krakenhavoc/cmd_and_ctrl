package db

import (
	"context"
	"testing"
)

// Tests for migration 0014 (bot_model_budget, #2980): the table exists
// after migrating a v13 database, and its key is unique.

func TestMigration0014CreatesTheBudgetTable(t *testing.T) {
	d := openAtVersion(t, 13)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v13 -> latest: %v", err)
	}
	mustExec(t, d, `INSERT INTO bot_model_budget (key, value, updated_at) VALUES ('day:2026-10-12', 5, 1)`)
	if _, err := d.Exec(`INSERT INTO bot_model_budget (key, value, updated_at) VALUES ('day:2026-10-12', 6, 2)`); err == nil {
		t.Fatal("a second row for one key must be refused: the store upserts")
	}
	var v int64
	if err := d.QueryRow(`SELECT value FROM bot_model_budget WHERE key = 'day:2026-10-12'`).Scan(&v); err != nil || v != 5 {
		t.Fatalf("value = %d, %v", v, err)
	}
}
