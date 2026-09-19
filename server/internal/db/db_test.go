package db

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenCreatesFileMode0600(t *testing.T) {
	dir := t.TempDir()
	d, err := Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	path := filepath.Join(dir, dirName, fileName)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("file mode = %04o, want 0600", perm)
	}
}

func TestOpenSetsWALMode(t *testing.T) {
	dir := t.TempDir()
	d, err := Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	var mode string
	if err := d.QueryRowContext(context.Background(), `PRAGMA journal_mode`).Scan(&mode); err != nil {
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Errorf("journal_mode = %q, want %q", mode, "wal")
	}
}

func TestOpenSetsForeignKeysOn(t *testing.T) {
	dir := t.TempDir()
	d, err := Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	var on int
	if err := d.QueryRowContext(context.Background(), `PRAGMA foreign_keys`).Scan(&on); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}
	if on != 1 {
		t.Errorf("foreign_keys = %d, want 1", on)
	}
}

func TestOpenEmptyDataDirErrors(t *testing.T) {
	if _, err := Open(context.Background(), ""); err == nil {
		t.Fatal("Open(\"\") succeeded, want an error")
	}
}

func TestOpenCreatesDataDir(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "does", "not", "exist", "yet")
	d, err := Open(context.Background(), nested)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	if _, err := os.Stat(filepath.Join(nested, dirName)); err != nil {
		t.Errorf("db dir not created: %v", err)
	}
}

func TestOpenTwiceReusesTheSameFile(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	d1, err := Open(ctx, dir)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	if _, err := d1.ExecContext(ctx, `CREATE TABLE probe (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatalf("create probe table: %v", err)
	}
	if _, err := d1.ExecContext(ctx, `INSERT INTO probe (id) VALUES (1)`); err != nil {
		t.Fatalf("insert probe row: %v", err)
	}
	if err := d1.Close(); err != nil {
		t.Fatalf("close first handle: %v", err)
	}

	d2, err := Open(ctx, dir)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	defer d2.Close()

	var count int
	if err := d2.QueryRowContext(ctx, `SELECT COUNT(*) FROM probe`).Scan(&count); err != nil {
		t.Fatalf("select probe count: %v", err)
	}
	if count != 1 {
		t.Errorf("probe row count = %d, want 1 (same file, second Open)", count)
	}
}

// TestOpenSchemaTooNewRefuses simulates a database written by a
// future binary: a schema_migrations row whose version this binary's
// embedded migrations do not reach. Open must refuse it loudly rather
// than run against an unrecognised schema (ADR 0051 decision 1,
// mirroring game.SnapshotSchemaVersion's ErrSchemaTooNew).
func TestOpenSchemaTooNewRefuses(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	d, err := Open(ctx, dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	future := currentMaxMigrationVersion(t) + 1
	if _, err := d.ExecContext(ctx,
		`INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, 'from_the_future', ?)`,
		future, time.Now().Unix()); err != nil {
		t.Fatalf("insert future migration row: %v", err)
	}
	if err := d.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	_, err = Open(ctx, dir)
	if err == nil {
		t.Fatal("Open on a too-new schema succeeded, want ErrSchemaTooNew")
	}
	if !errors.Is(err, ErrSchemaTooNew) {
		t.Errorf("Open error = %v, want it to wrap ErrSchemaTooNew", err)
	}
}

func currentMaxMigrationVersion(t *testing.T) int {
	t.Helper()
	migs, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations: %v", err)
	}
	if len(migs) == 0 {
		t.Fatal("no embedded migrations")
	}
	return migs[len(migs)-1].version
}
