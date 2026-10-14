package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/accountdelete"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// snapshotTree builds what `restic restore` lays down: a data dir with
// only the database copy, the ledger and the playmats.
func snapshotTree(t *testing.T, user uuid.UUID, withPlaymatFile bool) string {
	t.Helper()
	ctx := context.Background()
	src := t.TempDir()
	d, err := db.Open(ctx, src)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, 'Ian', 1, 1)`, user.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`INSERT INTO user_playmats (user_id, slot, playmat_id, created_at) VALUES (?, 1, ?, 1)`,
		user.String(), "11111111-1111-4111-8111-111111111111"); err != nil {
		t.Fatal(err)
	}
	if err := d.Backup(ctx); err != nil {
		t.Fatal(err)
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}

	tree := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tree, "db"), 0o700); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(src, "db", db.BackupFileName))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tree, "db", db.BackupFileName), b, 0o600); err != nil {
		t.Fatal(err)
	}
	if withPlaymatFile {
		if err := os.MkdirAll(filepath.Join(tree, "playmats"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tree, "playmats", "11111111-1111-4111-8111-111111111111.jpg"), []byte("jpg"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return tree
}

func TestVerifyPassesAndReappliesTheLedger(t *testing.T) {
	user := uuid.New()
	tree := snapshotTree(t, user, true)
	if err := accountdelete.NewLedger(tree).Append(user, time.Now()); err != nil {
		t.Fatal(err)
	}
	var out, errw bytes.Buffer
	if code := run([]string{"verify", "-dir", tree}, &out, &errw); code != 0 {
		t.Fatalf("verify exit %d\nout: %s\nerr: %s", code, out.String(), errw.String())
	}
	for _, want := range []string{"opened and migrated", "reapplied=1", "playmat_files_missing=0", "restorecheck: ok"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}
}

func TestVerifyFailsOnMissingPlaymatAndMissingDatabase(t *testing.T) {
	tree := snapshotTree(t, uuid.New(), false)
	var out, errw bytes.Buffer
	if code := run([]string{"verify", "-dir", tree}, &out, &errw); code != 1 || !strings.Contains(out.String(), "FAIL") {
		t.Errorf("missing playmat: exit %d, out %s", code, out.String())
	}
	empty := t.TempDir()
	if code := run([]string{"verify", "-dir", empty}, &out, &errw); code != 1 {
		t.Errorf("empty tree: exit %d", code)
	}
}

func TestVerifyFailsOnACorruptDatabase(t *testing.T) {
	tree := snapshotTree(t, uuid.New(), true)
	if err := os.WriteFile(filepath.Join(tree, "db", db.BackupFileName), []byte("not a database"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errw bytes.Buffer
	if code := run([]string{"verify", "-dir", tree}, &out, &errw); code != 1 {
		t.Errorf("corrupt database: exit %d", code)
	}
}

func TestReapplyDeletesPlaymatFiles(t *testing.T) {
	user := uuid.New()
	tree := snapshotTree(t, user, true)
	b, _ := os.ReadFile(filepath.Join(tree, "db", db.BackupFileName))
	if err := os.WriteFile(filepath.Join(tree, "db", "cmdctrl.sqlite"), b, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := accountdelete.NewLedger(tree).Append(user, time.Now()); err != nil {
		t.Fatal(err)
	}
	var out, errw bytes.Buffer
	if code := run([]string{"reapply", "-data-dir", tree}, &out, &errw); code != 0 {
		t.Fatalf("reapply exit %d: %s %s", code, out.String(), errw.String())
	}
	if _, err := os.Stat(filepath.Join(tree, "playmats", "11111111-1111-4111-8111-111111111111.jpg")); !os.IsNotExist(err) {
		t.Errorf("playmat file survived re-application: %v", err)
	}
	if !strings.Contains(out.String(), "reapplied=1") {
		t.Errorf("output: %s", out.String())
	}
}

func TestUsageErrors(t *testing.T) {
	var out, errw bytes.Buffer
	for _, args := range [][]string{{}, {"nope"}, {"verify"}} {
		if code := run(args, &out, &errw); code != 2 {
			t.Errorf("run(%v) = %d, want 2", args, code)
		}
	}
}
