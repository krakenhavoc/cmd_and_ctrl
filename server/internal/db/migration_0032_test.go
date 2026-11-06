package db

import (
	"context"
	"testing"
)

// Tests for migration 0032 (friends and blocks, ADR 0213 §2, §3 and
// §10): a pair has one friendship row in order, a request and a block
// cannot name the same person twice, every row needs real users, the
// settings take only their named values and default to the ADR's, and
// the hand rollback in docs/environments.md.

// friendsVersion is this migration's number.
const friendsVersion = 32

func TestMigration0032AddsTheFourTables(t *testing.T) {
	d := openAtVersion(t, friendsVersion-1)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('a', 'A', 1, 1), ('b', 'B', 1, 1)`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate to latest: %v", err)
	}
	mustExec(t, d, `INSERT INTO friendships (user_a, user_b, created_at) VALUES ('a', 'b', 5)`)
	mustExec(t, d, `INSERT INTO friend_requests (from_user, to_user, created_at) VALUES ('a', 'b', 5)`)
	mustExec(t, d, `INSERT INTO user_blocks (user_id, blocked_user_id, created_at) VALUES ('b', 'a', 5)`)
	mustExec(t, d, `INSERT INTO social_prefs (user_id, updated_at) VALUES ('a', 5)`)
	var presence, requests string
	if err := d.QueryRow(`SELECT presence, requests FROM social_prefs WHERE user_id = 'a'`).Scan(&presence, &requests); err != nil {
		t.Fatal(err)
	}
	if presence != "friends" || requests != "played_with" {
		t.Errorf("a new settings row: presence %q requests %q; want friends and played_with (ADR 0213 §2, §5)", presence, requests)
	}
	var ignored *int64
	if err := d.QueryRow(`SELECT ignored_at FROM friend_requests`).Scan(&ignored); err != nil {
		t.Fatal(err)
	}
	if ignored != nil {
		t.Errorf("a new request is ignored at %v; want NULL", *ignored)
	}
	assertForeignKeysOnEveryConn(t, d)
}

func TestMigration0032Constraints(t *testing.T) {
	d := openAtVersion(t, 0)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('a', 'A', 1, 1), ('b', 'B', 1, 1)`)
	mustExec(t, d, `INSERT INTO friendships (user_a, user_b, created_at) VALUES ('a', 'b', 5)`)
	mustExec(t, d, `INSERT INTO friend_requests (from_user, to_user, created_at) VALUES ('a', 'b', 5)`)
	mustExec(t, d, `INSERT INTO user_blocks (user_id, blocked_user_id, created_at) VALUES ('a', 'b', 5)`)
	mustExec(t, d, `INSERT INTO social_prefs (user_id, presence, requests, updated_at) VALUES ('a', 'nobody', 'nobody', 5)`)
	for _, bad := range []string{
		// One row per pair, smaller id first.
		`INSERT INTO friendships (user_a, user_b, created_at) VALUES ('b', 'a', 5)`,
		`INSERT INTO friendships (user_a, user_b, created_at) VALUES ('a', 'a', 5)`,
		`INSERT INTO friendships (user_a, user_b, created_at) VALUES ('a', 'b', 6)`,
		`INSERT INTO friendships (user_a, user_b, created_at) VALUES ('a', 'nobody', 5)`,
		`INSERT INTO friend_requests (from_user, to_user, created_at) VALUES ('a', 'a', 5)`,
		`INSERT INTO friend_requests (from_user, to_user, created_at) VALUES ('a', 'b', 6)`,
		`INSERT INTO friend_requests (from_user, to_user, created_at) VALUES ('nobody', 'b', 5)`,
		`INSERT INTO user_blocks (user_id, blocked_user_id, created_at) VALUES ('a', 'a', 5)`,
		`INSERT INTO user_blocks (user_id, blocked_user_id, created_at) VALUES ('a', 'b', 6)`,
		`INSERT INTO user_blocks (user_id, blocked_user_id, created_at) VALUES ('a', 'nobody', 5)`,
		`INSERT INTO social_prefs (user_id, updated_at) VALUES ('nobody', 5)`,
		`UPDATE social_prefs SET presence = 'everyone'`,
		`UPDATE social_prefs SET requests = 'friends'`,
	} {
		if _, err := d.Exec(bad); err == nil {
			t.Errorf("accepted: %s", bad)
		}
	}
	// The reverse request is a different row, and so is the reverse block.
	mustExec(t, d, `INSERT INTO friend_requests (from_user, to_user, created_at) VALUES ('b', 'a', 5)`)
	mustExec(t, d, `INSERT INTO user_blocks (user_id, blocked_user_id, created_at) VALUES ('b', 'a', 5)`)
	mustExec(t, d, `UPDATE social_prefs SET presence = 'played_with', requests = 'played_with'`)
}

func TestMigration0032RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, friendsVersion)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('a', 'A', 1, 1), ('b', 'B', 1, 1)`)
	mustExec(t, d, `INSERT INTO friendships (user_a, user_b, created_at) VALUES ('a', 'b', 5)`)
	mustExec(t, d, `INSERT INTO user_blocks (user_id, blocked_user_id, created_at) VALUES ('b', 'a', 5)`)
	mustExec(t, d, `DROP TABLE social_prefs`)
	mustExec(t, d, `DROP TABLE user_blocks`)
	mustExec(t, d, `DROP TABLE friend_requests`)
	mustExec(t, d, `DROP TABLE friendships`)
	mustExec(t, d, `DELETE FROM schema_migrations WHERE version = ?`, friendsVersion)
	if v, _ := currentSchemaVersion(ctx, d); v != friendsVersion-1 {
		t.Fatalf("after the rollback the version is %d, want %d", v, friendsVersion-1)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("roll forward again: %v", err)
	}
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM friendships`).Scan(&n); err != nil || n != 0 {
		t.Errorf("rolled forward: %d friendships (%v); want an empty table", n, err)
	}
}
