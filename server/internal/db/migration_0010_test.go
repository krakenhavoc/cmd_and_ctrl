package db

import (
	"context"
	"database/sql"
	"testing"
)

// Tests for migration 0010 (seats.agent_client, ADR 0124 §6): one
// nullable column, NULL on every existing row, no backfill.

func TestMigration0010FromEmpty(t *testing.T) {
	d := openAtVersion(t, 0)
	assertSeatAgentShape(t, d)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'G', 'lobby', 1)`)
	mustExec(t, d, `INSERT INTO seats (game_id, seat, player_id, guest_name) VALUES ('g1', 0, 'p1', 'Alice')`)
	mustExec(t, d, `INSERT INTO seats (game_id, seat, player_id, guest_name, agent_client) VALUES ('g1', 1, 'p2', 'Claude', 'claude-code')`)
	if got := agentClient(t, d, "p1"); got != nil {
		t.Errorf("a person's seat has agent_client %q, want NULL", *got)
	}
	if got := agentClient(t, d, "p2"); got == nil || *got != "claude-code" {
		t.Errorf("the agent seat's agent_client = %v, want claude-code", got)
	}
}

// Every seat that exists when this deploys reads NULL: there is no
// backfill, and nothing else about the row moves.
func TestMigration0010PreservesAPopulatedV9Database(t *testing.T) {
	d := openAtVersion(t, 9)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'G', 'active', 1)`)
	mustExec(t, d, `INSERT INTO seats (game_id, seat, player_id, guest_name, bot_tier, deck_name)
		VALUES ('g1', 0, 'p1', 'Alice', NULL, 'Atraxa'),
		       ('g1', 1, 'p2', 'Bot 1', 'heuristic', 'Mono Red')`)

	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v9 -> latest: %v", err)
	}
	if n := countRows(t, d, "seats"); n != 2 {
		t.Errorf("seats: %d rows, want 2", n)
	}
	for _, p := range []string{"p1", "p2"} {
		if got := agentClient(t, d, p); got != nil {
			t.Errorf("%s agent_client = %q after the migration, want NULL", p, *got)
		}
	}
	var tier, deck sql.NullString
	if err := d.QueryRow(`SELECT bot_tier, deck_name FROM seats WHERE player_id = 'p2'`).Scan(&tier, &deck); err != nil {
		t.Fatal(err)
	}
	if tier.String != "heuristic" || deck.String != "Mono Red" {
		t.Errorf("p2 lost its columns: bot_tier %v, deck_name %v", tier, deck)
	}
	assertSeatAgentShape(t, d)
}

// An older binary refuses a v10 database, and the hand rollback
// documented in docs/environments.md leaves a database that rolls
// forward cleanly, with every seat back at NULL.
func TestMigration0010RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, 10)
	if v, err := currentSchemaVersion(ctx, d); err != nil || v < 10 {
		t.Fatalf("version = %d (%v), want >= 10", v, err)
	}
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g1', 'G', 'lobby', 1)`)
	mustExec(t, d, `INSERT INTO seats (game_id, seat, player_id, guest_name, agent_client) VALUES ('g1', 0, 'p1', 'Claude', 'codex')`)
	for _, q := range []string{
		`ALTER TABLE seats DROP COLUMN agent_client`,
		`DELETE FROM schema_migrations WHERE version = 10`,
	} {
		mustExec(t, d, q)
	}
	if v, _ := currentSchemaVersion(ctx, d); v != 9 {
		t.Fatalf("after the rollback the version is %d, want 9", v)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	assertSeatAgentShape(t, d)
	if got := agentClient(t, d, "p1"); got != nil {
		t.Errorf("after the round trip agent_client = %q, want NULL (no backfill)", *got)
	}
}

func agentClient(t *testing.T, d *sql.DB, playerID string) *string {
	t.Helper()
	var c *string
	if err := d.QueryRow(`SELECT agent_client FROM seats WHERE player_id = ?`, playerID).Scan(&c); err != nil {
		t.Fatalf("read agent_client for %s: %v", playerID, err)
	}
	return c
}

func assertSeatAgentShape(t *testing.T, d *sql.DB) {
	t.Helper()
	var (
		typ     string
		notNull int
		dflt    *string
	)
	err := d.QueryRow(`SELECT type, "notnull", dflt_value FROM pragma_table_info('seats') WHERE name = 'agent_client'`).Scan(&typ, &notNull, &dflt)
	if err != nil {
		t.Fatalf("seats.agent_client missing: %v", err)
	}
	if typ != "TEXT" || notNull != 0 || dflt != nil {
		t.Errorf("seats.agent_client = %s notnull=%d default=%v, want a nullable TEXT with no default", typ, notNull, dflt)
	}
	assertForeignKeysOnEveryConn(t, d)
}
