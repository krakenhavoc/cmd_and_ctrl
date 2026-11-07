package db

import (
	"context"
	"testing"
)

// Tests for migration 0035 (the Discord Linked Roles grant, ADR 0163 §4):
// one nullable identities column, NULL ("not linked") on every row old
// and new, the sign-in token untouched, and the hand rollback in
// docs/environments.md.

// roleConnectionVersion is this migration's number.
const roleConnectionVersion = 35

func TestMigration0035AddsAnUnlinkedRoleConnectionToken(t *testing.T) {
	d := openAtVersion(t, roleConnectionVersion-1)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	mustExec(t, d, `INSERT INTO identities (provider, subject, user_id, display_name, refresh_token, scopes, linked_at)
		VALUES ('discord', '111', 'u1', 'A', x'01', 'identify', 1)`)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate to latest: %v", err)
	}
	var (
		rc      []byte
		refresh []byte
	)
	if err := d.QueryRow(`SELECT role_connection_token, refresh_token FROM identities WHERE subject = '111'`).Scan(&rc, &refresh); err != nil {
		t.Fatal(err)
	}
	if rc != nil {
		t.Errorf("an existing identity after the migration has role_connection_token %x; want NULL (not linked)", rc)
	}
	if len(refresh) != 1 || refresh[0] != 1 {
		t.Errorf("the sign-in refresh token changed: %x", refresh)
	}
	mustExec(t, d, `UPDATE identities SET role_connection_token = x'0203' WHERE subject = '111'`)
	assertForeignKeysOnEveryConn(t, d)
}

func TestMigration0035RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, roleConnectionVersion)
	mustExec(t, d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES ('u1', 'A', 1, 1)`)
	mustExec(t, d, `INSERT INTO identities (provider, subject, user_id, display_name, scopes, linked_at, role_connection_token)
		VALUES ('discord', '111', 'u1', 'A', 'identify', 1, x'02')`)
	mustExec(t, d, `ALTER TABLE identities DROP COLUMN role_connection_token`)
	mustExec(t, d, `DELETE FROM schema_migrations WHERE version = ?`, roleConnectionVersion)
	if v, _ := currentSchemaVersion(ctx, d); v != roleConnectionVersion-1 {
		t.Fatalf("after the rollback the version is %d, want %d", v, roleConnectionVersion-1)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("roll forward again: %v", err)
	}
	var rc []byte
	if err := d.QueryRow(`SELECT role_connection_token FROM identities WHERE subject = '111'`).Scan(&rc); err != nil {
		t.Fatal(err)
	}
	if rc != nil {
		t.Errorf("rolled forward: role_connection_token %x; want NULL, everyone unlinked", rc)
	}
}
