package db

import (
	"context"
	"testing"
)

// Tests for migration 0007 (games.outcome, ADR 0057 Decision 7, #1520):
// one nullable column, NULL on every existing row.

func TestMigration0007FromEmpty(t *testing.T) {
	d := openAtVersion(t, 0)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'G', 'lobby', 1)`)
	var outcome *string
	if err := d.QueryRow(`SELECT outcome FROM games WHERE id = 'g1'`).Scan(&outcome); err != nil {
		t.Fatalf("select outcome: %v", err)
	}
	if outcome != nil {
		t.Errorf("outcome = %q on a fresh row, want NULL", *outcome)
	}
}

func TestMigration0007PreservesAPopulatedV6Database(t *testing.T) {
	d := openAtVersion(t, 6)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at, ended_at, winner_seat) VALUES ('g1', 'Done', 'ended', 1, 2, 1)`)

	if err := migrateTo(context.Background(), d, 7); err != nil {
		t.Fatalf("migrateTo(7): %v", err)
	}
	var (
		winner  int
		outcome *string
	)
	if err := d.QueryRow(`SELECT winner_seat, outcome FROM games WHERE id = 'g1'`).Scan(&winner, &outcome); err != nil {
		t.Fatalf("select: %v", err)
	}
	if winner != 1 || outcome != nil {
		t.Errorf("winner_seat=%d outcome=%v, want 1 and NULL (unknown) for a pre-0007 row", winner, outcome)
	}
	mustExec(t, d, `UPDATE games SET outcome = 'draw' WHERE id = 'g1'`)
	assertForeignKeysOnEveryConn(t, d)
}
