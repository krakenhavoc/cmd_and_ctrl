package db

import (
	"context"
	"testing"
)

// Tests for migration 0034 (scheduled tables, ADR 0215 slice 2): the two
// tables, the values a plan and an RSVP may take, a hold and its end set
// together, one RSVP per person per plan, deleting a plan deletes its
// list, deleting its table only forgets it, and the hand rollback in
// docs/environments.md.

// scheduledVersion is this migration's number.
const scheduledVersion = 34

func TestMigration0034AddsTheTwoTables(t *testing.T) {
	d := openAtVersion(t, scheduledVersion-1)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('a', 'A', 1, 1)`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate to latest: %v", err)
	}
	mustExec(t, d, `INSERT INTO scheduled_tables (id, host_user_id, name, starts_at, length_minutes, zone, players, status, created_at, updated_at)
		VALUES ('p', 'a', 'Friday', 10, 180, 'UTC', 4, 'scheduled', 5, 5)`)
	mustExec(t, d, `INSERT INTO scheduled_rsvps (schedule_id, discord_id, user_id, source, state, rsvp_at, updated_at)
		VALUES ('p', '1', 'a', 'host', 'seat', 5, 5)`)
	var opts string
	var moves int
	if err := d.QueryRow(`SELECT options, moves FROM scheduled_tables`).Scan(&opts, &moves); err != nil {
		t.Fatal(err)
	}
	if opts != "{}" || moves != 0 {
		t.Errorf("a new plan: options %q moves %d; want {} and 0", opts, moves)
	}
	assertForeignKeysOnEveryConn(t, d)
}

func TestMigration0034Constraints(t *testing.T) {
	d := openAtVersion(t, 0)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('a', 'A', 1, 1)`)
	mustExec(t, d, `INSERT INTO games (id, name, state, created_at) VALUES ('g', 'T', 'lobby', 1)`)
	mustExec(t, d, `INSERT INTO scheduled_tables (id, host_user_id, name, starts_at, length_minutes, zone, players, status, game_id, created_at, updated_at)
		VALUES ('p', 'a', 'P', 10, 120, 'UTC', 2, 'open', 'g', 5, 5), ('q', 'a', 'Q', 10, 240, 'UTC', 3, 'scheduled', NULL, 5, 5)`)
	mustExec(t, d, `INSERT INTO scheduled_rsvps (schedule_id, discord_id, user_id, source, state, rsvp_at, held_until, updated_at)
		VALUES ('p', '1', 'a', 'host', 'held', 5, 20, 5), ('q', '2', NULL, 'discord', 'waiting', 5, NULL, 5)`)
	plan := `INSERT INTO scheduled_tables (id, host_user_id, name, starts_at, length_minutes, zone, players, status, created_at, updated_at) VALUES `
	rsvp := `INSERT INTO scheduled_rsvps (schedule_id, discord_id, user_id, source, state, rsvp_at, held_until, updated_at) VALUES `
	for _, bad := range []string{
		plan + `('r', 'nobody', 'R', 10, 180, 'UTC', 4, 'scheduled', 5, 5)`,
		plan + `('r', 'a', '', 10, 180, 'UTC', 4, 'scheduled', 5, 5)`,
		plan + `('r', 'a', 'R', 10, 90, 'UTC', 4, 'scheduled', 5, 5)`,
		plan + `('r', 'a', 'R', 10, 180, 'UTC', 5, 'scheduled', 5, 5)`,
		plan + `('r', 'a', 'R', 10, 180, 'UTC', 1, 'scheduled', 5, 5)`,
		plan + `('r', 'a', 'R', 10, 180, 'UTC', 4, 'later', 5, 5)`,
		`UPDATE scheduled_tables SET moves = 4 WHERE id = 'q'`,
		`UPDATE scheduled_tables SET game_id = 'nope' WHERE id = 'q'`,
		// One row per person per plan.
		rsvp + `('p', '1', 'a', 'site', 'waiting', 6, NULL, 6)`,
		rsvp + `('nope', '3', NULL, 'site', 'waiting', 6, NULL, 6)`,
		rsvp + `('q', '3', 'nobody', 'site', 'waiting', 6, NULL, 6)`,
		rsvp + `('q', '', NULL, 'site', 'waiting', 6, NULL, 6)`,
		rsvp + `('q', '3', NULL, 'bot', 'waiting', 6, NULL, 6)`,
		rsvp + `('q', '3', NULL, 'site', 'maybe', 6, NULL, 6)`,
		// A hold has an end, and only a hold has one.
		rsvp + `('q', '3', NULL, 'site', 'held', 6, NULL, 6)`,
		rsvp + `('q', '3', NULL, 'site', 'waiting', 6, 9, 6)`,
	} {
		if _, err := d.Exec(bad); err == nil {
			t.Errorf("accepted: %s", bad)
		}
	}
	// Deleting a plan's table forgets it; deleting a plan deletes its list.
	mustExec(t, d, `DELETE FROM games WHERE id = 'g'`)
	var game *string
	if err := d.QueryRow(`SELECT game_id FROM scheduled_tables WHERE id = 'p'`).Scan(&game); err != nil || game != nil {
		t.Errorf("after its table went: game_id %v (%v), want NULL", game, err)
	}
	mustExec(t, d, `DELETE FROM scheduled_tables WHERE id = 'q'`)
	for q, want := range map[string]int{
		`SELECT COUNT(*) FROM scheduled_rsvps WHERE schedule_id = 'q'`: 0,
		`SELECT COUNT(*) FROM scheduled_rsvps WHERE schedule_id = 'p'`: 1,
	} {
		var n int
		if err := d.QueryRow(q).Scan(&n); err != nil || n != want {
			t.Errorf("%s = %d (%v), want %d", q, n, err, want)
		}
	}
}

func TestMigration0034RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, scheduledVersion)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('a', 'A', 1, 1)`)
	mustExec(t, d, `INSERT INTO scheduled_tables (id, host_user_id, name, starts_at, length_minutes, zone, players, status, created_at, updated_at)
		VALUES ('p', 'a', 'P', 10, 180, 'UTC', 4, 'scheduled', 5, 5)`)
	mustExec(t, d, `DROP TABLE scheduled_rsvps`)
	mustExec(t, d, `DROP TABLE scheduled_tables`)
	mustExec(t, d, `DELETE FROM schema_migrations WHERE version = ?`, scheduledVersion)
	if v, _ := currentSchemaVersion(ctx, d); v != scheduledVersion-1 {
		t.Fatalf("after the rollback the version is %d, want %d", v, scheduledVersion-1)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("roll forward again: %v", err)
	}
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM scheduled_tables`).Scan(&n); err != nil || n != 0 {
		t.Errorf("rolled forward: %d plans (%v); want an empty table", n, err)
	}
}
