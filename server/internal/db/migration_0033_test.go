package db

import (
	"context"
	"testing"
)

// Tests for migration 0033 (pods, ADR 0220 §9): the four tables, a link
// hash and its expiry set or cleared together, one row per person per pod,
// deleting a pod deletes its members and invites, every row needs real
// users, and the hand rollback in docs/environments.md.

// podsVersion is this migration's number.
const podsVersion = 33

func TestMigration0033AddsTheFourTables(t *testing.T) {
	d := openAtVersion(t, podsVersion-1)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('a', 'A', 1, 1), ('b', 'B', 1, 1)`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate to latest: %v", err)
	}
	mustExec(t, d, `INSERT INTO pods (id, name, owner_id, created_at, updated_at) VALUES ('p', 'Thursday', 'a', 5, 5)`)
	mustExec(t, d, `INSERT INTO pod_members (pod_id, user_id, joined_at) VALUES ('p', 'a', 5)`)
	mustExec(t, d, `INSERT INTO pod_invites (pod_id, user_id, invited_by, created_at) VALUES ('p', 'b', 'a', 5)`)
	mustExec(t, d, `INSERT INTO pod_suggestion_dismissals (user_id, group_key, dismissed_at) VALUES ('a', 'k', 5)`)
	var hash, declined *string
	if err := d.QueryRow(`SELECT invite_hash FROM pods`).Scan(&hash); err != nil {
		t.Fatal(err)
	}
	if err := d.QueryRow(`SELECT declined_at FROM pod_invites`).Scan(&declined); err != nil {
		t.Fatal(err)
	}
	if hash != nil || declined != nil {
		t.Errorf("a new pod has a link %v, a new invite is declined %v; want NULL for both", hash, declined)
	}
	assertForeignKeysOnEveryConn(t, d)
}

func TestMigration0033Constraints(t *testing.T) {
	d := openAtVersion(t, 0)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('a', 'A', 1, 1), ('b', 'B', 1, 1)`)
	mustExec(t, d, `INSERT INTO pods (id, name, owner_id, invite_hash, invite_expires_at, created_at, updated_at)
		VALUES ('p', 'P', 'a', 'h1', 9, 5, 5), ('q', 'Q', 'b', NULL, NULL, 5, 5)`)
	mustExec(t, d, `INSERT INTO pod_members (pod_id, user_id, joined_at) VALUES ('p', 'a', 5), ('p', 'b', 6), ('q', 'b', 5)`)
	mustExec(t, d, `INSERT INTO pod_invites (pod_id, user_id, invited_by, created_at) VALUES ('q', 'a', 'b', 5)`)
	mustExec(t, d, `INSERT INTO pod_suggestion_dismissals (user_id, group_key, dismissed_at) VALUES ('a', 'k', 5)`)
	for _, bad := range []string{
		// A link hash and its expiry are set and cleared together.
		`INSERT INTO pods (id, name, owner_id, invite_hash, created_at, updated_at) VALUES ('r', 'R', 'a', 'h2', 5, 5)`,
		`INSERT INTO pods (id, name, owner_id, invite_expires_at, created_at, updated_at) VALUES ('r', 'R', 'a', 9, 5, 5)`,
		`UPDATE pods SET invite_hash = NULL WHERE id = 'p'`,
		// Two pods never share a link.
		`UPDATE pods SET invite_hash = 'h1', invite_expires_at = 9 WHERE id = 'q'`,
		`INSERT INTO pods (id, name, owner_id, created_at, updated_at) VALUES ('r', 'R', 'nobody', 5, 5)`,
		`INSERT INTO pods (id, owner_id, created_at, updated_at) VALUES ('r', 'a', 5, 5)`,
		`INSERT INTO pod_members (pod_id, user_id, joined_at) VALUES ('p', 'a', 7)`,
		`INSERT INTO pod_members (pod_id, user_id, joined_at) VALUES ('nope', 'a', 7)`,
		`INSERT INTO pod_members (pod_id, user_id, joined_at) VALUES ('q', 'nobody', 7)`,
		`INSERT INTO pod_invites (pod_id, user_id, invited_by, created_at) VALUES ('q', 'a', 'b', 6)`,
		`INSERT INTO pod_invites (pod_id, user_id, invited_by, created_at) VALUES ('q', 'a', 'nobody', 6)`,
		`INSERT INTO pod_invites (pod_id, user_id, invited_by, created_at) VALUES ('nope', 'a', 'b', 6)`,
		`INSERT INTO pod_suggestion_dismissals (user_id, group_key, dismissed_at) VALUES ('a', 'k', 6)`,
		`INSERT INTO pod_suggestion_dismissals (user_id, group_key, dismissed_at) VALUES ('nobody', 'k', 6)`,
	} {
		if _, err := d.Exec(bad); err == nil {
			t.Errorf("accepted: %s", bad)
		}
	}
	// Clearing both at once is the turned-off link.
	mustExec(t, d, `UPDATE pods SET invite_hash = NULL, invite_expires_at = NULL WHERE id = 'p'`)

	// Deleting a pod deletes its members and invites, and nothing else.
	mustExec(t, d, `DELETE FROM pods WHERE id = 'q'`)
	for q, want := range map[string]int{
		`SELECT COUNT(*) FROM pod_members WHERE pod_id = 'q'`: 0,
		`SELECT COUNT(*) FROM pod_invites`:                    0,
		`SELECT COUNT(*) FROM pod_members WHERE pod_id = 'p'`: 2,
		`SELECT COUNT(*) FROM pod_suggestion_dismissals`:      1,
	} {
		var n int
		if err := d.QueryRow(q).Scan(&n); err != nil || n != want {
			t.Errorf("%s = %d (%v), want %d", q, n, err, want)
		}
	}
}

func TestMigration0033RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, podsVersion)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('a', 'A', 1, 1)`)
	mustExec(t, d, `INSERT INTO pods (id, name, owner_id, created_at, updated_at) VALUES ('p', 'P', 'a', 5, 5)`)
	mustExec(t, d, `INSERT INTO pod_members (pod_id, user_id, joined_at) VALUES ('p', 'a', 5)`)
	mustExec(t, d, `DROP TABLE pod_suggestion_dismissals`)
	mustExec(t, d, `DROP TABLE pod_invites`)
	mustExec(t, d, `DROP TABLE pod_members`)
	mustExec(t, d, `DROP TABLE pods`)
	mustExec(t, d, `DELETE FROM schema_migrations WHERE version = ?`, podsVersion)
	if v, _ := currentSchemaVersion(ctx, d); v != podsVersion-1 {
		t.Fatalf("after the rollback the version is %d, want %d", v, podsVersion-1)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("roll forward again: %v", err)
	}
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM pods`).Scan(&n); err != nil || n != 0 {
		t.Errorf("rolled forward: %d pods (%v); want an empty table", n, err)
	}
}
